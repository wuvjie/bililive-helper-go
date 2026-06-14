// history.go 提供历史记录的持久化存储服务。
// 支持记录的增删查、分页查询、磁盘文件的原子写入和自动清理。
package service

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"bililive-helper/internal/config"
	"bililive-helper/internal/fsutil"
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
func (s *HistoryService) Add(task, streamer, status, detail, logID string) {
	s.AddWithStats(task, streamer, status, 0, 0, 0, 0, detail, logID)
}

// AddWithStats 添加一条带统计数据的历史记录（文件数、释放/合并字节数、耗时）。
// 采用乐观写入策略：先更新内存缓存，再持久化到磁盘。
// 磁盘写入失败时记录保留在内存中，下次成功写入时自动持久化。
func (s *HistoryService) AddWithStats(task, streamer, status string, filesCount int, freedBytes, mergedBytes int64, duration float64, detail, logID string) {
	s.ensureLoadedSafe()
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	record := model.HistoryRecord{
		ID:          fmt.Sprintf("%s_%s_%d", now.Format("20060102_150405"), task, now.UnixNano()),
		Time:        now.Format("2006-01-02 15:04:05"),
		Task:        task,
		Streamer:    streamer,
		Status:      status,
		FilesCount:  filesCount,
		FreedBytes:  freedBytes,
		MergedBytes: mergedBytes,
		Duration:    duration,
		Detail:      detail,
		LogID:       logID,
	}
	if record.Streamer == "" {
		record.Streamer = "全局"
	}

	records := append([]model.HistoryRecord{}, s.cache...)
	records = append(records, record)
	records = s.cleanupRecords(records)
	// 先更新内存缓存，确保即使磁盘写入失败记录也不丢失
	s.cache = records
	s.saveRecords(records)
}

// GetRecords 分页查询历史记录，支持按任务类型和主播名过滤，按时间倒序排列。
func (s *HistoryService) GetRecords(task, streamer string, page, perPage int) ([]model.HistoryRecord, int) {
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

	if streamer != "" {
		lower := strings.ToLower(streamer)
		var filtered []model.HistoryRecord
		for _, r := range records {
			if strings.Contains(strings.ToLower(r.Streamer), lower) {
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
		// C4: 备份损坏文件，避免数据永久丢失
		backupFile := file + ".corrupt." + time.Now().Format("20060102150405")
		if bErr := os.WriteFile(backupFile, data, 0600); bErr != nil {
			s.logger.Warn("备份损坏历史文件失败", zap.Error(bErr))
		} else {
			s.logger.Warn("历史记录解析失败，已备份损坏文件",
				zap.String("backup", backupFile), zap.Error(err))
		}
		s.loaded = true
		s.cache = nil
		return
	}
	s.cache = wrapper.Records
	s.loaded = true
}

// saveRecords 将历史记录原子写入磁盘（write → fsync → rename）。
// 成功后更新内存缓存。注意：AddWithStats 已采用乐观写入策略（先更新缓存再调用本函数），
// 因此本函数的"失败不更新缓存"仅对 CleanupOldRecords 等直接调用者生效。
// 调用者必须持有写锁。
func (s *HistoryService) saveRecords(records []model.HistoryRecord) {
	file := s.config.GetHistoryFile()
	wrapper := struct {
		Records []model.HistoryRecord `json:"records"`
	}{Records: records}
	data, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		s.logger.Warn("序列化历史记录失败", zap.Error(err))
		return
	}
	if err := fsutil.AtomicSave(file, data, 0600); err != nil {
		s.logger.Warn("原子写入历史记录失败", zap.Error(err))
		return
	}
	// 原子写入成功后才更新内存缓存
	s.cache = records
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
