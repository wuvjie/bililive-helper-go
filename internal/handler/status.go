package handler

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"bililive-helper-go/internal/config"
	"bililive-helper-go/internal/fsutil"
	"bililive-helper-go/internal/service"
	"bililive-helper-go/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// Status 返回简要系统状态：磁盘使用率、主播数量、总容量。
func (h *Handler) Status(c *gin.Context) {
	cfg := h.config.ToDTO()
	disk, err := utils.GetDiskUsage(cfg.TargetDir)
	if err != nil {
		failInternal(c, fmt.Sprintf("获取磁盘信息失败（%s）: %v", cfg.TargetDir, err))
		return
	}
	streamers, _, _ := h.scanAllStreamers(cfg.TargetDir)
	ok(c, gin.H{
		"disk_usage": disk.UsedPct,
		"streamers":  streamers,
		"total_gb":   round1(float64(disk.Total) / 1073741824),
	})
}

// StatusDetail 返回详细系统状态：磁盘用量、待合并文件、主播列表、调度状态。
func (h *Handler) StatusDetail(c *gin.Context) {
	cfg := h.config.ToDTO()
	disk, err := utils.GetDiskUsage(cfg.TargetDir)
	if err != nil {
		failInternal(c, fmt.Sprintf("获取磁盘信息失败（%s）: %v", cfg.TargetDir, err))
		return
	}

	streamers, pendingCount, pendingSize := h.scanAllStreamers(cfg.TargetDir)
	ok(c, gin.H{
		"disk": gin.H{
			"usage_pct": disk.UsedPct,
			"total_gb":  round1(float64(disk.Total) / 1073741824),
			"used_gb":   round1(float64(disk.Used) / 1073741824),
			"free_gb":   round1(float64(disk.Free) / 1073741824),
		},
		"pending": gin.H{
			"original_files":   pendingCount,
			"original_size_gb": round2(float64(pendingSize) / 1073741824),
		},
		"streamers": streamers,
		"schedule":  h.scheduler.GetStatus(),
	})
}

