package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Config struct {
	TargetDir          string   `json:"TARGET_DIR"`
	TriggerThreshold   float64  `json:"TRIGGER_THRESHOLD"`
	TargetThreshold    float64  `json:"TARGET_THRESHOLD"`
	MinKeepPerStreamer int      `json:"MIN_KEEP_PER_STREAMER"`
	SafeAgeMinutes     int      `json:"SAFE_AGE_MINUTES"`
	GapMinutes         int      `json:"GAP_MINUTES"`
	MergeAgeMinutes    int      `json:"MERGE_AGE_MINUTES"`
	WhitelistKeywords  []string `json:"WHITELIST_KEYWORDS"`
	SafeMode           string   `json:"SAFE_MODE"`
	SafeDays           int      `json:"SAFE_DAYS"`
	MaxDeletePerRun    int      `json:"MAX_DELETE_PER_RUN"`
	BackupStartHour    int      `json:"BACKUP_START_HOUR"`
	BackupStartMinute  int      `json:"BACKUP_START_MINUTE"`
	BackupEndHour      int      `json:"BACKUP_END_HOUR"`
	BackupEndMinute    int      `json:"BACKUP_END_MINUTE"`
	Port               int      `json:"PORT"`
	Password           string   `json:"PASSWORD"`
	SecretKey          string   `json:"SECRET_KEY"`
	LogDir             string   `json:"LOG_DIR"`
	ConfigFile         string   `json:"-"`
}

var (
	defaultConfig = Config{
		TargetDir:          "/vol2/1000/video/bililive-go/抖音",
		TriggerThreshold:   80,
		TargetThreshold:    60,
		MinKeepPerStreamer: 3,
		SafeAgeMinutes:     120,
		GapMinutes:         60,
		MergeAgeMinutes:    30,
		WhitelistKeywords:  []string{"留存", "纪念", "高能", "生日", "勿删"},
		SafeMode:           "hours",
		SafeDays:           1,
		MaxDeletePerRun:    10,
		BackupStartHour:    4,
		BackupStartMinute:  0,
		BackupEndHour:      12,
		BackupEndMinute:    0,
		Port:               5000,
		Password:           "",  // generated on first run
		SecretKey:          "",  // generated on first run
		LogDir:             "/vol1/1000/docker/bililive-helper-go",
	}
	mu sync.RWMutex
)

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Load() *Config {
	cfg := defaultConfig
	// Apply env overrides first so ConfigFile points to the correct path
	if v := os.Getenv("TARGET_DIR"); v != "" {
		cfg.TargetDir = v
	}
	if v := os.Getenv("PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("LOG_DIR"); v != "" {
		cfg.LogDir = v
	}
	if v := os.Getenv("SECRET_KEY"); v != "" {
		cfg.SecretKey = v
	}
	cfg.ConfigFile = filepath.Join(cfg.LogDir, "config.json")
	if data, err := os.ReadFile(cfg.ConfigFile); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			fmt.Printf("[WARN] 配置文件解析失败，使用默认配置: %v\n", err)
		}
	}
	// Generate random credentials on first run (no config file, no env var)
	firstRun := false
	if cfg.Password == "" {
		cfg.Password = randomHex(12) // 24-char hex, 96-bit entropy
		firstRun = true
	}
	if cfg.SecretKey == "" {
		cfg.SecretKey = randomHex(16) // 32-char hex
		firstRun = true
	}
	if firstRun {
		fmt.Printf("═══ 首次启动 ═══\n")
		fmt.Printf("登录密码: %s\n", cfg.Password)
		fmt.Printf("请登录后及时修改密码\n")
		fmt.Printf("═══════════════\n")
	}
	return &cfg
}

