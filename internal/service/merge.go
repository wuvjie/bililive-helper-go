package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bililive-helper/internal/config"
	"bililive-helper/internal/ffmpeg"
	"bililive-helper/internal/utils"

	"go.uber.org/zap"
)

type ProgressFunc func(msg string)

type MergeService struct {
	config        *config.Config
	logger        *zap.Logger
	history       *HistoryService
	streamerLocks sync.Map // streamer name -> *streamerLock
}

type streamerLock struct {
	mu        sync.Mutex
	createdAt time.Time
}

const lockTimeout = 4 * time.Hour

func (s *MergeService) tryLockStreamer(name string) bool {
	now := time.Now()
	val, _ := s.streamerLocks.LoadOrStore(name, &streamerLock{})
	sl := val.(*streamerLock)

	if sl.mu.TryLock() {
		sl.createdAt = now
		return true
	}

	// Safety net: if lock held longer than timeout, destroy and retry with fresh entry
	if now.Sub(sl.createdAt) > lockTimeout {
		s.logger.Warn("发现过期主播锁，销毁旧锁", zap.String("streamer", name))
		s.streamerLocks.Delete(name)
		// LoadOrStore ensures each goroutine gets its own mutex entry
		newVal, _ := s.streamerLocks.LoadOrStore(name, &streamerLock{})
		newSl := newVal.(*streamerLock)
		if newSl.mu.TryLock() {
			newSl.createdAt = now
			return true
		}
		return false
	}

	return false
}

func (s *MergeService) unlockStreamer(name string) {
	if val, ok := s.streamerLocks.Load(name); ok {
		val.(*streamerLock).mu.Unlock()
	}
}

func NewMergeService(config *config.Config, logger *zap.Logger, history *HistoryService) *MergeService {
	return &MergeService{config: config, logger: logger, history: history}
}

// logToFile appends a timestamped line to the task log file, rotating at midnight.
func (s *MergeService) logToFile(task, message string) {
	logToFile(s.config.LogDir, task, message, s.logger)
}

type MergeResult struct {
	Done    int
	TotalGB float64
}

func (s *MergeService) checkDiskSpaceForMerge(tasks []mergeTask, targetDir string) error {
	var totalSourceBytes int64
	for _, t := range tasks {
		totalSourceBytes += int64(t.SizeGB * 1073741824)
	}
	disk, err := utils.GetDiskUsage(targetDir)
	if err != nil {
		return fmt.Errorf("获取磁盘信息失败: %w", err)
	}
	// TS方案峰值空间: 源文件 + TS中间文件 + 输出文件 ≈ 3倍源文件 + 2GB系统底线
	needed := (totalSourceBytes * 3) + (2 * 1024 * 1024 * 1024)
	if int64(disk.Free) < needed {
		return fmt.Errorf("磁盘空间不足：需要 %.1f GB 可用以应对 TS 转换峰值，当前仅 %.1f GB",
			float64(needed)/1073741824, float64(disk.Free)/1073741824)
	}
	return nil
}

