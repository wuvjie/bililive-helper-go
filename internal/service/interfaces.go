// interfaces.go 定义 service 层的核心接口。
package service

import (
	"context"
	"time"
)

// TaskRunner 是所有可调度任务的通用接口。
// MergeService 和 CleanService 各自实现此接口，
// SchedulerService 通过注册 TaskRunner 列表来调度任务，
// 替代原来的硬编码 switch "merge"/"clean"。
type TaskRunner interface {
	// Name 返回任务名称（如 "merge"、"clean"），用于日志和调度。
	Name() string
	// Run 执行任务，通过 progress 回调报告进度。
	// 返回任务结果和错误。
	Run(ctx context.Context, streamer string, progress ProgressFunc) (*TaskResult, error)
}

// TaskResult 封装任务执行结果。
type TaskResult struct {
	TaskType  string        // 任务类型（merge/clean）
	Streamer  string        // 目标主播（空表示全局）
	Processed int           // 处理的文件数
	Duration  time.Duration // 执行耗时
	Details   string        // 摘要信息
	LogID     string        // 操作日志 ID
}