func (c *Config) Validate() error {
	if c.TriggerThreshold < 0 || c.TriggerThreshold > 100 {
		return fmt.Errorf("TRIGGER_THRESHOLD 必须在 0-100 之间")
	}
	if c.TargetThreshold < 0 || c.TargetThreshold > 100 {
		return fmt.Errorf("TARGET_THRESHOLD 必须在 0-100 之间")
	}
	if c.TargetThreshold >= c.TriggerThreshold {
		return fmt.Errorf("TARGET_THRESHOLD 必须小于 TRIGGER_THRESHOLD")
	}
	if c.MinKeepPerStreamer < 1 {
		return fmt.Errorf("MIN_KEEP_PER_STREAMER 不能小于 1")
	}
	if c.SafeAgeMinutes < 0 {
		return fmt.Errorf("SAFE_AGE_MINUTES 不能为负数")
	}
	if c.GapMinutes < 0 {
		return fmt.Errorf("GAP_MINUTES 不能为负数")
	}
	if c.MergeAgeMinutes < 0 {
		return fmt.Errorf("MERGE_AGE_MINUTES 不能为负数")
	}
	if c.SafeMode != "hours" && c.SafeMode != "days" {
		return fmt.Errorf("SAFE_MODE 只能是 hours 或 days")
	}
	if c.SafeDays < 1 {
		return fmt.Errorf("SAFE_DAYS 不能小于 1")
	}
	if c.MaxDeletePerRun < 1 {
		return fmt.Errorf("MAX_DELETE_PER_RUN 不能小于 1")
	}
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("PORT 必须在 0-65535 之间")
	}
	if c.BackupStartHour < 0 || c.BackupStartHour > 23 {
		return fmt.Errorf("BACKUP_START_HOUR 必须在 0-23 之间")
	}
	if c.BackupStartMinute < 0 || c.BackupStartMinute > 59 {
		return fmt.Errorf("BACKUP_START_MINUTE 必须在 0-59 之间")
	}
	if c.BackupEndHour < 0 || c.BackupEndHour > 23 {
		return fmt.Errorf("BACKUP_END_HOUR 必须在 0-23 之间")
	}
	if c.BackupEndMinute < 0 || c.BackupEndMinute > 59 {
		return fmt.Errorf("BACKUP_END_MINUTE 必须在 0-59 之间")
	}
	return nil
}

// IsBackupWindow returns true if the current time falls within the quiet window.
func (c *Config) IsBackupWindow() bool {
	now := time.Now()
	cur := now.Hour()*60 + now.Minute()
	start := c.BackupStartHour*60 + c.BackupStartMinute
	end := c.BackupEndHour*60 + c.BackupEndMinute
	if start <= end {
		return cur >= start && cur < end
	}
	// Wraps midnight, e.g. 22:30-6:15
	return cur >= start || cur < end
}

// Apply runs fn under the config write lock then persists to disk atomically.
// If fn returns an error, the changes are rolled back and not persisted.
func (c *Config) Apply(fn func() error) error {
	mu.Lock()
	defer mu.Unlock()
	// Snapshot for rollback on error (deep copy slice)
	old := *c
	old.WhitelistKeywords = append([]string(nil), c.WhitelistKeywords...)
	if err := fn(); err != nil {
		*c = old
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		*c = old
		return err
	}
	// Atomic write: write to tmp file then rename (crash-safe)
	tmp := c.ConfigFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		*c = old
		return err
	}
	if err := os.Rename(tmp, c.ConfigFile); err != nil {
		os.Remove(tmp)
		*c = old
		return err
	}
	return nil
}

func (c *Config) GetScheduleFile() string {
	return filepath.Join(c.LogDir, "schedule.json")
}

func (c *Config) GetHistoryFile() string {
	return filepath.Join(c.LogDir, "history.json")
}

