// Package config 提供应用程序配置管理功能。
// 支持从配置文件加载、环境变量覆盖、原子写入、事务回滚、并发安全的配置读写。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bililive-helper/internal/utils"
)

// Config 是应用程序的核心配置结构体。
// 字段通过 JSON 序列化持久化到 config.json，Password 和 SecretKey 使用 json:"-" 排除。
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
	Password           string   `json:"-"`
	SecretKey          string   `json:"-"`
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
		Password:           "",  // auto-generated on first run if not set
		SecretKey:          "",  // auto-generated on first run if not set
		LogDir:             "/vol1/1000/docker/bililive-helper-go",
	}
	mu sync.RWMutex
)

// ConfigExists returns true if config.json exists in the given log directory.
func ConfigExists(logDir string) bool {
	_, err := os.Stat(filepath.Join(logDir, "config.json"))
	return err == nil
}

// DefaultConfig returns a copy of the default configuration (without credentials).
func DefaultConfig() Config {
	cfg := defaultConfig
	cfg.WhitelistKeywords = append([]string(nil), defaultConfig.WhitelistKeywords...)
	return cfg
}

// Load 从配置文件和环境变量加载配置。
// 加载顺序：默认值 -> config.json 文件 -> 环境变量覆盖。
// 首次运行时自动生成密码和密钥并持久化到凭据文件。
func Load() *Config {
	cfg := defaultConfig

	// 优先从文件加载，然后用环境变量覆盖（环境变量始终优先）
	// LOG_DIR 需要先确定，因为它决定了配置文件的位置
	cfgFileDir := cfg.LogDir
	if v := os.Getenv("LOG_DIR"); v != "" {
		cfgFileDir = v
	}
	cfg.ConfigFile = filepath.Join(cfgFileDir, "config.json")
	if data, err := os.ReadFile(cfg.ConfigFile); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			fmt.Printf("[WARN] 配置文件解析失败，使用默认配置: %v\n", err)
		}
	}

	// 环境变量覆盖文件值
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
	// Update ConfigFile path in case LOG_DIR changed via env
	cfg.ConfigFile = filepath.Join(cfg.LogDir, "config.json")

	// 首次运行自动生成随机凭据（无配置文件且无环境变量时）
	firstRun := false
	if cfg.Password == "" {
		cfg.Password = cfg.LoadCredential()
	}
	if cfg.Password == "" {
		cfg.Password = utils.RandomHex(12) // 24-char hex string (~96 bits of entropy)
		firstRun = true
	}
	if cfg.SecretKey == "" {
		cfg.SecretKey = utils.RandomHex(16) // 32-char hex string (~128 bits of entropy)
		firstRun = true
	}
	if firstRun {
		fmt.Printf("═══ 首次启动 ═══\n")
		fmt.Printf("登录密码: %s\n", cfg.Password)
		fmt.Printf("请登录后及时修改密码\n")
		fmt.Printf("═══════════════\n")
		// Persist auto-generated password so it survives restarts
		if err := cfg.SaveCredential(); err != nil {
			fmt.Printf("[WARN] 密码持久化失败: %v\n", err)
		}
	}
	return &cfg
}

// Validate 校验配置字段的合法性。
// 检查阈值范围、保底数量、安全期、端口号等，返回第一个不合法字段的错误信息。
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

// IsBackupWindow returns true if the current time falls within the configured quiet window.
// Supports windows that wrap around midnight (e.g. 22:30 - 06:15).
func (c *Config) IsBackupWindow() bool {
	now := time.Now()
	cur := now.Hour()*60 + now.Minute()
	start := c.BackupStartHour*60 + c.BackupStartMinute
	end := c.BackupEndHour*60 + c.BackupEndMinute
	if start <= end {
		return cur >= start && cur < end
	}
	// Wraps midnight (e.g. 22:30 - 06:15)
	return cur >= start || cur < end
}

// Apply 在写锁保护下执行配置修改函数 fn，成功后原子写入磁盘。
// 如果 fn 返回错误或写入失败，所有修改将回滚到调用前的状态。
func (c *Config) Apply(fn func() error) error {
	mu.Lock()
	defer mu.Unlock()
	// Snapshot for rollback (deep copy slice to avoid aliasing)
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
	// Atomic write: write to tmp file then rename to prevent corruption on crash
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

// GetScheduleFile 返回定时调度配置文件路径。
func (c *Config) GetScheduleFile() string {
	return filepath.Join(c.LogDir, "schedule.json")
}

// GetHistoryFile 返回历史记录文件路径。
func (c *Config) GetHistoryFile() string {
	return filepath.Join(c.LogDir, "history.json")
}

// GetCredentialFile returns the path to the credentials file that stores
// the password outside config.json (since Password has json:"-").
func (c *Config) GetCredentialFile() string {
	return filepath.Join(c.LogDir, ".credentials.json")
}

// LoadCredential reads the persisted password from the credential file.
// Returns empty string if no credential file exists.
func (c *Config) LoadCredential() string {
	file := c.GetCredentialFile()
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	var cred struct {
		Password string `json:"password"`
	}
	if err := json.Unmarshal(data, &cred); err != nil {
		return ""
	}
	return cred.Password
}

// SaveCredential persists the current password to a credential file.
func (c *Config) SaveCredential() error {
	file := c.GetCredentialFile()
	data, err := json.Marshal(struct {
		Password string `json:"password"`
	}{Password: c.Password})
	if err != nil {
		return err
	}
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, file); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// ConfigDTO 是 Config 的安全数据传输对象，不包含密码和密钥等敏感字段。
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

// ToDTO 返回不含敏感字段的配置副本（线程安全）。
func (c *Config) ToDTO() ConfigDTO {
	mu.RLock()
	defer mu.RUnlock()
	return c.ToDTOSnapshot()
}

// Snapshot 返回配置的深拷贝副本，用于并发读取场景（线程安全）。
func (c *Config) Snapshot() Config {
	mu.RLock()
	defer mu.RUnlock()
	snap := *c
	snap.WhitelistKeywords = append([]string(nil), c.WhitelistKeywords...)
	return snap
}

// ApplyFromMap 从 JSON 请求体的 map 中应用部分配置更新。
// 只更新 map 中存在的字段，TARGET_DIR 会额外验证目录是否存在。
func (c *Config) ApplyFromMap(m map[string]interface{}) {
	if v, ok := m["TARGET_DIR"].(string); ok {
		if info, err := os.Stat(v); err != nil || !info.IsDir() {
			// Reject invalid directory paths silently
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

// DiffDTO 比较两个 ConfigDTO 快照，返回人类可读的变更摘要。
// 如果配置完全相同返回空字符串。
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

// ToDTOSnapshot 返回不含敏感字段的 DTO 快照。调用者必须持有读锁或写锁。
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
