// scanner.go 提供录制文件的扫描、分组和场次划分逻辑。
// 解析 bililive-go 文件名格式，按时间间隔将同一主播的文件划分为合并批次。
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
	"bililive-helper-go/internal/ffmpeg"
	"bililive-helper-go/internal/utils"

	"go.uber.org/zap"
)

// videoFile 保存单个视频文件的解析元数据。
type videoFile struct {
	Name     string
	Key      string
	Datetime time.Time
	EndTime  time.Time // 实际结束时间 = Datetime + ffprobe 时长；ffprobe 失败时回退为 Datetime
}

// mergeTask 表示一个待合并的任务（多个文件合并为一个输出）。
type mergeTask struct {
	Files  []string
	Folder string
	SizeGB float64
}

// pendingBatch 保存通过快速检查但尚未验证文件稳定性的候选批次。
type pendingBatch struct {
	Names     []string
	Folder    string
	Size      int64
	Streamer  string
	LastMtime time.Time
}

// convertTask 表示一个待转换的任务（FLV 转 MP4）。
type convertTask struct {
	FlvPath string
	Mp4Path string
	Folder  string
	Name    string
}

// ScanSummary 汇总一次扫描中的目录和主播状态信息。
type ScanSummary struct {
	TotalStreamers int // 发现的主播目录数
	ActiveStreams  int // 正在录制的主播数
	WaitingSafety  int // 处于下播安全期内的主播数
	UpToDate       int // 录像已处于最新状态（已合并或无需合并）的主播数
}

const (
	recordingActiveThreshold = 5 * time.Minute // 录制活跃判定阈值：最新文件在此时间内被认为还在录制
	defaultGapMinutes        = 30              // 场次间隔回退值：未配置时的默认分钟数
	defaultMergeAgeMinutes   = 30              // 合并安全期回退值：未配置时的默认分钟数
)

// isFileBeingWritten 通过比较两次文件大小来检测文件是否正在被写入。
// 文件缺失、不可访问或大小变化时返回 true。
func isFileBeingWritten(ctx context.Context, path string, interval time.Duration) bool {
	return !isFileSizeStable(ctx, path, interval)
}

// isFileSizeStable 检查文件大小在指定间隔内是否保持稳定。
// 最终安全检查 — 捕获挂起的写入进程。
func isFileSizeStable(ctx context.Context, path string, interval time.Duration) bool {
	info1, err := os.Stat(path)
	if err != nil {
		return false
	}
	select {
	case <-time.After(interval):
	case <-ctx.Done():
		return false
	}
	info2, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info1.Size() == info2.Size()
}

// streamerGroupKey 提取用于场次聚类的主播分组键。
// 对标准命名格式 "[主播名][标题]" 提取第一段 "[主播名]"，防止主播中途改标题导致同一场直播无法合并；
// 非标准命名文件保留原 "raw|..." 键各自独立处理。
func streamerGroupKey(key string) string {
	if strings.HasPrefix(key, "raw|") {
		return key
	}
	return utils.StreamerKeyFromKey(key)
}

// stabilityProbeInterval 根据文件最后修改时间决定可用性写锁探测窗口。
// 已静置超过录制活跃阈值（5 分钟）的历史文件仅做短时快速校验，避免无谓阻塞；近期修改的文件等待更长以捕获持续写入。
func stabilityProbeInterval(modTime time.Time) time.Duration {
	if !modTime.IsZero() && time.Since(modTime) >= recordingActiveThreshold {
		return 200 * time.Millisecond
	}
	return 1 * time.Second
}

// stabilityWaitDuration 返回 Phase 2 批量稳定性复核等待时长。
// 对静置满 5 分钟的文件使用 500ms 快速复核，避免手动合并时强制卡死 1 分钟。
func stabilityWaitDuration(modTime time.Time) time.Duration {
	if !modTime.IsZero() && time.Since(modTime) >= recordingActiveThreshold {
		return 500 * time.Millisecond
	}
	return 5 * time.Second
}

