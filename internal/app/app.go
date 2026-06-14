// Package app 封装应用程序的生命周期管理。
// 负责服务初始化、HTTP 服务器启动和优雅停机。
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"bililive-helper-go/internal/config"
	"bililive-helper-go/internal/handler"
	"bililive-helper-go/internal/middleware"
	"bililive-helper-go/internal/service"
	"bililive-helper-go/internal/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// App 管理应用程序的完整生命周期。
type App struct {
	server    *http.Server
	scheduler *service.SchedulerService
	logger    *zap.Logger
}

// New 初始化所有服务并返回 App 实例。
func New() (*App, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}

	cfg := config.Load()

	// 按依赖顺序创建服务
	historyService := service.NewHistoryService(cfg, logger)
	mergeService := service.NewMergeService(cfg, logger, historyService)
	cleanService := service.NewCleanService(cfg, logger, historyService)
	schedulerService := service.NewSchedulerService(cfg, logger, mergeService, cleanService, historyService)

	// 检测 FFmpeg/FFprobe
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg 未安装或不在 PATH 中")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return nil, fmt.Errorf("ffprobe 未安装或不在 PATH 中")
	}
	logger.Info("ffmpeg / ffprobe 检测通过")

	// 清理残留临时文件
	if n := mergeService.CleanupTempFiles(); n > 0 {
		logger.Info("清理残留临时文件", zap.Int("count", n))
	}

	schedulerService.Start()

	// 创建 Handler 并注册路由
	h := handler.NewHandler(cfg, logger, mergeService, cleanService, historyService, schedulerService)
	r := setupRouter(cfg, h)

	port := cfg.Port
	if port == 0 {
		port = 5000
	}

	utils.LogStartup(logger, utils.StartupInfo{
		Port:              port,
		TargetDir:         cfg.TargetDir,
		TriggerThreshold:  cfg.TriggerThreshold,
		TargetThreshold:   cfg.TargetThreshold,
		GapMinutes:        cfg.GapMinutes,
		MergeAgeMinutes:   cfg.MergeAgeMinutes,
		BackupStartHour:   cfg.BackupStartHour,
		BackupStartMinute: cfg.BackupStartMinute,
		BackupEndHour:     cfg.BackupEndHour,
		BackupEndMinute:   cfg.BackupEndMinute,
		SafeMode:          cfg.SafeMode,
		SafeDays:          cfg.SafeDays,
	})

	logger.Info("启动服务器", zap.Int("port", port))
	fmt.Printf("Bililive Helper 启动成功，访问 http://localhost:%d\n", port)

	return &App{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: r,
		},
		scheduler: schedulerService,
		logger:    logger,
	}, nil
}

// Run 启动 HTTP 服务器并阻塞等待中断信号，然后优雅停机。
func (a *App) Run() error {
	// 在独立 goroutine 中启动服务器
	errCh := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		fmt.Println("\n正在关闭服务器...")
	case err := <-errCh:
		return fmt.Errorf("服务器启动失败: %w", err)
	}

	return a.Shutdown()
}

// Shutdown 优雅关闭所有服务。
func (a *App) Shutdown() error {
	a.scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("服务器关闭失败: %w", err)
	}

	a.logger.Sync()
	fmt.Println("服务器已关闭")
	return nil
}

// setupRouter 配置 gin 路由、中间件和 session。
func setupRouter(cfg *config.Config, h *handler.Handler) *gin.Engine {
	r := gin.Default()

	if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		panic(fmt.Sprintf("设置可信代理失败: %v", err))
	}

	store := cookie.NewStore([]byte(cfg.SecretKey))
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

	r.LoadHTMLFiles("templates/login.html")

	// 注册所有路由（从 routes.go 集中管理）
	handler.RegisterRoutes(r, h, cfg)

	return r
}
