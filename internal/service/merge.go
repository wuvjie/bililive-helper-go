// merge.go 提供录制文件的合并服务。
// 支持多文件 TS 拼接合并、FLV 转 MP4、重编码 fallback，以及主播级锁防止并发合并。
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
	"bililive-helper-go/internal/fsutil"
	"bililive-helper-go/internal/utils"

	"go.uber.org/zap"
)

// ProgressFunc 是进度回调函数类型，用于 SSE 流式输出。
type ProgressFunc func(msg string)

const (
	minDiskFreeBytes       = 10 * 1024 * 1024 * 1024 // 10GB: 磁盘空间硬限，低于此值跳过所有操作
	minConvertFreeBytes    = 512 * 1024 * 1024       // 512MB: FLV 转换所需最小磁盘空间
	minMergeFreeBytes      = 1024 * 1024 * 1024      // 1GB: 合并操作所需最小磁盘空间
	mergeHeadroomBytes     = 1 * 1024 * 1024 * 1024  // 1GB: 直接合并/fallback 峰值空间余量
	maxDedupAttempts       = 1000                    // 文件名去重最大尝试次数
	oneGB                  = 1073741824              // 1GB 字节数，用于大小格式化换算
	minValidFileSize       = 10240                   // 10KB: 视频文件最小有效大小
)

// formatDuration 格式化耗时：>1h 显示 "1h23m"，>1m 显示 "2m3s"，否则显示 "45s"。
func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

type mp4CheckFingerprint struct {
	size    int64
	mtime   int64
	ruleSig string
}

func ruleFingerprint(rule config.StreamerRule) string {
	return fmt.Sprintf("%v|%s|%s", rule.TrimEnabled, strings.TrimSpace(rule.TrimStart), strings.TrimSpace(rule.TrimEnd))
}

// MergeService 提供录制文件合并功能。
// 使用 per-streamer 锁防止同一主播的并发合并，支持 FLV 转 MP4、多文件 TS 拼接和重编码 fallback。
type MergeService struct {
	config        *config.Config
	logger        *zap.Logger
	history       *HistoryService
	streamerLocks sync.Map // streamer name -> *streamerLock
	checkedMP4s   sync.Map // abs/full path -> mp4CheckFingerprint，避免单分片 MP4 重复触发 ffprobe
}

// streamerLock 用于 per-streamer 粒度的合并锁，防止同一主播并发合并。
type streamerLock struct {
	mu sync.Mutex
}

// tryLockStreamer 尝试获取指定主播的合并锁。
// 锁被占用时直接返回 false，不强制回收。
// 依赖 Context 超时（FFmpeg Run 已有 2 小时超时）自然终止长时间任务并释放锁。
func (s *MergeService) tryLockStreamer(name string) (bool, *streamerLock) {
	val, _ := s.streamerLocks.LoadOrStore(name, &streamerLock{})
	sl := val.(*streamerLock)

	if sl.mu.TryLock() {
		return true, sl
	}

	return false, nil
}

func (s *MergeService) unlockStreamer(sl *streamerLock) {
	if sl != nil {
		sl.mu.Unlock()
	}
}

// IsStreamerLocked 检查指定主播是否有正在进行的合并任务。
// 实现 StreamerLockChecker 接口，供 CleanService 调用。
func (s *MergeService) IsStreamerLocked(name string) bool {
	val, ok := s.streamerLocks.Load(name)
	if !ok {
		return false
	}
	sl := val.(*streamerLock)
	if sl.mu.TryLock() {
		sl.mu.Unlock() // 成功获取说明锁未被持有，立即释放
		return false
	}
	return true // TryLock 失败说明锁被其他 goroutine 持有
}

// NewMergeService 创建合并服务实例。
func NewMergeService(config *config.Config, logger *zap.Logger, history *HistoryService) *MergeService {
	return &MergeService{config: config, logger: logger, history: history}
}

// MergeResult 保存合并操作的结果。
type MergeResult struct {
	Done    int
	Failed  int
	TotalGB float64
}

// checkDiskSpaceForMerge 检查合并任务是否有足够的磁盘空间。
// 优先采用直接流复制拼接：源文件 + 输出文件 ≈ 2 倍源文件 + 1GB 余量。
func (s *MergeService) checkDiskSpaceForMerge(tasks []mergeTask, targetDir string) error {
	var maxTaskBytes int64
	for _, t := range tasks {
		taskBytes := int64(t.SizeGB * oneGB)
		if taskBytes > maxTaskBytes {
			maxTaskBytes = taskBytes
		}
	}
	disk, err := utils.GetDiskUsage(targetDir)
	if err != nil {
		return fmt.Errorf("获取磁盘信息失败（%s）: %w", targetDir, err)
	}
	// 合并任务为串行执行，峰值空间取决于最大单场任务所需空间：源文件 + 输出文件 ≈ 2 倍源文件 + 1GB 余量
	needed := (maxTaskBytes * 2) + mergeHeadroomBytes
	if int64(disk.Free) < needed {
		return fmt.Errorf("磁盘空间不足：最大待合并任务需要 %.1f GB 可用空间，当前仅 %.1f GB",
			float64(needed)/oneGB, float64(disk.Free)/oneGB)
	}
	return nil
}

// classifyMergeFailure 检查合并失败的批次文件和输出，判断最可能的失败原因。
func classifyMergeFailure(folder, firstFile string) string {
	output := utils.MakeOutputName(firstFile)
	outputPath := filepath.Join(folder, output)
	if _, err := os.Stat(outputPath); os.IsNotExist(err) && !strings.EqualFold(filepath.Ext(output), ".mp4") {
		mp4OutputPath := utils.MakeMP4Name(outputPath)
		if _, mp4Err := os.Stat(mp4OutputPath); mp4Err == nil {
			outputPath = mp4OutputPath
		}
	}

	// 检查输出文件是否存在（可能被校验探针删除）
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		// 检查源文件是否有问题
		firstPath := filepath.Join(folder, firstFile)
		if info, err := os.Stat(firstPath); err != nil {
			return "源文件不存在"
		} else if info.Size() < minValidFileSize {
			return fmt.Sprintf("源文件过小(%s)", utils.FormatSize(info.Size()))
		}
		return "FFmpeg 输出校验失败"
	}

	// 输出存在但过小
	if info, err := os.Stat(outputPath); err == nil {
		if info.Size() < minValidFileSize {
			return fmt.Sprintf("输出过小(%s)", utils.FormatSize(info.Size()))
		}
	}
	return "输出校验失败"
}

