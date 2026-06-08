package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bililive-helper/internal/config"
	"bililive-helper/internal/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	rateMu       sync.Mutex
	rateAttempts = map[string][]time.Time{}
	rateLastGC   = time.Time{}
)

// isRateLimited 检查指定 IP 是否在 5 分钟内登录尝试次数超过 5 次。
// 使用内存 map 存储尝试记录，定期 GC 清理过期条目防止内存泄漏。
func isRateLimited(ip string) bool {
	rateMu.Lock()
	defer rateMu.Unlock()

	now := time.Now()
	cutoff := now.Add(-5 * time.Minute)

	// 定期 GC：清理过期 IP 条目，防止内存无限增长
	if now.Sub(rateLastGC) > time.Minute {
		rateLastGC = now
		for k, v := range rateAttempts {
			var valid []time.Time
			for _, t := range v {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rateAttempts, k)
			} else {
				rateAttempts[k] = valid
			}
		}
	}

	attempts := rateAttempts[ip]
	var valid []time.Time
	for _, t := range attempts {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rateAttempts[ip] = valid
	return len(valid) >= 5
}

func recordAttempt(ip string) {
	rateMu.Lock()
	defer rateMu.Unlock()
	rateAttempts[ip] = append(rateAttempts[ip], time.Now())
}

func hashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func verifyPassword(hashed, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}

// LoginPage 渲染登录页面。
func (h *Handler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

// Index 返回 Vue SPA 的主入口页面。
func (h *Handler) Index(c *gin.Context) {
	c.File("./templates/index.html")
}

// Login 处理用户登录请求。
// 验证密码后设置 Session。bcrypt 本身提供常量时间比较，防止时序攻击。
func (h *Handler) Login(c *gin.Context) {
	ip := c.ClientIP()
	if isRateLimited(ip) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "登录尝试次数过多，请 5 分钟后再试"})
		return
	}

	// 限制请求体大小为 1KB，防止内存耗尽攻击
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误，请发送 JSON 数据"})
		return
	}

	h.passwordMu.RLock()
	passwordOK := verifyPassword(h.hashedPassword, req.Password)
	h.passwordMu.RUnlock()
	if !passwordOK {
		recordAttempt(ip)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	session := sessions.Default(c)
	// 清除旧 Session 数据防止 Session 固定攻击
	session.Clear()
	session.Set("authenticated", true)
	session.Set("login_time", time.Now().Unix())
	session.Set("session_version", h.config.Snapshot().SessionVersion)
	if err := session.Save(); err != nil {
		h.logger.Warn("session 保存失败", zap.Error(err))
	}
	c.JSON(http.StatusOK, gin.H{"message": "登录成功"})
}

// Logout 处理用户登出，清除 Session 并重定向到登录页。
func (h *Handler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1}) // MaxAge -1 强制浏览器删除 cookie
	if err := session.Save(); err != nil {
		h.logger.Warn("session 保存失败", zap.Error(err))
	}
	c.Redirect(http.StatusFound, "/login")
}

// Health 返回服务健康状态，用于 Docker 健康检查。
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ChangePassword 允许已认证用户修改密码。
// 验证旧密码后更新配置文件、凭据文件和运行时哈希。
// 递增 SessionVersion 使所有旧 Session 自动失效。
func (h *Handler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写旧密码和新密码"})
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码至少 6 个字符"})
		return
	}

	h.passwordMu.RLock()
	oldHashed := h.hashedPassword
	h.passwordMu.RUnlock()
	if !verifyPassword(oldHashed, req.OldPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "旧密码错误"})
		return
	}

	// 先在内存中设置新密码，然后持久化到凭据文件。
	// 这样如果凭据保存失败，不会留下版本不一致的状态。
	h.config.Password = req.NewPassword
	if err := h.config.SaveCredential(); err != nil {
		h.config.Password = "" // 回滚内存
		h.logger.Error("密码持久化失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码持久化失败，请检查磁盘空间和权限"})
		return
	}

	// 持久化成功，更新配置（递增 SessionVersion 使旧 Session 失效）
	if err := h.config.Apply(func() error {
		h.config.SessionVersion++
		return nil
	}); err != nil {
		// 凭据文件已更新为新密码，但配置更新失败。
		// 运行时仍使用旧密码（内存哈希未更新），用户可重试。
		// 重启后会加载新密码 + 旧 SessionVersion，不影响登录。
		h.logger.Error("密码配置写入失败（凭据已更新，可重试）", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码更新失败，请重试"})
		return
	}

	// 更新运行时密码哈希
	hashed, err := hashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码更新失败"})
		return
	}
	h.passwordMu.Lock()
	h.hashedPassword = hashed
	h.passwordMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": "密码已更新"})
}

