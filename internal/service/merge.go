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

// logToFile appends a timestamped line to the task log file, rotating at midnight.
func (s *MergeService) logToFile(task, message string) {
	logToFile(s.config.LogDir, task, message, s.logger)
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
	// TS pipeline peak space: source files + TS intermediates + output ~ 3x source + 2GB headroom
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

	// Check if output was created but then deleted by validation probe
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		// Check if source files have issues
		firstPath := filepath.Join(folder, firstFile)
		if info, err := os.Stat(firstPath); err != nil {
			return "源文件不存在"
		} else if info.Size() < minValidFileSize {
			return fmt.Sprintf("源文件过小(%s)", utils.FormatSize(info.Size()))
		}
		return "ffmpeg输出校验失败"
	}

	// Output exists but too small
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
func (s *MergeService) Run(ctx context.Context, streamer string, onProgress ProgressFunc) (*MergeResult, error) {
	start := time.Now()
	cfg := s.config.Snapshot()
	root := cfg.TargetDir

	if cfg.IsBackupWindow() {
		return nil, fmt.Errorf("当前处于静默时段（%d:%02d-%d:%02d），合并暂停", cfg.BackupStartHour, cfg.BackupStartMinute, cfg.BackupEndHour, cfg.BackupEndMinute)
	}
	if onProgress == nil {
		onProgress = func(string) {}
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, fmt.Errorf("路径不存在: %s", root)
	}

	tag := "[全局]"
	if streamer != "" {
		tag = fmt.Sprintf("[%s]", streamer)
	}
	s.logToFile("merge", "═══════════════════════════════════════════")
	s.logToFile("merge", fmt.Sprintf("▶ 开始 %s 合并", tag))
	onProgress(fmt.Sprintf("▶ 开始 %s 合并", tag))
	onProgress(fmt.Sprintf("📂 扫描 %s ...", root))

	tasks, convertTasks := s.scanTasks(ctx, root, streamer, cfg)
	if len(tasks) == 0 && len(convertTasks) == 0 {
		s.logToFile("merge", "ℹ 无待合并文件")
		s.history.Add("merge", streamer, "success", "扫描完成，无待合并文件")
		onProgress("ℹ 无待合并文件")
		return &MergeResult{}, nil
	}

	if len(tasks) > 0 {
		if err := s.checkDiskSpaceForMerge(tasks, root); err != nil {
			s.logToFile("merge", fmt.Sprintf("❌ %s", err.Error()))
			onProgress(fmt.Sprintf("❌ %s", err.Error()))
			return nil, err
		}
	}

	// 磁盘空间硬性检查 — 可用空间低于 10GB 时跳过所有操作
	disk, diskErr := utils.GetDiskUsage(root)
	if diskErr == nil && disk.Free < minDiskFreeBytes { // < 10GB free
		s.logToFile("merge", fmt.Sprintf("❌ 磁盘空间不足（仅剩 %.1f GB），跳过所有操作", float64(disk.Free)/oneGB))
		onProgress(fmt.Sprintf("❌ 磁盘空间不足（仅剩 %.1f GB），跳过", float64(disk.Free)/oneGB))
		s.history.Add("merge", streamer, "fail", fmt.Sprintf("磁盘空间不足: %.1f GB（使用率 %.1f%%）", float64(disk.Free)/oneGB, disk.UsedPct))
		return &MergeResult{}, nil
	}

	done := 0
	totalGB := 0.0
	convertDone := 0
	mergeDone := 0
	mergeFailed := 0
	failedReasons := make(map[string]int)
	for i, ct := range convertTasks {
		if cfg.IsBackupWindow() || ctx.Err() != nil {
			break
		}
		var flvSize int64
		if fi, err := os.Stat(ct.FlvPath); err == nil {
			flvSize = fi.Size()
		}
		onProgress(fmt.Sprintf("🔄 [%d/%d] %s ⚙ %s → %s", i+1, len(convertTasks), ct.Name, filepath.Base(ct.FlvPath), filepath.Base(ct.Mp4Path)))
		if s.convertFlvToMp4(ctx, ct.FlvPath, ct.Mp4Path, onProgress) {
			done++
			convertDone++
			totalGB += float64(flvSize) / oneGB
			onProgress(fmt.Sprintf("✅ [%s] 完成", ct.Name))
		} else {
			onProgress(fmt.Sprintf("❌ [%s] 转换失败", ct.Name))
		}
	}

	// Merge tasks
	for i, task := range tasks {
		if cfg.IsBackupWindow() {
			break
		}
		streamerName := filepath.Base(task.Folder)
		locked, sl := s.tryLockStreamer(streamerName)
		if !locked {
			continue
		}
		fileList := strings.Join(task.Files, " + ")
		onProgress(fmt.Sprintf("⚙ [%d/%d] %s ⚙ %s (%.1f GB)", i+1, len(tasks), streamerName, fileList, task.SizeGB))
		s.logToFile("merge", fmt.Sprintf("⚙ %s 合并 %d 个文件: %s", streamerName, len(task.Files), fileList))
		if s.doMerge(ctx, task.Files, task.Folder, onProgress) {
			done++
			mergeDone++
			totalGB += task.SizeGB
			s.logToFile("merge", fmt.Sprintf("✅ %s 完成 → %s", streamerName, utils.MakeOutputName(task.Files[0])))
		} else {
			mergeFailed++
			reason := classifyMergeFailure(task.Folder, task.Files[0])
			failedReasons[reason]++
			s.logToFile("merge", fmt.Sprintf("❌ %s 失败: %s", streamerName, reason))
			onProgress(fmt.Sprintf("❌ [%s] 失败: %s", streamerName, reason))
		}
		s.unlockStreamer(sl)
	}

	// Summary
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

	onProgress("───────────────────────────")
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
		s.logToFile("merge", msg)
		s.history.AddWithStats("merge", streamer, "success", done, 0, int64(totalGB*oneGB), duration, detail)
		onProgress(msg)
	} else if mergeFailed > 0 {
		var parts []string
		for reason, cnt := range failedReasons {
			parts = append(parts, fmt.Sprintf("%s x %d", reason, cnt))
		}
		msg := fmt.Sprintf("❌ 完成: 扫描 %d 个主播, 全部失败 %s", totalScanned, strings.Join(parts, ", "))
		s.logToFile("merge", msg)
		s.history.Add("merge", streamer, "fail", fmt.Sprintf("合并失败: %s", strings.Join(parts, ", ")))
		onProgress(msg)
	} else {
		msg := fmt.Sprintf("ℹ 完成: 扫描 %d 个主播, 无需合并", totalScanned)
		s.logToFile("merge", msg)
		s.history.Add("merge", streamer, "success", fmt.Sprintf("扫描 %d 个主播，无需合并", totalScanned))
		onProgress(msg)
	}
	onProgress("───────────────────────────")
	s.logToFile("merge", "═══════════════════════════════════════════")

	return &MergeResult{Done: done, TotalGB: totalGB}, nil
}

