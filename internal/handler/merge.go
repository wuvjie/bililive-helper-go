package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bililive-helper-go/internal/model"
	"bililive-helper-go/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// maxManualMergeFiles 手动合并/重试时允许的最大文件数量上限。
const maxManualMergeFiles = 100

// RunMerge 执行全局或指定主播的合并任务，通过 SSE 流式返回进度。
func (h *Handler) RunMerge(c *gin.Context) {
	var req struct {
		Streamer string `json:"streamer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && c.Request.ContentLength > 0 {
		failBadRequest(c, "参数格式错误")
		return
	}

	h.runSSE(c, "merge", func(ctx context.Context, onProgress func(string)) string {
		result, logID, err := h.merge.Run(ctx, req.Streamer, onProgress)
		if err != nil {
			h.logger.Error("合并失败", zap.Error(err), zap.String("logID", logID))
			return fmt.Sprintf("❌ 错误: %s", err.Error())
		}
		if result.Done > 0 {
			h.logger.Info("合并完成", zap.Int("done", result.Done), zap.Int("failed", result.Failed), zap.Float64("gb", result.TotalGB), zap.String("logID", logID))
			webhookMsg := fmt.Sprintf("手动合并完成：%d 场次 (%.1f GB)", result.Done, result.TotalGB)
			if result.Failed > 0 {
				webhookMsg += fmt.Sprintf("，失败 %d 项", result.Failed)
			}
			utils.NotifyWebhook(webhookMsg)
			if result.Failed > 0 {
				return fmt.Sprintf("✅ 完成: 合并 %d 场次 (%.1f GB)，失败 %d 项", result.Done, result.TotalGB, result.Failed)
			}
			return fmt.Sprintf("✅ 完成: 合并 %d 场次 (%.1f GB)", result.Done, result.TotalGB)
		}
		return "✅ 完成"
	})
}

// ManualMerge 手动合并指定主播的指定文件列表（至少 2 个文件）。
func (h *Handler) ManualMerge(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32*1024) // 32KB 限制（最多 100 个文件名）
	var req struct {
		Streamer string   `json:"streamer" binding:"required"`
		Files    []string `json:"files" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		failBadRequest(c, "参数错误")
		return
	}
	if !utils.ValidateFilename(req.Streamer) {
		failBadRequest(c, "主播名包含非法字符")
		return
	}
	if len(req.Files) < 2 {
		failBadRequest(c, "至少选择 2 个文件")
		return
	}
	if len(req.Files) > maxManualMergeFiles {
		failBadRequest(c, fmt.Sprintf("文件数量超限（最多%d个）", maxManualMergeFiles))
		return
	}
	h.runManualMergeSSE(c, req.Streamer, req.Files, "手动合并")
}

func (h *Handler) runManualMergeSSE(c *gin.Context, streamer string, files []string, label string) {
	h.runSSE(c, "merge", func(ctx context.Context, onProgress func(string)) string {
		logID, err := h.merge.ManualMerge(ctx, streamer, files, onProgress)
		if err != nil {
			h.logger.Error(label+"失败", zap.Error(err), zap.String("logID", logID))
			return fmt.Sprintf("❌ 错误: %s", err.Error())
		}
		_ = logID // logID 已在 ManualMerge 内部写入 history
		return fmt.Sprintf("✅ 完成: %s %d 个文件", label, len(files))
	})
}

// RunClean 执行全局或指定主播的清理任务，通过 SSE 流式返回进度。
func (h *Handler) RunClean(c *gin.Context) {
	var req struct {
		Streamer string `json:"streamer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && c.Request.ContentLength > 0 {
		failBadRequest(c, "参数格式错误")
		return
	}

	h.runSSE(c, "clean", func(ctx context.Context, onProgress func(string)) string {
		result, logID, err := h.clean.Run(ctx, req.Streamer, onProgress)
		if err != nil {
			h.logger.Error("清理失败", zap.Error(err), zap.String("logID", logID))
			return fmt.Sprintf("❌ 错误: %s", err.Error())
		}
		if result.Deleted > 0 {
			h.logger.Info("清理完成", zap.Int("deleted", result.Deleted), zap.Int64("freed", result.Freed), zap.String("logID", logID))
			utils.NotifyWebhook(fmt.Sprintf("手动清理完成：%d 文件，释放 %s", result.Deleted, utils.FormatSize(result.Freed)))
			return fmt.Sprintf("✅ 完成: 删除 %d 文件，释放 %s", result.Deleted, utils.FormatSize(result.Freed))
		}
		return "✅ 完成"
	})
}

// RunTask 通过调度器手动触发指定任务（merge 或 clean）。
func (h *Handler) RunTask(c *gin.Context) {
	task := c.Param("task")
	if err := h.scheduler.RunTask(task); err != nil {
		failBadRequest(c, err.Error())
		return
	}
	taskName := map[string]string{"merge": "合并", "clean": "清理"}[task]
	if taskName == "" {
		taskName = task
	}
	okMsg(c, fmt.Sprintf("%s任务已触发", taskName))
}

// GetSchedule 返回当前调度状态（间隔、启用状态、上次/下次执行时间）。
func (h *Handler) GetSchedule(c *gin.Context) {
	ok(c, h.scheduler.GetStatus())
}

// SaveSchedule 保存调度配置（合并/清理间隔、启用状态、静默时段）。
// 记录变更日志并异步写入历史。
func (h *Handler) SaveSchedule(c *gin.Context) {
	var req struct {
		MergeInterval     int  `json:"merge_interval"`
		CleanInterval     int  `json:"clean_interval"`
		MergeEnabled      bool `json:"merge_enabled"`
		CleanEnabled      bool `json:"clean_enabled"`
		BackupStartHour   *int `json:"backup_start_hour"`
		BackupStartMinute *int `json:"backup_start_minute"`
		BackupEndHour     *int `json:"backup_end_hour"`
		BackupEndMinute   *int `json:"backup_end_minute"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		failBadRequest(c, "参数错误")
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

	// 如果请求中包含静默时段字段，先更新配置再保存调度（保证原子性）
	backupChanged := false
	if req.BackupStartHour != nil || req.BackupStartMinute != nil || req.BackupEndHour != nil || req.BackupEndMinute != nil {
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
			failBadRequest(c, err.Error())
			return
		}
	}

	if err := h.scheduler.SaveSchedule(schedule); err != nil {
		failInternal(c, "保存失败")
		return
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
			h.history.Add("schedule", "", "success", detail, "")
		} else {
			detail := fmt.Sprintf("调度已保存: 合并 %s/%d分钟, 清理 %s/%d分钟",
				map[bool]string{true: "已启用", false: "已禁用"}[req.MergeEnabled], req.MergeInterval,
				map[bool]string{true: "已启用", false: "已禁用"}[req.CleanEnabled], req.CleanInterval)
			h.logger.Info(detail)
			h.history.Add("schedule", "", "success", detail, "")
		}
	}()
	ok(c, gin.H{"status": "success", "schedule": h.scheduler.GetStatus()})
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

	// 检查 ffmpeg 可用性
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		checks["ffmpeg_ok"] = true
		checks["ffmpeg_path"] = path
	}
	// 检查 ffprobe 可用性
	if path, err := exec.LookPath("ffprobe"); err == nil {
		checks["ffprobe_ok"] = true
		checks["ffprobe_path"] = path
	}
	// 测试 ffmpeg 进程组创建能力
	if checks["ffmpeg_ok"].(bool) {
		cmd := exec.Command("ffmpeg", "-version")
		setProcessGroup(cmd)
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

	ok(c, checks)
}
