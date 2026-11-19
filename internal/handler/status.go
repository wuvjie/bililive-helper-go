package handler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bililive-helper-go/internal/config"
	"bililive-helper-go/internal/ffmpeg"
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

	type dayStat struct {
		Date       string `json:"date"`
		MergeCount int    `json:"merge_count"`
		MergeBytes int64  `json:"merge_bytes"`
		CleanCount int    `json:"clean_count"`
		CleanBytes int64  `json:"clean_bytes"`
	}
	dailyStats := make([]dayStat, 7)
	dailyIndex := make(map[string]int, 7)
	for i := 6; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		idx := 6 - i
		dailyStats[idx] = dayStat{Date: d}
		dailyIndex[d] = idx
	}

	var todayMerge, todayClean, yesterdayMerge, yesterdayClean, monthMerge, monthClean, totalClean, totalMerge int
	var todayMergeBytes, todayCleanBytes, yesterdayMergeBytes, yesterdayCleanBytes, monthMergeBytes, monthCleanBytes, totalCleanBytes, totalMergeBytes int64

	for _, r := range records {
		isEffective := r.Status == "success" || r.Status == "partial"
		// 有效清理：状态成功且实际产生了磁盘释放（排除未达水位无文件删除的空跑扫描）
		isEffectiveClean := r.Task == "clean" && isEffective && (r.FreedBytes > 0 || r.FilesCount > 0)
		// 有效合并：状态成功/部分成功且实际发生了分片合并或产出体积（排除无待合并文件的空跑扫描）
		isEffectiveMerge := r.Task == "merge" && isEffective && (r.MergedBytes > 0 || r.FilesCount > 0)

		if isEffectiveClean {
			totalClean++
			totalCleanBytes += r.FreedBytes
		}
		if isEffectiveMerge {
			totalMerge++
			totalMergeBytes += r.MergedBytes
		}

		if len(r.Time) < 10 {
			continue
		}
		recordDay := r.Time[:10]
		recordMonth := r.Time[:7]

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
		if idx, ok := dailyIndex[recordDay]; ok {
			if isEffectiveMerge {
				dailyStats[idx].MergeCount++
				dailyStats[idx].MergeBytes += r.MergedBytes
			}
			if isEffectiveClean {
				dailyStats[idx].CleanCount++
				dailyStats[idx].CleanBytes += r.FreedBytes
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
			"merge_count": totalMerge,
			"merge_bytes": totalMergeBytes,
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

// liveActiveWindow 判定主播是否正在录制的磁盘文件活跃时间窗口（与后台扫描器保持一致）。
const liveActiveWindow = 5 * time.Minute

// streamerCacheTTL 状态/主播列表扫盘结果的短时内存缓存有效期，降低 NAS/机械硬盘频繁唤醒频率。
const streamerCacheTTL = 5 * time.Second

// buildStandaloneMP4Set 识别目录中已结束写入且无需合并的单场次独立 MP4 文件。
// 按主播分组后根据文件名时间顺序聚簇（相邻 <= 3 小时视为同一场次），
// 既避免“同一天早晚两场单分片被误判待合并”，也避免“跨午夜连续分片被误判已完成”。
func buildStandaloneMP4Set(entries []os.DirEntry, now time.Time) map[string]bool {
	const clusterGap = 3 * time.Hour

	type parsedVideo struct {
		name        string
		streamerKey string
		dt          time.Time
		modTime     time.Time
		isMP4       bool
	}
	byStreamer := make(map[string][]parsedVideo)
	standalone := make(map[string]bool)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !utils.IsVideoFile(name) || utils.IsProcessedFile(name) || utils.IsTempVideoArtifact(name) {
			continue
		}
		info, err := entry.Info()
		if err != nil || info == nil {
			continue
		}
		isMP4 := strings.EqualFold(filepath.Ext(name), ".mp4")
		key, dt, ok := utils.ParseFilename(name)
		if !ok {
			// 非标准命名文件各自独立成簇（仅检测活跃写入窗口）
			if isMP4 && now.Sub(info.ModTime()) > liveActiveWindow {
				standalone[name] = true
			}
			continue
		}
		sk := utils.StreamerKeyFromKey(key)
		byStreamer[sk] = append(byStreamer[sk], parsedVideo{
			name:        name,
			streamerKey: sk,
			dt:          dt,
			modTime:     info.ModTime(),
			isMP4:       isMP4,
		})
	}

	for _, items := range byStreamer {
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].dt.Equal(items[j].dt) {
				return items[i].name < items[j].name
			}
			return items[i].dt.Before(items[j].dt)
		})
		clusterStart := 0
		for i := 1; i <= len(items); i++ {
			if i == len(items) || items[i].dt.Sub(items[i-1].dt) > clusterGap {
				if i-clusterStart == 1 {
					only := items[clusterStart]
					if only.isMP4 && now.Sub(only.modTime) > liveActiveWindow {
						standalone[only.name] = true
					}
				}
				clusterStart = i
			}
		}
	}
	return standalone
}