// Run 执行合并任务主流程。
// 扫描录制目录，将 FLV 转 MP4 并合并多文件片段。
// 参数 streamer 为空表示全局合并；onProgress 用于 SSE 进度回调。
func (s *MergeService) Run(ctx context.Context, streamer string, onProgress ProgressFunc) (*MergeResult, string, error) {
	start := time.Now()

	setup, err := PrepareTask(s.config, s.logger, "merge_log", "merge", "合并", streamer, onProgress)
	if err != nil {
		return nil, "", err
	}
	defer setup.OpLog.Close()
	cfg := setup.Cfg
	progress := setup.Progress
	root := cfg.TargetDir

	// 磁盘状态快照（起始）
	diskStart, _ := utils.GetDiskUsage(root)
	if diskStart != nil {
		progress(fmt.Sprintf("▶ 开始 %s 录像合并 | 📊 磁盘空间使用率 %.1f%% (剩余 %.1f GB)", setup.Tag, diskStart.UsedPct, float64(diskStart.Free)/oneGB))
	} else {
		progress(fmt.Sprintf("▶ 开始 %s 录像合并", setup.Tag))
	}
	if streamer != "" {
		progress(fmt.Sprintf("⚙ 正在扫描主播「%s」的录像文件...", streamer))
	} else {
		progress("⚙ 正在扫描所有主播的录像存储目录...")
	}

	tasks, convertTasks, scanInfo := s.scanTasks(ctx, root, streamer, cfg)
	if len(tasks) == 0 && len(convertTasks) == 0 {
		duration := time.Since(start)
		var reason string
		if scanInfo.TotalStreamers == 0 {
			if streamer != "" {
				reason = fmt.Sprintf("未找到主播「%s」的录像目录", streamer)
			} else {
				reason = "录像存储目录中暂无任何主播文件夹"
			}
		} else if scanInfo.ActiveStreams > 0 && scanInfo.WaitingSafety == 0 {
			reason = fmt.Sprintf("已扫描 %d 位主播目录：当前正在直播录制中，暂不合并", scanInfo.TotalStreamers)
		} else if scanInfo.WaitingSafety > 0 && scanInfo.ActiveStreams == 0 {
			reason = fmt.Sprintf("已扫描 %d 位主播目录：录像处于关播合并等待期（需静置等待满 %d 分钟），暂不合并", scanInfo.TotalStreamers, cfg.MergeAgeMinutes)
		} else if scanInfo.ActiveStreams > 0 && scanInfo.WaitingSafety > 0 {
			reason = fmt.Sprintf("已扫描 %d 位主播目录：%d 位正在录制，%d 位处于关播合并等待期，暂不合并", scanInfo.TotalStreamers, scanInfo.ActiveStreams, scanInfo.WaitingSafety)
		} else {
			reason = fmt.Sprintf("已扫描 %d 位主播目录：暂无需要合并的分片录像", scanInfo.TotalStreamers)
		}

		progress(fmt.Sprintf("ℹ %s（耗时 %s）", reason, formatDuration(duration)))

		// 自动定时轮询空跑：不产生无用历史记录，清理空日志文件，保持列表纯净
		if !setup.IsManual {
			setup.OpLog.Discard()
			return &MergeResult{}, "", nil
		}

		// 手动触发时：保留历史记录以便用户确认操作反馈
		s.history.Add("merge", streamer, "success", reason, setup.LogID)
		return &MergeResult{}, setup.LogID, nil
	}

	// 任务预览
	totalTasks := len(convertTasks) + len(tasks)
	progress(fmt.Sprintf("⚙ 发现 %d 个待转换分片 + %d 场待合并录像（共 %d 项任务）", len(convertTasks), len(tasks), totalTasks))

	if len(tasks) > 0 {
		if err := s.checkDiskSpaceForMerge(tasks, root); err != nil {
			progress(fmt.Sprintf("❌ %s", err.Error()))
			return nil, setup.LogID, err
		}
	}

	// 磁盘空间硬性检查 — 可用空间低于 10GB 时跳过所有操作
	disk, diskErr := utils.GetDiskUsage(root)
	if diskErr == nil && disk.Free < minDiskFreeBytes {
		setup.OpLog.Log(fmt.Sprintf("❌ 磁盘剩余空间不足（仅剩 %.1f GB，需至少 10 GB），终止合并", float64(disk.Free)/oneGB))
		progress(fmt.Sprintf("❌ 磁盘剩余空间不足（仅剩 %.1f GB，需至少 10 GB），跳过本次合并", float64(disk.Free)/oneGB))
		s.history.Add("merge", streamer, "fail", fmt.Sprintf("磁盘空间不足: %.1f GB（使用率 %.1f%%）", float64(disk.Free)/oneGB, disk.UsedPct), setup.LogID)
		return &MergeResult{}, setup.LogID, nil
	}

	// 执行 FLV 转换和合并
	convertDone, convertBytes := s.runConversions(ctx, convertTasks, cfg, progress, setup.OpLog)
	mergeDone, mergeFailed, failedReasons, mergeBytes := s.runMerges(ctx, tasks, cfg, progress, setup.OpLog)
	totalBytes := convertBytes + mergeBytes
	totalGB := float64(totalBytes) / oneGB
	done := convertDone + mergeDone

	// 磁盘状态快照（结束）
	diskEnd, _ := utils.GetDiskUsage(root)

	// 结果汇总
	s.summarizeResults(ctx, root, streamer, done, convertDone, mergeDone, mergeFailed, failedReasons, convertBytes, mergeBytes, start, setup.LogID, progress, diskStart, diskEnd)
	return &MergeResult{Done: done, Failed: mergeFailed, TotalGB: totalGB}, setup.LogID, nil
}

