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

func (h *Handler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, h.config.ToDTO())
}

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

	// Content analysis (uses whitelist to exclude protected files)
	analysis := analyzeContent(cfg.TargetDir, cfg.WhitelistKeywords)

	dailyGB := analysis.DailyOutputGB
	daysUntilFull := 0.0
	if dailyGB > 0 {
		daysUntilFull = freeGB / dailyGB
	}

	// Risk level
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

	// Dynamic recommendation based on analysis
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

	// Adjust based on risk level
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

	// Fix #1: ensure trigger is always above current usage with a safe margin
	// to prevent immediate cleanup on config apply
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

	// Fix #6: streamer count adjustment — only when disk is under real pressure
	if analysis.StreamerCount > 15 && usedPct > 65 {
		maxDelete = maxDelete * 2
		if minKeep > 2 && riskLevel != "low" {
			minKeep = 2
		}
	}

	// Fix #7: GAP_MINUTES — use median gap from actual data, clamped to [10, 60]
	gapMinutes := 20
	if analysis.MedianGapMinutes > 0 {
		gapMinutes = int(analysis.MedianGapMinutes)
		if gapMinutes < 10 {
			gapMinutes = 10
		}
		if gapMinutes > 60 {
			gapMinutes = 60
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

	// Fix #8: validate recommended config before returning
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

// contentAnalysis holds scan results about the video library.
type contentAnalysis struct {
	StreamerCount    int
	TotalVideos      int
	MergedCount      int
	DailyOutputGB    float64
	MedianGapMinutes float64
}

// analyzeContent scans the target directory to understand the video library.
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
			// Skip whitelisted files — they won't be cleaned
			if containsAny(name, whitelist) || containsAny(entry.Name(), whitelist) {
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

			// Count recent files (last 7 days) for daily output estimate
			if now.Sub(mtime) < 7*24*time.Hour {
				streamerRecentSize += info.Size()
			}
		}

		// Per-streamer daily output from recent 7-day window
		if streamerRecentSize > 0 {
			result.DailyOutputGB += float64(streamerRecentSize) / 1073741824 / 7
		}

		// Compute per-streamer gaps between consecutive files
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

	// Median gap across all streamers
	if len(allGaps) > 0 {
		sort.Float64s(allGaps)
		mid := len(allGaps) / 2
		if len(allGaps)%2 == 0 {
			result.MedianGapMinutes = (allGaps[mid-1] + allGaps[mid]) / 2
		} else {
			result.MedianGapMinutes = allGaps[mid]
		}
	}

	// Fallback: no recent data, estimate from streamer count
	if result.DailyOutputGB < 1 {
		est := float64(result.StreamerCount) * 5
		if est < 5 {
			est = 5
		}
		result.DailyOutputGB = est
	}

	return result
}

func containsAny(s string, keywords []string) bool {
	s = strings.ToLower(s)
	for _, kw := range keywords {
		if strings.Contains(s, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func (h *Handler) DefaultConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"TRIGGER_THRESHOLD":     80,
		"TARGET_THRESHOLD":      60,
		"MIN_KEEP_PER_STREAMER": 3,
		"SAFE_AGE_MINUTES":     120,
		"GAP_MINUTES":           60,
		"MERGE_AGE_MINUTES":    30,
		"MAX_DELETE_PER_RUN":    10,
		"BACKUP_START_HOUR":    4,
		"BACKUP_START_MINUTE":  0,
		"BACKUP_END_HOUR":      12,
		"BACKUP_END_MINUTE":    0,
	})
}

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