// shouldTrimVideo 预判单个 MP4 文件是否确实需要按主播规则裁剪，避免无需裁剪的文件每轮扫描都入队。
func shouldTrimVideo(ctx context.Context, videoPath string, rule config.StreamerRule) bool {
	if !rule.TrimEnabled {
		return false
	}
	start := strings.TrimSpace(rule.TrimStart)
	if start == "" {
		start = "00:00:00"
	}
	end := strings.TrimSpace(rule.TrimEnd)
	startSec, startErr := ffmpeg.ParseTimeSeconds(start)
	endSec, endErr := ffmpeg.ParseTimeSeconds(end)

	hasStartTrim := startErr == nil && startSec > 0
	hasEndTrim := end != "" && endErr == nil && endSec > 0
	if !hasStartTrim && !hasEndTrim {
		return false
	}
	if hasStartTrim && hasEndTrim && endSec <= startSec {
		return false
	}

	if dur, err := ffmpeg.ProbeDuration(ctx, videoPath); err == nil && dur > 0 {
		if hasStartTrim && dur <= startSec {
			return false
		}
		if !hasStartTrim && hasEndTrim && dur <= endSec {
			return false
		}
	}
	return true
}

// getVideoFiles 返回目录中所有非已合并的视频文件及其解析元数据。
// 具有自愈能力：当同名 .flv/.ts 与 .mp4 共存时，通过 QuickProbe 验证 .mp4 是否真正完整；
// 若 .mp4 有效则清理冗余原片，若 .mp4 已损坏则清理半成品并保留完好原片重新转换；
// 同时兼容不符合标准命名规范的视频文件，使其与状态统计口径保持一致。
func (s *MergeService) getVideoFiles(ctx context.Context, folder string) []videoFile {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil
	}

	// 第一遍：收集同名 MP4 与源原片（FLV/TS）的共存映射
	mp4Files := make(map[string]string) // base -> mp4 filename
	rawBases := make(map[string]bool)   // base -> has .flv or .ts
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !utils.IsVideoFile(name) || utils.IsMergedFile(name) || utils.IsTempVideoArtifact(name) {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		base := strings.TrimSuffix(name, ext)
		if ext == ".mp4" {
			mp4Files[base] = name
		} else {
			rawBases[base] = true
		}
	}

	// 仅对与 .flv/.ts 共存的 .mp4 做 QuickProbe，区分“上次转换成功遗留原片”和“半成品损坏 MP4”
	validMP4Bases := make(map[string]bool)
	brokenMP4s := make(map[string]bool)
	for base, mp4Name := range mp4Files {
		if !rawBases[base] {
			continue
		}
		mp4Path := filepath.Join(folder, mp4Name)
		if ffmpeg.QuickProbe(ctx, mp4Path) == nil {
			validMP4Bases[base] = true
		} else {
			brokenMP4s[mp4Name] = true
			if err := utils.SafeUnlink(mp4Path); err != nil {
				s.logger.Warn("清理损坏 MP4 失败", zap.String("file", mp4Name), zap.Error(err))
			} else {
				s.logger.Info(fmt.Sprintf("🗑 清理损坏 MP4，保留源原片重新转换: %s", mp4Name))
			}
		}
	}

	// 第二遍：收集有效视频文件，清理已成功转封装后残留的 FLV/TS
	var videos []videoFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !utils.IsVideoFile(name) || utils.IsTempVideoArtifact(name) || utils.IsMergedFile(name) || brokenMP4s[name] {
			continue
		}

		ext := strings.ToLower(filepath.Ext(name))
		base := strings.TrimSuffix(name, ext)

		// 残留清理：仅当同名 MP4 经 QuickProbe 验证确认为完整成片时，才删除残留 FLV/TS
		if ext != ".mp4" && validMP4Bases[base] {
			path := filepath.Join(folder, name)
			if err := utils.SafeUnlink(path); err != nil {
				s.logger.Warn("清理残留原片失败", zap.String("file", name), zap.Error(err))
			} else {
				s.logger.Info(fmt.Sprintf("🗑 清理残留原片: %s", name))
			}
			continue
		}

		key, dt, ok := utils.ParseFilename(name)
		if !ok {
			// 兼容非标准命名的录制文件（如自定义命名模板或手动放入的原片），以文件 mtime 作为时间参与转换/统计
			info, infoErr := entry.Info()
			if infoErr != nil {
				continue
			}
			videos = append(videos, videoFile{Name: name, Key: "raw|" + name, Datetime: info.ModTime()})
			continue
		}
		videos = append(videos, videoFile{Name: name, Key: key, Datetime: dt})
	}
	return videos
}