// runConversions 执行所有单文件转换/裁剪任务（FLV/TS→MP4 或单 MP4 按规则裁剪）。返回完成数和实际生成成片总字节数。
func (s *MergeService) runConversions(ctx context.Context, convertTasks []convertTask, cfg config.Config, onProgress ProgressFunc, opLog *OpLogger) (int, int64) {
	done := 0
	var totalBytes int64
	lastStreamer := ""
	for _, ct := range convertTasks {
		if cfg.IsBackupWindow() || ctx.Err() != nil {
			break
		}
		streamerName := filepath.Base(filepath.Dir(ct.FlvPath))
		if streamerName != lastStreamer {
			onProgress(fmt.Sprintf("── %s ──", streamerName))
			lastStreamer = streamerName
		}
		var flvSize int64
		if fi, err := os.Stat(ct.FlvPath); err == nil {
			flvSize = fi.Size()
		}
		isTrimOnly := ct.FlvPath == ct.Mp4Path
		if isTrimOnly {
			onProgress(fmt.Sprintf("[%s] ✂ 正在按规则裁剪 MP4: %s (%s)", streamerName, filepath.Base(ct.Mp4Path), utils.FormatSize(flvSize)))
		} else {
			extLabel := strings.ToUpper(strings.TrimPrefix(filepath.Ext(ct.FlvPath), "."))
			onProgress(fmt.Sprintf("[%s] 🔄 正在转换 %s → MP4: %s (%s)", streamerName, extLabel, filepath.Base(ct.FlvPath), utils.FormatSize(flvSize)))
		}
		convertStart := time.Now()
		if outBytes, ok := s.convertFlvToMp4(ctx, ct.FlvPath, ct.Mp4Path, onProgress, opLog); ok {
			elapsed := time.Since(convertStart)
			done++
			effectiveBytes := outBytes
			if effectiveBytes <= 0 {
				effectiveBytes = flvSize
			}
			totalBytes += effectiveBytes
			if isTrimOnly {
				onProgress(fmt.Sprintf("[%s] ✂ ✅ 单分片裁剪完成: %s (%s，耗时 %s)", streamerName, filepath.Base(ct.Mp4Path), utils.FormatSize(effectiveBytes), formatDuration(elapsed)))
			} else if outBytes > 0 {
				onProgress(fmt.Sprintf("[%s] ✅ 格式转换完成: %s (%s，耗时 %s)", streamerName, filepath.Base(ct.Mp4Path), utils.FormatSize(outBytes), formatDuration(elapsed)))
			} else {
				onProgress(fmt.Sprintf("[%s] ✅ %s (%s)", streamerName, filepath.Base(ct.Mp4Path), formatDuration(elapsed)))
			}
		} else {
			if isTrimOnly {
				onProgress(fmt.Sprintf("[%s] ⚠ 单分片裁剪未生效", streamerName))
			} else {
				onProgress(fmt.Sprintf("[%s] ❌ 格式转换失败", streamerName))
			}
		}
	}
	return done, totalBytes
}

// runMerges 执行所有合并任务。返回成功数、失败数、失败原因统计和实际生成成片总字节数。
func (s *MergeService) runMerges(ctx context.Context, tasks []mergeTask, cfg config.Config, onProgress ProgressFunc, opLog *OpLogger) (int, int, map[string]int, int64) {
	done := 0
	failed := 0
	var totalBytes int64
	failedReasons := make(map[string]int)
	lastStreamer := ""
	for _, task := range tasks {
		if cfg.IsBackupWindow() {
			break
		}
		streamerName := filepath.Base(task.Folder)
		if streamerName != lastStreamer {
			onProgress(fmt.Sprintf("── %s ──", streamerName))
			lastStreamer = streamerName
		}
		func() {
			locked, sl := s.tryLockStreamer(streamerName)
			if !locked {
				return
			}
			defer s.unlockStreamer(sl)
			onProgress(fmt.Sprintf("[%s] ⚙ 正在合并 %d 个分片录像 (%s)", streamerName, len(task.Files), utils.FormatSize(int64(task.SizeGB*oneGB))))
			mergeStart := time.Now()
			if outBytes, ok := s.doMerge(ctx, task.Files, task.Folder, onProgress, opLog); ok {
				done++
				if outBytes > 0 {
					totalBytes += outBytes
				} else {
					totalBytes += int64(task.SizeGB * oneGB)
				}
				onProgress(fmt.Sprintf("[%s] ✅ 录像合并完成（耗时 %s）", streamerName, formatDuration(time.Since(mergeStart))))
			} else {
				failed++
				reason := classifyMergeFailure(task.Folder, task.Files[0])
				failedReasons[reason]++
				onProgress(fmt.Sprintf("[%s] ❌ 合并失败：%s", streamerName, reason))
			}
		}()
	}
	return done, failed, failedReasons, totalBytes
}

// summarizeResults 汇总合并结果并写入历史记录。
func (s *MergeService) summarizeResults(ctx context.Context, root, streamer string, done, convertDone, mergeDone, mergeFailed int, failedReasons map[string]int, convertBytes, mergeBytes int64, start time.Time, logID string, onProgress ProgressFunc, diskStart, diskEnd *utils.DiskUsage) {
	totalScanned := 0
	if dirs, err := fsutil.ScanStreamerDirs(root); err == nil {
		totalScanned = len(dirs)
	}
	if streamer != "" {
		totalScanned = 1
	}
	duration := time.Since(start)

	// 分隔线
	onProgress("───────────────────────────")
	onProgress("📊 结果摘要")

	// 扫描统计
	onProgress(fmt.Sprintf("  扫描：%d 位主播", totalScanned))

	if convertDone > 0 {
		onProgress(fmt.Sprintf("  转换：%d 个 FLV 分片 (%s)", convertDone, utils.FormatSize(convertBytes)))
	}
	if mergeDone > 0 {
		onProgress(fmt.Sprintf("  合并：%d 场录像 (%s)", mergeDone, utils.FormatSize(mergeBytes)))
	}
	if mergeFailed > 0 {
		var failParts []string
		for reason, cnt := range failedReasons {
			failParts = append(failParts, fmt.Sprintf("%s ×%d", reason, cnt))
		}
		onProgress(fmt.Sprintf("  失败：%d 项 (%s)", mergeFailed, strings.Join(failParts, ", ")))
	}

	// 耗时
	onProgress(fmt.Sprintf("  执行耗时：%s", formatDuration(duration)))

	// 磁盘变化
	if diskStart != nil && diskEnd != nil {
		freeChangeGB := float64(int64(diskEnd.Free)-int64(diskStart.Free)) / oneGB
		if freeChangeGB > 0 {
			onProgress(fmt.Sprintf("  磁盘空间：%.1f%% → %.1f%%（释放 %.1f GB）", diskStart.UsedPct, diskEnd.UsedPct, freeChangeGB))
		} else {
			onProgress(fmt.Sprintf("  磁盘空间：%.1f%% → %.1f%%（占用 %.1f GB）", diskStart.UsedPct, diskEnd.UsedPct, -freeChangeGB))
		}
	}

	onProgress("───────────────────────────")

	// 写入历史记录（统计按最终裁剪/合并后的成片字节数累加）
	totalBytes := convertBytes + mergeBytes
	detail := buildDetail(convertDone, mergeDone, mergeFailed, failedReasons, convertBytes, mergeBytes)
	status := "success"
	if mergeFailed > 0 {
		status = "partial"
	}
	if done > 0 {
		s.history.AddWithStats("merge", streamer, status, done, 0, totalBytes, duration.Seconds(), detail, logID)
	} else if mergeFailed > 0 {
		s.history.Add("merge", streamer, "fail", detail, logID)
	} else {
		s.history.Add("merge", streamer, "success", "无待合并文件", logID)
	}
}

