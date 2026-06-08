// Package service 提供核心业务逻辑。
// 包含合并服务、清理服务、历史记录服务、调度服务和日志管理。
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"bililive-helper/internal/config"
	"bililive-helper/internal/utils"

	"go.uber.org/zap"
)

// CleanService 提供智能清理功能。
// 根据磁盘使用率阈值触发清理，支持白名单保护、安全期保护和每主播保底数量。
type CleanService struct {
	config  *config.Config
	logger  *zap.Logger
	history *HistoryService
}

// NewCleanService 创建清理服务实例。
func NewCleanService(config *config.Config, logger *zap.Logger, history *HistoryService) *CleanService {
	return &CleanService{config: config, logger: logger, history: history}
}

// CleanResult 保存清理操作的结果。
type CleanResult struct {
	Deleted int
	Freed   int64
}

// Run 执行清理任务。
// 全局模式：检查磁盘阈值，未达阈值则跳过；已合并文件优先删除。
// 单主播模式：仅清理指定主播的文件。
// 参数 streamer 为空表示全局清理；onProgress 用于 SSE 进度回调。
func (s *CleanService) Run(ctx context.Context, streamer string, onProgress ProgressFunc) (*CleanResult, string, error) {
	start := time.Now()
	cfg := s.config.Snapshot()
	root := cfg.TargetDir

	opLog, err := NewOpLogger(filepath.Join(cfg.LogDir, "clean_log"), "clean")
	if err != nil {
		opLog = nil // 降级为 nil，不阻断操作
	}
	defer opLog.Close()

	if onProgress == nil {
		onProgress = func(string) {}
	}
	onProgress = opLog.ProgressFunc(onProgress)

	if cfg.IsBackupWindow() {
		return nil, opLog.LogID(), fmt.Errorf("当前处于静默时段（%d:%02d-%d:%02d），清理暂停", cfg.BackupStartHour, cfg.BackupStartMinute, cfg.BackupEndHour, cfg.BackupEndMinute)
	}

	if _, err := os.Stat(root); os.IsNotExist(err) {
		opLog.Log(fmt.Sprintf("❌ 路径不存在: %s", root))
		return nil, opLog.LogID(), fmt.Errorf("路径不存在: %s", root)
	}

	tag := "[全局]"
	if streamer != "" {
		tag = fmt.Sprintf("[%s]", streamer)
	}
	onProgress(fmt.Sprintf("▶ 开始 %s 清理", tag))

	disk, err := utils.GetDiskUsage(root)
	if err != nil {
		return nil, opLog.LogID(), err
	}

	if streamer == "" {
		if disk.UsedPct < cfg.TriggerThreshold {
			msg := fmt.Sprintf("📊 %s 磁盘 %.1f%%（未达阈值 %.0f%%）", tag, disk.UsedPct, cfg.TriggerThreshold)
			onProgress(msg)
			s.history.Add("clean", streamer, "success",
				fmt.Sprintf("磁盘 %.1f%% 未达阈值 %.0f%%", disk.UsedPct, cfg.TriggerThreshold), opLog.LogID())
			return &CleanResult{}, opLog.LogID(), nil
		}
		onProgress(fmt.Sprintf("📊 磁盘 %.1f%%（阈值 %.0f%%）", disk.UsedPct, cfg.TriggerThreshold))
		onProgress(fmt.Sprintf("⚙ %s 清理目标: %.1f%% → %.0f%%", tag, disk.UsedPct, cfg.TargetThreshold))
	} else {
		if disk.UsedPct > 95 {
			msg := fmt.Sprintf("❌ %s 磁盘 %.1f%% 超过 95%% 安全上限，请手动检查磁盘空间", tag, disk.UsedPct)
			onProgress(msg)
			return nil, opLog.LogID(), fmt.Errorf("磁盘使用率 %.1f%% 超过 95%% 安全上限，请手动清理磁盘空间", disk.UsedPct)
		}
		opLog.Log(fmt.Sprintf("▶ %s 清理", tag))
	}

	needToFree := s.calculateNeedToFree(disk, cfg)
	if needToFree > 0 {
		onProgress(fmt.Sprintf("📊 需释放 %s 才能达到 %.0f%%", utils.FormatSize(needToFree), cfg.TargetThreshold))
	}

	candidates, perStreamer := s.collectCandidates(root, streamer, cfg)

	// 按主播记录清理摘要日志
	for name, info := range perStreamer {
		opLog.Log(fmt.Sprintf("── %s ──", name))
		if info.total <= cfg.MinKeepPerStreamer {
			opLog.Log(fmt.Sprintf("ℹ %s → %d 个文件，全部保留（≤%d）", name, info.total, cfg.MinKeepPerStreamer))
		} else if info.skipped > 0 {
			opLog.Log(fmt.Sprintf("ℹ %s → %d 个文件，%d 个可清理，%d 个跳过（白名单/安全期）", name, info.total, info.candidate, info.skipped))
		} else {
			opLog.Log(fmt.Sprintf("ℹ %s → %d 个文件，%d 个可清理", name, info.total, info.candidate))
		}
	}

	if len(candidates) == 0 {
		msg := "ℹ 扫描完成，无符合条件的文件可删"
		onProgress(msg)
		s.history.Add("clean", streamer, "success", "无符合条件的文件可删", opLog.LogID())
		return &CleanResult{}, opLog.LogID(), nil
	}

	onProgress(fmt.Sprintf("ℹ 发现 %d 个候选文件", len(candidates)))

	sort.Slice(candidates, func(i, j int) bool {
		// 排序策略：已合并文件优先删除（安全），其次按文件年龄从老到新
		iMerged := utils.IsMergedFile(candidates[i].Name)
		jMerged := utils.IsMergedFile(candidates[j].Name)
		if iMerged != jMerged {
			return iMerged // 已合并文件优先（true > false）
		}
		return candidates[i].Mtime < candidates[j].Mtime
	})

	deleted, freed, truncated := s.deleteFiles(ctx, candidates, needToFree, cfg, onProgress)

	onProgress("───────────────────────────")

	duration := time.Since(start).Seconds()

	status := "success"
	statusMsg := fmt.Sprintf("删除 %d 文件，释放 %s", deleted, utils.FormatSize(freed))
	if truncated {
		if ctx.Err() != nil {
			status = "partial"
			statusMsg += "（任务被中断）"
		} else {
			statusMsg += fmt.Sprintf("（已达单次删除上限 %d）", cfg.MaxDeletePerRun)
		}
	}

	msg := fmt.Sprintf("✅ 完成: %s", statusMsg)
	onProgress(msg)

	// 删除后显示磁盘使用率
	if diskAfter, err := utils.GetDiskUsage(root); err == nil {
		onProgress(fmt.Sprintf("📊 当前磁盘 %.1f%%", diskAfter.UsedPct))
	}

	s.history.AddWithStats("clean", streamer, status, deleted, freed, 0, duration, statusMsg, opLog.LogID())

	return &CleanResult{Deleted: deleted, Freed: freed}, opLog.LogID(), nil
}