// isStreamActive 检查指定分组 key 的最新文件是否正在被录制。
// 判断依据：最新文件的修改时间在最近 5 分钟内 → 认为还在录制。
// 比 isFileBeingWritten 更可靠——不受短暂缓冲/网络波动影响。
func isStreamActive(folder string, batchKey string) bool {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return false
	}

	matchToken := batchKey
	if strings.HasPrefix(batchKey, "raw|") {
		matchToken = strings.TrimPrefix(batchKey, "raw|")
	}

	var newestMtime time.Time
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !utils.IsVideoFile(name) || utils.IsMergedFile(name) || utils.IsTempVideoArtifact(name) {
			continue
		}
		if !strings.Contains(name, matchToken) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(newestMtime) {
			newestMtime = info.ModTime()
		}
	}

	if newestMtime.IsZero() {
		return false
	}
	// 最新文件在最近 5 分钟内被修改 → 认为还在录制
	return time.Since(newestMtime) < recordingActiveThreshold
}

// buildSessionBatches 对同主播组内的分片按开播时间与文件名排序、并发探测相邻分片实际时长，
// 顺延累加同时间戳分片的 EndTime，并按 gapMinutes 间隔切分为待处理的场次批次。
func buildSessionBatches(ctx context.Context, folder string, items []videoFile, gapMinutes int) [][]videoFile {
	if len(items) == 0 {
		return nil
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Datetime.Equal(items[j].Datetime) {
			return items[i].Name < items[j].Name
		}
		return items[i].Datetime.Before(items[j].Datetime)
	})

	// 仅当同一组有多个文件且相邻文件开播时间相距 ≤ 12 小时时才用 ffprobe 探测实际时长，
	// 避免孤立的单个 MP4 或跨天独立场次触发无意义的 ffprobe 子进程。
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	for i := range items {
		needProbe := len(items) > 1 &&
			((i+1 < len(items) && items[i+1].Datetime.Sub(items[i].Datetime) <= 12*time.Hour) ||
				(i > 0 && items[i].Datetime.Sub(items[i-1].Datetime) <= 12*time.Hour))
		if !needProbe {
			items[i].EndTime = items[i].Datetime
			continue
		}
		wg.Add(1)
		sem <- struct{}{} // 获取信号量槽位
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量槽位
			path := filepath.Join(folder, items[idx].Name)
			dur, err := ffmpeg.ProbeDuration(ctx, path)
			if err == nil && dur > 0 {
				items[idx].EndTime = items[idx].Datetime.Add(time.Duration(dur * float64(time.Second)))
			} else {
				items[idx].EndTime = items[idx].Datetime
			}
		}(i)
	}
	wg.Wait()

	// 同时间戳的分片文件（bililive-go 场次开始时间相同），EndTime 顺序累加而非各自独立计算
	// 例：001.flv(30min) 002.flv(30min) 都标 21:33 → 001: 22:03, 002: 22:33
	for i := 1; i < len(items); i++ {
		if !items[i].Datetime.After(items[i-1].Datetime) {
			dur := items[i].EndTime.Sub(items[i].Datetime)
			if dur < 0 {
				dur = 0
			}
			items[i].EndTime = items[i-1].EndTime.Add(dur)
		}
	}

	// gap <= gapMinutes → 同场次合并；gap > gapMinutes → 切分不同场次
	batches := [][]videoFile{{items[0]}}
	for i := 0; i < len(items)-1; i++ {
		if items[i+1].Datetime.Sub(items[i].EndTime).Minutes() > float64(gapMinutes) {
			batches = append(batches, []videoFile{items[i+1]})
		} else {
			batches[len(batches)-1] = append(batches[len(batches)-1], items[i+1])
		}
	}
	return batches
}