// buildDetail 构建历史记录的详情字符串。
func buildDetail(convertDone, mergeDone, mergeFailed int, failedReasons map[string]int, convertBytes, mergeBytes int64) string {
	var parts []string
	if convertDone > 0 {
		parts = append(parts, fmt.Sprintf("转换 %d 个 FLV (%s)", convertDone, utils.FormatSize(convertBytes)))
	}
	if mergeDone > 0 {
		parts = append(parts, fmt.Sprintf("合并 %d 场次 (%s)", mergeDone, utils.FormatSize(mergeBytes)))
	}
	if mergeFailed > 0 {
		var failParts []string
		for reason, cnt := range failedReasons {
			failParts = append(failParts, fmt.Sprintf("%s ×%d", reason, cnt))
		}
		parts = append(parts, fmt.Sprintf("失败 %d 项(%s)", mergeFailed, strings.Join(failParts, ", ")))
	}
	if len(parts) == 0 {
		return "无待合并文件"
	}
	return strings.Join(parts, ", ")
}

// convertFlvToMp4 将单个 FLV/TS 文件无损转封装为 MP4，或在源/目标同为 MP4 时仅执行主播裁剪规则。
// 转换成功后保留原始录制时间，执行主播裁剪规则，并返回最终成片字节数。
func (s *MergeService) convertFlvToMp4(ctx context.Context, flvPath, mp4Path string, onProgress ProgressFunc, opLog *OpLogger) (int64, bool) {
	// 转换前读取原始录制时间戳与修改时间
	var flvMtime time.Time
	if info, err := os.Stat(flvPath); err == nil {
		flvMtime = info.ModTime()
	}

	// 跳过正在被录制软件锁定的文件
	if isFileBeingWritten(ctx, flvPath, stabilityProbeInterval(flvMtime)) {
		onProgress(fmt.Sprintf("⚠ %s 被占用，跳过", filepath.Base(flvPath)))
		return 0, false
	}

	streamerName := filepath.Base(filepath.Dir(mp4Path))

	// 若输入输出为同一 MP4 文件，则属于单分片 MP4 裁剪任务，跳过转封装与删源步骤
	if flvPath == mp4Path {
		trimmed := s.applyStreamerTrim(ctx, streamerName, mp4Path, flvMtime, onProgress, opLog)
		var outputSize int64
		if info, err := os.Stat(mp4Path); err == nil {
			outputSize = info.Size()
		}
		return outputSize, trimmed
	}

	// 检查磁盘空间
	disk, diskErr := utils.GetDiskUsage(filepath.Dir(flvPath))
	if diskErr == nil && disk.Free < minConvertFreeBytes {
		onProgress(fmt.Sprintf("⚠ 磁盘空间不足（仅剩 %.1f GB / 需要 %.1f GB），跳过转换", float64(disk.Free)/oneGB, float64(minConvertFreeBytes)/oneGB))
		return 0, false
	}

	if err := ffmpeg.ConvertViaTS(ctx, flvPath, mp4Path); err != nil {
		opLog.Log(fmt.Sprintf("❌ 格式转封装失败: %v，保留原始文件", err))
		onProgress(fmt.Sprintf("❌ 转换失败，保留 %s", filepath.Base(flvPath)))
		return 0, false
	}

	if err := ffmpeg.ValidateOutput(ctx, mp4Path); err != nil {
		opLog.Log(fmt.Sprintf("❌ MP4 输出校验失败: %v，保留原始文件", err))
		onProgress(fmt.Sprintf("❌ MP4 输出校验失败，保留 %s", filepath.Base(flvPath)))
		utils.SafeUnlink(mp4Path) // 清理校验失败的部分输出
		return 0, false
	}

	// 保留原始录制时间，防止因 mtime 变化导致合并分组错误
	if !flvMtime.IsZero() {
		if err := os.Chtimes(mp4Path, flvMtime, flvMtime); err != nil {
			opLog.Log(fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(mp4Path), err))
		}
	}

	// 执行可能配置的主播视频裁剪规则
	s.applyStreamerTrim(ctx, streamerName, mp4Path, flvMtime, onProgress, opLog)

	// 在裁剪完成且删除源分片前采集最终成片的真实磁盘大小
	var outputSize int64
	if info, err := os.Stat(mp4Path); err == nil {
		outputSize = info.Size()
	}

	// 删除原始文件（带重试）
	if err := utils.SafeUnlink(flvPath); err != nil {
		opLog.Log(fmt.Sprintf("⚠ 删除原始文件失败: %v", err))
	}
	return outputSize, true
}

// concatReencode 使用 ffmpeg concat filter 重编码合并文件。
// 作为 stream-copy 失败时的 fallback（编解码器不兼容、头部损坏等情况）。
func (s *MergeService) concatReencode(ctx context.Context, files []string, folder, outputPath string, onProgress ProgressFunc, opLog *OpLogger) (int64, bool) {
	return s.concatReencodeWithDims(ctx, files, folder, outputPath, nil, nil, onProgress, opLog)
}