// Stats 返回统计数据：今日/本月的合并清理次数和数据量，以及近 7 天每日趋势。
func (h *Handler) Stats(c *gin.Context) {
	records := h.history.GetAllRecords()
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	month := now.Format("2006-01")

	var todayMerge, todayClean, yesterdayMerge, yesterdayClean, monthMerge, monthClean, totalClean int
	var todayMergeBytes, todayCleanBytes, yesterdayMergeBytes, yesterdayCleanBytes, monthMergeBytes, monthCleanBytes, totalCleanBytes int64

	for _, r := range records {
		// 有效清理：状态成功且实际产生了磁盘释放（排除未达水位无文件删除的空跑扫描）
		isEffectiveClean := r.Task == "clean" && (r.Status == "success" || r.Status == "partial") && (r.FreedBytes > 0 || r.FilesCount > 0)
		if isEffectiveClean {
			totalClean++
			totalCleanBytes += r.FreedBytes
		}

		if len(r.Time) < 10 {
			continue
		}
		recordDay := r.Time[:10]
		recordMonth := r.Time[:7]

		// 有效合并：状态成功/部分成功且实际发生了分片合并或产出体积（排除无待合并文件的空跑扫描）
		isEffectiveMerge := r.Task == "merge" && (r.Status == "success" || r.Status == "partial") && (r.MergedBytes > 0 || r.FilesCount > 0)

		if recordDay == today {
			if isEffectiveMerge {
				todayMerge++
				todayMergeBytes += r.MergedBytes
			}
			if isEffectiveClean {
				todayClean++
				todayCleanBytes += r.FreedBytes
			}
		}
		if recordDay == yesterday {
			if isEffectiveMerge {
				yesterdayMerge++
				yesterdayMergeBytes += r.MergedBytes
			}
			if isEffectiveClean {
				yesterdayClean++
				yesterdayCleanBytes += r.FreedBytes
			}
		}
		if recordMonth == month {
			if isEffectiveMerge {
				monthMerge++
				monthMergeBytes += r.MergedBytes
			}
			if isEffectiveClean {
				monthClean++
				monthCleanBytes += r.FreedBytes
			}
		}
	}

	// 构建近 7 天每日统计
	type dayStat struct {
		Date       string `json:"date"`
		MergeCount int    `json:"merge_count"`
		MergeBytes int64  `json:"merge_bytes"`
		CleanCount int    `json:"clean_count"`
		CleanBytes int64  `json:"clean_bytes"`
	}
	dailyStats := make([]dayStat, 7)
	for i := 6; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		dailyStats[6-i] = dayStat{Date: d}
	}
	for _, r := range records {
		if len(r.Time) < 10 {
			continue
		}
		recordDay := r.Time[:10]
		isEffectiveMerge := r.Task == "merge" && (r.Status == "success" || r.Status == "partial") && (r.MergedBytes > 0 || r.FilesCount > 0)
		isEffectiveClean := r.Task == "clean" && (r.Status == "success" || r.Status == "partial") && (r.FreedBytes > 0 || r.FilesCount > 0)

		for i := range dailyStats {
			if dailyStats[i].Date == recordDay {
				if isEffectiveMerge {
					dailyStats[i].MergeCount++
					dailyStats[i].MergeBytes += r.MergedBytes
				}
				if isEffectiveClean {
					dailyStats[i].CleanCount++
					dailyStats[i].CleanBytes += r.FreedBytes
				}
			}
		}
	}

	var past7MergeBytes int64
	for _, d := range dailyStats {
		past7MergeBytes += d.MergeBytes
	}
	dailyOutputGB := float64(past7MergeBytes) / 1073741824 / 7.0

	daysUntilFull := 0.0
	cfg := h.config.Snapshot()
	if disk, err := utils.GetDiskUsage(cfg.TargetDir); err == nil && dailyOutputGB > 0 {
		freeGB := float64(disk.Free) / 1073741824
		daysUntilFull = freeGB / dailyOutputGB
	}

	ok(c, gin.H{
		"today": gin.H{
			"merge_count": todayMerge,
			"merge_bytes": todayMergeBytes,
			"clean_count": todayClean,
			"clean_bytes": todayCleanBytes,
		},
		"yesterday": gin.H{
			"merge_count": yesterdayMerge,
			"merge_bytes": yesterdayMergeBytes,
			"clean_count": yesterdayClean,
			"clean_bytes": yesterdayCleanBytes,
		},
		"month": gin.H{
			"merge_count": monthMerge,
			"merge_bytes": monthMergeBytes,
			"clean_count": monthClean,
			"clean_bytes": monthCleanBytes,
		},
		"total": gin.H{
			"clean_count": totalClean,
			"clean_bytes": totalCleanBytes,
		},
		"daily":           dailyStats,
		"daily_output_gb": round1(dailyOutputGB),
		"days_until_full": math.Round(daysUntilFull),
	})
}

// GetStreamers 返回所有主播列表（文件数、磁盘占用），按大小降序排列。
func (h *Handler) GetStreamers(c *gin.Context) {
	cfg := h.config.ToDTO()
	streamers, _, _ := h.scanAllStreamers(cfg.TargetDir)
	ok(c, streamers)
}