// convertFlvToMp4 将单个 FLV 文件转换为 MP4（通过 TS 中间格式）。
// 转换成功后保留原始录制时间，删除原始 FLV 文件。
func (s *MergeService) convertFlvToMp4(ctx context.Context, flvPath, mp4Path string, onProgress ProgressFunc) bool {
	// Check source file is not locked
	if isFileBeingWritten(flvPath, 1*time.Second) {
		onProgress(fmt.Sprintf("⚠ %s 被占用，跳过", filepath.Base(flvPath)))
		return false
	}

	// Check disk space
	disk, diskErr := utils.GetDiskUsage(filepath.Dir(flvPath))
	if diskErr == nil && disk.Free < minConvertFreeBytes {
		onProgress("⚠ 磁盘空间不足，跳过转换")
		return false
	}

	// Capture original recording time before conversion
	var flvMtime time.Time
	if info, err := os.Stat(flvPath); err == nil {
		flvMtime = info.ModTime()
	}

	if err := ffmpeg.ConvertViaTS(ctx, flvPath, mp4Path); err != nil {
		s.logToFile("merge", fmt.Sprintf("❌ FLV→MP4 失败: %v，保留原始文件", err))
		onProgress(fmt.Sprintf("❌ 转换失败，保留 %s", filepath.Base(flvPath)))
		return false
	}

	if err := ffmpeg.ValidateOutput(ctx, mp4Path); err != nil {
		s.logToFile("merge", fmt.Sprintf("❌ MP4 输出校验失败: %v，保留原始文件", err))
		return false
	}

	// 保留原始录制时间，防止因 mtime 变化导致合并分组错误
	if !flvMtime.IsZero() {
		if err := os.Chtimes(mp4Path, flvMtime, flvMtime); err != nil {
			s.logToFile("merge", fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(mp4Path), err))
		}
	}

	// Delete original with retry
	if err := utils.SafeUnlink(flvPath); err != nil {
		s.logToFile("merge", fmt.Sprintf("⚠ 删除原始文件失败: %v", err))
	}

	if info, err := os.Stat(mp4Path); err != nil {
		s.logToFile("merge", fmt.Sprintf("✅ FLV→MP4 完成: %s (大小未知)", filepath.Base(mp4Path)))
	} else {
		s.logToFile("merge", fmt.Sprintf("✅ FLV→MP4 完成: %s (%s)", filepath.Base(mp4Path), utils.FormatSize(info.Size())))
	}
	onProgress(fmt.Sprintf("✅ 完成: %s", filepath.Base(mp4Path)))
	return true
}