func (s *MergeService) concatReencodeWithDims(ctx context.Context, files []string, folder, outputPath string, widths, heights []int, onProgress ProgressFunc, opLog *OpLogger) (int64, bool) {
	// 检查输入文件总大小 — 大文件在低配硬件上跳过重编码
	var totalSize int64
	var latestSrcMtime time.Time
	for _, f := range files {
		if fi, err := os.Stat(filepath.Join(folder, f)); err == nil {
			totalSize += fi.Size()
			if fi.ModTime().After(latestSrcMtime) {
				latestSrcMtime = fi.ModTime()
			}
		}
	}
	if totalSize > ffmpeg.MaxReencodeSize {
		onProgress(fmt.Sprintf("⚠ 文件过大 (%s)，跳过重编码", utils.FormatSize(totalSize)))
		return 0, false
	}

	if err := ffmpeg.ReencodeWithDimensions(ctx, files, folder, outputPath, widths, heights, onProgress); err != nil {
		opLog.Log(fmt.Sprintf("❌ 重编码失败: %v", err))
		return 0, false
	}

	return s.finalizeMergedOutput(ctx, folder, files, outputPath, totalSize, latestSrcMtime, true, onProgress, opLog)
}

// finalizeMergedOutput 统一执行合并成片输出后的收尾流程：
// 恢复时间戳 -> 执行主播裁剪规则 -> 统计成片体积 -> 对齐合并弹幕与封面 -> 删除源分片与伴生文件。
func (s *MergeService) finalizeMergedOutput(ctx context.Context, folder string, files []string, outputPath string, fallbackSize int64, latestSrcMtime time.Time, reencoded bool, onProgress ProgressFunc, opLog *OpLogger) (int64, bool) {
	if !latestSrcMtime.IsZero() {
		if err := os.Chtimes(outputPath, latestSrcMtime, latestSrcMtime); err != nil {
			opLog.Log(fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(outputPath), err))
		}
	}

	// 先基于原分片时长合并/对齐伴生 XML 弹幕和封面，再执行视频裁剪（裁剪时将同步平移伴生 XML 时间轴）
	alignCompanionDanmaku(folder, files, outputPath)

	streamerName := filepath.Base(folder)
	s.applyStreamerTrim(ctx, streamerName, outputPath, latestSrcMtime, onProgress, opLog)

	var outputSize int64
	if info, err := os.Stat(outputPath); err != nil {
		outputSize = fallbackSize
		if reencoded {
			opLog.Log("✅ 重编码完成 (大小未知)")
		} else {
			onProgress(fmt.Sprintf("✅ → %s", filepath.Base(outputPath)))
		}
	} else {
		outputSize = info.Size()
		if reencoded {
			opLog.Log(fmt.Sprintf("✅ 重编码完成: %s", utils.FormatSize(outputSize)))
		}
		onProgress(fmt.Sprintf("✅ → %s (%s)", filepath.Base(outputPath), utils.FormatSize(outputSize)))
	}

	for _, f := range files {
		fullPath := filepath.Join(folder, f)
		if err := utils.SafeUnlink(fullPath); err != nil {
			opLog.Log(fmt.Sprintf("⚠ 删除源文件失败: %s (%v)", f, err))
		}
		CleanCompanionFiles(fullPath)
	}
	return outputSize, true
}

// checkFileAvailability 检查批次中的所有文件是否可访问且未被锁定。
// 文件不存在时，检查是否有对应的合并版（说明已被之前的合并处理）。
func checkFileAvailability(ctx context.Context, folder string, files []string) error {
	for _, f := range files {
		// 路径穿越防护：确保文件路径在预期目录内
		fullPath := filepath.Join(folder, f)
		cleanPath := filepath.Clean(fullPath)
		cleanFolder := filepath.Clean(folder)
		if !strings.HasPrefix(cleanPath, cleanFolder+string(os.PathSeparator)) {
			return fmt.Errorf("路径穿越检测: %s", f)
		}
		info, err := os.Stat(fullPath)
		if err != nil {
			// 文件不存在，检查是否已有合并版输出（兼容原扩展名与转封装后的 .mp4）
			mergedName := utils.MakeOutputName(f)
			mergedPath := filepath.Join(folder, mergedName)
			if _, merr := os.Stat(mergedPath); merr == nil {
				return fmt.Errorf("文件已合并: %s → %s", f, mergedName)
			}
			if !strings.EqualFold(filepath.Ext(mergedName), ".mp4") {
				mp4MergedName := utils.MakeMP4Name(mergedName)
				if _, merr := os.Stat(filepath.Join(folder, mp4MergedName)); merr == nil {
					return fmt.Errorf("文件已合并: %s → %s", f, mp4MergedName)
				}
			}
			return fmt.Errorf("文件不存在: %s", f)
		}
		if info.IsDir() {
			return fmt.Errorf("路径是目录而非文件: %s", f)
		}
		if isFileBeingWritten(ctx, fullPath, stabilityProbeInterval(info.ModTime())) {
			return fmt.Errorf("文件被占用: %s", f)
		}
	}
	return nil
}

