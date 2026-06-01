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

type HistoryService struct {
	config  *config.Config
	logger  *zap.Logger
	mu      sync.RWMutex
	cache   []model.HistoryRecord
	loaded  bool
	loadErr error
	once    sync.Once
}

func NewHistoryService(config *config.Config, logger *zap.Logger) *HistoryService {
	return &HistoryService{config: config, logger: logger}
}

func (s *HistoryService) Add(task, streamer, status, detail string) {
	s.AddWithStats(task, streamer, status, 0, 0, 0, 0, detail)
}

func (s *HistoryService) AddWithStats(task, streamer, status string, filesCount int, freedBytes, mergedBytes int64, duration float64, detail string) {
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

	s.ensureLoaded()
	records := append([]model.HistoryRecord{}, s.cache...)
	records = append(records, record)
	records = s.cleanupRecords(records)
	s.saveRecords(records)
}

func (s *HistoryService) GetRecords(task string, page, perPage int) ([]model.HistoryRecord, int) {
	s.mu.RLock()
	s.ensureLoaded()
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

func (s *HistoryService) GetAllRecords() []model.HistoryRecord {
	s.mu.RLock()
	s.ensureLoaded()
	records := append([]model.HistoryRecord{}, s.cache...)
	s.mu.RUnlock()
	if records == nil {
		return []model.HistoryRecord{}
	}
	return records
}

// ensureLoaded loads records from disk if not already loaded.
// Caller MUST hold at least RLock.
func (s *HistoryService) ensureLoaded() {
	if s.loaded {
		return
	}
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
	os.MkdirAll(filepath.Dir(file), 0755)
	wrapper := struct {
		Records []model.HistoryRecord `json:"records"`
	}{Records: records}
	data, _ := json.MarshalIndent(wrapper, "", "  ")
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err == nil {
		if err := os.Rename(tmp, file); err == nil {
			s.cache = records // only update cache after successful disk write
			return
		}
		os.Remove(tmp)
	} else {
		os.Remove(tmp)
	}
}

func (s *HistoryService) cleanupRecords(records []model.HistoryRecord) []model.HistoryRecord {
	cutoff := time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	var cleaned []model.HistoryRecord
	for _, r := range records {
		if r.Time > cutoff {
			cleaned = append(cleaned, r)
		}
	}
	if len(cleaned) > 100 {
		cleaned = cleaned[len(cleaned)-100:]
	}
	return cleaned
}

func (s *HistoryService) CleanupOldRecords() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureLoaded()
	records := append([]model.HistoryRecord{}, s.cache...)
	cleaned := s.cleanupRecords(records)
	if len(cleaned) != len(records) {
		s.saveRecords(cleaned)
	}
}

// Reload forces a fresh read from disk on next access.
func (s *HistoryService) Reload() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loaded = false
	s.cache = nil
}