// ConfigDTO is a safe representation of Config without sensitive fields.
type ConfigDTO struct {
	TargetDir          string   `json:"TARGET_DIR"`
	TriggerThreshold   float64  `json:"TRIGGER_THRESHOLD"`
	TargetThreshold    float64  `json:"TARGET_THRESHOLD"`
	MinKeepPerStreamer int      `json:"MIN_KEEP_PER_STREAMER"`
	SafeAgeMinutes     int      `json:"SAFE_AGE_MINUTES"`
	GapMinutes         int      `json:"GAP_MINUTES"`
	MergeAgeMinutes    int      `json:"MERGE_AGE_MINUTES"`
	WhitelistKeywords  []string `json:"WHITELIST_KEYWORDS"`
	SafeMode           string   `json:"SAFE_MODE"`
	SafeDays           int      `json:"SAFE_DAYS"`
	MaxDeletePerRun    int      `json:"MAX_DELETE_PER_RUN"`
	BackupStartHour    int      `json:"BACKUP_START_HOUR"`
	BackupStartMinute  int      `json:"BACKUP_START_MINUTE"`
	BackupEndHour      int      `json:"BACKUP_END_HOUR"`
	BackupEndMinute    int      `json:"BACKUP_END_MINUTE"`
	Port               int      `json:"PORT"`
	LogDir             string   `json:"LOG_DIR"`
}

func (c *Config) ToDTO() ConfigDTO {
	mu.RLock()
	defer mu.RUnlock()
	return c.ToDTOSnapshot()
}

// Snapshot returns a copy of the Config safe for concurrent reads.
func (c *Config) Snapshot() Config {
	mu.RLock()
	defer mu.RUnlock()
	snap := *c
	snap.WhitelistKeywords = append([]string(nil), c.WhitelistKeywords...)
	return snap
}

// ApplyFromMap applies partial config updates from a map (JSON request body).
// Only present keys are updated. Validates TARGET_DIR exists as a directory.
func (c *Config) ApplyFromMap(m map[string]interface{}) {
	if v, ok := m["TARGET_DIR"].(string); ok {
		if info, err := os.Stat(v); err != nil || !info.IsDir() {
			// Reject invalid directory — keep the old value
		} else {
			c.TargetDir = v
		}
	}
	if v, ok := m["TRIGGER_THRESHOLD"].(float64); ok {
		c.TriggerThreshold = v
	}
	if v, ok := m["TARGET_THRESHOLD"].(float64); ok {
		c.TargetThreshold = v
	}
	if v, ok := m["MIN_KEEP_PER_STREAMER"].(float64); ok {
		c.MinKeepPerStreamer = int(v)
	}
	if v, ok := m["SAFE_AGE_MINUTES"].(float64); ok {
		c.SafeAgeMinutes = int(v)
	}
	if v, ok := m["GAP_MINUTES"].(float64); ok {
		c.GapMinutes = int(v)
	}
	if v, ok := m["MERGE_AGE_MINUTES"].(float64); ok {
		c.MergeAgeMinutes = int(v)
	}
	if v, ok := m["SAFE_MODE"].(string); ok {
		c.SafeMode = v
	}
	if v, ok := m["SAFE_DAYS"].(float64); ok {
		c.SafeDays = int(v)
	}
	if v, ok := m["MAX_DELETE_PER_RUN"].(float64); ok {
		c.MaxDeletePerRun = int(v)
	}
	if v, ok := m["BACKUP_START_HOUR"].(float64); ok {
		c.BackupStartHour = int(v)
	}
	if v, ok := m["BACKUP_START_MINUTE"].(float64); ok {
		c.BackupStartMinute = int(v)
	}
	if v, ok := m["BACKUP_END_HOUR"].(float64); ok {
		c.BackupEndHour = int(v)
	}
	if v, ok := m["BACKUP_END_MINUTE"].(float64); ok {
		c.BackupEndMinute = int(v)
	}
	if v, ok := m["WHITELIST_KEYWORDS"].([]interface{}); ok {
		var keywords []string
		for _, kw := range v {
			if s, ok := kw.(string); ok {
				keywords = append(keywords, s)
			}
		}
		c.WhitelistKeywords = keywords
	}
	if v, ok := m["PORT"].(float64); ok {
		c.Port = int(v)
	}
}