// scanAllStreamers 一次性扫描所有主播目录，返回主播列表、待合并文件数和总大小。
// 每个主播包含 name、files、size_bytes、size_gb 和 mtime（最近视频的修改时间戳，无文件时为 0）。
func (h *Handler) scanAllStreamers(root string) ([]gin.H, int, int64) {
	type streamerInfo struct {
		name         string
		size         int64
		count        int
		pendingCount int
		pendingSize  int64
		latestTime   time.Time
	}

	dirs, err := fsutil.ScanStreamerDirs(root)
	if err != nil {
		return []gin.H{}, 0, 0
	}

	var infos []streamerInfo
	for _, dir := range dirs {
		var totalSize, pendSize int64
		var totalCount, pendCount int
		var latest time.Time
		for _, fe := range dir.Files {
			if fe.IsDir() {
				continue
			}
			name := fe.Name()
			if !utils.IsVideoFile(name) {
				continue
			}
			info, _ := fe.Info()
			if info == nil {
				continue
			}
			totalSize += info.Size()
			totalCount++
			fileTime := info.ModTime()
			if _, dt, ok := utils.ParseFilename(name); ok {
				fileTime = dt
			}
			if fileTime.After(latest) {
				latest = fileTime
			}
			if !utils.IsMergedFile(name) {
				pendSize += info.Size()
				pendCount++
			}
		}
		infos = append(infos, streamerInfo{
			name: dir.Name, size: totalSize, count: totalCount,
			pendingCount: pendCount, pendingSize: pendSize, latestTime: latest,
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].size > infos[j].size
	})

	var totalPendingCount int
	var totalPendingSize int64
	streamers := make([]gin.H, 0, len(infos))
	for _, si := range infos {
		totalPendingCount += si.pendingCount
		totalPendingSize += si.pendingSize
		var mtime int64
		if !si.latestTime.IsZero() {
			mtime = si.latestTime.Unix()
		}
		var rule config.StreamerRule
		hasRule := false
		if r, ok := h.config.GetStreamerRule(si.name); ok {
			rule = r
			hasRule = true
		}
		streamers = append(streamers, gin.H{
			"name":       si.name,
			"files":      si.count,
			"size_bytes": si.size,
			"size_gb":    round2(float64(si.size) / 1073741824),
			"mtime":      mtime,
			"rule":       rule,
			"has_rule":   hasRule,
		})
	}

	return streamers, totalPendingCount, totalPendingSize
}

// GetStreamerFiles 返回指定主播的所有视频文件信息。
func (h *Handler) GetStreamerFiles(c *gin.Context) {
	streamer := c.Param("name")
	if !utils.ValidateFilename(streamer) {
		failBadRequest(c, "主播名包含非法字符")
		return
	}
	cfg := h.config.ToDTO()
	folder := filepath.Join(cfg.TargetDir, streamer)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		failNotFound(c, fmt.Sprintf("主播 %s 的目录不存在", streamer))
		return
	}

	files := []gin.H{}
	entries, _ := os.ReadDir(folder)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !utils.IsVideoFile(name) {
			continue
		}
		info, err := entry.Info()
		if err != nil || info == nil {
			continue
		}
		fileTime := info.ModTime()
		if _, dt, ok := utils.ParseFilename(name); ok {
			fileTime = dt
		}
		files = append(files, gin.H{
			"name":      name,
			"path":      filepath.Join(folder, name),
			"rel_path":  filepath.Join(streamer, name),
			"size":      info.Size(),
			"size_str":  utils.FormatSize(info.Size()),
			"mtime":     fileTime.Unix(),
			"is_merged": utils.IsMergedFile(name),
		})
	}
	// 按修改/录制时间倒序排列（最新的录像文件排在最前）
	sort.Slice(files, func(i, j int) bool {
		mtimeI, _ := files[i]["mtime"].(int64)
		mtimeJ, _ := files[j]["mtime"].(int64)
		if mtimeI != mtimeJ {
			return mtimeI > mtimeJ
		}
		nameI, _ := files[i]["name"].(string)
		nameJ, _ := files[j]["name"].(string)
		return nameI > nameJ
	})
	ok(c, files)
}

// GetStreamerRule 返回指定主播的录像规则。
func (h *Handler) GetStreamerRule(c *gin.Context) {
	name := c.Param("name")
	if !utils.ValidateFilename(name) {
		failBadRequest(c, "主播名包含非法字符")
		return
	}
	rule, exists := h.config.GetStreamerRule(name)
	ok(c, gin.H{
		"streamer": name,
		"rule":     rule,
		"exists":   exists,
	})
}