// classifyMergeFailure inspects the batch files and output to determine why doMerge likely failed.
func classifyMergeFailure(folder, firstFile string) string {
	output := utils.MakeOutputName(firstFile)
	outputPath := filepath.Join(folder, output)

	// Check if output was created but deleted by probe
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		// Check if any source files have issues
		firstPath := filepath.Join(folder, firstFile)
		if info, err := os.Stat(firstPath); err != nil {
			return "源文件不存在"
		} else if info.Size() < 10240 {
			return fmt.Sprintf("源文件过小(%s)", utils.FormatSize(info.Size()))
		}
		return "ffmpeg输出校验失败"
	}

	// Output exists but too small
	if info, err := os.Stat(outputPath); err == nil {
		if info.Size() < 10240 {
			return fmt.Sprintf("输出过小(%s)", utils.FormatSize(info.Size()))
		}
	}
	return "输出校验失败"
}

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

	tasks, convertTasks := s.scanTasks(root, streamer, cfg)
	if len(tasks) == 0 && len(convertTasks) == 0 {
		s.logToFile("merge", "ℹ 无待合并文件")
		s.history.Add("merge", streamer, "success", "无待合并文件")
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

	// Disk space check before processing
	disk, diskErr := utils.GetDiskUsage(root)
	if diskErr == nil && disk.Free < 10*1024*1024*1024 { // < 10GB free
		s.logToFile("merge", fmt.Sprintf("❌ 磁盘空间不足（仅剩 %.1f GB），跳过所有操作", float64(disk.Free)/1073741824))
		onProgress(fmt.Sprintf("❌ 磁盘空间不足（仅剩 %.1f GB），跳过", float64(disk.Free)/1073741824))
		s.history.Add("merge", streamer, "fail", fmt.Sprintf("磁盘空间不足: %.1f GB", float64(disk.Free)/1073741824))
		return &MergeResult{}, nil
	}

	done := 0
	totalGB := 0.0
	convertDone := 0
	mergeDone := 0
	mergeFailed := 0
	failedReasons := make(map[string]int)
	for i, ct := range convertTasks {
		if cfg.IsBackupWindow() {
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
			totalGB += float64(flvSize) / 1073741824
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
		if !s.tryLockStreamer(streamerName) {
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
		s.unlockStreamer(streamerName)
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
		s.history.AddWithStats("merge", streamer, "success", done, 0, int64(totalGB*1073741824), duration, detail)
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
		s.history.Add("merge", streamer, "success", "无新合并")
		onProgress(msg)
	}
	onProgress("───────────────────────────")
	s.logToFile("merge", "═══════════════════════════════════════════")

	return &MergeResult{Done: done, TotalGB: totalGB}, nil
}

// convertFlvToMp4 converts a single FLV file to MP4 via TS (same pipeline as multi-file merge).
func (s *MergeService) convertFlvToMp4(ctx context.Context, flvPath, mp4Path string, onProgress ProgressFunc) bool {
	// Check source file is not locked
	if isFileBeingWritten(flvPath, 1*time.Second) {
		onProgress(fmt.Sprintf("⚠ %s 被占用，跳过", filepath.Base(flvPath)))
		return false
	}

	// Check disk space
	disk, diskErr := utils.GetDiskUsage(filepath.Dir(flvPath))
	if diskErr == nil && disk.Free < 512*1024*1024 {
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

	// Preserve original recording time on the converted MP4 to prevent
	// incorrect merge grouping caused by mtime-based gap calculations
	if !flvMtime.IsZero() {
		os.Chtimes(mp4Path, flvMtime, flvMtime)
	}

	// Delete original with retry
	if err := utils.SafeUnlink(flvPath); err != nil {
		s.logToFile("merge", fmt.Sprintf("⚠ 删除原始文件失败: %v", err))
	}

	info, _ := os.Stat(mp4Path)
	s.logToFile("merge", fmt.Sprintf("✅ FLV→MP4 完成: %s (%s)", filepath.Base(mp4Path), utils.FormatSize(info.Size())))
	onProgress(fmt.Sprintf("✅ 完成: %s", filepath.Base(mp4Path)))
	return true
}

// concatReencode merges files using ffmpeg concat filter with re-encoding.
// Used as fallback when -c copy fails or produces corrupted output.
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
		os.Chtimes(outputPath, latestSrcMtime, latestSrcMtime)
	}

	info, _ := os.Stat(outputPath)
	s.logToFile("merge", fmt.Sprintf("✅ 重编码完成: %s", utils.FormatSize(info.Size())))
	return true
}

// checkFileAvailability checks if all files in the batch are accessible and not being written to
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

	output := utils.MakeOutputName(files[0])
	outputPath := filepath.Join(folder, output)

	// Check disk space
	disk, diskErr := utils.GetDiskUsage(folder)
	if diskErr == nil && disk.Free < 1024*1024*1024 {
		s.logToFile("merge", fmt.Sprintf("❌ 磁盘空间不足（仅剩 %.1f GB），跳过", float64(disk.Free)/1073741824))
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

	// Step 1: Convert each input to TS (if not already TS)
	var tsFiles []string
	tmpDir := filepath.Join(folder, ".merge_tmp_"+time.Now().Format("20060102150405"))
	os.MkdirAll(tmpDir, 0755)
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
			os.RemoveAll(tmpDir)
			return s.concatReencode(ctx, files, folder, outputPath, onProgress)
		}
		tsFiles = append(tsFiles, filepath.ToSlash(tsPath))
	}

	// Step 2: Concat TS files → MP4
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

	// Step 3: Validate output
	if err := ffmpeg.ValidateOutput(ctx, concatOutputPath); err != nil {
		s.logToFile("merge", fmt.Sprintf("⚠ 输出校验失败: %v，切换重编码", err))
		utils.SafeUnlink(concatOutputPath)
		return s.concatReencode(ctx, files, folder, outputPath, onProgress)
	}

	// fsync
	if fd, err := os.OpenFile(concatOutputPath, os.O_RDWR, 0); err == nil {
		fd.Sync()
		fd.Close()
	}

	// FLV→MP4 conversion: only needed when concatOutputPath != outputPath
	// (i.e. concat wrote to a different file that needs renaming/converting)
	if outputIsFLV && concatOutputPath != filepath.Join(folder, utils.MakeMP4Name(output)) {
		mp4Name := utils.MakeMP4Name(output)
		mp4Path := filepath.Join(folder, mp4Name)
		onProgress(fmt.Sprintf("🔄 转换 FLV→MP4: %s", mp4Name))

		if err := ffmpeg.ConvertViaTS(ctx, concatOutputPath, mp4Path); err != nil {
			s.logToFile("merge", fmt.Sprintf("❌ FLV→MP4 失败: %v，保留 FLV", err))
		} else if err := ffmpeg.ValidateOutput(ctx, mp4Path); err != nil {
			s.logToFile("merge", "❌ MP4 输出损坏，保留 FLV")
			utils.SafeUnlink(mp4Path)
		} else {
			if !latestSrcMtime.IsZero() {
				os.Chtimes(mp4Path, latestSrcMtime, latestSrcMtime)
			}
			utils.SafeUnlink(concatOutputPath)
			s.logToFile("merge", fmt.Sprintf("✅ FLV→MP4 → %s", mp4Name))
		}
	} else if outputIsFLV {
		// ConcatTS already wrote the MP4 directly — just preserve timestamp
		if !latestSrcMtime.IsZero() {
			os.Chtimes(concatOutputPath, latestSrcMtime, latestSrcMtime)
		}
	} else {
		// Non-FLV output: preserve recording time on the merged file
		if !latestSrcMtime.IsZero() {
			os.Chtimes(concatOutputPath, latestSrcMtime, latestSrcMtime)
		}
	}

	// Delete originals — merge validated, originals no longer needed
	for _, f := range files {
		utils.SafeUnlink(filepath.Join(folder, f))
	}

	s.logToFile("merge", fmt.Sprintf("✅ 合并成功: %s (%d 个文件)", output, len(files)))
	return true
}

