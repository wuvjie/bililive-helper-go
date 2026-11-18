// Package service 提供核心业务逻辑。
// 包含合并服务、清理服务、历史记录服务、调度服务和日志管理。
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"bililive-helper-go/internal/config"
	"bililive-helper-go/internal/fsutil"
	"bililive-helper-go/internal/utils"

	"go.uber.org/zap"
)

// StreamerLockChecker 检查指定主播是否有正在进行的合并任务。
// 用于清理服务避免误删正在合并中的文件。
type StreamerLockChecker interface {
	IsStreamerLocked(name string) bool
}

// CleanService 提供智能清理功能。
// 根据磁盘使用率阈值触发清理，支持白名单保护、安全期保护和每主播保底数量。
type CleanService struct {
	config  *config.Config
	logger  *zap.Logger
	history *HistoryService
	locker  StreamerLockChecker // 可选，为 nil 时跳过锁检查
	runMu   sync.Mutex          // 防止定时清理与手动清理并发冲突
}

// NewCleanService 创建清理服务实例。
func NewCleanService(config *config.Config, logger *zap.Logger, history *HistoryService, locker StreamerLockChecker) *CleanService {
	return &CleanService{config: config, logger: logger, history: history, locker: locker}
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
	if !s.runMu.TryLock() {
		return nil, "", fmt.Errorf("磁盘清理任务正在执行中，请稍后再试")
	}
	defer s.runMu.Unlock()

	start := time.Now()

	setup, err := PrepareTask(s.config, s.logger, "clean_log", "clean", "清理", streamer, onProgress)
	if err != nil {
		return nil, "", err
	}
	defer setup.OpLog.Close()
	cfg := setup.Cfg
	progress := setup.Progress
	root := cfg.TargetDir

	disk, err := utils.GetDiskUsage(root)
	if err != nil {
		return nil, setup.LogID, err
	}
	progress(fmt.Sprintf("▶ 开始 %s 录像清理 | 📊 磁盘空间使用率 %.1f%% (剩余 %.1f GB)", setup.Tag, disk.UsedPct, float64(disk.Free)/oneGB))

	if streamer == "" {
		if disk.UsedPct < cfg.TriggerThreshold {
			msg := fmt.Sprintf("📊 磁盘空间使用率 %.1f%%（低于清理触发阈值 %.0f%%，无需清理）", disk.UsedPct, cfg.TriggerThreshold)
			progress(msg)
			if !setup.IsManual {
				setup.OpLog.Discard()
				return &CleanResult{}, "", nil
			}
			s.history.Add("clean", streamer, "success",
				fmt.Sprintf("磁盘 %.1f%% 低于阈值 %.0f%%", disk.UsedPct, cfg.TriggerThreshold), setup.LogID)
			return &CleanResult{}, setup.LogID, nil
		}
		progress(fmt.Sprintf("📊 当前磁盘空间使用率 %.1f%%（已达清理触发阈值 %.0f%%）", disk.UsedPct, cfg.TriggerThreshold))
		progress(fmt.Sprintf("⚙ 清理目标：将磁盘使用率从 %.1f%% 降至停止阈值 %.0f%%", disk.UsedPct, cfg.TargetThreshold))
	} else {
		if disk.UsedPct > 95 {
			msg := fmt.Sprintf("❌ 磁盘空间使用率已达 %.1f%%，超过 95%% 安全上限，请手动检查磁盘空间", disk.UsedPct)
			progress(msg)
			return nil, setup.LogID, fmt.Errorf("磁盘使用率 %.1f%% 超过 95%% 安全上限，请手动清理磁盘空间", disk.UsedPct)
		}
		setup.OpLog.Log(fmt.Sprintf("▶ %s 清理", setup.Tag))
	}

	var needToFree int64
	if streamer == "" {
		needToFree = s.calculateNeedToFree(disk, cfg)
		if needToFree > 0 {
			progress(fmt.Sprintf("📊 预计需释放 %s 磁盘空间以达到停止阈值 %.0f%%", utils.FormatSize(needToFree), cfg.TargetThreshold))
		}
	}

	candidates, perStreamer := s.collectCandidates(root, streamer, cfg)

	// 按主播记录清理摘要日志
	for name, info := range perStreamer {
		setup.OpLog.Log(fmt.Sprintf("── %s ──", name))
		if info.protected {
			setup.OpLog.Log(fmt.Sprintf("ℹ [%s] 共 %d 个录像文件，全部保留（已开启主播免删保护）", name, info.total))
		} else if info.total <= info.minKeep {
			setup.OpLog.Log(fmt.Sprintf("ℹ [%s] 共 %d 个录像文件，全部保留（处于最少保留场次限制内，≤%d 场）", name, info.total, info.minKeep))
		} else if info.skipped > 0 {
			setup.OpLog.Log(fmt.Sprintf("ℹ [%s] 共 %d 个录像文件，%d 个待清理，%d 个跳过（受关键词白名单或免删保护期保护）", name, info.total, info.candidate, info.skipped))
		} else {
			setup.OpLog.Log(fmt.Sprintf("ℹ [%s] 共 %d 个录像文件，%d 个待清理", name, info.total, info.candidate))
		}
	}

	if len(candidates) == 0 {
		msg := "ℹ 扫描完成，暂无符合清理条件的录像（所有录像均受最少保留场次或免删保护期保护）"
		progress(msg)
		if !setup.IsManual {
			setup.OpLog.Discard()
			return &CleanResult{}, "", nil
		}
		s.history.Add("clean", streamer, "success", "无需清理（录像均受保护）", setup.LogID)
		return &CleanResult{}, setup.LogID, nil
	}

	progress(fmt.Sprintf("ℹ 发现 %d 个待清理的候选录像文件", len(candidates)))

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Mtime != candidates[j].Mtime {
			return candidates[i].Mtime < candidates[j].Mtime
		}
		return candidates[i].Name < candidates[j].Name
	})

	deleted, freed, truncated := s.deleteFiles(ctx, candidates, needToFree, cfg, progress)

	duration := time.Since(start)

	status := "success"
	statusMsg := fmt.Sprintf("清理 %d 个录像文件，释放 %s", deleted, utils.FormatSize(freed))
	if truncated {
		if ctx.Err() != nil {
			status = "partial"
			statusMsg += "（任务被中断）"
		} else {
			statusMsg += fmt.Sprintf("（已达单次删除上限 %d）", cfg.MaxDeletePerRun)
		}
	}

	progress("───────────────────────────")
	progress("📊 结果摘要")
	progress(fmt.Sprintf("  清理：%d 个录像文件", deleted))
	progress(fmt.Sprintf("  释放空间：%s", utils.FormatSize(freed)))
	progress(fmt.Sprintf("  执行耗时：%s", formatDuration(duration)))

	if diskAfter, err := utils.GetDiskUsage(root); err == nil {
		freeChangeGB := float64(int64(diskAfter.Free)-int64(disk.Free)) / oneGB
		if freeChangeGB > 0 {
			progress(fmt.Sprintf("  磁盘空间：%.1f%% → %.1f%%（释放 %.1f GB）", disk.UsedPct, diskAfter.UsedPct, freeChangeGB))
		} else {
			progress(fmt.Sprintf("  磁盘空间：%.1f%% → %.1f%%（占用 %.1f GB）", disk.UsedPct, diskAfter.UsedPct, -freeChangeGB))
		}
	}
	progress("───────────────────────────")

	s.history.AddWithStats("clean", streamer, status, deleted, freed, 0, duration.Seconds(), statusMsg, setup.LogID)

	return &CleanResult{Deleted: deleted, Freed: freed}, setup.LogID, nil
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
	minKeep   int
	protected bool
}