// doMerge 执行多文件合并的完整流程：优先直接 concat demuxer 流复制拼接，
// 失败时自动降级为 TS 中转拼接（支持 H.264/HEVC），再失败则回退到重编码。
// 输出经校验、fsync、时间戳恢复、裁剪和伴生弹幕合并后删除原始分片。
// 调用方须保证 files 按时间升序排列（files[0] 为最早的文件）。
func (s *MergeService) doMerge(ctx context.Context, files []string, folder string, onProgress ProgressFunc, opLog *OpLogger) (int64, bool) {
	if len(files) < 2 {
		return 0, false
	}
	if ctx.Err() != nil {
		return 0, false
	}

	// 检查所有文件是否可访问且未被锁定
	if err := checkFileAvailability(ctx, folder, files); err != nil {
		onProgress(fmt.Sprintf("⚠ %v", err))
		return 0, false
	}

	// 输出文件名统一使用 .mp4 扩展名进行去重探测，避免同开播前缀多批次互相覆盖
	output := resolveUniqueMergeOutputName(folder, files[0])
	if output == "" {
		opLog.Log(fmt.Sprintf("❌ 无法找到可用的输出文件名（已尝试 %d 个后缀）", maxDedupAttempts))
		return 0, false
	}
	defaultOutput := utils.MakeMP4Name(utils.MakeOutputName(files[0]))
	if output != defaultOutput {
		opLog.Log(fmt.Sprintf("⚠ 输出文件已存在，自动重命名为: %s", output))
	}
	outputPath := filepath.Join(folder, output)

	// 检查磁盘空间
	disk, diskErr := utils.GetDiskUsage(folder)
	if diskErr == nil && disk.Free < minMergeFreeBytes {
		opLog.Log(fmt.Sprintf("❌ 磁盘空间不足（仅剩 %.1f GB / 需要 %.1f GB），跳过", float64(disk.Free)/oneGB, float64(minMergeFreeBytes)/oneGB))
		return 0, false
	}

	// 记录最新源文件的修改时间和总大小，用于时间戳校正和进度显示
	var latestSrcMtime time.Time
	var totalFileSize int64
	sameExt := true
	allTS := true
	firstExt := strings.ToLower(filepath.Ext(files[0]))
	directPaths := make([]string, len(files))
	for i, f := range files {
		srcPath := filepath.Join(folder, f)
		if info, err := os.Stat(srcPath); err == nil {
			totalFileSize += info.Size()
			if info.ModTime().After(latestSrcMtime) {
				latestSrcMtime = info.ModTime()
			}
		}
		ext := strings.ToLower(filepath.Ext(f))
		if ext != firstExt {
			sameExt = false
		}
		if ext != ".ts" {
			allTS = false
		}
		directPaths[i] = filepath.ToSlash(srcPath)
	}

	// 第 0 步：检测跨分片分辨率是否一致；若主播中途改分辨率/横竖屏，直接跳过流复制走归一化重编码，并缓存尺寸供重编码复用
	widths := make([]int, len(files))
	heights := make([]int, len(files))
	reencodeFallback := func() (int64, bool) {
		return s.concatReencodeWithDims(ctx, files, folder, outputPath, widths, heights, onProgress, opLog)
	}
	refW, refH := 0, 0
	var mismatchDesc string
	for i, f := range files {
		w, h := ffmpeg.ProbeVideoDimensions(ctx, filepath.Join(folder, f))
		widths[i], heights[i] = w, h
		if w > 0 && h > 0 {
			if refW == 0 {
				refW, refH = w, h
			} else if (w != refW || h != refH) && mismatchDesc == "" {
				mismatchDesc = fmt.Sprintf("%dx%d → %dx%d", refW, refH, w, h)
			}
		}
	}
	if mismatchDesc != "" {
		opLog.Log(fmt.Sprintf("⚠ 检测到分片分辨率不一致（%s），直接启用归一化重编码合并", mismatchDesc))
		onProgress(fmt.Sprintf("⚠ 检测到分片分辨率变化（%s），启用归一化重编码合并…", mismatchDesc))
		return reencodeFallback()
	}

	// 第 1 级：同容器分片优先直接 concat demuxer 流复制拼接（零 TS 中间文件、兼容 H.264/HEVC）
	mergedByCopy := false
	if sameExt && firstExt != "" {
		onProgress("⚙ 直接流复制合并分片…")
		if err := ffmpeg.ConcatTS(ctx, directPaths, outputPath, onProgress); err == nil {
			if vErr := ffmpeg.ValidateOutput(ctx, outputPath); vErr == nil {
				mergedByCopy = true
			} else {
				opLog.Log(fmt.Sprintf("⚠ 直接合并输出校验失败: %v，尝试 TS 中转拼接", vErr))
				utils.SafeUnlink(outputPath)
			}
		} else {
			opLog.Log(fmt.Sprintf("⚠ 直接流复制合并失败: %v，尝试 TS 中转拼接", err))
			utils.SafeUnlink(outputPath)
		}
	}

	if !mergedByCopy && allTS {
		onProgress("⚠ TS 直接拼接失败，切换重编码…")
		return reencodeFallback()
	}

	// 第 2 级：降级通过 TS 中转容器拼接（按视频编码动态选择 BSF）
	if !mergedByCopy {
		tmpDir, err := os.MkdirTemp(folder, ".merge_tmp_")
		if err != nil {
			onProgress(fmt.Sprintf("❌ 创建临时目录失败: %v", err))
			return 0, false
		}
		defer os.RemoveAll(tmpDir)

		var tsFiles []string
		for _, f := range files {
			if strings.EqualFold(filepath.Ext(f), ".ts") {
				tsFiles = append(tsFiles, filepath.ToSlash(filepath.Join(folder, f)))
				continue
			}
			srcPath := filepath.Join(folder, f)
			tsPath := filepath.Join(tmpDir, fmt.Sprintf("seg_%d.ts", len(tsFiles)))
			onProgress(fmt.Sprintf("🔄 转换 %s → TS…", filepath.Base(f)))
			codec := ffmpeg.ProbeVideoCodec(ctx, srcPath)
			tsArgs := []string{"-nostdin", "-i", filepath.ToSlash(srcPath), "-c", "copy"}
			tsArgs = append(tsArgs, ffmpeg.VideoBSFForCodec(codec)...)
			tsArgs = append(tsArgs, "-y", "-loglevel", "error", tsPath)
			if err := ffmpeg.Run(ctx, ffmpeg.Options{Args: tsArgs}); err != nil {
				opLog.Log(fmt.Sprintf("❌ 转 TS 失败 %s: %v", filepath.Base(f), err))
				onProgress(fmt.Sprintf("❌ 转换失败: %s，切换重编码…", filepath.Base(f)))
				return reencodeFallback()
			}
			tsFiles = append(tsFiles, filepath.ToSlash(tsPath))
		}

		onProgress("⚙ 拼接 TS 文件…")
		if err := ffmpeg.ConcatTS(ctx, tsFiles, outputPath, onProgress); err != nil {
			opLog.Log(fmt.Sprintf("⚠ TS 拼接失败: %v，切换重编码", err))
			onProgress("⚠ 拼接失败，切换重编码…")
			utils.SafeUnlink(outputPath)
			return reencodeFallback()
		}

		if err := ffmpeg.ValidateOutput(ctx, outputPath); err != nil {
			opLog.Log(fmt.Sprintf("⚠ 输出校验失败: %v，切换重编码", err))
			utils.SafeUnlink(outputPath)
			return reencodeFallback()
		}
	}

	// fsync — 确保数据刷入持久存储后再删除原始文件
	fd, openErr := os.OpenFile(outputPath, os.O_RDONLY, 0)
	if openErr != nil {
		opLog.Log(fmt.Sprintf("❌ 无法打开输出文件进行 fsync: %v，尝试重编码", openErr))
		return reencodeFallback()
	}
	syncErr := fd.Sync()
	_ = fd.Close()
	if syncErr != nil {
		opLog.Log(fmt.Sprintf("❌ fsync 失败: %v，尝试重编码", syncErr))
		utils.SafeUnlink(outputPath) // 删除不可靠输出
		return reencodeFallback()
	}

	return s.finalizeMergedOutput(ctx, folder, files, outputPath, totalFileSize, latestSrcMtime, false, onProgress, opLog)
}

