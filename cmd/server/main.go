// Package main 是 Bililive Helper 应用程序的入口。
// 负责初始化配置、创建业务服务、注册 HTTP 路由、启动服务器并处理优雅停机。
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"bililive-helper/internal/config"
	"bililive-helper/internal/handler"
	"bililive-helper/internal/middleware"
	"bililive-helper/internal/service"
	"bililive-helper/internal/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// main 是应用程序入口函数。
// 负责加载配置、初始化服务、注册路由、启动 HTTP 服务器，并监听系统信号实现优雅停机。
func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 加载配置：config.json -> 环境变量覆盖 -> 自动生成凭据
	cfg := config.Load()

	// 按依赖顺序创建服务：history -> merge/clean -> scheduler
	historyService := service.NewHistoryService(cfg, logger)
	mergeService := service.NewMergeService(cfg, logger, historyService)
	cleanService := service.NewCleanService(cfg, logger, historyService)
	schedulerService := service.NewSchedulerService(cfg, logger, mergeService, cleanService, historyService)

	// 启动前检测 FFmpeg/FFprobe 是否可用，避免运行时才发现缺失
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Fatal("ffmpeg 未安装或不在 PATH 中，请先安装 ffmpeg")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		log.Fatal("ffprobe 未安装或不在 PATH 中，请先安装 ffprobe（通常与 ffmpeg 一起安装）")
	}
	logger.Info("ffmpeg / ffprobe 检测通过")

	// 清理上次异常退出遗留的临时文件（.merge_tmp_*、.concat_* 等）
	if n := mergeService.CleanupTempFiles(); n > 0 {
		logger.Info("清理残留临时文件", zap.Int("count", n))
	}

	schedulerService.Start()
	defer schedulerService.Stop()

	h := handler.NewHandler(cfg, logger, mergeService, cleanService, historyService, schedulerService)

	r := gin.Default()

	store := cookie.NewStore([]byte(cfg.SecretKey))
	// Session 安全配置：HttpOnly 防 XSS、SameSiteLax 防 CSRF、可选 Secure 标志
	secure := os.Getenv("COOKIE_SECURE") == "true"
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7 * 86400,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("session", store))
	r.Use(middleware.SecurityHeaders())

	// 加载登录页模板
	r.LoadHTMLFiles("templates/login.html")

	// 提供 Vue SPA 静态资源
	r.Static("/assets", "./templates/assets")

	// Vue Router History 模式：非 API 路由统一返回 index.html
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "API not found"})
			return
		}
		c.File("./templates/index.html")
	})

	r.GET("/", h.Index)
	r.GET("/login", h.LoginPage)
	r.GET("/logout", h.Logout)
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	// 注册 API 路由
	api := r.Group("/api")
	{
		// 公开接口（无需认证）
		api.POST("/login", h.Login)
		api.GET("/health", h.Health)
		api.GET("/setup/status", h.SetupStatus)
		api.POST("/setup/init", h.SetupInit)

		// 需要认证的接口
		auth := api.Group("")
		auth.Use(middleware.AuthRequired())
		auth.Use(middleware.RateLimiter(60)) // 已认证接口：60 次 POST/分钟/IP
		{
			auth.GET("/auth/check", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
			auth.POST("/auth/change-password", h.ChangePassword)
			auth.GET("/status", h.Status)
			auth.GET("/status/detail", h.StatusDetail)
			auth.GET("/stats", h.Stats)
			auth.GET("/config", h.GetConfig)
			auth.POST("/config", h.SaveConfig)
			auth.GET("/config/recommend", h.RecommendConfig)
			auth.GET("/config/defaults", h.DefaultConfig)
			auth.GET("/config/export", h.ExportConfig)
			auth.POST("/config/import", h.ImportConfig)
			auth.GET("/schedule", h.GetSchedule)
			auth.POST("/schedule", h.SaveSchedule)
			auth.POST("/schedule/run/:task", h.RunTask)
			auth.GET("/streamers", h.GetStreamers)
			auth.GET("/streamers/:name/files", h.GetStreamerFiles)
			auth.GET("/files/:name", h.GetStreamerFiles)
			auth.POST("/merge", h.RunMerge)
			auth.POST("/merge/manual", h.ManualMerge)
			auth.POST("/merge/retry", h.MergeRetry)
			auth.POST("/clean", h.RunClean)
			auth.GET("/clean/estimate", h.CleanEstimate)
			auth.POST("/clean/emergency", h.EmergencyClean)
			auth.GET("/run/:task", h.RunTaskSSE)
			auth.GET("/history", h.GetHistory)
			auth.GET("/history/export", h.ExportHistory)
			auth.GET("/logs/list/:task", h.GetLogList)
			auth.GET("/logs/content/:task", h.GetLogContent)
			auth.GET("/setup/check", h.SetupCheck)
		}
	}

	port := cfg.Port
	if port == 0 {
		port = 5000
	}

	// 打印启动摘要（配置、FFmpeg 信息、端口等）
	utils.LogStartup(logger, cfg.TargetDir, cfg.TriggerThreshold, cfg.TargetThreshold,
		cfg.GapMinutes, cfg.MergeAgeMinutes,
		cfg.BackupStartHour, cfg.BackupStartMinute, cfg.BackupEndHour, cfg.BackupEndMinute,
		cfg.SafeMode, cfg.SafeDays, port)

	logger.Info("启动服务器", zap.Int("port", port))
	fmt.Printf("Bililive Helper 启动成功，访问 http://localhost:%d\n", port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	// 在独立 goroutine 中启动 HTTP 服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号（SIGINT/SIGTERM），实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\n正在关闭服务器...")

	schedulerService.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}
	fmt.Println("服务器已关闭")
}
