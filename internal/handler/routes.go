// routes.go 集中管理所有路由注册。
package handler

import (
	"net/http"
	"strings"

	"bililive-helper-go/internal/config"
	"bililive-helper-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由到 gin.Engine。
// 包括公开接口、需认证的 API 接口、静态文件和 Vue Router fallback。
func RegisterRoutes(r *gin.Engine, h *Handler, cfg *config.Config) {
	// 静态文件
	r.Static("/assets", "./templates/assets")

	// Vue Router Hash 模式：非 API 路由统一返回 index.html
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			fail(c, 404, "API not found")
			return
		}
		c.File("./templates/index.html")
	})

	// 页面路由
	r.GET("/", h.Index)
	r.GET("/login", h.LoginPage)
	r.GET("/logout", h.Logout)
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	// API 路由
	api := r.Group("/api")
	{
		// 公开接口（无需认证）
		api.POST("/login", h.Login)
		api.GET("/health", h.Health)
		api.GET("/setup/status", h.SetupStatus)
		api.POST("/setup/init", h.SetupInit)

		// 需要认证的接口
		auth := api.Group("")
		auth.Use(middleware.AuthRequired(cfg))
		rateLimiter, stopRateLimiter := middleware.RateLimiter(60)
		auth.Use(rateLimiter)
		defer stopRateLimiter()
		{
			// 认证
			auth.GET("/auth/check", func(c *gin.Context) { ok(c, gin.H{"ok": true}) })
			auth.POST("/auth/change-password", h.ChangePassword)

			// 状态
			auth.GET("/status", h.Status)
			auth.GET("/status/detail", h.StatusDetail)
			auth.GET("/stats", h.Stats)

			// 配置
			auth.GET("/config", h.GetConfig)
			auth.POST("/config", h.SaveConfig)
			auth.GET("/config/recommend", h.RecommendConfig)
			auth.GET("/config/defaults", h.DefaultConfig)
			auth.GET("/config/export", h.ExportConfig)
			auth.POST("/config/import", h.ImportConfig)

			// 调度
			auth.GET("/schedule", h.GetSchedule)
			auth.POST("/schedule", h.SaveSchedule)
			auth.POST("/schedule/run/:task", h.RunTask)

			// 主播
			auth.GET("/streamers", h.GetStreamers)
			auth.GET("/streamers/:name/files", h.GetStreamerFiles)

			// 任务
			auth.POST("/merge", h.RunMerge)
			auth.POST("/merge/manual", h.ManualMerge)
			auth.POST("/merge/retry", h.MergeRetry)
			auth.POST("/clean", h.RunClean)
			auth.GET("/clean/estimate", h.CleanEstimate)
			auth.POST("/clean/emergency", h.EmergencyClean)

			// 历史
			auth.GET("/history", h.GetHistory)
			auth.GET("/history/export", h.ExportHistory)
			auth.GET("/logs/content/:task", h.GetLogContent)

			// 诊断
			auth.GET("/setup/check", h.SetupCheck)
		}
	}
}
