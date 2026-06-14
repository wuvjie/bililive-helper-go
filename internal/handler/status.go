package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"bililive-helper-go/internal/fsutil"
	"bililive-helper-go/internal/utils"

	"github.com/gin-gonic/gin"
)

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
		"total_gb":   float64(int(float64(disk.Total)/1073741824*10)) / 10,
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
			"total_gb":  float64(int(float64(disk.Total)/1073741824*10)) / 10,
			"used_gb":   float64(int(float64(disk.Used)/1073741824*10)) / 10,
			"free_gb":   float64(int(float64(disk.Free)/1073741824*10)) / 10,
		},
		"pending": gin.H{
			"original_files":   pendingCount,
			"original_size_gb": float64(int(float64(pendingSize)/1073741824*100)) / 100,
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
	month := now.Format("2006-01")

	var todayMerge, todayClean, monthMerge, monthClean int
	var todayMergeBytes, todayCleanBytes, monthMergeBytes, monthCleanBytes int64

	for _, r := range records {
		if len(r.Time) < 10 {
			continue
		}
		recordDay := r.Time[:10]
		recordMonth := r.Time[:7]

		if recordDay == today {
			if r.Task == "merge" && r.Status == "success" {
				todayMerge++
				todayMergeBytes += r.MergedBytes
			}
			if r.Task == "clean" && r.Status == "success" {
				todayClean++
				todayCleanBytes += r.FreedBytes
			}
		}
		if recordMonth == month {
			if r.Task == "merge" && r.Status == "success" {
				monthMerge++
				monthMergeBytes += r.MergedBytes
			}
			if r.Task == "clean" && r.Status == "success" {
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
		for i := range dailyStats {
			if dailyStats[i].Date == recordDay && r.Status == "success" {
				if r.Task == "merge" {
					dailyStats[i].MergeCount++
					dailyStats[i].MergeBytes += r.MergedBytes
				}
				if r.Task == "clean" {
					dailyStats[i].CleanCount++
					dailyStats[i].CleanBytes += r.FreedBytes
				}
			}
		}
	}

	ok(c, gin.H{
		"today": gin.H{
			"merge_count": todayMerge,
			"merge_bytes": todayMergeBytes,
			"clean_count": todayClean,
			"clean_bytes": todayCleanBytes,
		},
		"month": gin.H{
			"merge_count": monthMerge,
			"merge_bytes": monthMergeBytes,
			"clean_count": monthClean,
			"clean_bytes": monthCleanBytes,
		},
		"daily": dailyStats,
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
			if info.ModTime().After(latest) {
				latest = info.ModTime()
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
		streamers = append(streamers, gin.H{
			"name":       si.name,
			"files":      si.count,
			"size_bytes": si.size,
			"size_gb":    float64(int(float64(si.size)/1073741824*100)) / 100,
			"mtime":      mtime,
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
		files = append(files, gin.H{
			"name":      name,
			"size":      info.Size(),
			"size_str":  utils.FormatSize(info.Size()),
			"mtime":     info.ModTime().Unix(),
			"is_merged": utils.IsMergedFile(name),
		})
	}
	ok(c, files)
}
