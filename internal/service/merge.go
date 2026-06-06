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
	"sync/atomic"
	"time"

	"bililive-helper/internal/config"
	"bililive-helper/internal/ffmpeg"
	"bililive-helper/internal/utils"

	"go.uber.org/zap"
)

// ProgressFunc 是进度回调函数类型，用于 SSE 流式输出。
type ProgressFunc func(msg string)

const (
	minDiskFreeBytes     = 10 * 1024 * 1024 * 1024 // 10GB: 磁盘空间硬限，低于此值跳过所有操作
	minConvertFreeBytes  = 512 * 1024 * 1024        // 512MB: FLV 转换所需最小磁盘空间
	minMergeFreeBytes    = 1024 * 1024 * 1024       // 1GB: 合并操作所需最小磁盘空间
	tsMergeHeadroomBytes = 2 * 1024 * 1024 * 1024   // 2GB: TS 管线峰值空间余量
	maxDedupAttempts     = 1000                      // 文件名去重最大尝试次数
	oneGB                = 1073741824                // 1GB 字节数，用于大小格式化换算
	minValidFileSize     = 10240                     // 10KB: 视频文件最小有效大小
)

// MergeService 提供录制文件合并功能。
// 使用 per-streamer 锁防止同一主播的并发合并，支持 FLV 转 MP4、多文件 TS 拼接和重编码 fallback。
type MergeService struct {
	config        *config.Config
	logger        *zap.Logger
	history       *HistoryService
	streamerLocks sync.Map // streamer name -> *streamerLock
}

// streamerLock 用于 per-streamer 粒度的合并锁，防止同一主播并发合并。
type streamerLock struct {
	mu        sync.Mutex
	createdAt atomic.Int64 // Unix 时间戳，避免 time.Time 的 data race
}

const lockTimeout = 4 * time.Hour

// tryLockStreamer 尝试获取指定主播的合并锁。
// 如果锁被持有超过 lockTimeout（4 小时），尝试回收过期锁。
func (s *MergeService) tryLockStreamer(name string) (bool, *streamerLock) {
	now := time.Now()
	val, _ := s.streamerLocks.LoadOrStore(name, &streamerLock{})
	sl := val.(*streamerLock)

	if sl.mu.TryLock() {
		sl.createdAt.Store(now.Unix())
		return true, sl
	}

	// 安全网：锁持有超过超时时间后尝试回收。
	// CompareAndDelete 避免销毁仍在合法持有锁的 goroutine 的锁
	// — 如果 map 条目在 LoadOrStore 和 Delete 之间被替换，静默退让。
	if now.Sub(time.Unix(sl.createdAt.Load(), 0)) > lockTimeout {
		if s.streamerLocks.CompareAndDelete(name, sl) {
			s.logger.Warn("回收过期主播锁", zap.String("streamer", name))
			newVal, _ := s.streamerLocks.LoadOrStore(name, &streamerLock{})
			newSl := newVal.(*streamerLock)
			if newSl.mu.TryLock() {
				newSl.createdAt.Store(now.Unix())
				return true, newSl
			}
		}
		return false, nil
	}

	return false, nil
}

func (s *MergeService) unlockStreamer(sl *streamerLock) {
	if sl != nil {
		sl.mu.Unlock()
	}
}

// NewMergeService 创建合并服务实例。
func NewMergeService(config *config.Config, logger *zap.Logger, history *HistoryService) *MergeService {
	return &MergeService{config: config, logger: logger, history: history}
}

// MergeResult 保存合并操作的结果。
type MergeResult struct {
	Done    int
	TotalGB float64
}

// checkDiskSpaceForMerge 检查合并任务是否有足够的磁盘空间。
// TS 管线峰值空间：源文件 + TS 中间文件 + 输出文件 ≈ 3 倍源文件 + 2GB 余量。
func (s *MergeService) checkDiskSpaceForMerge(tasks []mergeTask, targetDir string) error {
	var totalSourceBytes int64
	for _, t := range tasks {
		totalSourceBytes += int64(t.SizeGB * oneGB)
	}
	disk, err := utils.GetDiskUsage(targetDir)
	if err != nil {
		return fmt.Errorf("获取磁盘信息失败: %w", err)
	}
	// TS 管线峰值空间：源文件 + TS 中间文件 + 输出文件 ≈ 3 倍源文件 + 2GB 余量
	needed := (totalSourceBytes * 3) + tsMergeHeadroomBytes
	if int64(disk.Free) < needed {
		return fmt.Errorf("磁盘空间不足：需要 %.1f GB 可用以应对 TS 转换峰值，当前仅 %.1f GB",
			float64(needed)/oneGB, float64(disk.Free)/oneGB)
	}
	return nil
}

