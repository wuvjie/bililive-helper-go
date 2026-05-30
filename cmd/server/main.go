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

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()

	historyService := service.NewHistoryService(cfg, logger)
	mergeService := service.NewMergeService(cfg, logger, historyService)
	cleanService := service.NewCleanService(cfg, logger, historyService)
	schedulerService := service.NewSchedulerService(cfg, logger, mergeService, cleanService, historyService)

	// Check ffmpeg availability at startup
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Fatal("ffmpeg 未安装或不在 PATH 中，请先安装 ffmpeg")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		log.Fatal("ffprobe 未安装或不在 PATH 中，请先安装 ffprobe（通常与 ffmpeg 一起安装）")
	}
	logger.Info("ffmpeg / ffprobe 检测通过")

	// Clean up leftover temp files from previous crashes
	if n := mergeService.CleanupTempFiles(); n > 0 {
		logger.Info("清理残留临时文件", zap.Int("count", n))
	}

	schedulerService.Start()
	defer schedulerService.Stop()

	h := handler.NewHandler(cfg, logger, mergeService, cleanService, historyService, schedulerService)

	r := gin.Default()

	store := cookie.NewStore([]byte(cfg.SecretKey))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7 * 86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("session", store))
	r.Use(middleware.SecurityHeaders())

	r.LoadHTMLFiles("templates/login.html")

	// Serve frontend assets (Vue SPA build)
	r.Static("/assets", "./templates/assets")

	// Serve index.html for Vue Router History mode
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

	api := r.Group("/api")
	{
		api.POST("/login", h.Login)
		api.GET("/health", h.Health)

		auth := api.Group("")
		auth.Use(middleware.AuthRequired())
		{
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

	// Startup summary
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

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

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
