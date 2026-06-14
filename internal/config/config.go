// Package config 提供应用程序配置管理功能。
// 支持从配置文件加载、环境变量覆盖、原子写入、事务回滚、并发安全的配置读写。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"bililive-helper/internal/fsutil"
	"bililive-helper/internal/utils"
)

// Config 是应用程序的核心配置结构体。
// 字段通过 JSON 序列化持久化到 config.json，Password 和 SecretKey 使用 json:"-" 排除。
type Config struct {
	mu                 *sync.RWMutex `json:"-"` // 实例级读写锁，保护并发读写
	TargetDir          string        `json:"TARGET_DIR"`
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
	SessionVersion     int      `json:"SESSION_VERSION,omitempty"` // 改密时递增，使旧 Session 失效
}

var (
	defaultConfig = Config{
		mu:                 &sync.RWMutex{},
		TargetDir:          "/vol2/1000/video/bililive-go/抖音",
		TriggerThreshold:   80,
		TargetThreshold:    60,
		MinKeepPerStreamer: 3,
		SafeAgeMinutes:     120,
		GapMinutes:         30,
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
		Password:           "",  // 首次运行时自动生成（如未设置）
		SecretKey:          "",  // 首次运行时自动生成（如未设置）
		LogDir:             "/vol1/1000/docker/bililive-helper-go",
	}
)

// ConfigExists 检查指定日志目录下是否存在 config.json 文件。
func ConfigExists(logDir string) bool {
	_, err := os.Stat(filepath.Join(logDir, "config.json"))
	return err == nil
}

