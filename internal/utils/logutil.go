// Package utils 提供通用工具函数。
// 包含磁盘查询、文件操作、视频解析、字符串处理和 Webhook 通知等功能。
package utils

import (
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// FormatSize 将字节数格式化为人类可读的大小字符串（KB/MB/GB）。
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	default:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	}
}

// LogStartup 打印系统配置和 FFmpeg 可用性的启动摘要日志。
func LogStartup(logger *zap.Logger, cfgTargetDir string, triggerThreshold, targetThreshold float64, gapMinutes, mergeAgeMinutes, backupStartH, backupStartM, backupEndH, backupEndM int, safeMode string, safeDays, port int) {
	logger.Info("═══ Bililive Helper 启动 ═══")

	ffmpegOK, ffmpegVer := false, ""
	videoEncoder := "libx264"
	if _, err := exec.LookPath("ffmpeg"); err == nil {
		ffmpegOK = true
		if out, err := exec.Command("ffmpeg", "-version").Output(); err == nil {
			lines := strings.SplitN(string(out), "\n", 2)
			if len(lines) > 0 {
				ffmpegVer = strings.TrimSpace(lines[0])
				if len(ffmpegVer) > 50 {
					ffmpegVer = ffmpegVer[:50]
				}
			}
		}
		if out, err := exec.Command("ffmpeg", "-encoders").Output(); err == nil {
			if strings.Contains(string(out), "h264_rkmpp") {
				videoEncoder = "h264_rkmpp"
			}
		}
	}

	logger.Info("系统摘要",
		zap.Int("port", port),
		zap.String("target_dir", cfgTargetDir),
		zap.Bool("ffmpeg", ffmpegOK),
		zap.String("ffmpeg_ver", ffmpegVer),
		zap.String("encoder", videoEncoder),
		zap.Float64("trigger", triggerThreshold),
		zap.Float64("target", targetThreshold),
		zap.Int("gap_min", gapMinutes),
		zap.Int("merge_age_min", mergeAgeMinutes),
		zap.String("safe_mode", safeMode),
		zap.Int("safe_days", safeDays),
		zap.String("quiet_window", fmt.Sprintf("%02d:%02d-%02d:%02d", backupStartH, backupStartM, backupEndH, backupEndM)),
	)
	logger.Info("═══════════════════════════")
}
