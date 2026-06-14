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

	"bililive-helper/internal/config"
	"bililive-helper/internal/ffmpeg"
	"bililive-helper/internal/utils"

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
	Names    []string
	Folder   string
	Size     int64
	Streamer string
}

// convertTask 表示一个待转换的任务（FLV 转 MP4）。
type convertTask struct {
	FlvPath string
	Mp4Path string
	Folder  string
	Name    string
}

const (
	recordingActiveThreshold = 5 * time.Minute // 录制活跃判定阈值：最新文件在此时间内被认为还在录制
	defaultGapMinutes        = 30              // 场次间隔回退值：未配置时的默认分钟数
	defaultMergeAgeMinutes   = 30              // 合并安全期回退值：未配置时的默认分钟数
)

// isFileBeingWritten 通过比较两次文件大小来检测文件是否正在被写入。
// 文件缺失、不可访问或大小变化时返回 true。
func isFileBeingWritten(ctx context.Context, path string, interval time.Duration) bool {
	info1, err := os.Stat(path)
	if err != nil {
		return true
	}
	select {
	case <-time.After(interval):
	case <-ctx.Done():
		return true
	}
	info2, err := os.Stat(path)
	if err != nil {
		return true
	}
	return info1.Size() != info2.Size()
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

// getVideoFiles 返回目录中所有非已合并的视频文件及其解析元数据。
// 具有自愈能力：如果存在有效的 MP4 文件，会清理同名的残留 FLV/TS 文件。
func (s *MergeService) getVideoFiles(folder string) []videoFile {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil
	}

	// 第一遍：索引有有效 MP4 文件的基础文件名
	mp4Bases := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if utils.IsMergedFile(name) {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".mp4" {
			// 仅保留足够大的 MP4 文件（排除损坏的小文件）
			if info, err := entry.Info(); err == nil && info.Size() >= minValidFileSize {
				base := strings.TrimSuffix(name, ext)
				mp4Bases[base] = true
			}
		}
	}

	// 第二遍：收集视频文件，清理上次删除失败的残留
	var videos []videoFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !utils.IsVideoFile(name) {
			continue
		}
		if utils.IsMergedFile(name) {
			continue
		}

		ext := strings.ToLower(filepath.Ext(name))
		base := strings.TrimSuffix(name, ext)

		// 残留清理：如果已存在有效的 MP4，同名的 FLV/TS 是上次删除失败的残留
		if ext != ".mp4" && mp4Bases[base] {
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
			continue
		}
		videos = append(videos, videoFile{Name: name, Key: key, Datetime: dt})
	}
	return videos
}