// DefaultConfig 返回默认配置的副本（不含凭据信息）。
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
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 && p <= 65535 {
			cfg.Port = p
		}
	}
	// LOG_DIR 通过环境变量变更后，更新 ConfigFile 路径
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
		// 持久化自动生成的密码，使其在重启后仍然有效
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
	if c.Port < 1 || c.Port > 65535 {
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

// IsBackupWindow 判断当前时间是否在静默时段内。
// 支持跨午夜的时段配置（如 22:30 - 06:15）。
func (c *Config) IsBackupWindow() bool {
	now := time.Now()
	cur := now.Hour()*60 + now.Minute()
	start := c.BackupStartHour*60 + c.BackupStartMinute
	end := c.BackupEndHour*60 + c.BackupEndMinute
	if start <= end {
		return cur >= start && cur < end
	}
	// 跨午夜时段（如 22:30 - 06:15）
	return cur >= start || cur < end
}

// Apply 在写锁保护下执行配置修改函数 fn，成功后原子写入磁盘。
// 如果 fn 返回错误或写入失败，所有修改将回滚到调用前的状态。
func (c *Config) Apply(fn func() error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 快照用于回滚（深拷贝切片避免别名问题）
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
	// 原子写入：先写临时文件再 sync+rename，防止崩溃时数据损坏
	if err := fsutil.AtomicSave(c.ConfigFile, data, 0600); err != nil {
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

// GetCredentialFile 返回凭据文件路径。
// 密码通过 json:"-" 排除在 config.json 之外，单独存储在此文件中。
func (c *Config) GetCredentialFile() string {
	return filepath.Join(c.LogDir, ".credentials.json")
}

// LoadCredential 从凭据文件读取持久化的密码。
// 凭据文件不存在时返回空字符串。
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

// SaveCredential 将当前密码持久化到凭据文件。
func (c *Config) SaveCredential() error {
	file := c.GetCredentialFile()
	data, err := json.Marshal(struct {
		Password string `json:"password"`
	}{Password: c.Password})
	if err != nil {
		return err
	}
	return fsutil.AtomicSave(file, data, 0600)
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
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ToDTOSnapshot()
}

// Snapshot 返回配置的深拷贝副本，用于并发读取场景（线程安全）。
func (c *Config) Snapshot() Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	snap := *c
	snap.WhitelistKeywords = append([]string(nil), c.WhitelistKeywords...)
	return snap
}

// ApplyFromJSON 从 JSON 请求体应用部分配置更新。
// 使用 JSON 反序列化直接更新 Config struct，类型安全，替代原来的 map[string]interface{} 模式。
// TARGET_DIR 会额外验证目录是否存在。只更新 JSON 中存在的字段（零值字段不覆盖）。
func (c *Config) ApplyFromJSON(data []byte) {
	// 先解析为 map 检查哪些字段存在，再逐个应用
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		fmt.Printf("[WARN] 配置 JSON 解析失败: %v\n", err)
		return
	}

	if raw, ok := m["TARGET_DIR"]; ok {
		var v string
		json.Unmarshal(raw, &v)
		if info, err := os.Stat(v); err != nil || !info.IsDir() {
			fmt.Printf("[WARN] TARGET_DIR 无效，已忽略: %s (err=%v)\n", v, err)
		} else {
			c.TargetDir = v
		}
	}
	if raw, ok := m["TRIGGER_THRESHOLD"]; ok {
		json.Unmarshal(raw, &c.TriggerThreshold)
	}
	if raw, ok := m["TARGET_THRESHOLD"]; ok {
		json.Unmarshal(raw, &c.TargetThreshold)
	}
	if raw, ok := m["MIN_KEEP_PER_STREAMER"]; ok {
		json.Unmarshal(raw, &c.MinKeepPerStreamer)
	}
	if raw, ok := m["SAFE_AGE_MINUTES"]; ok {
		json.Unmarshal(raw, &c.SafeAgeMinutes)
	}
	if raw, ok := m["GAP_MINUTES"]; ok {
		json.Unmarshal(raw, &c.GapMinutes)
	}
	if raw, ok := m["MERGE_AGE_MINUTES"]; ok {
		json.Unmarshal(raw, &c.MergeAgeMinutes)
	}
	if raw, ok := m["SAFE_MODE"]; ok {
		json.Unmarshal(raw, &c.SafeMode)
	}
	if raw, ok := m["SAFE_DAYS"]; ok {
		json.Unmarshal(raw, &c.SafeDays)
	}
	if raw, ok := m["MAX_DELETE_PER_RUN"]; ok {
		json.Unmarshal(raw, &c.MaxDeletePerRun)
	}
	if raw, ok := m["BACKUP_START_HOUR"]; ok {
		json.Unmarshal(raw, &c.BackupStartHour)
	}
	if raw, ok := m["BACKUP_START_MINUTE"]; ok {
		json.Unmarshal(raw, &c.BackupStartMinute)
	}
	if raw, ok := m["BACKUP_END_HOUR"]; ok {
		json.Unmarshal(raw, &c.BackupEndHour)
	}
	if raw, ok := m["BACKUP_END_MINUTE"]; ok {
		json.Unmarshal(raw, &c.BackupEndMinute)
	}
	if raw, ok := m["WHITELIST_KEYWORDS"]; ok {
		json.Unmarshal(raw, &c.WhitelistKeywords)
	}
	if raw, ok := m["PORT"]; ok {
		json.Unmarshal(raw, &c.Port)
	}
}

// DiffDTO 比较两个 ConfigDTO 快照，返回人类可读的变更摘要。
// 如果配置完全相同返回空字符串。
func DiffDTO(old, new ConfigDTO) string {
	var changes []string

	// 辅助函数：比较单个字段，不同时追加变更描述
	addStr := func(label, a, b string) {
		if a != b {
			changes = append(changes, fmt.Sprintf("%s: %s→%s", label, a, b))
		}
	}
	addFloat := func(label string, a, b float64) {
		if a != b {
			changes = append(changes, fmt.Sprintf("%s: %.0f→%.0f", label, a, b))
		}
	}
	addInt := func(label string, a, b int) {
		if a != b {
			changes = append(changes, fmt.Sprintf("%s: %d→%d", label, a, b))
		}
	}

	addStr("目录", filepath.Base(old.TargetDir), filepath.Base(new.TargetDir))
	addFloat("触发阈值", old.TriggerThreshold, new.TriggerThreshold)
	addFloat("目标阈值", old.TargetThreshold, new.TargetThreshold)
	addInt("保底数量", old.MinKeepPerStreamer, new.MinKeepPerStreamer)
	addInt("安全期", old.SafeAgeMinutes, new.SafeAgeMinutes)
	addInt("间隔", old.GapMinutes, new.GapMinutes)
	addInt("合并等待", old.MergeAgeMinutes, new.MergeAgeMinutes)
	addStr("安全模式", old.SafeMode, new.SafeMode)
	addInt("安全期天数", old.SafeDays, new.SafeDays)
	addInt("单次删除上限", old.MaxDeletePerRun, new.MaxDeletePerRun)
	addInt("端口", old.Port, new.Port)
	addStr("日志目录", filepath.Base(old.LogDir), filepath.Base(new.LogDir))

	// 复合字段特殊处理
	if old.BackupStartHour != new.BackupStartHour || old.BackupStartMinute != new.BackupStartMinute ||
		old.BackupEndHour != new.BackupEndHour || old.BackupEndMinute != new.BackupEndMinute {
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
// 通过 JSON 序列化/反序列化自动生成，排除 Password(json:"-")、SecretKey(json:"-")、
// ConfigFile(json:"-")、SessionVersion(omitempty+零值)、mu(未导出)。
func (c *Config) ToDTOSnapshot() ConfigDTO {
	data, err := json.Marshal(c)
	if err != nil {
		// 不应发生（marshal 本地 struct），降级为零值
		return ConfigDTO{}
	}
	var dto ConfigDTO
	json.Unmarshal(data, &dto)
	return dto
}
