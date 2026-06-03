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

// convertTask 表示一个待转换的任务（FLV 转 MP4）。
type convertTask struct {
	FlvPath string
	Mp4Path string
	Folder  string
	Name    string
}

// isFileBeingWritten 通过比较两次文件大小来检测文件是否正在被写入。
// 文件缺失、不可访问或大小变化时返回 true。

func isFileBeingWritten(path string, interval time.Duration) bool {
	info1, err := os.Stat(path)
	if err != nil {
		return true
	}
	time.Sleep(interval)
	info2, err := os.Stat(path)
	if err != nil {
		return true
	}
	return info1.Size() != info2.Size()
}

// isFileSizeStable 检查文件大小在指定间隔内是否保持稳定。
// 最终安全检查 — 捕获挂起的写入进程。
func isFileSizeStable(path string, interval time.Duration) bool {
	info1, err := os.Stat(path)
	if err != nil {
		return false
	}
	time.Sleep(interval)
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
			// Only consider MP4 files that are large enough to be valid
			if info, err := entry.Info(); err == nil && info.Size() >= 10240 {
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
				s.logToFile("merge", fmt.Sprintf("🗑 清理残留原片: %s", name))
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

// isStreamActive 检查指定 key 的最新文件是否正在被录制写入。
// 只探测最新文件，避免因新批次存在而导致旧批次被错误跳过。
func isStreamActive(folder string, batchKey string) bool {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return false
	}

	var newestPath string
	var maxMtime time.Time

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
		if info.ModTime().After(maxMtime) {
			maxMtime = info.ModTime()
			newestPath = filepath.Join(folder, name)
		}
	}

	if newestPath == "" {
		return false
	}
	return isFileBeingWritten(newestPath, 1*time.Second)
}

// scanTasks 扫描录制目录，将文件按主播分组、按时间排序、按间隔分场次。
// 返回待合并任务列表和待转换任务列表（FLV→MP4）。
func (s *MergeService) scanTasks(root, streamer string, cfg config.Config) ([]mergeTask, []convertTask) {
	var tasks []mergeTask
	var convertTasks []convertTask
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil
	}

	gapMinutes := cfg.GapMinutes
	if gapMinutes <= 0 {
		gapMinutes = 20
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

			// 对每个文件用 ffprobe 探测实际时长，计算结束时间
			for i := range items {
				path := filepath.Join(folder, items[i].Name)
				dur, err := utils.GetVideoDuration(path)
				if err == nil && dur > 0 {
					items[i].EndTime = items[i].Datetime.Add(time.Duration(dur * float64(time.Second)))
				} else {
					items[i].EndTime = items[i].Datetime
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
					if info, err := os.Stat(mp4Path); err == nil && info.Size() >= 10240 {
						actualOutputPath = mp4Path
					}
				}

				if info, err := os.Stat(actualOutputPath); err == nil && info.Size() >= 10240 {
					if ffmpeg.QuickProbe(context.Background(), actualOutputPath) == nil {
						for _, v := range batch {
							utils.SafeUnlink(filepath.Join(folder, v.Name))
						}
						s.logToFile("merge", fmt.Sprintf("[%s] ✅ %s → 已合并，清理原片", entry.Name(), filepath.Base(actualOutputPath)))
					} else {
						utils.SafeUnlink(actualOutputPath)
						s.logToFile("merge", fmt.Sprintf("[%s] ⚠ %s → 输出损坏，将重新合并", entry.Name(), filepath.Base(actualOutputPath)))
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
					if !utils.IsVideoHealthy(path) {
						s.logToFile("merge", fmt.Sprintf("⏭ [%s] 跳过损坏文件: %s", entry.Name(), v.Name))
						continue
					}
					names = append(names, v.Name)
					size += info.Size()
				}

				if len(names) == 0 {
					s.logToFile("merge", fmt.Sprintf("[%s] ⏭ 全部文件无效，跳过", entry.Name()))
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
							s.logToFile("merge", fmt.Sprintf("[%s] ⏭ %s → 录制中，跳过", entry.Name(), singleName))
							continue
						}

						ageMin := time.Since(flvInfo.ModTime()).Minutes()
						mergeAgeMin := float64(cfg.MergeAgeMinutes)
						if mergeAgeMin <= 0 {
							mergeAgeMin = 30
						}
						if ageMin < mergeAgeMin {
							s.logToFile("merge", fmt.Sprintf("[%s] ⏭ %s → 落盘等待（%.0f分钟前，需%.0f分钟）", entry.Name(), singleName, ageMin, mergeAgeMin))
							continue
						}

						mp4Name := utils.MakeMP4Name(singleName)
						mp4Path := filepath.Join(folder, mp4Name)
						if mp4Info, err := os.Stat(mp4Path); err == nil && mp4Info.Size() >= 10240 {
							if ffmpeg.QuickProbe(context.Background(), mp4Path) == nil {
								utils.SafeUnlink(flvPath)
								s.logToFile("merge", fmt.Sprintf("[%s] ✅ %s → 已有MP4，清理FLV", entry.Name(), singleName))
								continue
							}
							utils.SafeUnlink(mp4Path)
							s.logToFile("merge", fmt.Sprintf("[%s] ⚠ %s → MP4损坏，重新转换", entry.Name(), singleName))
						}
						convertTasks = append(convertTasks, convertTask{FlvPath: flvPath, Mp4Path: mp4Path, Folder: folder, Name: entry.Name()})
						s.logToFile("merge", fmt.Sprintf("[%s] 🔄 %s → 待转换 FLV→MP4", entry.Name(), singleName))
					} else if ext == ".ts" {
						s.logToFile("merge", fmt.Sprintf("[%s] ⏭ %s → 孤立TS，等待清理", entry.Name(), singleName))
					}
					// 单个 MP4 无需操作，静默跳过
					continue
				}

				// 多文件合并（过滤后 >=2 个文件）
				lastFile := filepath.Join(folder, names[len(names)-1])
				if isFileBeingWritten(lastFile, 2*time.Second) {
					s.logToFile("merge", fmt.Sprintf("[%s] ⏭ %d个文件 → 录制中，跳过", entry.Name(), len(names)))
					continue
				}

				latestMtime := time.Now()
				if lfi, err := os.Stat(lastFile); err == nil {
					latestMtime = lfi.ModTime()
				}
				ageMinutes := time.Since(latestMtime).Minutes()
				mergeAgeMinutes := float64(cfg.MergeAgeMinutes)
				if mergeAgeMinutes <= 0 {
					mergeAgeMinutes = 30
				}

				streamActive := isStreamActive(folder, key)
				if streamActive || ageMinutes < mergeAgeMinutes {
					s.logToFile("merge", fmt.Sprintf("[%s] ⏭ %d个文件 → 落盘等待（%.0f分钟前，需%.0f分钟）", entry.Name(), len(names), ageMinutes, mergeAgeMinutes))
					continue
				}

				if !isFileSizeStable(lastFile, 1*time.Minute) {
					s.logToFile("merge", fmt.Sprintf("[%s] ⏭ %d个文件 → 文件大小仍在变化", entry.Name(), len(names)))
					continue
				}

				s.logToFile("merge", fmt.Sprintf("[%s] 🔗 %d个文件 (%.1f GB) → 待合并", entry.Name(), len(names), float64(size)/1073741824))
				tasks = append(tasks, mergeTask{
					Files:  names,
					Folder: folder,
					SizeGB: float64(size) / 1073741824,
				})
			}
		}
	}
	return tasks, convertTasks
}