func (h *Handler) attachRulesToCachedRows(rows []cachedStreamerRow) []gin.H {
	streamers := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		var rule config.StreamerRule
		hasRule := false
		if cfgRule, ok := h.config.GetStreamerRule(r.Name); ok {
			rule = cfgRule
			hasRule = true
		}
		streamers = append(streamers, gin.H{
			"name":       r.Name,
			"files":      r.Files,
			"size_bytes": r.SizeBytes,
			"size_gb":    r.SizeGB,
			"mtime":      r.Mtime,
			"is_live":    r.IsLive,
			"rule":       rule,
			"has_rule":   hasRule,
		})
	}
	return streamers
}

// scanAllStreamers 扫描所有主播目录并返回主播列表、待合并文件数和总大小。
// 内置 5 秒 TTL 内存快照缓存以减少轮询导致的机械硬盘唤醒；命中缓存时实时读取最新的主播个性化规则。
func (h *Handler) scanAllStreamers(root string) ([]gin.H, int, int64) {
	now := time.Now()
	h.scanCache.mu.RLock()
	if h.scanCache.root == root && !h.scanCache.fetchedAt.IsZero() && now.Sub(h.scanCache.fetchedAt) < streamerCacheTTL {
		rows := make([]cachedStreamerRow, len(h.scanCache.rows))
		copy(rows, h.scanCache.rows)
		pendingCount := h.scanCache.pendingCount
		pendingSize := h.scanCache.pendingSize
		h.scanCache.mu.RUnlock()
		return h.attachRulesToCachedRows(rows), pendingCount, pendingSize
	}
	h.scanCache.mu.RUnlock()

	streamers, pendingCount, pendingSize := h.scanAllStreamersAt(root, now)
	rows := make([]cachedStreamerRow, 0, len(streamers))
	for _, s := range streamers {
		rows = append(rows, cachedStreamerRow{
			Name:      s["name"].(string),
			Files:     s["files"].(int),
			SizeBytes: s["size_bytes"].(int64),
			SizeGB:    s["size_gb"].(float64),
			Mtime:     s["mtime"].(int64),
			IsLive:    s["is_live"].(bool),
		})
	}

	h.scanCache.mu.Lock()
	h.scanCache.root = root
	h.scanCache.fetchedAt = now
	h.scanCache.rows = rows
	h.scanCache.pendingCount = pendingCount
	h.scanCache.pendingSize = pendingSize
	h.scanCache.mu.Unlock()

	return streamers, pendingCount, pendingSize
}

