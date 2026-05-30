package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	expectedToken := os.Getenv("API_TOKEN")
	expectedTokenBytes := []byte(expectedToken)

	return func(c *gin.Context) {
		// Check session first
		session := sessions.Default(c)
		if session.Get("authenticated") == true {
			c.Next()
			return
		}

		// Check API token (Bearer) — constant-time comparison
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if expectedToken != "" && subtle.ConstantTimeCompare([]byte(token), expectedTokenBytes) == 1 {
				c.Next()
				return
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

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