func resolveStreamerCleanPolicy(cfg config.Config, streamerName string) (minKeep int, protected bool) {
	minKeep = cfg.MinKeepPerStreamer
	if rule, ok := cfg.StreamerRules[streamerName]; ok {
		if rule.Protected {
			return minKeep, true
		}
		if rule.MinKeep > 0 {
			minKeep = rule.MinKeep
		}
	}
	return minKeep, false
}

// collectCandidates 扫描所有主播目录，收集可清理的候选文件。
// 返回候选文件列表和每个主播的统计信息（总数、候选数、跳过数及个性化策略）。
func (s *CleanService) collectCandidates(root, streamer string, cfg config.Config) ([]candidateFile, map[string]streamerStats) {
	var candidates []candidateFile
	perStreamer := make(map[string]streamerStats)
	dirs, err := fsutil.ScanStreamerDirs(root)
	if err != nil {
		s.logger.Warn("读取根目录失败，跳过清理", zap.Error(err))
		return nil, nil
	}
	for _, dir := range dirs {
		if streamer != "" && dir.Name != streamer {
			continue
		}
		// 跳过正在合并的主播，避免误删合并中的文件
		if s.locker != nil && s.locker.IsStreamerLocked(dir.Name) {
			s.logger.Info("跳过正在合并的主播", zap.String("streamer", dir.Name))
			continue
		}
		before := len(candidates)
		streamerCandidates := s.collectStreamerCandidates(dir.Path, dir.Name, cfg)
		candidates = append(candidates, streamerCandidates...)
		after := len(candidates)
		total := s.countVideos(dir.Path)
		minKeep, protected := resolveStreamerCleanPolicy(cfg, dir.Name)
		skipped := max(0, total-(after-before)-min(total, minKeep))
		if protected {
			skipped = total
		}
		perStreamer[dir.Name] = streamerStats{
			total:     total,
			candidate: after - before,
			skipped:   skipped,
			minKeep:   minKeep,
			protected: protected,
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
		if !entry.IsDir() && utils.IsVideoFile(entry.Name()) && !utils.IsTempVideoArtifact(entry.Name()) {
			count++
		}
	}
	return count
}

// collectStreamerCandidates 收集单个主播目录下的可清理候选文件。
// 优先剔除处于待合并保护期内的未合并原片（不参与清理也不占用成片的 minKeep 保底名额），
// 再按场次维度保留最新的 minKeep 个场次，最后应用白名单与安全期过滤。
func (s *CleanService) collectStreamerCandidates(folder, streamerName string, cfg config.Config) []candidateFile {
	minKeep, protected := resolveStreamerCleanPolicy(cfg, streamerName)
	if protected {
		return nil // 免清理保护生效
	}
	wl := cfg.WhitelistKeywords

	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil
	}

	now := time.Now()
	mergeAgeMin := cfg.MergeAgeMinutes
	if mergeAgeMin <= 0 {
		mergeAgeMin = defaultMergeAgeMinutes
	}
	mergeProtectWindow := time.Duration(mergeAgeMin) * time.Minute

	type cleanUnit struct {
		isCompleted bool
		streamerKey string
		files       []candidateFile
		maxMtime    int64
	}

	var completedUnits []cleanUnit
	var staleRawFiles []candidateFile

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !utils.IsVideoFile(name) || utils.IsTempVideoArtifact(name) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}

		ext := strings.ToLower(filepath.Ext(name))
		isCompleted := utils.IsMergedFile(name) || ext == ".mp4"

		// 未合并原片（.flv / .ts）若在合并静置保护窗口内，视为正在等待合并/转封装，
		// 直接跳过：既不得删除，也不挤占已完成录像的 minKeep 保底名额。
		if !isCompleted && now.Sub(info.ModTime()) < mergeProtectWindow {
			continue
		}

		fileTime := info.ModTime()
		sKey := ""
		if parsedKey, dt, ok := utils.ParseFilename(name); ok {
			fileTime = dt
			sKey = utils.StreamerKeyFromKey(parsedKey)
		}

		cf := candidateFile{
			Path:  filepath.Join(folder, name),
			Name:  name,
			Size:  info.Size(),
			Mtime: fileTime.Unix(),
		}

		if isCompleted {
			completedUnits = append(completedUnits, cleanUnit{
				isCompleted: true,
				streamerKey: sKey,
				files:       []candidateFile{cf},
				maxMtime:    cf.Mtime,
			})
		} else {
			staleRawFiles = append(staleRawFiles, cf)
		}
	}

	// 将已超出合并等待期的历史未合并分片（.flv / .ts）按场次时间间隔聚类（同主播 2 小时内视作同一场次），
	// 避免单场多个碎片各自占用 1 个 minKeep 场次名额。
	sort.Slice(staleRawFiles, func(i, j int) bool {
		return staleRawFiles[i].Mtime < staleRawFiles[j].Mtime
	})
	var rawUnits []cleanUnit
	const rawSessionGapSec = int64(2 * 3600)
	for _, rf := range staleRawFiles {
		sKey := ""
		if parsedKey, _, ok := utils.ParseFilename(rf.Name); ok {
			sKey = utils.StreamerKeyFromKey(parsedKey)
		}
		if len(rawUnits) > 0 && rawUnits[len(rawUnits)-1].streamerKey == sKey &&
			rf.Mtime-rawUnits[len(rawUnits)-1].maxMtime <= rawSessionGapSec {
			last := &rawUnits[len(rawUnits)-1]
			last.files = append(last.files, rf)
			if rf.Mtime > last.maxMtime {
				last.maxMtime = rf.Mtime
			}
		} else {
			rawUnits = append(rawUnits, cleanUnit{
				isCompleted: false,
				streamerKey: sKey,
				files:       []candidateFile{rf},
				maxMtime:    rf.Mtime,
			})
		}
	}

	// 优先以已完成成片（.mp4 / -合并版）满足 minKeep 保底名额；
	// 仅当成片数量不足 minKeep 时，才用较新的历史未合并场次补足保底名额，
	// 确保未合并分片永远不会将旧的完整成片挤出保底区。
	sort.Slice(completedUnits, func(i, j int) bool {
		return completedUnits[i].maxMtime < completedUnits[j].maxMtime
	})
	sort.Slice(rawUnits, func(i, j int) bool {
		return rawUnits[i].maxMtime < rawUnits[j].maxMtime
	})

	var reclaimableUnits []cleanUnit
	if len(completedUnits) > minKeep {
		reclaimableUnits = append(reclaimableUnits, completedUnits[:len(completedUnits)-minKeep]...)
	}
	remainingKeep := minKeep - len(completedUnits)
	if remainingKeep < 0 {
		remainingKeep = 0
	}
	if len(rawUnits) > remainingKeep {
		reclaimableUnits = append(reclaimableUnits, rawUnits[:len(rawUnits)-remainingKeep]...)
	}
	sort.Slice(reclaimableUnits, func(i, j int) bool {
		return reclaimableUnits[i].maxMtime < reclaimableUnits[j].maxMtime
	})

	var result []candidateFile
	for _, u := range reclaimableUnits {
		for _, v := range u.files {
			if utils.ContainsAny(v.Name, wl) || utils.ContainsAny(streamerName, wl) {
				continue
			}
			if cfg.SafeMode == "days" {
				cutoff := now.AddDate(0, 0, -cfg.SafeDays)
				if time.Unix(v.Mtime, 0).After(cutoff) {
					continue
				}
			} else {
				cutoff := now.Add(-time.Duration(cfg.SafeAgeMinutes) * time.Minute)
				if time.Unix(v.Mtime, 0).After(cutoff) {
					continue
				}
			}
			result = append(result, v)
		}
	}
	return result
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
			s.logger.Info("⚠ 任务被取消，终止清理")
			truncated = true
			break
		}
		if deleted >= cfg.MaxDeletePerRun {
			s.logger.Info(fmt.Sprintf("ℹ 已达单次清理上限 %d 个文件，停止本次清理", cfg.MaxDeletePerRun))
			truncated = true
			break
		}
		if needToFree > 0 && freed >= needToFree {
			s.logger.Info(fmt.Sprintf("ℹ 已释放 %s 磁盘空间，达到清理目标", utils.FormatSize(freed)))
			break
		}
		if _, err := os.Stat(f.Path); os.IsNotExist(err) {
			continue
		}

		// 对比两次快照，检测正在写入的文件
		s1, ok1 := sizeMap1[f.Path]
		s2, ok2 := sizeMap2[f.Path]
		if ok1 && ok2 && s1 != s2 {
			s.logger.Info(fmt.Sprintf("⏭ 跳过（录像文件正在录制写入中）: %s", f.Name))
			continue
		}

		// 执行删除
		if err := utils.SafeUnlink(f.Path); err != nil {
			s.logger.Info(fmt.Sprintf("⚠ 清理失败: %s (%v)", f.Name, err))
			continue
		}

		// 级联清理同名伴生文件（XML 弹幕、封面图、元数据等），防止产生孤儿文件
		CleanCompanionFiles(f.Path)

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