// resolveUniqueMergeOutputName 根据首分片文件名计算唯一的合并输出 .mp4 文件名，若重名自动附加 -2、-3 后缀。
func resolveUniqueMergeOutputName(folder, firstFile string) string {
	output := utils.MakeMP4Name(utils.MakeOutputName(firstFile))
	if _, err := os.Stat(filepath.Join(folder, output)); os.IsNotExist(err) {
		return output
	}
	ext := filepath.Ext(output)
	stem := strings.TrimSuffix(output, ext)
	for i := 2; i <= maxDedupAttempts; i++ {
		candidate := fmt.Sprintf("%s-%d%s", stem, i, ext)
		if _, err := os.Stat(filepath.Join(folder, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}
	return ""
}

// ManualMerge 手动合并指定主播的指定文件列表。
// 获取主播锁后校验文件合法性，按文件名时间排序后调用 doMerge 执行合并。
// ctx 用于控制合并过程的取消。
func (s *MergeService) ManualMerge(ctx context.Context, streamer string, files []string, onProgress ProgressFunc) (string, error) {
	name := streamer
	locked, sl := s.tryLockStreamer(name)
	if !locked {
		return "", fmt.Errorf("%s 合并任务正在执行中，请稍后再试", name)
	}
	defer s.unlockStreamer(sl)

	cfg := s.config.Snapshot()
	opLog, err := NewOpLogger(filepath.Join(cfg.LogDir, "merge_log"), "merge")
	if err != nil {
		opLog = nil // 降级为 nil，不阻断操作
	}
	defer opLog.Close()

	folder := filepath.Join(cfg.TargetDir, streamer)
	if cfg.IsBackupWindow() {
		opLog.Discard()
		return "", fmt.Errorf("当前处于静默时段（%d:%02d-%d:%02d），合并暂停", cfg.BackupStartHour, cfg.BackupStartMinute, cfg.BackupEndHour, cfg.BackupEndMinute)
	}
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		opLog.Discard()
		return "", fmt.Errorf("目录不存在")
	}

	if onProgress == nil {
		onProgress = func(string) {}
	}
	onProgress = opLog.ProgressFunc(onProgress)

	var validFiles []string
	var totalInputBytes int64
	var skipped []string
	for _, f := range files {
		if !utils.ValidateFilename(f) {
			opLog.Discard()
			return "", fmt.Errorf("非法文件名: %s", f)
		}
		path := filepath.Join(folder, f)
		if info, err := os.Stat(path); err == nil {
			validFiles = append(validFiles, f)
			totalInputBytes += info.Size()
		} else {
			skipped = append(skipped, f)
		}
	}

	if len(validFiles) < 2 {
		opLog.Discard()
		if len(skipped) > 0 {
			return "", fmt.Errorf("有效文件不足2个（%d 个文件已不存在，请刷新文件列表）", len(skipped))
		}
		return "", fmt.Errorf("有效文件不足2个")
	}

	// 磁盘空间前置检查：确保可用空间满足直接合并（2 倍体积 + 1GB 余量）
	disk, diskErr := utils.GetDiskUsage(folder)
	if diskErr == nil {
		needed := (totalInputBytes * 2) + mergeHeadroomBytes
		if int64(disk.Free) < needed {
			opLog.Discard()
			return "", fmt.Errorf("磁盘空间不足：需要 %.1f GB 可用空间，当前仅 %.1f GB",
				float64(needed)/oneGB, float64(disk.Free)/oneGB)
		}
	}

	// 按文件名中的日期时间升序排序（确保合并后时间线正确）
	SortByFilename(validFiles)

	if len(skipped) > 0 {
		onProgress(fmt.Sprintf("⚠ %d 个文件已不存在，使用剩余 %d 个文件继续", len(skipped), len(validFiles)))
	}

	onProgress(fmt.Sprintf("⚙ 手动合并 %d 个文件 (%s)", len(validFiles), utils.FormatSize(totalInputBytes)))
	hasOriginal := false
	for _, f := range validFiles {
		if !utils.IsMergedFile(f) {
			hasOriginal = true
			break
		}
	}
	if !hasOriginal {
		opLog.Discard()
		return "", fmt.Errorf("所选文件全部是合并版，请至少选择一个原始文件")
	}

	start := time.Now()
	if outBytes, ok := s.doMerge(ctx, validFiles, folder, onProgress, opLog); ok {
		duration := time.Since(start).Seconds()
		effectiveBytes := outBytes
		if effectiveBytes <= 0 {
			effectiveBytes = totalInputBytes
		}
		s.history.AddWithStats("merge", streamer, "success", len(validFiles), 0, effectiveBytes, duration, fmt.Sprintf("手动合并 %d 个文件 (%s)", len(validFiles), utils.FormatSize(effectiveBytes)), opLog.LogID())
		onProgress(fmt.Sprintf("✅ 手动合并完成: %d 个文件 (%s)", len(validFiles), utils.FormatSize(effectiveBytes)))
		return opLog.LogID(), nil
	}

	return opLog.LogID(), fmt.Errorf("合并失败")
}

// applyStreamerTrim 检查并执行针对指定主播的视频时段裁剪规则。
// 若成功裁剪，将原子替换原视频并恢复时间戳；若失败或无需截取，安全保留原视频。
func (s *MergeService) applyStreamerTrim(ctx context.Context, streamerName, videoPath string, originalMtime time.Time, onProgress ProgressFunc, opLog *OpLogger) bool {
	rule, ok := s.config.GetStreamerRule(streamerName)
	if !ok || !rule.TrimEnabled {
		return false
	}

	start := strings.TrimSpace(rule.TrimStart)
	if start == "" {
		start = "00:00:00"
	}
	end := strings.TrimSpace(rule.TrimEnd)

	// 时长预检：检查视频总时长是否小于等于起始裁剪时间，或从头截取时是否小于等于截取上限
	startSec, startErr := ffmpeg.ParseTimeSeconds(start)
	endSec, endErr := ffmpeg.ParseTimeSeconds(end)
	needDurationCheck := (startErr == nil && startSec > 0) || (end != "" && endErr == nil && endSec > 0)
	if needDurationCheck {
		if dur, durErr := ffmpeg.ProbeDuration(ctx, videoPath); durErr == nil && dur > 0 {
			if startErr == nil && startSec > 0 && dur <= startSec {
				if opLog != nil {
					opLog.Log(fmt.Sprintf("[%s] 视频时长 (%.1f秒) ≤ 裁剪起始时间 (%s)，保留原视频", streamerName, dur, start))
				}
				onProgress(fmt.Sprintf("[%s] ℹ 视频时长短于裁剪起始时间，无需裁剪，保留完整视频", streamerName))
				return true
			}
			if (start == "0" || start == "00:00:00" || (startErr == nil && startSec == 0)) && end != "" && endErr == nil && endSec > 0 && dur <= endSec {
				if opLog != nil {
					opLog.Log(fmt.Sprintf("[%s] 视频时长 (%.1f秒) ≤ 截取上限 (%s)，保留原视频", streamerName, dur, end))
				}
				onProgress(fmt.Sprintf("[%s] ℹ 视频时长短于设定时段，无需裁剪，保留完整视频", streamerName))
				return true
			}
		}
	}

	trimDesc := fmt.Sprintf("[%s - %s]", start, func() string {
		if end == "" {
			return "结束"
		}
		return end
	}())
	onProgress(fmt.Sprintf("[%s] ✂ 正在按规则裁剪: %s", streamerName, trimDesc))

	trimmedTmp := videoPath + ".trim.mp4"
	defer utils.SafeUnlink(trimmedTmp)

	err := ffmpeg.TrimVideo(ctx, videoPath, trimmedTmp, start, end)
	if err != nil {
		if opLog != nil {
			opLog.Log(fmt.Sprintf("⚠ [%s] 视频裁剪未生效或失败: %v，已安全保留原完整视频", streamerName, err))
		}
		onProgress(fmt.Sprintf("[%s] ⚠ 裁剪未生效，已保留原完整视频", streamerName))
		return false
	}

	// 裁剪成功：替换原文件
	if err := fsutil.ReplaceFile(trimmedTmp, videoPath); err != nil {
		if opLog != nil {
			opLog.Log(fmt.Sprintf("❌ [%s] 替换裁剪文件失败: %v", streamerName, err))
		}
		return false
	}

	// 同步裁剪/平移同名 XML 弹幕时间轴，使成片与弹幕保持一致
	trimStartSec := 0.0
	if startErr == nil && startSec > 0 {
		trimStartSec = startSec
	}
	trimEndSec := 0.0
	if end != "" && endErr == nil && endSec > 0 {
		trimEndSec = endSec
	}
	trimCompanionDanmaku(videoPath, trimStartSec, trimEndSec)

	// 恢复时间戳
	if !originalMtime.IsZero() {
		_ = os.Chtimes(videoPath, originalMtime, originalMtime)
	}

	if info, err := os.Stat(videoPath); err == nil {
		s.checkedMP4s.Store(videoPath, mp4CheckFingerprint{
			size:    info.Size(),
			mtime:   info.ModTime().UnixNano(),
			ruleSig: ruleFingerprint(rule),
		})
		onProgress(fmt.Sprintf("[%s] ✂ ✅ 裁剪完成（%s，大小: %s）", streamerName, trimDesc, utils.FormatSize(info.Size())))
	} else {
		onProgress(fmt.Sprintf("[%s] ✂ ✅ 裁剪完成（%s）", streamerName, trimDesc))
	}
	return true
}

// SortByFilename 按文件名中的日期时间升序排列文件列表。
// 预解析所有文件名后再排序，避免比较器中重复调用 ParseFilename（含正则）。
// 可解析的文件按日期排序，同日期按文件名字典序（001 < 002）；
// 不可解析的文件（零值 time）排在最前，按字典序排列。
func SortByFilename(files []string) {
	type item struct {
		name string
		dt   time.Time
	}
	items := make([]item, len(files))
	for i, f := range files {
		_, dt, ok := utils.ParseFilename(f)
		if ok {
			items[i] = item{name: f, dt: dt}
		} else {
			items[i] = item{name: f}
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].dt.Equal(items[j].dt) {
			return items[i].name < items[j].name
		}
		return items[i].dt.Before(items[j].dt)
	})
	for i, it := range items {
		files[i] = it.name
	}
}

