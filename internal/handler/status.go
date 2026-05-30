package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bililive-helper/internal/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Status(c *gin.Context) {
	cfg := h.config.ToDTO()
	disk, err := utils.GetDiskUsage(cfg.TargetDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取磁盘信息失败"})
		return
	}
	streamers, _, _ := h.scanAllStreamers(cfg.TargetDir)
	c.JSON(http.StatusOK, gin.H{
		"disk_usage": disk.UsedPct,
		"streamers":  streamers,
		"total_gb":   float64(int(float64(disk.Total)/1073741824*10)) / 10,
	})
}

func (h *Handler) StatusDetail(c *gin.Context) {
	cfg := h.config.ToDTO()
	disk, err := utils.GetDiskUsage(cfg.TargetDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取磁盘信息失败"})
		return
	}

	streamers, pendingCount, pendingSize := h.scanAllStreamers(cfg.TargetDir)
	c.JSON(http.StatusOK, gin.H{
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

	// Build 7-day daily stats
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

	c.JSON(http.StatusOK, gin.H{
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

func (h *Handler) GetStreamers(c *gin.Context) {
	cfg := h.config.ToDTO()
	streamers, _, _ := h.scanAllStreamers(cfg.TargetDir)
	c.JSON(http.StatusOK, streamers)
}

// scanAllStreamers performs a single scan of all streamer directories,
// returning the streamer list, pending file count, and pending total size.
func (h *Handler) scanAllStreamers(root string) ([]gin.H, int, int64) {
	type streamerInfo struct {
		name        string
		size        int64
		count       int
		pendingCount int
		pendingSize  int64
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return []gin.H{}, 0, 0
	}

	var infos []streamerInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		folder := filepath.Join(root, entry.Name())
		folderEntries, _ := os.ReadDir(folder)
		var totalSize, pendSize int64
		var totalCount, pendCount int
		for _, fe := range folderEntries {
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
			if !utils.IsMergedFile(name) {
				pendSize += info.Size()
				pendCount++
			}
		}
		infos = append(infos, streamerInfo{
			name: entry.Name(), size: totalSize, count: totalCount,
			pendingCount: pendCount, pendingSize: pendSize,
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
		streamers = append(streamers, gin.H{
			"name":       si.name,
			"files":      si.count,
			"size_bytes": si.size,
			"size_gb":    float64(int(float64(si.size)/1073741824*100)) / 100,
		})
	}

	return streamers, totalPendingCount, totalPendingSize
}

func (h *Handler) GetStreamerFiles(c *gin.Context) {
	streamer := c.Param("name")
	if !utils.ValidateFilename(streamer) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "非法主播名"})
		return
	}
	cfg := h.config.ToDTO()
	folder := filepath.Join(cfg.TargetDir, streamer)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "目录不存在"})
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
			"mtime":     float64(info.ModTime().Unix()),
			"is_merged": strings.Contains(name, "-合并版"),
		})
	}
	c.JSON(http.StatusOK, files)
}