// SetupStatus 返回当前是否为首次运行（config.json 不存在）。
// 此接口无需认证，前端据此决定是否显示初始化向导。
func (h *Handler) SetupStatus(c *gin.Context) {
	firstRun := !config.ConfigExists(h.config.LogDir)
	c.JSON(http.StatusOK, gin.H{
		"first_run": firstRun,
		"log_dir":   h.config.LogDir,
	})
}

// SetupInit 处理首次运行的初始化请求：校验目录、保存配置、自动登录。
func (h *Handler) SetupInit(c *gin.Context) {
	// 互斥锁防止并发初始化竞态
	h.setupMu.Lock()
	defer h.setupMu.Unlock()

	// 如果配置已存在则拒绝（防止重复初始化）
	if config.ConfigExists(h.config.LogDir) {
		c.JSON(http.StatusConflict, gin.H{"error": "系统已完成初始化"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请设置密码"})
		return
	}

	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少 6 个字符"})
		return
	}

	// 目录来自环境变量（docker-compose），使用运行时值而非默认值
	cfg := config.DefaultConfig()
	cfg.TargetDir = h.config.TargetDir // 使用运行时的 TARGET_DIR（来自环境变量）
	cfg.LogDir = h.config.LogDir       // 使用运行时的 LOG_DIR（来自环境变量）
	cfg.Password = req.Password
	cfg.SecretKey = utils.RandomHex(16)
	cfg.ConfigFile = filepath.Join(cfg.LogDir, "config.json")

	// 原子写入 config.json（通过 Apply 内部的 atomicWriteFile + fsync）
	if err := h.config.Apply(func() error {
		h.config.TargetDir = cfg.TargetDir
		h.config.LogDir = cfg.LogDir
		h.config.ConfigFile = cfg.ConfigFile
		h.config.Password = cfg.Password
		h.config.SecretKey = cfg.SecretKey
		return nil
	}); err != nil {
		h.logger.Error("初始化配置写入失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置写入失败，请检查目录权限"})
		return
	}

	// 持久化密码到凭据文件（Password 字段 json:"-" 不会写入 config.json）
	if err := h.config.SaveCredential(); err != nil {
		h.logger.Error("密码持久化失败，初始化回滚", zap.Error(err))
		// 回滚：删除已写入的 config.json，恢复首次运行状态
		os.Remove(h.config.ConfigFile)
		h.config.Apply(func() error {
			h.config.Password = ""
			h.config.SecretKey = ""
			return nil
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码持久化失败，请检查磁盘空间和权限"})
		return
	}

	// 重新哈希密码用于运行时登录验证
	hashed, err := hashPassword(cfg.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码哈希失败"})
		return
	}
	h.passwordMu.Lock()
	h.hashedPassword = hashed
	h.passwordMu.Unlock()

	// 自动登录：设置 Session
	session := sessions.Default(c)
	session.Clear()
	session.Set("authenticated", true)
	session.Set("login_time", time.Now().Unix())
	session.Set("session_version", h.config.Snapshot().SessionVersion)
	if err := session.Save(); err != nil {
		h.logger.Warn("session 保存失败", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{"message": "初始化成功"})
}