// CleanupTempFiles 清理录制目录中的残留临时文件。
// 包括：concat 清单文件、.mp4.tmp/.flv.tmp/.tmp.ts/.trim.mp4 临时中间文件、崩溃残留的 .merge_tmp_* 目录。
// 注意：不再误删主播目录下的独立正常 .ts 录制原片。
func (s *MergeService) CleanupTempFiles() int {
	cfg := s.config.Snapshot()
	root := cfg.TargetDir
	dirs, err := fsutil.ScanStreamerDirs(root)
	if err != nil {
		return 0
	}
	cleaned := 0
	for _, dir := range dirs {
		folder := dir.Path
		fileEntries := dir.Files

		for _, fe := range fileEntries {
			name := fe.Name()

			// 清理崩溃残留的 .merge_tmp_* 临时目录
			if fe.IsDir() && strings.HasPrefix(name, ".merge_tmp_") {
				tmpPath := filepath.Join(folder, name)
				if err := os.RemoveAll(tmpPath); err == nil {
					cleaned++
					s.logger.Info(fmt.Sprintf("🗑 清理残留临时目录: %s", name))
				}
				continue
			}

			if fe.IsDir() {
				continue
			}

			// 清理残留的 concat 清单文件
			if strings.HasPrefix(name, ".concat_") && strings.HasSuffix(name, ".txt") {
				path := filepath.Join(folder, name)
				if err := os.Remove(path); err == nil {
					cleaned++
					s.logger.Info(fmt.Sprintf("🗑 清理临时文件: %s", name))
				}
				continue
			}

			// 清理中断写入或崩溃残留的内部临时文件（.mp4.tmp / .flv.tmp / .tmp.ts / .trim.mp4）
			if strings.HasSuffix(name, ".mp4.tmp") ||
				strings.HasSuffix(name, ".flv.tmp") ||
				strings.HasSuffix(name, ".tmp.ts") ||
				strings.HasSuffix(name, ".trim.mp4") {
				path := filepath.Join(folder, name)
				if err := utils.SafeUnlink(path); err == nil {
					cleaned++
					s.logger.Info(fmt.Sprintf("🗑 清理临时文件: %s", name))
				}
			}
		}
	}
	return cleaned
}
