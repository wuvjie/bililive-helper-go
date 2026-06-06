// Package middleware 提供 HTTP 中间件。
// 包含认证检查（Session/Bearer Token）、请求频率限制、安全响应头等功能。
package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"bililive-helper/internal/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// RateLimiter 提供基于 IP 的令牌桶限流中间件，仅对 POST 请求生效。
// 适用于局域网部署场景，不适合公网暴露。
// 返回中间件处理函数和一个 stop 函数，用于在应用关闭时停止 GC 协程。
func RateLimiter(maxPerMinute int) (gin.HandlerFunc, func()) {
	type bucket struct {
		tokens   int
		lastFill time.Time
	}
	var (
		mu      sync.Mutex
		buckets = make(map[string]*bucket)
	)
	// 定期 GC：清理超过 2 分钟未活动的 IP 条目，防止内存无限增长
	stopCh := make(chan struct{})
	stopOnce := sync.Once{}
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				mu.Lock()
				cutoff := time.Now().Add(-2 * time.Minute)
				for ip, b := range buckets {
					if b.lastFill.Before(cutoff) {
						delete(buckets, ip)
					}
				}
				mu.Unlock()
			}
		}
	}()

	handler := func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}
		ip := c.ClientIP()
		mu.Lock()
		b, exists := buckets[ip]
		if !exists {
			b = &bucket{tokens: maxPerMinute, lastFill: time.Now()}
			buckets[ip] = b
		}
		// 基于流逝时间补充令牌，实现分钟内细粒度限流
		elapsed := time.Since(b.lastFill)
		refill := int(elapsed.Seconds() * float64(maxPerMinute) / 60)
		if refill > 0 {
			b.tokens += refill
			if b.tokens > maxPerMinute {
				b.tokens = maxPerMinute
			}
		}
		b.lastFill = time.Now()
		if b.tokens <= 0 {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
			c.Abort()
			return
		}
		b.tokens--
		mu.Unlock()
		c.Next()
	}
	return handler, func() { stopOnce.Do(func() { close(stopCh) }) }
}

// AuthRequired 认证中间件，支持两种认证方式：
// 1. Session Cookie 认证（浏览器端），校验 session_version 防止旧 session 被利用
// 2. Bearer Token 认证（API 调用，constant-time 比较防时序攻击）
func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	expectedToken := os.Getenv("API_TOKEN")
	expectedTokenBytes := []byte(expectedToken)

	return func(c *gin.Context) {
		// Session 认证（浏览器）
		session := sessions.Default(c)
		if session.Get("authenticated") == true {
			// 校验 session_version：改密后旧 session 自动失效
			sessionVer, _ := session.Get("session_version").(int)
			currentVer := cfg.Snapshot().SessionVersion
			if sessionVer != currentVer {
				session.Clear()
				session.Options(sessions.Options{MaxAge: -1})
				session.Save()
				if strings.HasPrefix(c.Request.URL.Path, "/api") {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已过期，请重新登录"})
					c.Abort()
					return
				}
				c.Redirect(http.StatusFound, "/login")
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// API Token 认证（Bearer）— 使用 constant-time 比较防止时序攻击
		// 仅在 API_TOKEN 已配置时启用 Bearer 认证，避免空 token 绕过
		if expectedToken != "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if subtle.ConstantTimeCompare([]byte(token), expectedTokenBytes) == 1 {
					c.Next()
					return
				}
			}
		}

		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或登录已过期"})
			c.Abort()
			return
		}
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
	}
}

// SecurityHeaders 注入安全响应头（防 XSS、防点击劫持、CSP 策略等）。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// 注意：login.html 包含内联 <script>，需要 'unsafe-inline'。
		// 如需更严格的 CSP，需将 login.html 中的内联脚本提取为外部文件。
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; base-uri 'self'; form-action 'self'; frame-ancestors 'self'")
		c.Next()
	}
}