// isStreamActive 检查指定 key 的最新文件是否正在被录制。
// 判断依据：最新文件的修改时间在最近 5 分钟内 → 认为还在录制。
// 比 isFileBeingWritten 更可靠——不受短暂缓冲/网络波动影响。
func isStreamActive(folder string, batchKey string) bool {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return false
	}

	var newestMtime time.Time
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if utils.IsMergedFile(name) {
			continue
		}
		if !strings.Contains(name, batchKey) {
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

// scanTasks 扫描录制目录，将文件按主播分组、按时间排序、按间隔分场次。
// 返回待合并任务列表和待转换任务列表（FLV→MP4）。
func (s *MergeService) scanTasks(ctx context.Context, root, streamer string, cfg config.Config) ([]mergeTask, []convertTask) {
	var tasks []mergeTask
	var convertTasks []convertTask
	var pending []pendingBatch // Phase 1: 收集候选，Phase 2: 并发验证稳定性
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil
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

		folder := filepath.Join(root, entry.Name())
		videos := s.getVideoFiles(folder)
		if len(videos) == 0 {
			continue
		}

		// 按文件名中的 key（主播+标题）分组
		groups := make(map[string][]videoFile)
		for _, v := range videos {
			groups[v.Key] = append(groups[v.Key], v)
		}

		// 每组按文件名时间排序，同时间戳按文件名排序（001 < 002 < 003）
		for key, items := range groups {
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].Datetime.Equal(items[j].Datetime) {
					return items[i].Name < items[j].Name
				}
				return items[i].Datetime.Before(items[j].Datetime)
			})

			// 对每个文件并发用 ffprobe 探测实际时长，计算结束时间
			// 使用带缓冲 channel 作为 semaphore 限制并发数为 4
			var wg sync.WaitGroup
			sem := make(chan struct{}, 4)
			for i := range items {
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

			// 修正：同时间戳的分片文件，EndTime 应顺序累加而非各自独立计算。
			// bililive-go 文件名中的时间是场次开始时间，不是每个分片的开始时间。
			// 例：001.flv(30min) 002.flv(30min) 都标 21:33，
			//     正确 EndTime: 001→22:03, 002→22:33（而非都是 22:03）
			for i := 1; i < len(items); i++ {
				if !items[i].Datetime.After(items[i-1].Datetime) {
					dur := items[i].EndTime.Sub(items[i].Datetime)
					if dur <= 0 {
						dur = 0
					}
					items[i].EndTime = items[i-1].EndTime.Add(dur)
				}
			}

			// 用前一个文件的结束时间 vs 下一个文件的开始时间 计算 gap
			// gap < GapMinutes → 同场次，合并；gap > GapMinutes → 不同场次，分批
			batches := [][]videoFile{{items[0]}}
			for i := 0; i < len(items)-1; i++ {
				gapMin := items[i+1].Datetime.Sub(items[i].EndTime).Minutes()
				if gapMin < 0 {
					gapMin = 0
				}
				if gapMin > float64(gapMinutes) {
					batches = append(batches, []videoFile{items[i+1]})
				} else {
					batches[len(batches)-1] = append(batches[len(batches)-1], items[i+1])
				}
			}

			for _, batch := range batches {
				outputName := utils.MakeOutputName(batch[0].Name)
				outputPath := filepath.Join(folder, outputName)

				// FLV 输入合并后输出为 MP4，需要检查 MP4 版本
				actualOutputPath := outputPath
				if strings.HasSuffix(outputName, ".flv") {
					mp4Path := strings.TrimSuffix(outputPath, ".flv") + ".mp4"
					if info, err := os.Stat(mp4Path); err == nil && info.Size() >= minValidFileSize {
						actualOutputPath = mp4Path
					}
				}

				if info, err := os.Stat(actualOutputPath); err == nil && info.Size() >= minValidFileSize {
					if ffmpeg.QuickProbe(ctx, actualOutputPath) == nil {
						for _, v := range batch {
							if err := utils.SafeUnlink(filepath.Join(folder, v.Name)); err != nil {
								s.logger.Warn("清理原片失败", zap.String("file", v.Name), zap.Error(err))
							}
						}
						s.logger.Info(fmt.Sprintf("[%s] ✅ %s → 已合并，清理原片", entry.Name(), filepath.Base(actualOutputPath)))
					} else {
						if err := utils.SafeUnlink(actualOutputPath); err != nil {
							s.logger.Warn("清理损坏输出失败", zap.String("file", filepath.Base(actualOutputPath)), zap.Error(err))
						}
						s.logger.Info(fmt.Sprintf("[%s] ⚠ %s → 输出损坏，将重新合并", entry.Name(), filepath.Base(actualOutputPath)))
					}
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

				// 过滤后只剩单文件 — 按格式路由（FLV 需转换，MP4 跳过）
				if len(names) == 1 {
					singleName := names[0]
					ext := strings.ToLower(filepath.Ext(singleName))

					// FLV → MP4 转换路径
					if ext == ".flv" {
						flvPath := filepath.Join(folder, singleName)
						flvInfo, flvErr := os.Stat(flvPath)
						if flvErr != nil {
							continue
						}

						if isStreamActive(folder, key) {
							s.logger.Info(fmt.Sprintf("[%s] ⏭ %s → 录制中，跳过", entry.Name(), singleName))
							continue
						}

						ageMin := time.Since(flvInfo.ModTime()).Minutes()
						mergeAgeMin := float64(cfg.MergeAgeMinutes)
						if mergeAgeMin <= 0 {
							mergeAgeMin = defaultMergeAgeMinutes
						}
						if ageMin < mergeAgeMin {
							s.logger.Info(fmt.Sprintf("[%s] ⏭ %s → 等待安全期（%.0f分钟前，需%.0f分钟）", entry.Name(), singleName, ageMin, mergeAgeMin))
							continue
						}

						mp4Name := utils.MakeMP4Name(singleName)
						mp4Path := filepath.Join(folder, mp4Name)
						if mp4Info, err := os.Stat(mp4Path); err == nil && mp4Info.Size() >= minValidFileSize {
							if ffmpeg.QuickProbe(ctx, mp4Path) == nil {
								utils.SafeUnlink(flvPath)
								s.logger.Info(fmt.Sprintf("[%s] ✅ %s → 已有MP4，清理FLV", entry.Name(), singleName))
								continue
							}
							utils.SafeUnlink(mp4Path)
							s.logger.Info(fmt.Sprintf("[%s] ⚠ %s → MP4损坏，重新转换", entry.Name(), singleName))
						}
						convertTasks = append(convertTasks, convertTask{FlvPath: flvPath, Mp4Path: mp4Path, Folder: folder, Name: entry.Name()})
						s.logger.Info(fmt.Sprintf("[%s] 🔄 %s → 待转换 FLV→MP4", entry.Name(), singleName))
					} else if ext == ".ts" {
						s.logger.Info(fmt.Sprintf("[%s] ⏭ %s → 孤立TS，等待清理", entry.Name(), singleName))
					}
					// 单个 MP4 无需操作，静默跳过
					continue
				}

				// 多文件合并（过滤后 >=2 个文件）
				// 主播正在录制时，跳过所有批次（避免合并未完成的场次）
				if isStreamActive(folder, key) {
					s.logger.Info(fmt.Sprintf("[%s] ⏭ %d个文件 → 主播正在录制，跳过", entry.Name(), len(names)))
					continue
				}

				lastFile := filepath.Join(folder, names[len(names)-1])
				if isFileBeingWritten(ctx, lastFile, 2*time.Second) {
					s.logger.Info(fmt.Sprintf("[%s] ⏭ %d个文件 → 文件写入中，跳过", entry.Name(), len(names)))
					continue
				}

				latestMtime := time.Now()
				if lfi, err := os.Stat(lastFile); err == nil {
					latestMtime = lfi.ModTime()
				}
				ageMinutes := time.Since(latestMtime).Minutes()
				mergeAgeMinutes := float64(cfg.MergeAgeMinutes)
				if mergeAgeMinutes <= 0 {
					mergeAgeMinutes = defaultMergeAgeMinutes
				}

				if ageMinutes < mergeAgeMinutes {
					s.logger.Info(fmt.Sprintf("[%s] ⏭ %d个文件 → 等待安全期（%.0f分钟前，需%.0f分钟）", entry.Name(), len(names), ageMinutes, mergeAgeMinutes))
					continue
				}

				// Phase 1: 通过快速检查，加入候选列表（不阻塞）
				s.logger.Info(fmt.Sprintf("[%s] 🔗 %d个文件 (%.1f GB) → 候选待验证", entry.Name(), len(names), float64(size)/oneGB))
				pending = append(pending, pendingBatch{
					Names:    names,
					Folder:   folder,
					Size:     size,
					Streamer: entry.Name(),
				})
			}
		}
	}
	// Phase 2: 并发验证候选批次的文件稳定性（避免串行阻塞 1 分钟/批次）
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
				if isFileSizeStable(ctx, lastFile, 1*time.Minute) {
					mu.Lock()
					tasks = append(tasks, mergeTask{
						Files:  pb.Names,
						Folder: pb.Folder,
						SizeGB: float64(pb.Size) / oneGB,
					})
					mu.Unlock()
					s.logger.Info(fmt.Sprintf("[%s] ✅ %d个文件 (%.1f GB) → 稳定，待合并", pb.Streamer, len(pb.Names), float64(pb.Size)/oneGB))
				} else {
					s.logger.Info(fmt.Sprintf("[%s] ⏭ %d个文件 → 文件大小仍在变化，跳过", pb.Streamer, len(pb.Names)))
				}
			}(&pending[i])
		}
		wg.Wait()
	}
	return tasks, convertTasks
}
