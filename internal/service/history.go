// history.go 提供历史记录的持久化存储服务。
// 支持记录的增删查、分页查询、磁盘文件的原子写入和自动清理。
package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"bililive-helper/internal/config"
	"bililive-helper/internal/model"

	"go.uber.org/zap"
)

// HistoryService 提供历史记录的持久化存储。
// 使用内存缓存 + 磁盘文件的模式，首次访问时从磁盘加载，所有变更原子写入。
type HistoryService struct {
	config *config.Config
	logger *zap.Logger
	mu     sync.RWMutex
	cache  []model.HistoryRecord
	loaded bool
}

// NewHistoryService 创建历史记录服务实例。
func NewHistoryService(config *config.Config, logger *zap.Logger) *HistoryService {
	return &HistoryService{config: config, logger: logger}
}

// Add 添加一条简要历史记录（无统计数据）。
func (s *HistoryService) Add(task, streamer, status, detail string) {
	s.AddWithStats(task, streamer, status, 0, 0, 0, 0, detail)
}

// AddWithStats 添加一条带统计数据的历史记录（文件数、释放/合并字节数、耗时）。
func (s *HistoryService) AddWithStats(task, streamer, status string, filesCount int, freedBytes, mergedBytes int64, duration float64, detail string) {
	s.ensureLoadedSafe()
	s.mu.Lock()
	defer s.mu.Unlock()

	record := model.HistoryRecord{
		ID:          fmt.Sprintf("%s_%s_%d", time.Now().Format("20060102_150405"), task, time.Now().UnixNano()),
		Time:        time.Now().Format("2006-01-02 15:04:05"),
		Task:        task,
		Streamer:    streamer,
		Status:      status,
		FilesCount:  filesCount,
		FreedBytes:  freedBytes,
		MergedBytes: mergedBytes,
		Duration:    duration,
		Detail:      detail,
	}
	if record.Streamer == "" {
		record.Streamer = "全局"
	}

	records := append([]model.HistoryRecord{}, s.cache...)
	records = append(records, record)
	records = s.cleanupRecords(records)
	s.saveRecords(records)
}

// GetRecords 分页查询历史记录，支持按任务类型过滤，按时间倒序排列。
func (s *HistoryService) GetRecords(task string, page, perPage int) ([]model.HistoryRecord, int) {
	s.ensureLoadedSafe()
	s.mu.RLock()
	records := append([]model.HistoryRecord{}, s.cache...) // copy under lock
	s.mu.RUnlock()

	if records == nil {
		records = []model.HistoryRecord{}
	}

	if task != "" {
		var filtered []model.HistoryRecord
		for _, r := range records {
			if r.Task == task {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Time > records[j].Time
	})

	total := len(records)
	start := (page - 1) * perPage
	if start >= total {
		return []model.HistoryRecord{}, total
	}
	end := start + perPage
	if end > total {
		end = total
	}
	return records[start:end], total
}

// GetAllRecords 返回全部历史记录（用于导出和统计）。
func (s *HistoryService) GetAllRecords() []model.HistoryRecord {
	s.ensureLoadedSafe()
	s.mu.RLock()
	records := append([]model.HistoryRecord{}, s.cache...)
	s.mu.RUnlock()
	if records == nil {
		return []model.HistoryRecord{}
	}
	return records
}

// ensureLoadedSafe 使用双重检查锁模式确保历史记录在首次访问时从磁盘加载。
func (s *HistoryService) ensureLoadedSafe() {
	s.mu.RLock()
	loaded := s.loaded
	s.mu.RUnlock()
	if loaded {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loaded {
		return
	}
	s.doLoad()
}

// doLoad 从磁盘读取历史记录文件。调用者必须持有写锁。
func (s *HistoryService) doLoad() {
	file := s.config.GetHistoryFile()
	data, err := os.ReadFile(file)
	if err != nil {
		s.loaded = true
		s.cache = nil
		return
	}
	var wrapper struct {
		Records []model.HistoryRecord `json:"records"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		s.logger.Warn("历史记录解析失败，重建空记录", zap.Error(err))
		s.loaded = true
		s.cache = nil
		return
	}
	s.cache = wrapper.Records
	s.loaded = true
}

func (s *HistoryService) saveRecords(records []model.HistoryRecord) {
	file := s.config.GetHistoryFile()
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		s.logger.Warn("创建历史记录目录失败", zap.Error(err))
		return
	}
	wrapper := struct {
		Records []model.HistoryRecord `json:"records"`
	}{Records: records}
	data, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		s.logger.Warn("序列化历史记录失败", zap.Error(err))
		return
	}
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err == nil {
		if err := os.Rename(tmp, file); err == nil {
			// 原子 rename 成功后才更新内存缓存
			s.cache = records
			return
		}
		os.Remove(tmp)
	} else {
		os.Remove(tmp)
	}
}

// cleanupRecords 清理超过 90 天的记录，并限制最大记录数为 1000。
func (s *HistoryService) cleanupRecords(records []model.HistoryRecord) []model.HistoryRecord {
	cutoff := time.Now().AddDate(0, 0, -90).Format("2006-01-02 15:04:05")
	var cleaned []model.HistoryRecord
	for _, r := range records {
		if r.Time > cutoff {
			cleaned = append(cleaned, r)
		}
	}
	if len(cleaned) > 1000 {
		cleaned = cleaned[len(cleaned)-1000:]
	}
	return cleaned
}

// CleanupOldRecords 执行历史记录清理（由调度器每日调用）。
func (s *HistoryService) CleanupOldRecords() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.loaded {
		s.doLoad()
	}
	records := append([]model.HistoryRecord{}, s.cache...)
	cleaned := s.cleanupRecords(records)
	if len(cleaned) != len(records) {
		s.saveRecords(cleaned)
	}
}