// scanTasks 扫描录制目录，将文件按主播分组、按时间排序、按间隔分场次。
// 返回待合并任务列表和待转换任务列表（FLV→MP4）。
func (s *MergeService) scanTasks(ctx context.Context, root, streamer string, cfg config.Config) ([]mergeTask, []convertTask, ScanSummary) {
	var stats ScanSummary
	var tasks []mergeTask
	var convertTasks []convertTask
	var pending []pendingBatch // Phase 1: 收集候选，Phase 2: 并发验证稳定性
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil, stats
	}

	gapMinutes := cfg.GapMinutes
	if gapMinutes <= 0 {
		gapMinutes = defaultGapMinutes
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if streamer != "" && entry.Name() != streamer {
			continue
		}
		stats.TotalStreamers++

		folder := filepath.Join(root, entry.Name())
		videos := s.getVideoFiles(ctx, folder)
		if len(videos) == 0 {
			stats.UpToDate++
			continue
		}

		streamerActive := false
		streamerWaitingSafety := false
		initConvertCount := len(convertTasks)
		initPendingCount := len(pending)

		// 按主播标识（剥离标题）聚合文件，避免主播中途改房间标题导致同场次分片无法合并；
		// 非标准命名文件保留 "raw|<name>" 键各自独立检查。
		groups := make(map[string][]videoFile)
		for _, v := range videos {
			gk := streamerGroupKey(v.Key)
			groups[gk] = append(groups[gk], v)
		}

		for key, items := range groups {
			batches := buildSessionBatches(ctx, folder, items, gapMinutes)

			mergeAgeMin := float64(cfg.MergeAgeMinutes)
			if mergeAgeMin <= 0 {
				mergeAgeMin = defaultMergeAgeMinutes
			}
			streamerRule, hasRule := cfg.StreamerRules[entry.Name()]

			enqueueSingleMP4Trim := func(fileName string) {
				if !hasRule || !streamerRule.TrimEnabled {
					return
				}
				mp4Path := filepath.Join(folder, fileName)
				if ct, active, waiting, ok := s.checkSingleMP4Trim(ctx, folder, entry.Name(), key, mp4Path, streamerRule, mergeAgeMin); ok {
					convertTasks = append(convertTasks, ct)
					s.logger.Info(fmt.Sprintf("[%s] ✂ %s → 单分片 MP4 符合裁剪规则，加入裁剪队列", entry.Name(), fileName))
				} else if active {
					streamerActive = true
				} else if waiting {
					streamerWaitingSafety = true
				}
			}

			for _, batch := range batches {
				// 检查对应默认输出路径是否存在损坏的半成品文件；
				// 若已存在有效的 "-合并版.mp4"，不再盲删本批次分片（同开播时间戳前缀可能对应不同批次，
				// 交由后续 doMerge 自动去重命名 "-合并版-2.mp4"，防止误删下半场录像）。
				outputName := utils.MakeOutputName(batch[0].Name)
				outputPath := filepath.Join(folder, outputName)
				actualOutputPath := outputPath
				extOut := strings.ToLower(filepath.Ext(outputName))
				if extOut == ".flv" || extOut == ".ts" {
					mp4Candidate := utils.MakeMP4Name(outputPath)
					if info, err := os.Stat(mp4Candidate); err == nil && info.Size() >= minValidFileSize {
						actualOutputPath = mp4Candidate
					}
				}
				if info, err := os.Stat(actualOutputPath); err == nil && info.Size() >= minValidFileSize {
					if ffmpeg.QuickProbe(ctx, actualOutputPath) != nil {
						if err := utils.SafeUnlink(actualOutputPath); err != nil {
							s.logger.Warn("清理损坏输出失败", zap.String("file", filepath.Base(actualOutputPath)), zap.Error(err))
						}
						s.logger.Info(fmt.Sprintf("[%s] ⚠ %s → 输出文件损坏，将重新合并", entry.Name(), filepath.Base(actualOutputPath)))
					}
				}

				// 独立的单个 .mp4：未启用裁剪规则时直接跳过（避免无意义 ProbeHealth 调用）；
				// 已启用裁剪规则时纳入裁剪检测，确保单文件直录也受主播裁剪规则控制。
				if len(batch) == 1 && strings.EqualFold(filepath.Ext(batch[0].Name), ".mp4") {
					enqueueSingleMP4Trim(batch[0].Name)
					continue
				}

				// 文件健康检查：跳过空文件和结构损坏的文件（如缺少 moov atom）
				// 保留小文件（可能是分段之间的过渡内容）
				var names []string
				var size int64
				for _, v := range batch {
					path := filepath.Join(folder, v.Name)
					info, _ := os.Stat(path)
					if info == nil || info.Size() == 0 {
						continue
					}
					if !ffmpeg.ProbeHealth(ctx, path) {
						s.logger.Info(fmt.Sprintf("[%s] ⏭ 跳过损坏文件: %s", entry.Name(), v.Name))
						continue
					}
					names = append(names, v.Name)
					size += info.Size()
				}

				if len(names) == 0 {
					s.logger.Info(fmt.Sprintf("[%s] ⏭ %d 个文件全部无效，跳过", entry.Name(), len(batch)))
					continue
				}

				// 过滤后只剩单文件 — FLV / TS 统一转封装为 MP4，MP4 按需裁剪
				if len(names) == 1 {
					singleName := names[0]
					ext := strings.ToLower(filepath.Ext(singleName))

					if ext == ".flv" || ext == ".ts" {
						srcPath := filepath.Join(folder, singleName)
						srcInfo, srcErr := os.Stat(srcPath)
						if srcErr != nil {
							continue
						}

						if isStreamActive(folder, key) {
							streamerActive = true
							s.logger.Info(fmt.Sprintf("[%s] ⏭ %s → 正在直播录制中，跳过", entry.Name(), singleName))
							continue
						}

						ageMin := time.Since(srcInfo.ModTime()).Minutes()
						if ageMin < mergeAgeMin {
							streamerWaitingSafety = true
							s.logger.Info(fmt.Sprintf("[%s] ⏭ %s → 处于关播合并等待期（录制结束 %.0f 分钟，需等待满 %.0f 分钟）", entry.Name(), singleName, ageMin, mergeAgeMin))
							continue
						}

						mp4Name := utils.MakeMP4Name(singleName)
						mp4Path := filepath.Join(folder, mp4Name)
						if mp4Info, err := os.Stat(mp4Path); err == nil && mp4Info.Size() >= minValidFileSize {
							if ffmpeg.QuickProbe(ctx, mp4Path) == nil {
								utils.SafeUnlink(srcPath)
								s.logger.Info(fmt.Sprintf("[%s] ✅ %s → 已存在对应 MP4，清理源 %s 文件", entry.Name(), singleName, strings.ToUpper(strings.TrimPrefix(ext, "."))))
								continue
							}
							utils.SafeUnlink(mp4Path)
							s.logger.Info(fmt.Sprintf("[%s] ⚠ %s → MP4 文件损坏，重新执行格式转换", entry.Name(), singleName))
						}
						convertTasks = append(convertTasks, convertTask{FlvPath: srcPath, Mp4Path: mp4Path, Folder: folder, Name: entry.Name()})
						s.logger.Info(fmt.Sprintf("[%s] 🔄 %s → 待转换 %s → MP4", entry.Name(), singleName, strings.ToUpper(strings.TrimPrefix(ext, "."))))
					} else if ext == ".mp4" {
						enqueueSingleMP4Trim(singleName)
					}
					continue
				}

				// 多文件合并（过滤后 >=2 个文件）
				// 主播正在录制时，跳过所有批次（避免合并未完成的场次）
				if isStreamActive(folder, key) {
					streamerActive = true
					s.logger.Info(fmt.Sprintf("[%s] ⏭ %d 个文件 → 主播正在直播录制中，跳过", entry.Name(), len(names)))
					continue
				}

				lastFile := filepath.Join(folder, names[len(names)-1])
				latestMtime := time.Now()
				if lfi, err := os.Stat(lastFile); err == nil {
					latestMtime = lfi.ModTime()
				}
				ageMinutes := time.Since(latestMtime).Minutes()
				if ageMinutes < mergeAgeMin {
					streamerWaitingSafety = true
					s.logger.Info(fmt.Sprintf("[%s] ⏭ %d 个分片 → 处于关播合并等待期（录制结束 %.0f 分钟，需等待满 %.0f 分钟）", entry.Name(), len(names), ageMinutes, mergeAgeMin))
					continue
				}

				if isFileBeingWritten(ctx, lastFile, stabilityProbeInterval(latestMtime)) {
					streamerActive = true
					s.logger.Info(fmt.Sprintf("[%s] ⏭ %d 个文件 → 录像文件正在写入中，跳过", entry.Name(), len(names)))
					continue
				}

				// Phase 1: 通过快速检查，加入候选列表（不阻塞）
				s.logger.Info(fmt.Sprintf("[%s] 🔗 %d 个分片 (%.1f GB) → 候选分片待验证稳定性", entry.Name(), len(names), float64(size)/oneGB))
				pending = append(pending, pendingBatch{
					Names:     names,
					Folder:    folder,
					Size:      size,
					Streamer:  entry.Name(),
					LastMtime: latestMtime,
				})
			}
		}
		if len(convertTasks) > initConvertCount || len(pending) > initPendingCount {
			// 有待处理任务
		} else if streamerActive {
			stats.ActiveStreams++
		} else if streamerWaitingSafety {
			stats.WaitingSafety++
		} else {
			stats.UpToDate++
		}
	}
	// Phase 2: 并发验证候选批次的文件稳定性（按文件静置时长自适应等待窗口）
	if len(pending) > 0 {
		s.logger.Info(fmt.Sprintf("⏳ 并发验证 %d 个候选批次的文件稳定性...", len(pending)))
		var wg sync.WaitGroup
		sem := make(chan struct{}, 4) // 限制并发数为 4
		var mu sync.Mutex
		for i := range pending {
			wg.Add(1)
			sem <- struct{}{}
			go func(pb *pendingBatch) {
				defer wg.Done()
				defer func() { <-sem }()
				lastFile := filepath.Join(pb.Folder, pb.Names[len(pb.Names)-1])
				waitInterval := stabilityWaitDuration(pb.LastMtime)
				if isFileSizeStable(ctx, lastFile, waitInterval) {
					mu.Lock()
					tasks = append(tasks, mergeTask{
						Files:  pb.Names,
						Folder: pb.Folder,
						SizeGB: float64(pb.Size) / oneGB,
					})
					mu.Unlock()
					s.logger.Info(fmt.Sprintf("[%s] ✅ %d 个分片 (%.1f GB) → 文件已稳定，准备合并", pb.Streamer, len(pb.Names), float64(pb.Size)/oneGB))
				} else {
					s.logger.Info(fmt.Sprintf("[%s] ⏭ %d 个分片 → 文件大小仍在变化中，跳过", pb.Streamer, len(pb.Names)))
				}
			}(&pending[i])
		}
		wg.Wait()
	}
	return tasks, convertTasks, stats
}

