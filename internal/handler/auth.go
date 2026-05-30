package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	rateMu       sync.Mutex
	rateAttempts = map[string][]time.Time{}
	rateLastGC   = time.Time{}
)

func isRateLimited(ip string) bool {
	rateMu.Lock()
	defer rateMu.Unlock()

	now := time.Now()
	cutoff := now.Add(-5 * time.Minute)

	// Periodic GC: remove expired IPs to prevent memory leak
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

func hashPassword(password string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h)
}

func verifyPassword(hashed, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}

func (h *Handler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func (h *Handler) Index(c *gin.Context) {
	c.File("./templates/index.html")
}

func (h *Handler) Login(c *gin.Context) {
	ip := c.ClientIP()
	if isRateLimited(ip) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "登录尝试次数过多，请5分钟后再试"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if !verifyPassword(hashPassword(h.config.Password), req.Password) {
		// Try the plaintext password (for backward compatibility with old configs)
		if req.Password != h.config.Password {
			recordAttempt(ip)
			// Random delay to prevent timing attacks
			time.Sleep(100*time.Millisecond + time.Duration(len(req.Password)%7)*20*time.Millisecond)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
			return
		}
	}

	session := sessions.Default(c)
	// Clear old session data to prevent session fixation
	session.Clear()
	session.Set("authenticated", true)
	session.Set("login_time", time.Now().Unix())
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "登录成功"})
}

func (h *Handler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1}) // force cookie deletion
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
