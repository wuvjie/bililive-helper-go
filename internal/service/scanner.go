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

// videoFile holds parsed metadata for a single video file.
type videoFile struct {
	Name     string
	Key      string
	Datetime time.Time
}

type mergeTask struct {
	Files  []string
	Folder string
	SizeGB float64
}

type convertTask struct {
	FlvPath string
	Mp4Path string
	Folder  string
	Name    string
}

// isFileBeingWritten checks if file size changes over an interval.
// Returns true if file is being written, missing, or size changed.
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

// isFileSizeStable checks if file size hasn't changed over an interval.
// Used as final safety check before processing — catches hung write processes.
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

// getVideoFiles returns all non-merged video files with parsed metadata.
// Also self-heals: if a valid MP4 exists, leftover FLV/TS with the same base name are deleted.
func (s *MergeService) getVideoFiles(folder string) []videoFile {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil
	}

	// Pass 1: index base names that have valid-sized MP4 files
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

	// Pass 2: collect videos, clean up residuals
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

		// Residual cleanup: if MP4 exists, the FLV/TS is a leftover from a previous failed delete
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

// isStreamActive checks if the LATEST file with the same key is currently being written to.
// Only probes the newest file — avoids the "collateral skip" where older batches are
// blocked by the mere existence of newer batches from a different session.
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

		// Group by key
		groups := make(map[string][]videoFile)
		for _, v := range videos {
			groups[v.Key] = append(groups[v.Key], v)
		}

		// For each group: sort by datetime (from filename), split by time gap
		// Use Datetime exclusively — mtime is unreliable (changes on copy/touch/scan)
		for key, items := range groups {
			sort.Slice(items, func(i, j int) bool {
				return items[i].Datetime.Before(items[j].Datetime)
			})

			batches := [][]videoFile{{items[0]}}
			for i := 0; i < len(items)-1; i++ {
				curr := items[i]
				next := items[i+1]
				gapMin := next.Datetime.Sub(curr.Datetime).Minutes()
				if gapMin < 0 {
					gapMin = 0
				}
				if gapMin > float64(gapMinutes) {
					batches = append(batches, []videoFile{next})
				} else {
					batches[len(batches)-1] = append(batches[len(batches)-1], next)
				}
			}

			for _, batch := range batches {
				outputName := utils.MakeOutputName(batch[0].Name)
				outputPath := filepath.Join(folder, outputName)
				if info, err := os.Stat(outputPath); err == nil && info.Size() >= 10240 {
					if ffmpeg.QuickProbe(context.Background(), outputPath) == nil {
						for _, v := range batch {
							utils.SafeUnlink(filepath.Join(folder, v.Name))
						}
						s.logToFile("merge", fmt.Sprintf("[%s] ✅ %s → 已合并，清理原片", entry.Name(), outputName))
					} else {
						utils.SafeUnlink(outputPath)
						s.logToFile("merge", fmt.Sprintf("[%s] ⚠ %s → 输出损坏，将重新合并", entry.Name(), outputName))
					}
					continue
				}

				// Filter: remove corrupt files (< 1MB or < 5s) BEFORE routing
				var names []string
				var size int64
				for _, v := range batch {
					path := filepath.Join(folder, v.Name)
					info, _ := os.Stat(path)
					if info == nil {
						continue
					}
					sz := info.Size()
					if sz < 1048576 {
						s.logToFile("merge", fmt.Sprintf("⏭ [%s] 跳过过小文件: %s (%s)", entry.Name(), v.Name, utils.FormatSize(sz)))
						continue
					}
					if dur, err := utils.GetVideoDuration(path); err != nil || dur < 5 {
						s.logToFile("merge", fmt.Sprintf("⏭ [%s] 跳过无效文件: %s (时长%.0fs)", entry.Name(), v.Name, dur))
						continue
					}
					names = append(names, v.Name)
					size += sz
				}

				if len(names) == 0 {
					s.logToFile("merge", fmt.Sprintf("[%s] ⏭ 全部文件无效，跳过", entry.Name()))
					continue
				}

				// Single file after filtering — route by format
				if len(names) == 1 {
					singleName := names[0]
					ext := strings.ToLower(filepath.Ext(singleName))

					if ext == ".flv" {
						// FLV → MP4 conversion path
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
					// Single MP4: no action needed, skip silently
					continue
				}

				// Multi-file merge (len(names) >= 2)
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