// classifyMergeFailure 检查合并失败的批次文件和输出，判断最可能的失败原因。
func classifyMergeFailure(folder, firstFile string) string {
	output := utils.MakeOutputName(firstFile)
	outputPath := filepath.Join(folder, output)

	// 检查输出文件是否存在（可能被校验探针删除）
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		// 检查源文件是否有问题
		firstPath := filepath.Join(folder, firstFile)
		if info, err := os.Stat(firstPath); err != nil {
			return "源文件不存在"
		} else if info.Size() < minValidFileSize {
			return fmt.Sprintf("源文件过小(%s)", utils.FormatSize(info.Size()))
		}
		return "ffmpeg输出校验失败"
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
	cfg := s.config.Snapshot()
	root := cfg.TargetDir

	opLog, err := NewOpLogger(filepath.Join(cfg.LogDir, "merge_log"), "merge")
	if err != nil {
		opLog = nil // 降级为 nil，不阻断操作
	}
	defer opLog.Close()

	if cfg.IsBackupWindow() {
		return nil, opLog.LogID(), fmt.Errorf("当前处于静默时段（%d:%02d-%d:%02d），合并暂停", cfg.BackupStartHour, cfg.BackupStartMinute, cfg.BackupEndHour, cfg.BackupEndMinute)
	}
	if onProgress == nil {
		onProgress = func(string) {}
	}
	onProgress = opLog.ProgressFunc(onProgress)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, opLog.LogID(), fmt.Errorf("路径不存在: %s", root)
	}

	tag := "[全局]"
	if streamer != "" {
		tag = fmt.Sprintf("[%s]", streamer)
	}
	onProgress(fmt.Sprintf("▶ 开始 %s 合并", tag))
	onProgress(fmt.Sprintf("⚙ 扫描 %s ...", root))

	tasks, convertTasks := s.scanTasks(ctx, root, streamer, cfg)
	if len(tasks) == 0 && len(convertTasks) == 0 {
		s.history.Add("merge", streamer, "success", "扫描完成，无待合并文件", opLog.LogID())
		onProgress("ℹ 无待合并文件")
		return &MergeResult{}, opLog.LogID(), nil
	}

	if len(tasks) > 0 {
		if err := s.checkDiskSpaceForMerge(tasks, root); err != nil {
			onProgress(fmt.Sprintf("❌ %s", err.Error()))
			return nil, opLog.LogID(), err
		}
	}

	// 磁盘空间硬性检查 — 可用空间低于 10GB 时跳过所有操作
	disk, diskErr := utils.GetDiskUsage(root)
	if diskErr == nil && disk.Free < minDiskFreeBytes { // < 10GB free
		opLog.Log(fmt.Sprintf("❌ 磁盘空间不足（仅剩 %.1f GB），跳过所有操作", float64(disk.Free)/oneGB))
		onProgress(fmt.Sprintf("❌ 磁盘空间不足（仅剩 %.1f GB / 需要 10 GB），跳过", float64(disk.Free)/oneGB))
		s.history.Add("merge", streamer, "fail", fmt.Sprintf("磁盘空间不足: %.1f GB（使用率 %.1f%%）", float64(disk.Free)/oneGB, disk.UsedPct), opLog.LogID())
		return &MergeResult{}, opLog.LogID(), nil
	}

	done := 0
	totalGB := 0.0
	convertDone := 0
	mergeDone := 0
	mergeFailed := 0
	failedReasons := make(map[string]int)
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
		onProgress(fmt.Sprintf("[%s] 🔄 FLV→MP4: %s → %s", streamerName, filepath.Base(ct.FlvPath), filepath.Base(ct.Mp4Path)))
		if s.convertFlvToMp4(ctx, ct.FlvPath, ct.Mp4Path, onProgress, opLog) {
			done++
			convertDone++
			totalGB += float64(flvSize) / oneGB
			if info, err := os.Stat(ct.Mp4Path); err == nil {
				onProgress(fmt.Sprintf("[%s] ✅ → %s (%s)", streamerName, filepath.Base(ct.Mp4Path), utils.FormatSize(info.Size())))
			} else {
				onProgress(fmt.Sprintf("[%s] ✅ → %s", streamerName, filepath.Base(ct.Mp4Path)))
			}
		} else {
			onProgress(fmt.Sprintf("[%s] ❌ 转换失败", streamerName))
		}
	}

	// 处理待合并任务
	for _, task := range tasks {
		if cfg.IsBackupWindow() {
			break
		}
		streamerName := filepath.Base(task.Folder)
		if streamerName != lastStreamer {
			onProgress(fmt.Sprintf("── %s ──", streamerName))
			lastStreamer = streamerName
		}
		locked, sl := s.tryLockStreamer(streamerName)
		if !locked {
			continue
		}
		onProgress(fmt.Sprintf("[%s] ⚙ 合并 %d 个文件 (%s)", streamerName, len(task.Files), utils.FormatSize(int64(task.SizeGB*oneGB))))
		if s.doMerge(ctx, task.Files, task.Folder, onProgress, opLog) {
			done++
			mergeDone++
			totalGB += task.SizeGB
			outputName := utils.MakeOutputName(task.Files[0])
			if strings.HasSuffix(outputName, ".flv") {
				outputName = strings.TrimSuffix(outputName, ".flv") + ".mp4"
			}
			onProgress(fmt.Sprintf("[%s] ✅ → %s", streamerName, outputName))
		} else {
			mergeFailed++
			reason := classifyMergeFailure(task.Folder, task.Files[0])
			failedReasons[reason]++
			onProgress(fmt.Sprintf("[%s] ❌ 失败: %s", streamerName, reason))
		}
		s.unlockStreamer(sl)
	}

	// 合并结果汇总
	totalScanned := 0
	if dirs, err := os.ReadDir(root); err == nil {
		for _, d := range dirs {
			if d.IsDir() {
				totalScanned++
			}
		}
	}
	if streamer != "" {
		totalScanned = 1
	}
	duration := time.Since(start).Seconds()

	if done > 0 {
		var parts []string
		if convertDone > 0 {
			parts = append(parts, fmt.Sprintf("转换 %d 个FLV", convertDone))
		}
		if mergeDone > 0 {
			parts = append(parts, fmt.Sprintf("合并 %d 场次 (%.1f GB)", mergeDone, totalGB))
		}
		detail := strings.Join(parts, ", ")
		if detail == "" {
			detail = fmt.Sprintf("完成 %d 项", done)
		}
		msg := fmt.Sprintf("✅ 完成: 扫描 %d 个主播, %s", totalScanned, detail)
		s.history.AddWithStats("merge", streamer, "success", done, 0, int64(totalGB*oneGB), duration, detail, opLog.LogID())
		onProgress(msg)
	} else if mergeFailed > 0 {
		var parts []string
		for reason, cnt := range failedReasons {
			parts = append(parts, fmt.Sprintf("%s x %d", reason, cnt))
		}
		msg := fmt.Sprintf("❌ 全部失败: 扫描 %d 个主播, %s", totalScanned, strings.Join(parts, ", "))
		s.history.Add("merge", streamer, "fail", fmt.Sprintf("合并失败: %s", strings.Join(parts, ", ")), opLog.LogID())
		onProgress(msg)
	} else {
		msg := fmt.Sprintf("ℹ 完成: 扫描 %d 个主播, 无需合并", totalScanned)
		s.history.Add("merge", streamer, "success", fmt.Sprintf("扫描 %d 个主播，无需合并", totalScanned), opLog.LogID())
		onProgress(msg)
	}
	return &MergeResult{Done: done, TotalGB: totalGB}, opLog.LogID(), nil
}