// DiffDTO compares two ConfigDTO snapshots and returns a human-readable change summary.
// Returns empty string if no changes.
func DiffDTO(old, new ConfigDTO) string {
	var changes []string
	if old.TargetDir != new.TargetDir {
		changes = append(changes, fmt.Sprintf("目录: %s→%s", filepath.Base(old.TargetDir), filepath.Base(new.TargetDir)))
	}
	if old.TriggerThreshold != new.TriggerThreshold {
		changes = append(changes, fmt.Sprintf("触发阈值: %.0f→%.0f", old.TriggerThreshold, new.TriggerThreshold))
	}
	if old.TargetThreshold != new.TargetThreshold {
		changes = append(changes, fmt.Sprintf("目标阈值: %.0f→%.0f", old.TargetThreshold, new.TargetThreshold))
	}
	if old.MinKeepPerStreamer != new.MinKeepPerStreamer {
		changes = append(changes, fmt.Sprintf("保底数量: %d→%d", old.MinKeepPerStreamer, new.MinKeepPerStreamer))
	}
	if old.SafeAgeMinutes != new.SafeAgeMinutes {
		changes = append(changes, fmt.Sprintf("安全期: %d→%d分钟", old.SafeAgeMinutes, new.SafeAgeMinutes))
	}
	if old.GapMinutes != new.GapMinutes {
		changes = append(changes, fmt.Sprintf("间隔: %d→%d分钟", old.GapMinutes, new.GapMinutes))
	}
	if old.MergeAgeMinutes != new.MergeAgeMinutes {
		changes = append(changes, fmt.Sprintf("合并等待: %d→%d分钟", old.MergeAgeMinutes, new.MergeAgeMinutes))
	}
	if old.SafeMode != new.SafeMode {
		changes = append(changes, fmt.Sprintf("安全模式: %s→%s", old.SafeMode, new.SafeMode))
	}
	if old.MaxDeletePerRun != new.MaxDeletePerRun {
		changes = append(changes, fmt.Sprintf("单次删除上限: %d→%d", old.MaxDeletePerRun, new.MaxDeletePerRun))
	}
	if old.BackupStartHour != new.BackupStartHour || old.BackupStartMinute != new.BackupStartMinute || old.BackupEndHour != new.BackupEndHour || old.BackupEndMinute != new.BackupEndMinute {
		changes = append(changes, fmt.Sprintf("静默时段: %d:%02d-%d:%02d→%d:%02d-%d:%02d",
			old.BackupStartHour, old.BackupStartMinute, old.BackupEndHour, old.BackupEndMinute,
			new.BackupStartHour, new.BackupStartMinute, new.BackupEndHour, new.BackupEndMinute))
	}
	if !equalSlice(old.WhitelistKeywords, new.WhitelistKeywords) {
		changes = append(changes, fmt.Sprintf("白名单: [%s]→[%s]",
			strings.Join(old.WhitelistKeywords, ","), strings.Join(new.WhitelistKeywords, ",")))
	}
	if len(changes) == 0 {
		return ""
	}
	return fmt.Sprintf("配置变更: %s", strings.Join(changes, ", "))
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ToDTOSnapshot returns a DTO snapshot without locking (caller must hold lock).
func (c *Config) ToDTOSnapshot() ConfigDTO {
	return ConfigDTO{
		TargetDir:          c.TargetDir,
		TriggerThreshold:   c.TriggerThreshold,
		TargetThreshold:    c.TargetThreshold,
		MinKeepPerStreamer: c.MinKeepPerStreamer,
		SafeAgeMinutes:     c.SafeAgeMinutes,
		GapMinutes:         c.GapMinutes,
		MergeAgeMinutes:    c.MergeAgeMinutes,
		WhitelistKeywords:  c.WhitelistKeywords,
		SafeMode:           c.SafeMode,
		SafeDays:           c.SafeDays,
		MaxDeletePerRun:    c.MaxDeletePerRun,
		BackupStartHour:    c.BackupStartHour,
		BackupStartMinute:  c.BackupStartMinute,
		BackupEndHour:      c.BackupEndHour,
		BackupEndMinute:    c.BackupEndMinute,
		Port:               c.Port,
		LogDir:             c.LogDir,
	}
}