func (s *MergeService) ManualMerge(ctx context.Context, streamer string, files []string, onProgress ProgressFunc) error {
	name := streamer
	if !s.tryLockStreamer(name) {
		return fmt.Errorf("%s 合并任务正在执行中", name)
	}
	defer s.unlockStreamer(name)

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
	for _, f := range files {
		if !utils.ValidateFilename(f) {
			return fmt.Errorf("非法文件名: %s", f)
		}
		path := filepath.Join(folder, f)
		if _, err := os.Stat(path); err == nil {
			validFiles = append(validFiles, f)
		}
	}

	if len(validFiles) < 2 {
		return fmt.Errorf("有效文件不足2个")
	}

	onProgress(fmt.Sprintf("⏳ 手动合并 %d 个文件", len(validFiles)))

	if s.doMerge(ctx, validFiles, folder, onProgress) {
		s.history.Add("merge", streamer, "success", fmt.Sprintf("手动合并 %d 个文件", len(validFiles)))
		onProgress(fmt.Sprintf("✅ 手动合并 %d 个文件完成", len(validFiles)))
		return nil
	}

	return fmt.Errorf("合并失败")
}

func (s *MergeService) CleanupTempFiles() int {
	root := s.config.TargetDir
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

		// Collect valid MP4 base names for orphan detection
		mp4Bases := make(map[string]bool)
		for _, fe := range fileEntries {
			if fe.IsDir() {
				continue
			}
			name := fe.Name()
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".mp4" {
				if info, err := fe.Info(); err == nil && info.Size() >= 10240 {
					base := strings.TrimSuffix(name, ext)
					mp4Bases[base] = true
				}
			}
		}

		for _, fe := range fileEntries {
			name := fe.Name()

			// Clean up concat temp files
			if strings.HasPrefix(name, ".concat_") && strings.HasSuffix(name, ".txt") {
				path := filepath.Join(folder, name)
				if err := os.Remove(path); err == nil {
					cleaned++
					s.logToFile("merge", fmt.Sprintf("🗑 清理临时文件: %s", name))
				}
			}

			// Clean up MP4/FLV temp files (.tmp suffix)
			if strings.HasSuffix(name, ".tmp") && (strings.HasSuffix(name, ".mp4.tmp") || strings.HasSuffix(name, ".flv.tmp")) {
				path := filepath.Join(folder, name)
				if err := os.Remove(path); err == nil {
					cleaned++
					s.logToFile("merge", fmt.Sprintf("🗑 清理临时文件: %s", name))
				}
			}

			// Clean orphaned .ts files: no corresponding valid MP4 with same base name
			if fe.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".ts" && !strings.HasPrefix(name, ".") {
				base := strings.TrimSuffix(name, ext)
				if !mp4Bases[base] {
					path := filepath.Join(folder, name)
					if info, err := fe.Info(); err == nil && info.Size() >= 10240 {
						if err := utils.SafeUnlink(path); err == nil {
							cleaned++
							s.logToFile("merge", fmt.Sprintf("🗑 清理孤立TS: %s", name))
						}
					}
				}
			}
		}

		// Clean crash-residual .merge_tmp_* directories
		tmpEntries, _ := os.ReadDir(folder)
		for _, te := range tmpEntries {
			if te.IsDir() && strings.HasPrefix(te.Name(), ".merge_tmp_") {
				tmpPath := filepath.Join(folder, te.Name())
				if err := os.RemoveAll(tmpPath); err == nil {
					cleaned++
					s.logToFile("merge", fmt.Sprintf("🗑 清理残留临时目录: %s", te.Name()))
				}
			}
		}
	}
	return cleaned
}