// convertFlvToMp4 将单个 FLV 文件转换为 MP4（通过 TS 中间格式）。
// 转换成功后保留原始录制时间，删除原始 FLV 文件。
func (s *MergeService) convertFlvToMp4(ctx context.Context, flvPath, mp4Path string, onProgress ProgressFunc, opLog *OpLogger) bool {
	// 跳过正在被录制软件锁定的文件
	if isFileBeingWritten(flvPath, 1*time.Second) {
		onProgress(fmt.Sprintf("⚠ %s 被占用，跳过", filepath.Base(flvPath)))
		return false
	}

	// 检查磁盘空间
	disk, diskErr := utils.GetDiskUsage(filepath.Dir(flvPath))
	if diskErr == nil && disk.Free < minConvertFreeBytes {
		onProgress(fmt.Sprintf("⚠ 磁盘空间不足（仅剩 %.1f GB / 需要 %.1f GB），跳过转换", float64(disk.Free)/oneGB, float64(minConvertFreeBytes)/oneGB))
		return false
	}

	// 转换前保留原始录制时间戳
	var flvMtime time.Time
	if info, err := os.Stat(flvPath); err == nil {
		flvMtime = info.ModTime()
	}

	if err := ffmpeg.ConvertViaTS(ctx, flvPath, mp4Path); err != nil {
		opLog.Log(fmt.Sprintf("❌ FLV→MP4 失败: %v，保留原始文件", err))
		onProgress(fmt.Sprintf("❌ 转换失败，保留 %s", filepath.Base(flvPath)))
		return false
	}

	if err := ffmpeg.ValidateOutput(ctx, mp4Path); err != nil {
		opLog.Log(fmt.Sprintf("❌ MP4 输出校验失败: %v，保留原始文件", err))
		return false
	}

	// 保留原始录制时间，防止因 mtime 变化导致合并分组错误
	if !flvMtime.IsZero() {
		if err := os.Chtimes(mp4Path, flvMtime, flvMtime); err != nil {
			opLog.Log(fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(mp4Path), err))
		}
	}

	// 删除原始文件（带重试）
	if err := utils.SafeUnlink(flvPath); err != nil {
		opLog.Log(fmt.Sprintf("⚠ 删除原始文件失败: %v", err))
	}

	if info, err := os.Stat(mp4Path); err != nil {
		opLog.Log(fmt.Sprintf("✅ FLV→MP4 完成: %s (大小未知)", filepath.Base(mp4Path)))
	} else {
		opLog.Log(fmt.Sprintf("✅ FLV→MP4 完成: %s (%s)", filepath.Base(mp4Path), utils.FormatSize(info.Size())))
	}
	onProgress(fmt.Sprintf("✅ 完成: %s", filepath.Base(mp4Path)))
	return true
}

