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

// StartupInfo 封装启动摘要所需的所有配置信息。
type StartupInfo struct {
	Port              int
	TargetDir         string
	TriggerThreshold  float64
	TargetThreshold   float64
	GapMinutes        int
	MergeAgeMinutes   int
	BackupStartHour   int
	BackupStartMinute int
	BackupEndHour     int
	BackupEndMinute   int
	SafeMode          string
	SafeDays          int
}

// LogStartup 打印系统配置和 FFmpeg 可用性的启动摘要日志。
func LogStartup(logger *zap.Logger, info StartupInfo) {
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
		zap.Int("port", info.Port),
		zap.String("target_dir", info.TargetDir),
		zap.Bool("ffmpeg", ffmpegOK),
		zap.String("ffmpeg_ver", ffmpegVer),
		zap.String("encoder", videoEncoder),
		zap.Float64("trigger", info.TriggerThreshold),
		zap.Float64("target", info.TargetThreshold),
		zap.Int("gap_min", info.GapMinutes),
		zap.Int("merge_age_min", info.MergeAgeMinutes),
		zap.String("safe_mode", info.SafeMode),
		zap.Int("safe_days", info.SafeDays),
		zap.String("quiet_window", fmt.Sprintf("%02d:%02d-%02d:%02d", info.BackupStartHour, info.BackupStartMinute, info.BackupEndHour, info.BackupEndMinute)),
	)
	logger.Info("═══════════════════════════")
}
