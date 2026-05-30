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

type SchedulerService struct {
	config  *config.Config
	logger  *zap.Logger
	merge   *MergeService
	clean   *CleanService
	history *HistoryService

	tickCh       chan struct{}
	stopCh       chan struct{}
	wg           sync.WaitGroup
	lastRun      map[string]time.Time
	running      map[string]bool
	scheduleMu   sync.RWMutex
	scheduleConf model.ScheduleConfig
	mu           sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
}

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

	// Initialize lastRun to now — first task runs after full interval
	now := time.Now()
	s.lastRun = map[string]time.Time{
		"merge": now,
		"clean": now,
	}
	return s
}

func (s *SchedulerService) Start() {
	go s.loop()
	s.logger.Info("调度器启动")
}

func (s *SchedulerService) Stop() {
	s.cancel()
	close(s.stopCh)
	s.wg.Wait()
}

func (s *SchedulerService) loop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// Daily cleanup at midnight
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

func (s *SchedulerService) runDueTasks() {
	schedule := s.getSchedule()
	now := time.Now()
	cfg := s.config.Snapshot()
	s.mu.Lock()
	defer s.mu.Unlock()

	// Skip all tasks during quiet window
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

// runTask executes a scheduled task and updates status on completion.
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
		res, err := s.clean.Run("", nil)
		if err != nil {
			s.logToFile(task, fmt.Sprintf("❌ 清理失败: %v", err))
		} else if res != nil && res.Deleted > 0 {
			utils.NotifyWebhook(fmt.Sprintf("自动清理完成：%d 文件，释放 %s", res.Deleted, utils.FormatSize(res.Freed)))
		}
	}
}

func (s *SchedulerService) logToFile(task, message string) {
	// Use the same log rotation as merge/clean
	logToFile(s.config.LogDir, task, message, s.logger)
}

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

func (s *SchedulerService) SaveSchedule(schedule model.ScheduleConfig) error {
	schedule.MergeInterval = max(10, min(1440, schedule.MergeInterval))
	schedule.CleanInterval = max(10, min(1440, schedule.CleanInterval))

	file := s.config.GetScheduleFile()
	os.MkdirAll(filepath.Dir(file), 0755)
	data, err := json.MarshalIndent(schedule, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化调度配置失败: %w", err)
	}
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, file); err != nil {
		os.Remove(tmp)
		return err
	}

	// Update in-memory cache so next tick picks up changes immediately
	s.scheduleMu.Lock()
	s.scheduleConf = schedule
	s.scheduleMu.Unlock()

	// Nudge the loop to re-evaluate sooner
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
		MergeInterval: 120,
		CleanInterval: 360,
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
		return defaultSchedule
	}
	return schedule
}
