// Package model 定义应用程序的核心数据模型。
// 包含历史记录、调度配置和任务状态等数据结构。
package model

// HistoryRecord 表示一条操作历史记录（合并、清理、配置变更等）。
type HistoryRecord struct {
	ID          string  `json:"id"`
	Time        string  `json:"time"`
	Task        string  `json:"task"`
	Streamer    string  `json:"streamer"`
	Status      string  `json:"status"`
	FilesCount  int     `json:"files_count"`
	FreedBytes  int64   `json:"freed_bytes"`
	MergedBytes int64   `json:"merged_bytes"`
	Duration    float64 `json:"duration"`
	Detail      string  `json:"detail"`
}

// ScheduleConfig 表示定时调度配置（合并/清理间隔和启用状态）。
type ScheduleConfig struct {
	MergeInterval int  `json:"merge_interval"`
	CleanInterval int  `json:"clean_interval"`
	MergeEnabled  bool `json:"merge_enabled"`
	CleanEnabled  bool `json:"clean_enabled"`
}

// ScheduleStatus 表示调度器的运行状态（各任务的启用状态、执行时间等）。
type ScheduleStatus struct {
	Running       bool       `json:"running"`
	MergeEnabled  bool       `json:"merge_enabled"`
	MergeInterval int        `json:"merge_interval"`
	CleanEnabled  bool       `json:"clean_enabled"`
	CleanInterval int        `json:"clean_interval"`
	Merge         TaskStatus `json:"merge"`
	Clean         TaskStatus `json:"clean"`
}

// TaskStatus 表示单个任务的运行状态。
type TaskStatus struct {
	Enabled   bool    `json:"enabled"`
	Interval  int     `json:"interval"`
	LastRun   float64 `json:"last_run"`
	NextRun   float64 `json:"next_run"`
	IsRunning bool    `json:"is_running"`
}
