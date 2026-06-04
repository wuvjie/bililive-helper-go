// scheduler.go 提供定时任务调度器。
// 按可配置间隔自动触发合并和清理任务，支持静默时段和手动触发。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bililive-helper/internal/config"
	"bililive-helper/internal/model"
	"bililive-helper/internal/utils"

	"go.uber.org/zap"
)

// SchedulerService 提供定时任务调度功能。
// 按可配置间隔自动触发合并和清理任务，支持静默时段、手动触发和每日历史清理。
type SchedulerService struct {
	config  *config.Config
	logger  *zap.Logger
	merge   *MergeService
	clean   *CleanService
	history *HistoryService

	tickCh       chan struct{}
	stopCh       chan struct{}
	stopOnce     sync.Once
	startOnce    sync.Once
	wg           sync.WaitGroup
	lastRun      map[string]time.Time
	running      map[string]bool
	scheduleMu   sync.RWMutex
	scheduleConf model.ScheduleConfig
	mu           sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewSchedulerService 创建调度服务实例。
// 初始化调度配置，设置 lastRun 为当前时间（冷启动后需等待一个完整间隔才执行）。
func NewSchedulerService(config *config.Config, logger *zap.Logger, merge *MergeService, clean *CleanService, history *HistoryService) *SchedulerService {
	ctx, cancel := context.WithCancel(context.Background())

	s := &SchedulerService{
		config:  config,
		logger:  logger,
		merge:   merge,
		clean:   clean,
		history: history,
		tickCh:  make(chan struct{}, 1),
		stopCh:  make(chan struct{}),
		running: make(map[string]bool),
		ctx:     ctx,
		cancel:  cancel,
	}
	s.scheduleConf = s.loadSchedule()

	// 初始化 lastRun 为当前时间 — 首次任务需等待一个完整间隔后才执行
	now := time.Now()
	s.lastRun = map[string]time.Time{
		"merge": now,
		"clean": now,
	}
	return s
}

// Start 启动调度器主循环。
func (s *SchedulerService) Start() {
	s.startOnce.Do(func() {
		go s.loop()
		s.logger.Info("调度器启动")
	})
}

// Stop 停止调度器，等待所有正在运行的任务完成。
func (s *SchedulerService) Stop() {
	s.stopOnce.Do(func() {
		s.cancel()
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *SchedulerService) loop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// 每日午夜自动清理过期历史记录
	var lastDay string

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.runDueTasks()

			// Daily cleanup
			today := time.Now().Format("2006-01-02")
			if today != lastDay {
				lastDay = today
				go s.history.CleanupOldRecords()
			}
		case <-s.tickCh:
			s.runDueTasks()
		}
	}
}

// runDueTasks 检查到期的任务并触发执行。
// 静默时段内跳过所有任务。
func (s *SchedulerService) runDueTasks() {
	schedule := s.getSchedule()
	now := time.Now()
	cfg := s.config.Snapshot()
	s.mu.Lock()
	defer s.mu.Unlock()

	// 静默时段内跳过所有任务
	if cfg.IsBackupWindow() {
		return
	}

	if schedule.MergeEnabled {
		elapsed := now.Sub(s.lastRun["merge"])
		if elapsed < 0 {
			elapsed = 0
		}
		if elapsed >= time.Duration(schedule.MergeInterval)*time.Minute {
			if !s.running["merge"] {
				s.running["merge"] = true
				s.wg.Add(1)
				go s.runTask("merge")
			}
		}
	}
	if schedule.CleanEnabled {
		elapsed := now.Sub(s.lastRun["clean"])
		if elapsed < 0 {
			elapsed = 0
		}
		if elapsed >= time.Duration(schedule.CleanInterval)*time.Minute {
			if !s.running["clean"] {
				s.running["clean"] = true
				s.wg.Add(1)
				go s.runTask("clean")
			}
		}
	}
}

// RunTask 手动触发指定任务（merge 或 clean）。
// 如果任务正在运行中返回错误。
func (s *SchedulerService) RunTask(task string) error {
	if task != "merge" && task != "clean" {
		return fmt.Errorf("无效任务: %s", task)
	}
	s.mu.Lock()
	if s.running[task] {
		s.mu.Unlock()
		return fmt.Errorf("%s 正在运行中", task)
	}
	s.running[task] = true
	s.wg.Add(1)
	s.mu.Unlock()
	go s.runTask(task)
	return nil
}