// concatReencode 使用 ffmpeg concat filter 重编码合并文件。
// 作为 stream-copy 失败时的 fallback（编解码器不兼容、头部损坏等情况）。
func (s *MergeService) concatReencode(ctx context.Context, files []string, folder, outputPath string, onProgress ProgressFunc) bool {
	// Check total input size — skip re-encode for large files on weak hardware
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
		s.logToFile("merge", fmt.Sprintf("⚠ 文件过大 (%s)，跳过重编码（硬件性能不足）", utils.FormatSize(totalSize)))
		onProgress(fmt.Sprintf("⚠ 文件过大 (%s)，跳过重编码", utils.FormatSize(totalSize)))
		return false
	}

	if err := ffmpeg.Reencode(ctx, files, folder, outputPath, onProgress); err != nil {
		s.logToFile("merge", fmt.Sprintf("❌ %v", err))
		return false
	}

	// Preserve original recording time
	if !latestSrcMtime.IsZero() {
		if err := os.Chtimes(outputPath, latestSrcMtime, latestSrcMtime); err != nil {
			s.logToFile("merge", fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(outputPath), err))
		}
	}

	if info, err := os.Stat(outputPath); err != nil {
		s.logToFile("merge", "✅ 重编码完成 (大小未知)")
	} else {
		s.logToFile("merge", fmt.Sprintf("✅ 重编码完成: %s", utils.FormatSize(info.Size())))
	}
	return true
}

// checkFileAvailability 检查批次中的所有文件是否可访问且未被锁定。
func checkFileAvailability(folder string, files []string) error {
	for _, f := range files {
		path := filepath.Join(folder, f)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("文件不存在: %s", f)
		}
		// Try to open for reading — if another process has it locked for writing,
		// the file might still be readable, but check if it's being written to
		if isFileBeingWritten(path, 1*time.Second) {
			return fmt.Errorf("文件被占用: %s", f)
		}
	}
	return nil
}

