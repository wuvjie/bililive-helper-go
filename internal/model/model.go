package model

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

type ScheduleConfig struct {
	MergeInterval int  `json:"merge_interval"`
	CleanInterval int  `json:"clean_interval"`
	MergeEnabled  bool `json:"merge_enabled"`
	CleanEnabled  bool `json:"clean_enabled"`
}

type ScheduleStatus struct {
	Running       bool       `json:"running"`
	MergeEnabled  bool       `json:"merge_enabled"`
	MergeInterval int        `json:"merge_interval"`
	CleanEnabled  bool       `json:"clean_enabled"`
	CleanInterval int        `json:"clean_interval"`
	Merge         TaskStatus `json:"merge"`
	Clean         TaskStatus `json:"clean"`
}

type TaskStatus struct {
	Enabled   bool    `json:"enabled"`
	Interval  int     `json:"interval"`
	LastRun   float64 `json:"last_run"`
	NextRun   float64 `json:"next_run"`
	IsRunning bool    `json:"is_running"`
}