type candidateFile struct {
	Path  string
	Name  string
	Size  int64
	Mtime int64 // Unix 时间戳（秒）
}

// calculateNeedToFree 计算需要释放的字节数以达到目标阈值。
func (s *CleanService) calculateNeedToFree(disk *utils.DiskUsage, cfg config.Config) int64 {
	if disk.UsedPct <= cfg.TargetThreshold {
		return 0
	}
	targetSize := float64(disk.Total) * cfg.TargetThreshold / 100
	return int64(float64(disk.Used) - targetSize)
}

type streamerStats struct {
	total     int
	candidate int
	skipped   int
}

// collectCandidates 扫描所有主播目录，收集可清理的候选文件。
// 返回候选文件列表和每个主播的统计信息（总数、候选数、跳过数）。
func (s *CleanService) collectCandidates(root, streamer string, cfg config.Config) ([]candidateFile, map[string]streamerStats) {
	var candidates []candidateFile
	perStreamer := make(map[string]streamerStats)
	entries, err := os.ReadDir(root)
	if err != nil {
		s.logger.Warn("读取根目录失败，跳过清理", zap.Error(err))
		return nil, nil
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if streamer != "" && entry.Name() != streamer {
			continue
		}
		folder := filepath.Join(root, entry.Name())
		before := len(candidates)
		s.collectStreamerCandidates(folder, entry.Name(), &candidates, cfg)
		after := len(candidates)
		total := s.countVideos(folder)
		perStreamer[entry.Name()] = streamerStats{
			total:     total,
			candidate: after - before,
			skipped:   max(0, total-(after-before)-min(total, cfg.MinKeepPerStreamer)),
		}
	}
	return candidates, perStreamer
}

func (s *CleanService) countVideos(folder string) int {
	count := 0
	entries, err := os.ReadDir(folder)
	if err != nil {
		return 0
	}
	for _, entry := range entries {
		if !entry.IsDir() && utils.IsVideoFile(entry.Name()) {
			count++
		}
	}
	return count
}