// checkSingleMP4Trim 检查单个 MP4 是否满足主播个性化裁剪入队条件。
// 通过 checkedMP4s 指纹缓存（大小 + mtime + 裁剪规则）避免已裁剪或无需裁剪的独立 MP4 在每轮扫描时重复拉起 ffprobe。
func (s *MergeService) checkSingleMP4Trim(ctx context.Context, folder, streamerName, batchKey, mp4Path string, rule config.StreamerRule, mergeAgeMin float64) (task convertTask, active bool, waiting bool, ok bool) {
	info, err := os.Stat(mp4Path)
	if err != nil || info.Size() < minValidFileSize {
		return convertTask{}, false, false, false
	}

	fp := mp4CheckFingerprint{
		size:    info.Size(),
		mtime:   info.ModTime().UnixNano(),
		ruleSig: ruleFingerprint(rule),
	}
	if cached, hit := s.checkedMP4s.Load(mp4Path); hit && cached == fp {
		return convertTask{}, false, false, false
	}

	if isStreamActive(folder, batchKey) {
		return convertTask{}, true, false, false
	}
	if time.Since(info.ModTime()).Minutes() < mergeAgeMin {
		return convertTask{}, false, true, false
	}
	if !ffmpeg.ProbeHealth(ctx, mp4Path) {
		return convertTask{}, false, false, false
	}
	if ffmpeg.IsVideoTrimmed(ctx, mp4Path) || !shouldTrimVideo(ctx, mp4Path, rule) {
		s.checkedMP4s.Store(mp4Path, fp)
		return convertTask{}, false, false, false
	}
	return convertTask{
		FlvPath: mp4Path,
		Mp4Path: mp4Path,
		Folder:  folder,
		Name:    streamerName,
	}, false, false, true
}
