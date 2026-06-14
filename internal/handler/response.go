// response.go 提供统一的 HTTP 响应辅助函数。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ok 返回成功响应（HTTP 200）。
func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// okMsg 返回成功消息响应（HTTP 200）。
func okMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

// okStatus 返回成功状态响应（HTTP 200）。
func okStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// fail 返回错误响应。httpCode 为 HTTP 状态码，msg 为中文错误消息。
func fail(c *gin.Context, httpCode int, msg string) {
	c.JSON(httpCode, gin.H{"error": msg})
}

// failBadRequest 返回 400 错误响应。
func failBadRequest(c *gin.Context, msg string) {
	fail(c, http.StatusBadRequest, msg)
}

// failUnauthorized 返回 401 错误响应。
func failUnauthorized(c *gin.Context, msg string) {
	fail(c, http.StatusUnauthorized, msg)
}

// failConflict 返回 409 错误响应。
func failConflict(c *gin.Context, msg string) {
	fail(c, http.StatusConflict, msg)
}

// failTooMany 返回 429 错误响应。
func failTooMany(c *gin.Context, msg string) {
	fail(c, http.StatusTooManyRequests, msg)
}

// failInternal 返回 500 错误响应。
func failInternal(c *gin.Context, msg string) {
	fail(c, http.StatusInternalServerError, msg)
}