// doMerge 执行多文件合并的完整流程：FLV→TS→拼接→MP4→校验→删除原始文件。
// 合并失败时自动 fallback 到重编码模式。
// 调用方须保证 files 按时间升序排列（files[0] 为最早的文件）。
func (s *MergeService) doMerge(ctx context.Context, files []string, folder string, onProgress ProgressFunc) bool {
	if len(files) < 2 {
		return false
	}
	if ctx.Err() != nil {
		return false
	}

	// Check all files are accessible and not locked
	if err := checkFileAvailability(folder, files); err != nil {
		s.logToFile("merge", fmt.Sprintf("⚠ %v，跳过合并", err))
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
			s.logToFile("merge", fmt.Sprintf("❌ 无法找到可用的输出文件名（已尝试 %d 个后缀）", maxDedupAttempts))
			return false
		}
		s.logToFile("merge", fmt.Sprintf("⚠ 输出文件已存在，自动重命名为: %s", output))
	}

	// Check disk space
	disk, diskErr := utils.GetDiskUsage(folder)
	if diskErr == nil && disk.Free < minMergeFreeBytes {
		s.logToFile("merge", fmt.Sprintf("❌ 磁盘空间不足（仅剩 %.1f GB），跳过", float64(disk.Free)/oneGB))
		return false
	}

	// Capture latest source mtime for timestamp correction after merge
	var latestSrcMtime time.Time
	for _, f := range files {
		if info, err := os.Stat(filepath.Join(folder, f)); err == nil {
			if info.ModTime().After(latestSrcMtime) {
				latestSrcMtime = info.ModTime()
			}
		}
	}

	onProgress(fmt.Sprintf("⚙ 合并 %d 个文件…", len(files)))

	// 步骤 1：将每个输入文件转换为 TS 格式（已是 TS 的跳过）
	var tsFiles []string
	tmpDir := filepath.Join(folder, ".merge_tmp_"+time.Now().Format("20060102150405"))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		s.logToFile("merge", fmt.Sprintf("❌ 创建临时目录失败: %v", err))
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
			s.logToFile("merge", fmt.Sprintf("❌ FLV→TS 转换失败 %s: %v", filepath.Base(f), err))
			onProgress(fmt.Sprintf("❌ 转换失败: %s", filepath.Base(f)))
			return s.concatReencode(ctx, files, folder, outputPath, onProgress)
		}
		tsFiles = append(tsFiles, filepath.ToSlash(tsPath))
	}

	// 步骤 2：拼接 TS 文件 → MP4
	onProgress("⚙ 拼接 TS 文件…")
	outputIsFLV := strings.HasSuffix(output, ".flv")
	concatOutputPath := outputPath
	if outputIsFLV {
		concatOutputPath = strings.TrimSuffix(outputPath, ".flv") + ".mp4"
	}

	if err := ffmpeg.ConcatTS(ctx, tsFiles, concatOutputPath, onProgress); err != nil {
		s.logToFile("merge", fmt.Sprintf("⚠ TS 拼接失败: %v，切换重编码", err))
		onProgress("⚠ 拼接失败，切换重编码…")
		utils.SafeUnlink(concatOutputPath)
		return s.concatReencode(ctx, files, folder, outputPath, onProgress)
	}

	// 步骤 3：校验输出文件
	if err := ffmpeg.ValidateOutput(ctx, concatOutputPath); err != nil {
		s.logToFile("merge", fmt.Sprintf("⚠ 输出校验失败: %v，切换重编码", err))
		utils.SafeUnlink(concatOutputPath)
		return s.concatReencode(ctx, files, folder, outputPath, onProgress)
	}

	// fsync — 确保数据刷入持久存储后再删除原始文件
	if fd, err := os.OpenFile(concatOutputPath, os.O_RDONLY, 0); err == nil {
		if syncErr := fd.Sync(); syncErr != nil {
			fd.Close()
			s.logToFile("merge", fmt.Sprintf("❌ fsync 失败: %v，删除不可靠输出，保留原始文件", syncErr))
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
			s.logToFile("merge", fmt.Sprintf("❌ FLV→MP4 失败: %v，保留 FLV", err))
			return false
		} else if err := ffmpeg.ValidateOutput(ctx, mp4Path); err != nil {
			s.logToFile("merge", "❌ MP4 输出损坏，保留 FLV")
			utils.SafeUnlink(mp4Path)
			return false
		} else {
			if !latestSrcMtime.IsZero() {
				if err := os.Chtimes(mp4Path, latestSrcMtime, latestSrcMtime); err != nil {
					s.logToFile("merge", fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(mp4Path), err))
				}
			}
			utils.SafeUnlink(concatOutputPath)
			s.logToFile("merge", fmt.Sprintf("✅ FLV→MP4 → %s", mp4Name))
		}
	} else if outputIsFLV {
		// ConcatTS 已直接输出 MP4 — 只需保留录制时间戳
		if !latestSrcMtime.IsZero() {
			if err := os.Chtimes(concatOutputPath, latestSrcMtime, latestSrcMtime); err != nil {
				s.logToFile("merge", fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(concatOutputPath), err))
			}
		}
	} else {
		// 非 FLV 输出：保留合并文件的录制时间
		if !latestSrcMtime.IsZero() {
			if err := os.Chtimes(concatOutputPath, latestSrcMtime, latestSrcMtime); err != nil {
				s.logToFile("merge", fmt.Sprintf("⚠ 设置时间戳失败 %s: %v", filepath.Base(concatOutputPath), err))
			}
		}
	}

	// 合并校验通过后删除原始文件
	for _, f := range files {
		utils.SafeUnlink(filepath.Join(folder, f))
	}

	s.logToFile("merge", fmt.Sprintf("✅ 合并成功: %s (%d 个文件)", output, len(files)))
	return true
}

