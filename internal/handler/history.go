package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"bililive-helper/internal/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetHistory(c *gin.Context) {
	task := c.Query("task")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	records, total := h.history.GetRecords(task, page, perPage)
	pages := (total + perPage - 1) / perPage

	c.JSON(http.StatusOK, gin.H{
		"items":    records,
		"total":    total,
		"page":     page,
		"per_page": perPage,
		"pages":    pages,
	})
}

func (h *Handler) ExportHistory(c *gin.Context) {
	c.JSON(http.StatusOK, h.history.GetAllRecords())
}

func (h *Handler) GetLogList(c *gin.Context) {
	task := c.Param("task")
	if !utils.ValidateFilename(task) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "非法任务名"})
		return
	}
	c.JSON(http.StatusOK, h.listLogFiles(task))
}

func (h *Handler) listLogFiles(task string) []gin.H {
	logDir := filepath.Join(h.config.LogDir, task+"_log")
	baseName := task + "_videos.log"

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil
	}

	var files []gin.H
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, baseName) {
			continue
		}
		info, err := entry.Info()
		if err != nil || info.Size() == 0 {
			continue
		}
		label := ""
		if name == baseName {
			label = time.Now().Format("2006-01-02") + " (最新)"
		} else if strings.HasPrefix(name, baseName+".") {
			datePart := strings.TrimPrefix(name, baseName+".")
			label = datePart
		} else {
			continue
		}
		files = append(files, gin.H{
			"date":     label,
			"filename": name,
			"mtime":    info.ModTime().Unix(),
			"task":     task,
		})
	}

	// Sort newest first
	sort.Slice(files, func(i, j int) bool {
		mi, _ := files[i]["mtime"].(int64)
		mj, _ := files[j]["mtime"].(int64)
		return mi > mj
	})

	return files
}

func (h *Handler) GetLogContent(c *gin.Context) {
	task := c.Param("task")
	if !utils.ValidateFilename(task) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "非法任务名"})
		return
	}
	file := c.Query("file")

	if file == "" || !strings.HasPrefix(file, task+"_videos.log") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效文件"})
		return
	}
	if strings.Contains(file, "..") || strings.Contains(file, "/") || strings.Contains(file, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效文件"})
		return
	}

	path := filepath.Join(h.config.LogDir, task+"_log", file)
	content, err := os.ReadFile(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 200 {
		lines = lines[len(lines)-200:]
	}
	c.String(http.StatusOK, strings.Join(lines, "\n"))
}
