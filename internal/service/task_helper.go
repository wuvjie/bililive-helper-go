// task_helper.go 提取 MergeService 和 CleanService 共享的启动样板代码。
package service

import (
	"fmt"
	"os"
	"path/filepath"

	"bililive-helper-go/internal/config"

	"go.uber.org/zap"
)

// TaskSetup 封装任务启动时的通用初始化逻辑。
// 包括：配置快照、操作日志创建、进度回调包装、静默时段检查、路径存在性检查。
type TaskSetup struct {
	Cfg      config.Config
	OpLog    *OpLogger
	Progress ProgressFunc
	LogID    string
	Tag      string // "[全局]" 或 "[主播名]"
}

// PrepareTask 执行任务启动前的通用准备工作。
// taskType: 英文任务类型（"merge"/"clean"），用于 logID 文件名；
// displayName: 中文显示名（"合并"/"清理"），用于错误消息；
// streamer 为空表示全局模式。
func PrepareTask(cfg *config.Config, logger *zap.Logger, logDir, taskType, displayName, streamer string, onProgress ProgressFunc) (*TaskSetup, error) {
	snap := cfg.Snapshot()

	// 创建操作日志（失败时降级为 nil）
	opLog, err := NewOpLogger(filepath.Join(snap.LogDir, logDir), taskType)
	if err != nil {
		opLog = nil
	}

	// 包装进度回调
	if onProgress == nil {
		onProgress = func(string) {}
	}
	progress := opLog.ProgressFunc(onProgress)

	// 静默时段检查
	if snap.IsBackupWindow() {
		opLog.Close()
		return nil, fmt.Errorf("当前处于静默时段（%d:%02d-%d:%02d），%s暂停",
			snap.BackupStartHour, snap.BackupStartMinute,
			snap.BackupEndHour, snap.BackupEndMinute,
			displayName)
	}

	// 路径存在性检查
	if _, err := os.Stat(snap.TargetDir); os.IsNotExist(err) {
		opLog.Close()
		return nil, fmt.Errorf("路径不存在: %s", snap.TargetDir)
	}

	tag := "[全局]"
	if streamer != "" {
		tag = fmt.Sprintf("[%s]", streamer)
	}

	return &TaskSetup{
		Cfg:      snap,
		OpLog:    opLog,
		Progress: progress,
		LogID:    opLog.LogID(),
		Tag:      tag,
	}, nil
}