// ManualMerge 手动合并指定主播的指定文件列表。
// 获取主播锁后校验文件合法性，然后调用 doMerge 执行合并。
func (s *MergeService) ManualMerge(ctx context.Context, streamer string, files []string, onProgress ProgressFunc) error {
	name := streamer
	locked, sl := s.tryLockStreamer(name)
	if !locked {
		return fmt.Errorf("%s 合并任务正在执行中", name)
	}
	defer s.unlockStreamer(sl)

	cfg := s.config.Snapshot()
	folder := filepath.Join(cfg.TargetDir, streamer)
	if cfg.IsBackupWindow() {
		return fmt.Errorf("当前处于静默时段（%d:%02d-%d:%02d），合并暂停", cfg.BackupStartHour, cfg.BackupStartMinute, cfg.BackupEndHour, cfg.BackupEndMinute)
	}
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return fmt.Errorf("目录不存在")
	}

	if onProgress == nil {
		onProgress = func(string) {}
	}

	var validFiles []string
	var totalInputBytes int64
	for _, f := range files {
		if !utils.ValidateFilename(f) {
			return fmt.Errorf("非法文件名: %s", f)
		}
		path := filepath.Join(folder, f)
		if info, err := os.Stat(path); err == nil {
			validFiles = append(validFiles, f)
			totalInputBytes += info.Size()
		}
	}

	if len(validFiles) < 2 {
		return fmt.Errorf("有效文件不足2个")
	}

	// 按文件名中的日期时间升序排序（确保合并后时间线正确）
	SortByFilename(validFiles)

	onProgress(fmt.Sprintf("⏳ 手动合并 %d 个文件", len(validFiles)))
	hasOriginal := false
	for _, f := range validFiles {
		if !utils.IsMergedFile(f) {
			hasOriginal = true
			break
		}
	}
	if !hasOriginal {
		return fmt.Errorf("所选文件全部是合并版，请至少选择一个原始文件")
	}

	start := time.Now()
	if s.doMerge(ctx, validFiles, folder, onProgress) {
		duration := time.Since(start).Seconds()
		s.history.AddWithStats("merge", streamer, "success", len(validFiles), 0, totalInputBytes, duration, fmt.Sprintf("手动合并 %d 个文件 (%s)", len(validFiles), utils.FormatSize(totalInputBytes)))
		onProgress(fmt.Sprintf("✅ 手动合并 %d 个文件完成", len(validFiles)))
		return nil
	}

	return fmt.Errorf("合并失败")
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
					s.logToFile("merge", fmt.Sprintf("🗑 清理残留临时目录: %s", name))
				}
				continue
			}

			// 清理残留的 concat 临时文件
			if strings.HasPrefix(name, ".concat_") && strings.HasSuffix(name, ".txt") {
				path := filepath.Join(folder, name)
				if err := os.Remove(path); err == nil {
					cleaned++
					s.logToFile("merge", fmt.Sprintf("🗑 清理临时文件: %s", name))
				}
			}

			// 清理 MP4/FLV 的 .tmp 后缀文件（中断写入的残留）
			if strings.HasSuffix(name, ".tmp") && (strings.HasSuffix(name, ".mp4.tmp") || strings.HasSuffix(name, ".flv.tmp")) {
				path := filepath.Join(folder, name)
				if err := os.Remove(path); err == nil {
					cleaned++
					s.logToFile("merge", fmt.Sprintf("🗑 清理临时文件: %s", name))
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
							s.logToFile("merge", fmt.Sprintf("🗑 清理孤立TS: %s", name))
						}
					}
				}
			}
		}
	}
	return cleaned
}
