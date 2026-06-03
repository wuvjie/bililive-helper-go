package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bililive-helper/internal/config"
	"bililive-helper/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetConfig 返回当前配置（不含敏感字段）。
func (h *Handler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, h.config.ToDTO())
}

// SaveConfig 保存配置更新（部分更新，只修改请求中包含的字段）。
// 记录变更日志并异步写入历史记录。
func (h *Handler) SaveConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var changeDetail string
	if err := h.config.Apply(func() error {
		old := h.config.ToDTOSnapshot()
		h.config.ApplyFromMap(req)
		if err := h.config.Validate(); err != nil {
			return err
		}
		changeDetail = config.DiffDTO(old, h.config.ToDTOSnapshot())
		return nil
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if changeDetail != "" {
		h.logger.Info(changeDetail)
		go h.history.Add("config", "", "success", changeDetail)
	} else {
		go h.history.Add("config", "", "success", "配置未变更")
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// RecommendConfig 基于磁盘大小、内容分析和风险评估，返回智能推荐配置。
// 分析维度包括：主播数量、日产出量、磁盘满载天数、中位文件间隔等。
func (h *Handler) RecommendConfig(c *gin.Context) {
	cfg := h.config.ToDTO()
	disk, err := utils.GetDiskUsage(cfg.TargetDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取磁盘信息失败"})
		return
	}

	totalGB := float64(disk.Total) / 1073741824
	freeGB := float64(disk.Free) / 1073741824
	usedPct := disk.UsedPct

	// 内容分析 — 排除白名单文件的统计
	analysis := analyzeContent(cfg.TargetDir, cfg.WhitelistKeywords)

	dailyGB := analysis.DailyOutputGB
	daysUntilFull := 0.0
	if dailyGB > 0 {
		daysUntilFull = freeGB / dailyGB
	}

	// 风险评估：根据磁盘满载天数判断风险等级
	riskLevel := "low"
	if daysUntilFull > 0 && daysUntilFull < 7 {
		riskLevel = "critical"
	} else if daysUntilFull > 0 && daysUntilFull < 30 {
		riskLevel = "high"
	} else if usedPct > 85 {
		riskLevel = "high"
	} else if usedPct > 70 {
		riskLevel = "normal"
	}

	// 动态推荐：基于磁盘容量的基础策略，再根据风险等级调整
	var trigger, target float64
	var minKeep, safeAge, mergeAge, maxDelete int
	safeMode := "hours"

	// Base strategy from disk size
	if totalGB < 256 {
		trigger, target = 75, 60
		minKeep, safeAge, mergeAge, maxDelete = 2, 120, 20, 5
	} else if totalGB < 512 {
		trigger, target = 80, 65
		minKeep, safeAge, mergeAge, maxDelete = 3, 120, 30, 10
	} else {
		trigger, target = 85, 70
		minKeep, safeAge, mergeAge, maxDelete = 3, 120, 30, 15
	}

	// Adjust strategy based on risk level
	switch riskLevel {
	case "critical":
		trigger = usedPct - 3
		if trigger < 65 {
			trigger = 65
		}
		target = trigger - 15
		maxDelete = 30
	case "high":
		trigger = usedPct - 3
		if trigger < 70 {
			trigger = 70
		}
		target = trigger - 10
		maxDelete = max(maxDelete, 20)
	}

	// 确保触发阈值始终高于当前使用率并保留安全余量
	// 防止应用配置后立即触发清理
	minTrigger := usedPct + 3
	if minTrigger > 95 {
		minTrigger = 95
	}
	if trigger < minTrigger {
		trigger = minTrigger
	}
	if target >= trigger {
		target = trigger - 5
	}
	if target < 30 {
		target = 30
	}

	// 主播数量调整 — 仅在磁盘压力较大时生效
	if analysis.StreamerCount > 15 && usedPct > 65 {
		maxDelete = maxDelete * 2
		if minKeep > 2 && riskLevel != "low" {
			minKeep = 2
		}
	}

	// GAP_MINUTES — 使用实际文件间隔的中位数，限制在 [10, 30] 范围
	gapMinutes := 20
	if analysis.MedianGapMinutes > 0 {
		gapMinutes = int(analysis.MedianGapMinutes)
		if gapMinutes < 10 {
			gapMinutes = 10
		}
		if gapMinutes > 30 {
			gapMinutes = 30
		}
	}

	// Build reason
	var reasonParts []string
	if totalGB < 256 {
		reasonParts = append(reasonParts, fmt.Sprintf("%.0fGB小盘", totalGB))
	} else if totalGB < 512 {
		reasonParts = append(reasonParts, fmt.Sprintf("%.0fGB中盘", totalGB))
	} else {
		reasonParts = append(reasonParts, fmt.Sprintf("%.0fGB大盘", totalGB))
	}
	reasonParts = append(reasonParts, fmt.Sprintf("%d个主播", analysis.StreamerCount))
	if analysis.MergedCount > 0 {
		reasonParts = append(reasonParts, fmt.Sprintf("日产出~%.0fGB", dailyGB))
	}
	reasonParts = append(reasonParts, fmt.Sprintf("可维持%.0f天", daysUntilFull))
	reason := strings.Join(reasonParts, " · ")

	needToFreeGB := 0.0
	if usedPct > target {
		targetSize := float64(disk.Total) * target / 100
		needToFreeGB = (float64(disk.Used) - targetSize) / 1073741824
	}

	// 验证推荐配置的内部一致性
	rec := config.Config{
		TriggerThreshold:   trigger,
		TargetThreshold:    target,
		MinKeepPerStreamer: minKeep,
		SafeAgeMinutes:     safeAge,
		GapMinutes:         gapMinutes,
		MergeAgeMinutes:    mergeAge,
		MaxDeletePerRun:    maxDelete,
		SafeMode:           safeMode,
		SafeDays:           1,
		Port:               cfg.Port,
		BackupStartHour:    cfg.BackupStartHour,
		BackupStartMinute:  cfg.BackupStartMinute,
		BackupEndHour:      cfg.BackupEndHour,
		BackupEndMinute:    cfg.BackupEndMinute,
	}
	if err := rec.Validate(); err != nil {
		h.logger.Warn("推荐配置校验失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "推荐配置生成异常"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"TRIGGER_THRESHOLD":     trigger,
		"TARGET_THRESHOLD":      target,
		"MIN_KEEP_PER_STREAMER": minKeep,
		"SAFE_AGE_MINUTES":     safeAge,
		"SAFE_MODE":             safeMode,
		"MERGE_AGE_MINUTES":    mergeAge,
		"MAX_DELETE_PER_RUN":    maxDelete,
		"GAP_MINUTES":           gapMinutes,
		"analysis": gin.H{
			"streamer_count":  analysis.StreamerCount,
			"total_videos":    analysis.TotalVideos,
			"merged_count":    analysis.MergedCount,
			"daily_output_gb": dailyGB,
			"days_until_full": daysUntilFull,
		},
		"risk_level":      riskLevel,
		"reason":          reason,
		"current_usage":   usedPct,
		"total_gb":        totalGB,
		"free_gb":         freeGB,
		"need_to_free_gb": needToFreeGB,
	})
}

// contentAnalysis 保存视频库扫描分析结果。
type contentAnalysis struct {
	StreamerCount    int
	TotalVideos      int
	MergedCount      int
	DailyOutputGB    float64
	MedianGapMinutes float64
}

// analyzeContent 扫描目标目录，分析视频库内容（主播数、文件数、日产出、间隔等）。
func analyzeContent(root string, whitelist []string) contentAnalysis {
	var result contentAnalysis
	entries, err := os.ReadDir(root)
	if err != nil {
		return result
	}

	now := time.Now()
	var allGaps []float64

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		result.StreamerCount++
		folder := filepath.Join(root, entry.Name())

		var streamerModTimes []time.Time
		var streamerRecentSize int64
		folderEntries, _ := os.ReadDir(folder)
		for _, fe := range folderEntries {
			if fe.IsDir() {
				continue
			}
			name := fe.Name()
			if !utils.IsVideoFile(name) {
				continue
			}
			// 跳过白名单文件 — 它们不会被清理，排除在分析之外
			if utils.ContainsAny(name, whitelist) || utils.ContainsAny(entry.Name(), whitelist) {
				continue
			}
			result.TotalVideos++
			if utils.IsMergedFile(name) {
				result.MergedCount++
			}
			info, err := fe.Info()
			if err != nil {
				continue
			}
			mtime := info.ModTime()
			streamerModTimes = append(streamerModTimes, mtime)

			// 统计近 7 天的文件大小用于估算日产出
			if now.Sub(mtime) < 7*24*time.Hour {
				streamerRecentSize += info.Size()
			}
		}

		// Per-streamer daily output from recent 7-day window
		if streamerRecentSize > 0 {
			result.DailyOutputGB += float64(streamerRecentSize) / 1073741824 / 7
		}

		// 计算每个主播的文件间隔，用于估算中位间隔
		if len(streamerModTimes) >= 2 {
			sort.Slice(streamerModTimes, func(i, j int) bool {
				return streamerModTimes[i].Before(streamerModTimes[j])
			})
			for i := 1; i < len(streamerModTimes); i++ {
				gap := streamerModTimes[i].Sub(streamerModTimes[i-1]).Minutes()
				if gap > 0 {
					allGaps = append(allGaps, gap)
				}
			}
		}
	}

	// 计算全局中位间隔 — 用于 GAP_MINUTES 推荐
	if len(allGaps) > 0 {
		sort.Float64s(allGaps)
		mid := len(allGaps) / 2
		if len(allGaps)%2 == 0 {
			result.MedianGapMinutes = (allGaps[mid-1] + allGaps[mid]) / 2
		} else {
			result.MedianGapMinutes = allGaps[mid]
		}
	}

	// 无近期数据时的兜底估算：基于主播数量
	if result.DailyOutputGB < 1 {
		est := float64(result.StreamerCount) * 5
		if est < 5 {
			est = 5
		}
		result.DailyOutputGB = est
	}

	return result
}

// DefaultConfig 返回默认配置值。
func (h *Handler) DefaultConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"TRIGGER_THRESHOLD":     80,
		"TARGET_THRESHOLD":      60,
		"MIN_KEEP_PER_STREAMER": 3,
		"SAFE_AGE_MINUTES":     120,
		"GAP_MINUTES":           30,
		"MERGE_AGE_MINUTES":    30,
		"MAX_DELETE_PER_RUN":    10,
		"BACKUP_START_HOUR":    4,
		"BACKUP_START_MINUTE":  0,
		"BACKUP_END_HOUR":      12,
		"BACKUP_END_MINUTE":    0,
	})
}

// ExportConfig 导出完整配置数据（含调度配置和最近 100 条历史记录）。
func (h *Handler) ExportConfig(c *gin.Context) {
	cfg := h.config.ToDTO()
	schedule := h.scheduler.GetStatus()
	records := h.history.GetAllRecords()
	if len(records) > 100 {
		records = records[len(records)-100:]
	}

	c.JSON(http.StatusOK, gin.H{
		"version":     "2.0.0",
		"exported_at": time.Now().Format(time.RFC3339),
		"config":      cfg,
		"schedule":    schedule,
		"history":     records,
	})
}

// ImportConfig 从导入数据中恢复配置。
func (h *Handler) ImportConfig(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	cfgData, ok := data["config"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少配置数据"})
		return
	}

	if err := h.config.Apply(func() error {
		h.config.ApplyFromMap(cfgData)
		return h.config.Validate()
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go h.history.Add("config", "", "success", "配置已导入")
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "配置导入成功"})
}