// concatReencode 使用 ffmpeg concat filter 重编码合并文件。
// 作为 stream-copy 失败时的 fallback（编解码器不兼容、头部损坏等情况）。
func (s *MergeService) concatReencode(ctx context.Context, files []string, folder, outputPath string, onProgress ProgressFunc, opLog *OpLogger) bool {
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
		return false
	}

	if err := ffmpeg.Reencode(ctx, files, folder, outputPath, onProgress); err != nil {
		opLog.Log(fmt.Sprintf("❌ 重编码失败: %v", err))
		return false
	}

	// 保留原始录制时间
	if !latestSrcMtime.IsZero() {
		if err := os.Chtimes(outputPath, latestSrcMtime, latestSrcMtime); err != nil {
			opLog.Log(fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(outputPath), err))
		}
	}

	if info, err := os.Stat(outputPath); err != nil {
		opLog.Log("✅ 重编码完成 (大小未知)")
	} else {
		opLog.Log(fmt.Sprintf("✅ 重编码完成: %s", utils.FormatSize(info.Size())))
	}
	return true
}

// checkFileAvailability 检查批次中的所有文件是否可访问且未被锁定。
// 文件不存在时，检查是否有对应的合并版（说明已被之前的合并处理）。
func checkFileAvailability(folder string, files []string) error {
	for _, f := range files {
		path := filepath.Join(folder, f)
		if _, err := os.Stat(path); err != nil {
			// 文件不存在，检查是否已有合并版输出
			mergedName := utils.MakeOutputName(f)
			mergedPath := filepath.Join(folder, mergedName)
			if _, merr := os.Stat(mergedPath); merr == nil {
				return fmt.Errorf("文件已合并: %s → %s", f, mergedName)
			}
			return fmt.Errorf("文件不存在: %s", f)
		}
		if isFileBeingWritten(path, 1*time.Second) {
			return fmt.Errorf("文件被占用: %s", f)
		}
	}
	return nil
}