// runTask 执行调度任务并在完成后更新状态。
func (s *SchedulerService) runTask(task string) {
	defer func() {
		s.mu.Lock()
		s.running[task] = false
		s.lastRun[task] = time.Now()
		s.mu.Unlock()
		s.wg.Done()
	}()

	s.logToFile(task, fmt.Sprintf("▶ 调度触发 → %s", map[string]string{"merge": "合并", "clean": "清理"}[task]))

	switch task {
	case "merge":
		res, err := s.merge.Run(s.ctx, "", nil)
		if err != nil {
			s.logToFile(task, fmt.Sprintf("❌ 合并失败: %v", err))
		} else if res != nil && res.Done > 0 {
			utils.NotifyWebhook(fmt.Sprintf("自动合并完成：%d 场次 (%.1f GB)", res.Done, res.TotalGB))
		}
	case "clean":
		res, err := s.clean.Run(s.ctx, "", nil)
		if err != nil {
			s.logToFile(task, fmt.Sprintf("❌ 清理失败: %v", err))
		} else if res != nil && res.Deleted > 0 {
			utils.NotifyWebhook(fmt.Sprintf("自动清理完成：%d 文件，释放 %s", res.Deleted, utils.FormatSize(res.Freed)))
		}
	}
}

func (s *SchedulerService) logToFile(task, message string) {
	// Use the same log rotation as merge/clean tasks
	logToFile(s.config.LogDir, task, message, s.logger)
}

// GetStatus 返回当前调度状态（各任务的启用状态、间隔、上次/下次执行时间、是否运行中）。
func (s *SchedulerService) GetStatus() model.ScheduleStatus {
	schedule := s.getSchedule()
	s.mu.Lock()
	defer s.mu.Unlock()

	hasRunning := s.running["merge"] || s.running["clean"]

	return model.ScheduleStatus{
		Running:       hasRunning,
		MergeEnabled:  schedule.MergeEnabled,
		MergeInterval: schedule.MergeInterval,
		CleanEnabled:  schedule.CleanEnabled,
		CleanInterval: schedule.CleanInterval,
		Merge: model.TaskStatus{
			Enabled:   schedule.MergeEnabled,
			Interval:  schedule.MergeInterval,
			LastRun:   float64(s.lastRun["merge"].Unix()),
			NextRun:   float64(s.lastRun["merge"].Add(time.Duration(schedule.MergeInterval) * time.Minute).Unix()),
			IsRunning: s.running["merge"],
		},
		Clean: model.TaskStatus{
			Enabled:   schedule.CleanEnabled,
			Interval:  schedule.CleanInterval,
			LastRun:   float64(s.lastRun["clean"].Unix()),
			NextRun:   float64(s.lastRun["clean"].Add(time.Duration(schedule.CleanInterval) * time.Minute).Unix()),
			IsRunning: s.running["clean"],
		},
	}
}

// SaveSchedule 保存调度配置到文件，并立即通知调度循环重新评估。
func (s *SchedulerService) SaveSchedule(schedule model.ScheduleConfig) error {
	schedule.MergeInterval = max(10, min(1440, schedule.MergeInterval))
	schedule.CleanInterval = max(10, min(1440, schedule.CleanInterval))

	file := s.config.GetScheduleFile()
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return fmt.Errorf("创建调度目录失败: %w", err)
	}
	data, err := json.MarshalIndent(schedule, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化调度配置失败: %w", err)
	}
	// Atomic write: tmp file then rename to prevent corruption on crash
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, file); err != nil {
		os.Remove(tmp)
		return err
	}

	// 更新内存缓存，下次 tick 立即生效
	s.scheduleMu.Lock()
	s.scheduleConf = schedule
	s.scheduleMu.Unlock()

	// 通知调度循环立即重新评估到期任务
	select {
	case s.tickCh <- struct{}{}:
	default:
	}

	return nil
}

func (s *SchedulerService) getSchedule() model.ScheduleConfig {
	s.scheduleMu.RLock()
	defer s.scheduleMu.RUnlock()
	return s.scheduleConf
}

func (s *SchedulerService) loadSchedule() model.ScheduleConfig {
	defaultSchedule := model.ScheduleConfig{
		MergeInterval: 360,
		CleanInterval: 720,
		MergeEnabled:  true,
		CleanEnabled:  true,
	}
	file := s.config.GetScheduleFile()
	data, err := os.ReadFile(file)
	if err != nil {
		return defaultSchedule
	}
	var schedule model.ScheduleConfig
	if err := json.Unmarshal(data, &schedule); err != nil {
		s.logger.Warn("调度配置解析失败，使用默认值", zap.Error(err), zap.String("file", file))
		return defaultSchedule
	}
	return schedule
}