// SaveStreamerRule 保存或更新指定主播的录像规则。
func (h *Handler) SaveStreamerRule(c *gin.Context) {
	name := c.Param("name")
	if !utils.ValidateFilename(name) {
		failBadRequest(c, "主播名包含非法字符")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8192)
	var rule config.StreamerRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		failBadRequest(c, "请求格式错误: "+err.Error())
		return
	}

	// 标准化字段并与当前规则比对：若未发生实质变更，保持幂等，避免无谓磁盘写盘与虚假审计历史
	normRule := rule
	if normRule.TrimStart == "" && normRule.TrimEnabled {
		normRule.TrimStart = "00:00:00"
	}
	normRule.TrimMode = ""
	normRule.DurationMinutes = 0
	normRule.StartTime = ""
	normRule.EndTime = ""

	if oldRule, exists := h.config.GetStreamerRule(name); exists {
		if oldRule.Protected == normRule.Protected &&
			oldRule.MinKeep == normRule.MinKeep &&
			oldRule.TrimEnabled == normRule.TrimEnabled &&
			oldRule.TrimStart == normRule.TrimStart &&
			oldRule.TrimEnd == normRule.TrimEnd {
			ok(c, gin.H{"streamer": name, "rule": rule, "unchanged": true})
			return
		}
	}

	if err := h.config.SetStreamerRule(name, rule); err != nil {
		h.logger.Error("保存主播规则失败", zap.String("streamer", name), zap.Error(err))
		fail(c, http.StatusInternalServerError, "保存主播规则失败")
		return
	}
	h.history.Add("config", name, "success", "更新主播录像规则", "")
	ok(c, gin.H{"streamer": name, "rule": rule})
}

// PlayStreamerVideo 支持 HTTP Range 流式点播播放录播视频（带防路径穿越校验）。
func (h *Handler) PlayStreamerVideo(c *gin.Context) {
	streamer := c.Param("name")
	file := c.Query("file")
	if !utils.ValidateFilename(streamer) || !utils.ValidateFilename(file) {
		failBadRequest(c, "参数包含非法字符")
		return
	}
	if !utils.IsVideoFile(file) {
		failBadRequest(c, "仅支持视频点播")
		return
	}

	cfg := h.config.Snapshot()
	fullPath := filepath.Join(cfg.TargetDir, streamer, file)
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		failNotFound(c, "视频文件不存在")
		return
	}

	c.Header("Content-Disposition", "inline; filename="+filepath.Base(file))
	http.ServeFile(c.Writer, c.Request, fullPath)
}
// DeleteStreamerFile 删除指定主播的单个录像文件（防路径穿越 + 记录释放空间历史）。
func (h *Handler) DeleteStreamerFile(c *gin.Context) {
	streamer := c.Param("name")
	file := c.Query("file")
	if !utils.ValidateFilename(streamer) || !utils.ValidateFilename(file) {
		failBadRequest(c, "参数包含非法字符")
		return
	}
	if !utils.IsVideoFile(file) {
		failBadRequest(c, "仅支持删除视频文件")
		return
	}

	// 检查该主播是否正在进行合并，避免删除正在被 FFmpeg 读取的文件
	if h.merge.IsStreamerLocked(streamer) {
		failBadRequest(c, fmt.Sprintf("主播 %s 正在进行合并处理中，暂时无法删除文件", streamer))
		return
	}

	cfg := h.config.Snapshot()
	fullPath := filepath.Join(cfg.TargetDir, streamer, file)
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		failNotFound(c, "待删除的视频文件不存在")
		return
	}

	fileSize := info.Size()
	if err := os.Remove(fullPath); err != nil {
		h.logger.Error("删除视频文件失败", zap.String("path", fullPath), zap.Error(err))
		fail(c, http.StatusInternalServerError, "删除文件失败: "+err.Error())
		return
	}

	// 级联清理同名伴生文件（XML 弹幕、封面图、元数据等），防止产生孤儿文件
	service.CleanCompanionFiles(fullPath)

	// 记录到操作历史，并计入 clean 任务以计入累计释放磁盘空间
	detail := fmt.Sprintf("手动删除视频: %s (%s)", file, utils.FormatSize(fileSize))
	h.history.AddWithStats("clean", streamer, "success", 1, fileSize, 0, 0, detail, "")

	ok(c, gin.H{
		"deleted":     file,
		"freed_bytes": fileSize,
		"freed_str":   utils.FormatSize(fileSize),
	})
}