// doMerge 执行多文件合并的完整流程：FLV→TS→拼接→MP4→校验→删除原始文件。
// 合并失败时自动 fallback 到重编码模式。
// 调用方须保证 files 按时间升序排列（files[0] 为最早的文件）。
func (s *MergeService) doMerge(ctx context.Context, files []string, folder string, onProgress ProgressFunc, opLog *OpLogger) bool {
	if len(files) < 2 {
		return false
	}
	if ctx.Err() != nil {
		return false
	}

	// 检查所有文件是否可访问且未被锁定
	if err := checkFileAvailability(folder, files); err != nil {
		onProgress(fmt.Sprintf("⚠ %v", err))
		return false
	}

	// 输出文件名：取最早的文件生成
	output := utils.MakeOutputName(files[0])
	outputPath := filepath.Join(folder, output)

	// 手动合并时：如果输出文件已存在，自动加序号避免覆盖
	if _, err := os.Stat(outputPath); err == nil {
		ext := filepath.Ext(output)
		stem := strings.TrimSuffix(output, ext)
		found := false
		for i := 2; i <= maxDedupAttempts; i++ {
			candidate := fmt.Sprintf("%s-%d%s", stem, i, ext)
			if _, err := os.Stat(filepath.Join(folder, candidate)); os.IsNotExist(err) {
				output = candidate
				outputPath = filepath.Join(folder, output)
				found = true
				break
			}
		}
		if !found {
			opLog.Log(fmt.Sprintf("❌ 无法找到可用的输出文件名（已尝试 %d 个后缀）", maxDedupAttempts))
			return false
		}
		opLog.Log(fmt.Sprintf("⚠ 输出文件已存在，自动重命名为: %s", output))
	}

	// 检查磁盘空间
	disk, diskErr := utils.GetDiskUsage(folder)
	if diskErr == nil && disk.Free < minMergeFreeBytes {
		opLog.Log(fmt.Sprintf("❌ 磁盘空间不足（仅剩 %.1f GB / 需要 %.1f GB），跳过", float64(disk.Free)/oneGB, float64(minMergeFreeBytes)/oneGB))
		return false
	}

	// 记录最新源文件的修改时间和总大小，用于时间戳校正和进度显示
	var latestSrcMtime time.Time
	var totalFileSize int64
	for _, f := range files {
		if info, err := os.Stat(filepath.Join(folder, f)); err == nil {
			totalFileSize += info.Size()
			if info.ModTime().After(latestSrcMtime) {
				latestSrcMtime = info.ModTime()
			}
		}
	}

	// 步骤 1：将每个输入文件转换为 TS 格式（已是 TS 的跳过）
	var tsFiles []string
	tmpDir := filepath.Join(folder, ".merge_tmp_"+time.Now().Format("20060102150405"))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		onProgress(fmt.Sprintf("❌ 创建临时目录失败: %v", err))
		return false
	}
	defer os.RemoveAll(tmpDir)

	for _, f := range files {
		if strings.HasSuffix(f, ".ts") {
			tsFiles = append(tsFiles, filepath.ToSlash(filepath.Join(folder, f)))
			continue
		}
		tsName := fmt.Sprintf("seg_%d.ts", len(tsFiles))
		tsPath := filepath.Join(tmpDir, tsName)
		onProgress(fmt.Sprintf("🔄 转换 %s → TS…", filepath.Base(f)))
		err := ffmpeg.Run(ctx, ffmpeg.Options{
			Args: []string{"-nostdin", "-i", filepath.ToSlash(filepath.Join(folder, f)),
				"-c", "copy", "-bsf:v", "h264_mp4toannexb", "-y", "-loglevel", "error", tsPath},
		})
		if err != nil {
			opLog.Log(fmt.Sprintf("❌ FLV→TS 转换失败 %s: %v", filepath.Base(f), err))
			onProgress(fmt.Sprintf("❌ 转换失败: %s", filepath.Base(f)))
			return s.concatReencode(ctx, files, folder, outputPath, onProgress, opLog)
		}
		tsFiles = append(tsFiles, filepath.ToSlash(tsPath))
	}

	// 步骤 2：拼接 TS 文件 → MP4
	onProgress("⚙ 拼接 TS 文件…")
	outputIsFLV := strings.HasSuffix(output, ".flv")
	concatOutputPath := outputPath
	if outputIsFLV {
		concatOutputPath = strings.TrimSuffix(outputPath, ".flv") + ".mp4"
		output = strings.TrimSuffix(output, ".flv") + ".mp4"
	}

	if err := ffmpeg.ConcatTS(ctx, tsFiles, concatOutputPath, onProgress); err != nil {
		opLog.Log(fmt.Sprintf("⚠ TS 拼接失败: %v，切换重编码", err))
		onProgress("⚠ 拼接失败，切换重编码…")
		utils.SafeUnlink(concatOutputPath)
		return s.concatReencode(ctx, files, folder, outputPath, onProgress, opLog)
	}

	// 步骤 3：校验输出文件
	if err := ffmpeg.ValidateOutput(ctx, concatOutputPath); err != nil {
		opLog.Log(fmt.Sprintf("⚠ 输出校验失败: %v，切换重编码", err))
		utils.SafeUnlink(concatOutputPath)
		return s.concatReencode(ctx, files, folder, outputPath, onProgress, opLog)
	}

	// fsync — 确保数据刷入持久存储后再删除原始文件
	if fd, err := os.OpenFile(concatOutputPath, os.O_RDONLY, 0); err == nil {
		if syncErr := fd.Sync(); syncErr != nil {
			fd.Close()
			opLog.Log(fmt.Sprintf("❌ fsync 失败: %v，删除不可靠输出，保留原始文件", syncErr))
			utils.SafeUnlink(concatOutputPath) // 删除不可靠输出，防止下次被误判为已合并
			os.RemoveAll(tmpDir)
			return false
		}
		fd.Close()
	}

	// FLV->MP4 转换：仅在 concat 输出到不同文件路径时需要。
	// 注意：此条件依赖 ConcatTS 将 .flv 输出自动转为 .mp4 的行为（concatOutputPath 已是 .mp4）。
	// 若修改 MakeMP4Name 或 ConcatTS 的命名/输出规则，需同步检查此处条件。
	if outputIsFLV && concatOutputPath != filepath.Join(folder, utils.MakeMP4Name(output)) {
		mp4Name := utils.MakeMP4Name(output)
		mp4Path := filepath.Join(folder, mp4Name)
		onProgress(fmt.Sprintf("🔄 转换 FLV→MP4: %s", mp4Name))

		if err := ffmpeg.ConvertViaTS(ctx, concatOutputPath, mp4Path); err != nil {
			opLog.Log(fmt.Sprintf("❌ FLV→MP4 失败: %v，保留 FLV", err))
			return false
		} else if err := ffmpeg.ValidateOutput(ctx, mp4Path); err != nil {
			opLog.Log(fmt.Sprintf("❌ MP4 输出损坏，保留 FLV: %s", mp4Name))
			utils.SafeUnlink(mp4Path)
			return false
		} else {
			if !latestSrcMtime.IsZero() {
				if err := os.Chtimes(mp4Path, latestSrcMtime, latestSrcMtime); err != nil {
					opLog.Log(fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(mp4Path), err))
				}
			}
			utils.SafeUnlink(concatOutputPath)
			opLog.Log(fmt.Sprintf("✅ FLV→MP4 完成: %s", mp4Name))
		}
	} else if outputIsFLV {
		// ConcatTS 已直接输出 MP4 — 只需保留录制时间戳
		if !latestSrcMtime.IsZero() {
			if err := os.Chtimes(concatOutputPath, latestSrcMtime, latestSrcMtime); err != nil {
				opLog.Log(fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(concatOutputPath), err))
			}
		}
	} else {
		// 非 FLV 输出：保留合并文件的录制时间
		if !latestSrcMtime.IsZero() {
			if err := os.Chtimes(concatOutputPath, latestSrcMtime, latestSrcMtime); err != nil {
				opLog.Log(fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(concatOutputPath), err))
			}
		}
	}

	// 合并校验通过后删除原始文件
	for _, f := range files {
		utils.SafeUnlink(filepath.Join(folder, f))
	}

	return true
}

