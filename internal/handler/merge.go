package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"bililive-helper/internal/model"
	"bililive-helper/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RunMerge 执行全局或指定主播的合并任务，通过 SSE 流式返回进度。
func (h *Handler) RunMerge(c *gin.Context) {
	var req struct {
		Streamer string `json:"streamer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && c.Request.ContentLength > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	h.runSSE(c, "merge", func(ctx context.Context, onProgress func(string)) string {
		result, err := h.merge.Run(ctx, req.Streamer, onProgress)
		if err != nil {
			h.logger.Error("合并失败", zap.Error(err))
			return fmt.Sprintf("❌ 错误: %s", err.Error())
		}
		if result.Done > 0 {
			h.logger.Info("合并完成", zap.Int("done", result.Done), zap.Float64("gb", result.TotalGB))
			utils.NotifyWebhook(fmt.Sprintf("手动合并完成：%d 场次 (%.1f GB)", result.Done, result.TotalGB))
			return fmt.Sprintf("✅ 完成: 合并 %d 场次 (%.1f GB)", result.Done, result.TotalGB)
		}
		return "✅ 完成"
	})
}

// ManualMerge 手动合并指定主播的指定文件列表（至少 2 个文件）。
func (h *Handler) ManualMerge(c *gin.Context) {
	var req struct {
		Streamer   string   `json:"streamer" binding:"required"`
		Files      []string `json:"files" binding:"required"`
		OutputName string   `json:"output_name"` // 可选：自定义输出文件名
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if len(req.Files) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "至少选择2个文件"})
		return
	}
	h.runManualMergeSSEWithName(c, req.Streamer, req.Files, req.OutputName, "手动合并")
}

// MergeRetry 重试之前失败的合并操作。
func (h *Handler) MergeRetry(c *gin.Context) {
	var req struct {
		Streamer string   `json:"streamer" binding:"required"`
		Files    []string `json:"files" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	h.runManualMergeSSE(c, req.Streamer, req.Files, "重试")
}

func (h *Handler) runManualMergeSSE(c *gin.Context, streamer string, files []string, label string) {
	h.runManualMergeSSEWithName(c, streamer, files, "", label)
}

func (h *Handler) runManualMergeSSEWithName(c *gin.Context, streamer string, files []string, outputName string, label string) {
	h.runSSE(c, "merge", func(ctx context.Context, onProgress func(string)) string {
		if err := h.merge.ManualMerge(ctx, streamer, files, outputName, onProgress); err != nil {
			h.logger.Error(label+"失败", zap.Error(err))
			return fmt.Sprintf("❌ 错误: %s", err.Error())
		}
		return fmt.Sprintf("✅ 完成: %s %d 个文件", label, len(files))
	})
}

// RunClean 执行全局或指定主播的清理任务，通过 SSE 流式返回进度。
func (h *Handler) RunClean(c *gin.Context) {
	var req struct {
		Streamer string `json:"streamer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && c.Request.ContentLength > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	h.runSSE(c, "clean", func(ctx context.Context, onProgress func(string)) string {
		result, err := h.clean.Run(req.Streamer, onProgress)
		if err != nil {
			h.logger.Error("清理失败", zap.Error(err))
			return fmt.Sprintf("❌ 错误: %s", err.Error())
		}
		if result.Deleted > 0 {
			h.logger.Info("清理完成", zap.Int("deleted", result.Deleted), zap.Int64("freed", result.Freed))
			utils.NotifyWebhook(fmt.Sprintf("手动清理完成：%d 文件，释放 %s", result.Deleted, utils.FormatSize(result.Freed)))
			return fmt.Sprintf("✅ 完成: 删除 %d 文件，释放 %s", result.Deleted, utils.FormatSize(result.Freed))
		}
		return "✅ 完成"
	})
}

// RunTaskSSE 通过 GET 请求以 SSE 方式执行合并或清理任务。
func (h *Handler) RunTaskSSE(c *gin.Context) {
	task := c.Param("task")
	streamer := c.Query("streamer")

	if task != "merge" && task != "clean" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效任务"})
		return
	}

	h.runSSE(c, task, func(ctx context.Context, onProgress func(string)) string {
		if task == "merge" {
			result, err := h.merge.Run(ctx, streamer, onProgress)
			if err != nil {
				return fmt.Sprintf("❌ 错误: %s", err.Error())
			}
			if result.Done > 0 {
				utils.NotifyWebhook(fmt.Sprintf("合并完成：%d 场次 (%.1f GB)", result.Done, result.TotalGB))
				return fmt.Sprintf("✅ 完成: 合并 %d 场次 (%.1f GB)", result.Done, result.TotalGB)
			}
			return "✅ 完成"
		}
		result, err := h.clean.Run(streamer, onProgress)
		if err != nil {
			return fmt.Sprintf("❌ 错误: %s", err.Error())
		}
		if result.Deleted > 0 {
			utils.NotifyWebhook(fmt.Sprintf("清理完成：%d 文件，释放 %s", result.Deleted, utils.FormatSize(result.Freed)))
			return fmt.Sprintf("✅ 完成: 删除 %d 文件，释放 %s", result.Deleted, utils.FormatSize(result.Freed))
		}
		return "✅ 完成"
	})
}

// runSSE 同步执行 fn 并通过 Server-Sent Events 流式传输进度消息。
// 进度更新会合并 — 每次 tick/notify 只发送最新消息，避免消息积压。
func (h *Handler) runSSE(c *gin.Context, task string, fn func(ctx context.Context, onProgress func(string)) string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	var latest atomic.Value
	notify := make(chan struct{}, 1)
	onProgress := func(msg string) {
		latest.Store(msg)
		select {
		case notify <- struct{}{}:
		default:
		}
	}

	done := make(chan string, 1)
	ctx := c.Request.Context()
	go func() {
		done <- fn(ctx, onProgress)
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastSent string
	for {
		select {
		case <-ctx.Done():
			return
		case <-notify:
			if v := latest.Load(); v != nil {
				if msg := v.(string); msg != "" && msg != lastSent {
					fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
					c.Writer.Flush()
					lastSent = msg
				}
			}
		case <-ticker.C:
			if v := latest.Load(); v != nil {
				if msg := v.(string); msg != "" && msg != lastSent {
					fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
					c.Writer.Flush()
					lastSent = msg
				}
			}
		case result := <-done:
			fmt.Fprintf(c.Writer, "data: %s\n\n", result)
			fmt.Fprintf(c.Writer, "data: [END]\n\n")
			c.Writer.Flush()
			return
		}
	}
}

// RunTask 通过调度器手动触发指定任务（merge 或 clean）。
func (h *Handler) RunTask(c *gin.Context) {
	task := c.Param("task")
	if err := h.scheduler.RunTask(task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": fmt.Sprintf("%s 已触发", task)})
}

// GetSchedule 返回当前调度状态（间隔、启用状态、上次/下次执行时间）。
func (h *Handler) GetSchedule(c *gin.Context) {
	c.JSON(http.StatusOK, h.scheduler.GetStatus())
}

// SaveSchedule 保存调度配置（合并/清理间隔、启用状态、静默时段）。
// 记录变更日志并异步写入历史。
func (h *Handler) SaveSchedule(c *gin.Context) {
	var req struct {
		MergeInterval    int    `json:"merge_interval"`
		CleanInterval    int    `json:"clean_interval"`
		MergeEnabled     bool   `json:"merge_enabled"`
		CleanEnabled     bool   `json:"clean_enabled"`
		BackupStartHour  *int   `json:"BACKUP_START_HOUR"`
		BackupStartMin   *int   `json:"BACKUP_START_MINUTE"`
		BackupEndHour    *int   `json:"BACKUP_END_HOUR"`
		BackupEndMin     *int   `json:"BACKUP_END_MINUTE"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 保存前记录旧状态用于变更日志
	oldStatus := h.scheduler.GetStatus()
	oldCfg := h.config.Snapshot()

	schedule := model.ScheduleConfig{
		MergeInterval: req.MergeInterval,
		CleanInterval: req.CleanInterval,
		MergeEnabled:  req.MergeEnabled,
		CleanEnabled:  req.CleanEnabled,
	}

	if err := h.scheduler.SaveSchedule(schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}

	// 如果请求中包含静默时段字段，同步更新到配置
	backupChanged := false
	if req.BackupStartHour != nil || req.BackupStartMin != nil || req.BackupEndHour != nil || req.BackupEndMin != nil {
		if err := h.config.Apply(func() error {
			if req.BackupStartHour != nil {
				h.config.BackupStartHour = *req.BackupStartHour
			}
			if req.BackupStartMin != nil {
				h.config.BackupStartMinute = *req.BackupStartMin
			}
			if req.BackupEndHour != nil {
				h.config.BackupEndHour = *req.BackupEndHour
			}
			if req.BackupEndMin != nil {
				h.config.BackupEndMinute = *req.BackupEndMin
			}
			backupChanged = true
			return h.config.Validate()
		}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	go func() {
		var changes []string
		if oldStatus.MergeInterval != req.MergeInterval {
			changes = append(changes, fmt.Sprintf("合并间隔: %d→%d分钟", oldStatus.MergeInterval, req.MergeInterval))
		}
		if oldStatus.CleanInterval != req.CleanInterval {
			changes = append(changes, fmt.Sprintf("清理间隔: %d→%d分钟", oldStatus.CleanInterval, req.CleanInterval))
		}
		if oldStatus.MergeEnabled != req.MergeEnabled {
			changes = append(changes, fmt.Sprintf("自动合并: %v→%v", oldStatus.MergeEnabled, req.MergeEnabled))
		}
		if oldStatus.CleanEnabled != req.CleanEnabled {
			changes = append(changes, fmt.Sprintf("自动清理: %v→%v", oldStatus.CleanEnabled, req.CleanEnabled))
		}
		if backupChanged {
			newCfg := h.config.ToDTO()
			if oldCfg.BackupStartHour != newCfg.BackupStartHour || oldCfg.BackupStartMinute != newCfg.BackupStartMinute || oldCfg.BackupEndHour != newCfg.BackupEndHour || oldCfg.BackupEndMinute != newCfg.BackupEndMinute {
				changes = append(changes, fmt.Sprintf("静默时段: %d:%02d-%d:%02d→%d:%02d-%d:%02d",
					oldCfg.BackupStartHour, oldCfg.BackupStartMinute, oldCfg.BackupEndHour, oldCfg.BackupEndMinute,
					newCfg.BackupStartHour, newCfg.BackupStartMinute, newCfg.BackupEndHour, newCfg.BackupEndMinute))
			}
		}
		if len(changes) > 0 {
			detail := fmt.Sprintf("调度变更: %s", strings.Join(changes, ", "))
			h.logger.Info(detail)
			h.history.Add("schedule", "", "success", detail)
		} else {
			detail := fmt.Sprintf("调度已保存: 合并%v/%d分钟, 清理%v/%d分钟", req.MergeEnabled, req.MergeInterval, req.CleanEnabled, req.CleanInterval)
			h.logger.Info(detail)
			h.history.Add("schedule", "", "success", detail)
		}
	}()
	c.JSON(http.StatusOK, gin.H{"status": "success", "schedule": h.scheduler.GetStatus()})
}

// CleanEstimate 预估可清理的文件数量和大小（不含白名单和已合并文件）。
func (h *Handler) CleanEstimate(c *gin.Context) {
	cfg := h.config.ToDTO()
	root := cfg.TargetDir
	entries, err := os.ReadDir(root)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目录不存在"})
		return
	}

	var totalSize int64
	count := 0
	wl := cfg.WhitelistKeywords
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		folder := filepath.Join(root, entry.Name())
		folderEntries, _ := os.ReadDir(folder)
		for _, fe := range folderEntries {
			if fe.IsDir() {
				continue
			}
			name := fe.Name()
			if !utils.IsVideoFile(name) {
				continue
			}
			if utils.IsMergedFile(name) {
				continue
			}
			isWhitelisted := false
			lowerName := strings.ToLower(name)
			lowerEntryName := strings.ToLower(entry.Name())
			for _, kw := range wl {
				lowerKw := strings.ToLower(kw)
				if strings.Contains(lowerName, lowerKw) || strings.Contains(lowerEntryName, lowerKw) {
					isWhitelisted = true
					break
				}
			}
			if isWhitelisted {
				continue
			}
			info, err := fe.Info()
			if err == nil {
				totalSize += info.Size()
				count++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"file_count":    count,
		"total_size_gb": float64(int(float64(totalSize)/1073741824*100)) / 100,
	})
}

// EmergencyClean 紧急清理，可临时覆盖目标阈值，通过 SSE 流式返回进度。
func (h *Handler) EmergencyClean(c *gin.Context) {
	var req struct {
		TargetPct float64 `json:"target_pct"`
		Confirm   bool    `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || !req.Confirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需要确认"})
		return
	}

	h.runSSE(c, "clean", func(ctx context.Context, onProgress func(string)) string {
		// 临时覆盖目标阈值（紧急清理结束后自动恢复原值）
		if req.TargetPct > 0 && req.TargetPct < 100 {
			originalTarget := h.config.Snapshot().TargetThreshold
			h.config.Apply(func() error {
				h.config.TargetThreshold = req.TargetPct
				return nil
			})
			defer h.config.Apply(func() error {
				h.config.TargetThreshold = originalTarget
				return nil
			})
		}
		result, err := h.clean.Run("", onProgress)
		if err != nil {
			return fmt.Sprintf("❌ 错误: %s", err.Error())
		}
		if result.Deleted > 0 {
			return fmt.Sprintf("✅ 完成: 删除 %d 文件，释放 %s", result.Deleted, utils.FormatSize(result.Freed))
		}
		return "ℹ 无需清理"
	})
}


// SetupCheck 执行系统诊断检查：FFmpeg 可用性、目录权限、磁盘空间、文件统计。
func (h *Handler) SetupCheck(c *gin.Context) {
	cfg := h.config.ToDTO()
	td := cfg.TargetDir
	checks := gin.H{
		"target_dir_exists":   false,
		"target_dir_writable": false,
		"target_dir":          td,
		"streamer_count":      0,
		"video_count":         0,
		"total_size_gb":       0.0,
		"ffmpeg_ok":           false,
		"ffprobe_ok":          false,
	}

	// Check ffmpeg
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		checks["ffmpeg_ok"] = true
		checks["ffmpeg_path"] = path
	}
	// Check ffprobe
	if path, err := exec.LookPath("ffprobe"); err == nil {
		checks["ffprobe_ok"] = true
		checks["ffprobe_path"] = path
	}
	// Test ffmpeg process group creation
	if checks["ffmpeg_ok"].(bool) {
		cmd := exec.Command("ffmpeg", "-version")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := cmd.Run(); err == nil {
			checks["ffmpeg_process_group_ok"] = true
		} else {
			checks["ffmpeg_process_group_ok"] = false
			checks["ffmpeg_process_group_error"] = err.Error()
		}
	}

	if info, err := os.Stat(td); err == nil {
		checks["target_dir_exists"] = true
		checks["target_dir_writable"] = info.Mode().Perm()&0200 != 0

		var totalSize float64
		entries, _ := os.ReadDir(td)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			checks["streamer_count"] = checks["streamer_count"].(int) + 1
			folder := filepath.Join(td, entry.Name())
			folderEntries, _ := os.ReadDir(folder)
			for _, fe := range folderEntries {
				if fe.IsDir() {
					continue
				}
				if utils.IsVideoFile(fe.Name()) {
					checks["video_count"] = checks["video_count"].(int) + 1
					if info, err := fe.Info(); err == nil {
						totalSize += float64(info.Size())
					}
				}
			}
		}
		checks["total_size_gb"] = float64(int(totalSize/1073741824*100)) / 100
	}

	disk, err := utils.GetDiskUsage(td)
	if err == nil {
		checks["disk_total_gb"] = float64(int(float64(disk.Total)/1073741824*10)) / 10
		checks["disk_free_gb"] = float64(int(float64(disk.Free)/1073741824*10)) / 10
		checks["disk_usage_pct"] = disk.UsedPct
	}

	c.JSON(http.StatusOK, checks)
}