// collectStreamerCandidates 收集单个主播目录下的可清理候选文件。
// 应用保底数量、白名单过滤和安全期保护规则。
func (s *CleanService) collectStreamerCandidates(folder, streamerName string, candidates *[]candidateFile, cfg config.Config) {
	minKeep := cfg.MinKeepPerStreamer
	wl := cfg.WhitelistKeywords

	var videos []candidateFile
	entries, err := os.ReadDir(folder)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !utils.IsVideoFile(name) {
			continue
		}
		// 包含已合并文件（它们也会被清理）
		info, err := entry.Info()
		if err != nil {
			continue
		}
		videos = append(videos, candidateFile{
			Path:  filepath.Join(folder, name),
			Name:  name,
			Size:  info.Size(),
			Mtime: info.ModTime().Unix(),
		})
	}

	sort.Slice(videos, func(i, j int) bool {
		return videos[i].Mtime < videos[j].Mtime
	})

	if len(videos) <= minKeep {
		return
	}
	videos = videos[:len(videos)-minKeep]

	for _, v := range videos {
		if utils.ContainsAny(v.Name, wl) || utils.ContainsAny(streamerName, wl) {
			continue
		}
		if cfg.SafeMode == "days" {
			cutoff := time.Now().AddDate(0, 0, -cfg.SafeDays)
			if time.Unix(v.Mtime, 0).After(cutoff) {
				continue
			}
		} else {
			cutoff := time.Now().Add(-time.Duration(cfg.SafeAgeMinutes) * time.Minute)
			if time.Unix(v.Mtime, 0).After(cutoff) {
				continue
			}
		}
		*candidates = append(*candidates, v)
	}
}

// deleteFiles 执行文件删除，使用双快照检测跳过正在写入的文件。
// 受单次删除上限和目标释放量双重约束。
// 返回 (deleted, freed, truncated)，truncated 表示因限制条件提前终止。
func (s *CleanService) deleteFiles(ctx context.Context, candidates []candidateFile, needToFree int64, cfg config.Config, onProgress ProgressFunc) (int, int64, bool) {
	deleted := 0
	freed := int64(0)
	truncated := false

	// 1. 记录所有候选文件的大小（第一次快照）
	sizeMap1 := make(map[string]int64)
	for _, f := range candidates {
		if info, err := os.Stat(f.Path); err == nil {
			sizeMap1[f.Path] = info.Size()
		}
	}

	// 2. 等待 1 秒后再次记录大小 — 正在写入的文件大小会不同
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
		return 0, 0, true
	}

	// 3. 记录第二次快照
	sizeMap2 := make(map[string]int64)
	for _, f := range candidates {
		if info, err := os.Stat(f.Path); err == nil {
			sizeMap2[f.Path] = info.Size()
		}
	}

	// 4. 执行删除 — 跳过两次快照间大小变化的文件（正在写入）
	for _, f := range candidates {
		if ctx.Err() != nil {
			s.logger.Info("⚠ 上下文取消，终止清理")
			truncated = true
			break
		}
		if deleted >= cfg.MaxDeletePerRun {
			s.logger.Info(fmt.Sprintf("ℹ 已达单次删除上限 %d 个文件", cfg.MaxDeletePerRun))
			truncated = true
			break
		}
		if needToFree > 0 && freed >= needToFree {
			s.logger.Info(fmt.Sprintf("ℹ 已释放 %s，达到目标", utils.FormatSize(freed)))
			break
		}
		if _, err := os.Stat(f.Path); os.IsNotExist(err) {
			continue
		}

		// 对比两次快照，检测正在写入的文件
		s1, ok1 := sizeMap1[f.Path]
		s2, ok2 := sizeMap2[f.Path]
		if ok1 && ok2 && s1 != s2 {
			s.logger.Info(fmt.Sprintf("⏭ 跳过（被占用正在写入）: %s", f.Name))
			continue
		}

		// 执行删除
		if err := utils.SafeUnlink(f.Path); err != nil {
			s.logger.Info(fmt.Sprintf("⚠ 删除失败: %s (%v)", f.Name, err))
			continue
		}

		deleted++
		freed += f.Size
		sn := f.Name
		runes := []rune(sn)
		if len(runes) > 35 {
			sn = string(runes[:10]) + "…" + string(runes[len(runes)-22:])
		}
		s.logger.Info(fmt.Sprintf("🗑 [%s] %s (%s)", filepath.Base(filepath.Dir(f.Path)), sn, utils.FormatSize(f.Size)))
		onProgress(fmt.Sprintf("🗑 [%s] %s (%s)", filepath.Base(filepath.Dir(f.Path)), sn, utils.FormatSize(f.Size)))
	}

	return deleted, freed, truncated
}