// ManualMerge 手动合并指定主播的指定文件列表。
// 获取主播锁后校验文件合法性，按文件名时间排序后调用 doMerge 执行合并。
// ctx 用于控制合并过程的取消。
func (s *MergeService) ManualMerge(ctx context.Context, streamer string, files []string, onProgress ProgressFunc) (string, error) {
	name := streamer
	locked, sl := s.tryLockStreamer(name)
	if !locked {
		return "", fmt.Errorf("%s 合并任务正在执行中", name)
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
		return opLog.LogID(), fmt.Errorf("当前处于静默时段（%d:%02d-%d:%02d），合并暂停", cfg.BackupStartHour, cfg.BackupStartMinute, cfg.BackupEndHour, cfg.BackupEndMinute)
	}
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return opLog.LogID(), fmt.Errorf("目录不存在")
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
			return opLog.LogID(), fmt.Errorf("非法文件名: %s", f)
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
		if len(skipped) > 0 {
			return opLog.LogID(), fmt.Errorf("有效文件不足2个（%d 个文件已不存在，请刷新文件列表）", len(skipped))
		}
		return opLog.LogID(), fmt.Errorf("有效文件不足2个")
	}

	// 按文件名中的日期时间升序排序（确保合并后时间线正确）
	SortByFilename(validFiles)

	if len(skipped) > 0 {
		onProgress(fmt.Sprintf("⚠ %d 个文件已不存在，使用剩余 %d 个文件继续", len(skipped), len(validFiles)))
	}

	onProgress(fmt.Sprintf("⚙ 手动合并 %d 个文件", len(validFiles)))
	hasOriginal := false
	for _, f := range validFiles {
		if !utils.IsMergedFile(f) {
			hasOriginal = true
			break
		}
	}
	if !hasOriginal {
		return opLog.LogID(), fmt.Errorf("所选文件全部是合并版，请至少选择一个原始文件")
	}

	start := time.Now()
	if s.doMerge(ctx, validFiles, folder, onProgress, opLog) {
		duration := time.Since(start).Seconds()
		s.history.AddWithStats("merge", streamer, "success", len(validFiles), 0, totalInputBytes, duration, fmt.Sprintf("手动合并 %d 个文件 (%s)", len(validFiles), utils.FormatSize(totalInputBytes)), opLog.LogID())
		onProgress(fmt.Sprintf("✅ 手动合并完成: %d 个文件 (%s)", len(validFiles), utils.FormatSize(totalInputBytes)))
		return opLog.LogID(), nil
	}

	return opLog.LogID(), fmt.Errorf("合并失败")
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
// 包括：concat 临时文件、.mp4.tmp/.flv.tmp 文件、孤立 TS 文件、崩溃残留的 .merge_tmp_* 目录。
func (s *MergeService) CleanupTempFiles() int {
	cfg := s.config.Snapshot()
	root := cfg.TargetDir
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	cleaned := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		folder := filepath.Join(root, entry.Name())
		fileEntries, err := os.ReadDir(folder)
		if err != nil {
			continue
		}

		// 收集有效 MP4 文件名用于孤立文件检测
		mp4Bases := make(map[string]bool)
		for _, fe := range fileEntries {
			if fe.IsDir() {
				continue
			}
			name := fe.Name()
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".mp4" {
				if info, err := fe.Info(); err == nil && info.Size() >= minValidFileSize {
					base := strings.TrimSuffix(name, ext)
					mp4Bases[base] = true
				}
			}
		}

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

			// 清理残留的 concat 临时文件
			if strings.HasPrefix(name, ".concat_") && strings.HasSuffix(name, ".txt") {
				path := filepath.Join(folder, name)
				if err := os.Remove(path); err == nil {
					cleaned++
					s.logger.Info(fmt.Sprintf("🗑 清理临时文件: %s", name))
				}
			}

			// 清理 MP4/FLV 的 .tmp 后缀文件（中断写入的残留）
			if strings.HasSuffix(name, ".tmp") && (strings.HasSuffix(name, ".mp4.tmp") || strings.HasSuffix(name, ".flv.tmp")) {
				path := filepath.Join(folder, name)
				if err := os.Remove(path); err == nil {
					cleaned++
					s.logger.Info(fmt.Sprintf("🗑 清理临时文件: %s", name))
				}
			}

			// 清理孤立 TS 文件：没有对应的有效 MP4 文件
			if fe.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".ts" && !strings.HasPrefix(name, ".") {
				base := strings.TrimSuffix(name, ext)
				if !mp4Bases[base] {
					path := filepath.Join(folder, name)
					if info, err := fe.Info(); err == nil && info.Size() >= minValidFileSize {
						if err := utils.SafeUnlink(path); err == nil {
							cleaned++
							s.logger.Info(fmt.Sprintf("🗑 清理孤立TS: %s", name))
						}
					}
				}
			}
		}
	}
	return cleaned
}