func (h *Handler) scanAllStreamersAt(root string, now time.Time) ([]gin.H, int, int64) {
	type streamerInfo struct {
		name            string
		size            int64
		count           int
		pendingCount    int
		pendingSize     int64
		latestTime      time.Time
		latestWriteTime time.Time
	}

	dirs, err := fsutil.ScanStreamerDirs(root)
	if err != nil {
		return []gin.H{}, 0, 0
	}

	var infos []streamerInfo
	for _, dir := range dirs {
		var totalSize, pendSize int64
		var totalCount, pendCount int
		var latest, latestWrite time.Time
		standaloneMP4s := buildStandaloneMP4Set(dir.Files, now)
		for _, fe := range dir.Files {
			if fe.IsDir() {
				continue
			}
			name := fe.Name()
			if !utils.IsVideoFile(name) || utils.IsTempVideoArtifact(name) {
				continue
			}
			info, _ := fe.Info()
			if info == nil {
				continue
			}
			totalSize += info.Size()
			totalCount++
			modTime := info.ModTime()
			fileTime := modTime
			if _, dt, ok := utils.ParseFilename(name); ok {
				fileTime = dt
			}
			if fileTime.After(latest) {
				latest = fileTime
			}
			isCompleted := utils.IsProcessedFile(name) || standaloneMP4s[name]
			if !isCompleted {
				pendSize += info.Size()
				pendCount++
			}
			if !utils.IsProcessedFile(name) {
				// 直播状态必须基于磁盘真实写入时间 ModTime，避免被文件名里的开播时间覆盖导致开播超过 30 分钟熄灭
				if modTime.After(latestWrite) {
					latestWrite = modTime
				}
			}
		}
		infos = append(infos, streamerInfo{
			name: dir.Name, size: totalSize, count: totalCount,
			pendingCount: pendCount, pendingSize: pendSize,
			latestTime: latest, latestWriteTime: latestWrite,
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
		isLive := !si.latestWriteTime.IsZero() && now.Sub(si.latestWriteTime) < liveActiveWindow
		displayTime := si.latestTime
		if isLive && si.latestWriteTime.After(displayTime) {
			displayTime = si.latestWriteTime
		}
		var mtime int64
		if !displayTime.IsZero() {
			mtime = displayTime.Unix()
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
			"is_live":    isLive,
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
	if abs, err := filepath.Abs(folder); err == nil {
		folder = abs
	}
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		failNotFound(c, fmt.Sprintf("主播 %s 的目录不存在", streamer))
		return
	}

	files := []gin.H{}
	entries, _ := os.ReadDir(folder)
	standaloneMP4s := buildStandaloneMP4Set(entries, time.Now())
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !utils.IsVideoFile(name) || utils.IsTempVideoArtifact(name) {
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
			"is_merged":       utils.IsProcessedFile(name) || standaloneMP4s[name],
			"is_trimmed_file": utils.IsTrimmedFile(name),
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

	oldRule, exists := h.config.GetStreamerRule(name)
	if exists {
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
	h.history.Add("config", name, "success", config.DiffStreamerRule(oldRule, normRule, exists), "")
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

// ProbeStreamerFile 探测指定录像文件的元数据（实际时长、是否含 bh_trimmed 裁剪标记及伴生弹幕状态）。
// 用于手动单文件剪辑弹窗的预检提示。
func (h *Handler) ProbeStreamerFile(c *gin.Context) {
	streamer := c.Param("name")
	file := c.Query("file")
	if !utils.ValidateFilename(streamer) || !utils.ValidateFilename(file) {
		failBadRequest(c, "参数包含非法字符")
		return
	}
	if !utils.IsVideoFile(file) || utils.IsTempVideoArtifact(file) {
		failBadRequest(c, "仅支持探测有效视频文件")
		return
	}

	cfg := h.config.Snapshot()
	folder := filepath.Join(cfg.TargetDir, streamer)
	fullPath := filepath.Join(folder, file)
	absFolder, err1 := filepath.Abs(folder)
	absPath, err2 := filepath.Abs(fullPath)
	if err1 != nil || err2 != nil || !strings.HasPrefix(absPath, absFolder+string(os.PathSeparator)) {
		failBadRequest(c, "文件路径越界")
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		failNotFound(c, "目标视频文件不存在")
		return
	}

	probeCtx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	dur, durErr := ffmpeg.ProbeDuration(probeCtx, fullPath)
	if durErr != nil || dur < 0 {
		dur = 0
	}
	isTrimmed := ffmpeg.IsVideoTrimmed(probeCtx, fullPath)

	ok(c, gin.H{
		"file":            file,
		"size":            info.Size(),
		"size_str":        utils.FormatSize(info.Size()),
		"duration_sec":    math.Round(dur*10) / 10,
		"probe_ok":        durErr == nil && dur > 0,
		"is_trimmed":      isTrimmed,
		"is_trimmed_file": utils.IsTrimmedFile(file),
		"has_danmaku":     service.HasCompanionDanmaku(fullPath),
	})
}

// ManualTrimStreamerFile 对指定主播的单个 .mp4 录像执行按需手动无损剪辑。
// 仅用户显式调用时触发，完全隔离于定时自动扫描任务。
func (h *Handler) ManualTrimStreamerFile(c *gin.Context) {
	streamer := c.Param("name")
	if !utils.ValidateFilename(streamer) {
		failBadRequest(c, "主播名包含非法字符")
		return
	}

	var req struct {
		File      string `json:"file" binding:"required"`
		TrimStart string `json:"trim_start"`
		TrimEnd   string `json:"trim_end"`
		Mode      string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		failBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	res, err := h.merge.ManualTrimFile(c.Request.Context(), streamer, req.File, req.TrimStart, req.TrimEnd, req.Mode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTrimNotFound):
			failNotFound(c, err.Error())
		case errors.Is(err, service.ErrTrimConflict):
			failConflict(c, err.Error())
		case errors.Is(err, service.ErrTrimBadRequest):
			failBadRequest(c, err.Error())
		default:
			failInternal(c, "手动剪辑失败: "+err.Error())
		}
		return
	}

	h.InvalidateStreamerCache()
	ok(c, res)
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
	h.InvalidateStreamerCache()

	// 记录到操作历史，并计入 clean 任务以计入累计释放磁盘空间
	detail := fmt.Sprintf("手动删除视频: %s (%s)", file, utils.FormatSize(fileSize))
	h.history.AddWithStats("clean", streamer, "success", 1, fileSize, 0, 0, detail, "")

	ok(c, gin.H{
		"deleted":     file,
		"freed_bytes": fileSize,
		"freed_str":   utils.FormatSize(fileSize),
	})
}
