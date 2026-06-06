package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"bililive-helper/internal/utils"

	"github.com/gin-gonic/gin"
)

// validLogID 校验操作日志 ID 格式：{type}_{YYYYMMDD}_{HHMMSS}_{4位hex}
var validLogID = regexp.MustCompile(`^[a-z]+_\d{8}_\d{6}_[0-9a-f]{4}$`)

// GetHistory 分页查询历史记录，支持按任务类型和主播名过滤。
func (h *Handler) GetHistory(c *gin.Context) {
	task := c.Query("task")
	streamer := c.Query("streamer")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	records, total := h.history.GetRecords(task, streamer, page, perPage)
	pages := (total + perPage - 1) / perPage

	c.JSON(http.StatusOK, gin.H{
		"items":    records,
		"total":    total,
		"page":     page,
		"per_page": perPage,
		"pages":    pages,
	})
}

// ExportHistory 导出全部历史记录。
func (h *Handler) ExportHistory(c *gin.Context) {
	c.JSON(http.StatusOK, h.history.GetAllRecords())
}

// GetLogContent 根据 log_id 返回对应的操作日志内容（最近 200 行）。
func (h *Handler) GetLogContent(c *gin.Context) {
	task := c.Param("task")
	if !utils.ValidateFilename(task) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "非法任务名"})
		return
	}

	logID := c.Query("log_id")
	if logID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少日志 ID"})
		return
	}
	if !validLogID.MatchString(logID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无效的日志 ID: %s", logID)})
		return
	}

	path := filepath.Join(h.config.LogDir, task+"_log", "op_"+logID+".log")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.String(http.StatusOK, "[系统提示] 该操作日志已超过 30 天，已被自动清理")
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取日志失败"})
		return
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 200 {
		lines = lines[len(lines)-200:]
	}
	c.String(http.StatusOK, strings.Join(lines, "\n"))
}
