package config

import (
	"strings"
	"sync"
	"testing"
)

// 特征测试：记录当前 Validate 行为作为基线
// 这些测试记录的是"当前行为"，不一定是"正确行为"
// 重构后如果测试失败，说明行为发生了变化

func TestValidate_ValidConfig(t *testing.T) {
	cfg := validConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("合法配置不应报错，得到: %v", err)
	}
}

func TestValidate_TriggerThreshold_Boundaries(t *testing.T) {
	tests := []struct {
		name        string
		trigger     float64
		target      float64
		wantError   bool
	}{
		{"正常值 trigger=90 target=80", 90, 80, false},
		{"上界 trigger=100 target=90", 100, 90, false},
		{"trigger 负数", -1, 0, true},
		{"trigger 超上界", 101, 90, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.TriggerThreshold = tt.trigger
			cfg.TargetThreshold = tt.target
			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Trigger=%v Target=%v, wantError=%v, got err=%v", tt.trigger, tt.target, tt.wantError, err)
			}
		})
	}
}

func TestValidate_TargetThreshold_MustBeLessThanTrigger(t *testing.T) {
	cfg := validConfig()
	cfg.TriggerThreshold = 80
	cfg.TargetThreshold = 90.0 // 大于 Trigger
	if err := cfg.Validate(); err == nil {
		t.Error("TargetThreshold >= TriggerThreshold 应该报错")
	}
}

func TestValidate_TargetThreshold_EqualToTrigger(t *testing.T) {
	cfg := validConfig()
	cfg.TriggerThreshold = 80
	cfg.TargetThreshold = 80.0 // 等于 Trigger
	if err := cfg.Validate(); err == nil {
		t.Error("TargetThreshold == TriggerThreshold 应该报错")
	}
}

func TestValidate_MinKeepPerStreamer(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		wantError bool
	}{
		{"最小值 1", 1, false},
		{"正常值 3", 3, false},
		{"零", 0, true},
		{"负数", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.MinKeepPerStreamer = tt.value
			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("MinKeepPerStreamer=%d, wantError=%v, got err=%v", tt.value, tt.wantError, err)
			}
		})
	}
}

func TestValidate_SafeMode(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"hours", "hours", false},
		{"days", "days", false},
		{"无效值 minutes", "minutes", true},
		{"空字符串", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.SafeMode = tt.value
			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("SafeMode=%q, wantError=%v, got err=%v", tt.value, tt.wantError, err)
			}
		})
	}
}

func TestValidate_Port(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		wantError bool
	}{
		{"最小值 1", 1, false},
		{"正常值 8080", 8080, false},
		{"最大值 65535", 65535, false},
		{"零", 0, true},
		{"超上界 65536", 65536, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.Port = tt.value
			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Port=%d, wantError=%v, got err=%v", tt.value, tt.wantError, err)
			}
		})
	}
}

func TestValidate_BackupWindow(t *testing.T) {
	tests := []struct {
		name      string
		startH    int
		startM    int
		endH      int
		endM      int
		wantError bool
	}{
		{"正常范围", 22, 0, 6, 0, false},
		{"startH 越界", 24, 0, 6, 0, true},
		{"startM 越界", 22, 60, 6, 0, true},
		{"endH 越界", 22, 0, -1, 0, true},
		{"endM 越界", 22, 0, 6, 60, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.BackupStartHour = tt.startH
			cfg.BackupStartMinute = tt.startM
			cfg.BackupEndHour = tt.endH
			cfg.BackupEndMinute = tt.endM
			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("backup(%d:%d-%d:%d), wantError=%v, got err=%v",
					tt.startH, tt.startM, tt.endH, tt.endM, tt.wantError, err)
			}
		})
	}
}

func TestValidate_NegativeFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"SafeAgeMinutes 负数", func(c *Config) { c.SafeAgeMinutes = -1 }},
		{"GapMinutes 负数", func(c *Config) { c.GapMinutes = -1 }},
		{"MergeAgeMinutes 负数", func(c *Config) { c.MergeAgeMinutes = -1 }},
		{"SafeDays 零", func(c *Config) { c.SafeDays = 0 }},
		{"MaxDeletePerRun 零", func(c *Config) { c.MaxDeletePerRun = 0 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.mutate(cfg)
			if err := cfg.Validate(); err == nil {
				t.Error("应该报错但没有")
			}
		})
	}
}

func TestConfig_StreamerRules_Snapshot(t *testing.T) {
	cfg := validConfig()
	cfg.StreamerRules = map[string]StreamerRule{
		"streamer_a": {
			TrimEnabled:     true,
			TrimMode:        "head",
			DurationMinutes: 45,
			Protected:       true,
		},
	}

	snap := cfg.Snapshot()
	rule, ok := snap.GetStreamerRule("streamer_a")
	if !ok || !rule.TrimEnabled || rule.DurationMinutes != 45 || !rule.Protected {
		t.Errorf("快照未正确复制主播规则: %+v", rule)
	}

	// 验证深拷贝隔离
	cfg.StreamerRules["streamer_a"] = StreamerRule{DurationMinutes: 10}
	ruleAfter, _ := snap.GetStreamerRule("streamer_a")
	if ruleAfter.DurationMinutes != 45 {
		t.Errorf("快照应隔离底层修改，期望 45，得到 %d", ruleAfter.DurationMinutes)
	}
}

func TestConfig_StreamerRules_ApplyFromJSON(t *testing.T) {
	cfg := validConfig()
	jsonData := []byte(`{
		"STREAMER_RULES": {
			"streamer_b": {
				"trim_enabled": true,
				"trim_mode": "range",
				"start_time": "00:05:00",
				"end_time": "00:35:00",
				"min_keep": 5
			}
		}
	}`)

	cfg.ApplyFromJSON(jsonData)
	rule, ok := cfg.GetStreamerRule("streamer_b")
	if !ok {
		t.Fatalf("应成功加载 streamer_b 规则")
	}
	if !rule.TrimEnabled || rule.TrimStart != "00:05:00" || rule.TrimEnd != "00:35:00" || rule.MinKeep != 5 {
		t.Errorf("解析结果与预期不符: %+v", rule)
	}
}

// validConfig 返回一个能通过 Validate 的合法配置
func validConfig() *Config {
	return &Config{
		mu:                 &sync.RWMutex{},
		TriggerThreshold:   90.0,
		TargetThreshold:    80.0,
		MinKeepPerStreamer: 3,
		SafeAgeMinutes:     60,
		GapMinutes:         5,
		MergeAgeMinutes:    5,
		SafeMode:           "hours",
		SafeDays:           7,
		MaxDeletePerRun:    50,
		Port:               8080,
		BackupStartHour:    0,
		BackupStartMinute:  0,
		BackupEndHour:      0,
		BackupEndMinute:    0,
	}
}

func TestDiffDTO(t *testing.T) {
	oldDTO := ConfigDTO{
		TriggerThreshold:   80,
		TargetThreshold:    60,
		MinKeepPerStreamer: 3,
		SafeMode:           "hours",
		SafeDays:           7,
		SafeAgeMinutes:     120,
		GapMinutes:         30,
		MergeAgeMinutes:    30,
		MaxDeletePerRun:    50,
	}
	newDTO := oldDTO
	newDTO.TriggerThreshold = 85
	newDTO.TargetThreshold = 65
	newDTO.MinKeepPerStreamer = 5
	newDTO.SafeMode = "days"
	newDTO.SafeDays = 5
	newDTO.SafeAgeMinutes = 7200
	newDTO.GapMinutes = 45
	newDTO.MergeAgeMinutes = 60
	newDTO.MaxDeletePerRun = 30
	newDTO.BackupStartHour = 2
	newDTO.BackupEndHour = 8

	diff := DiffDTO(oldDTO, newDTO)
	expectedSubs := []string{
		"触发阈值: 80%→85%",
		"停止阈值: 60%→65%",
		"最少保留场次: 3场→5场",
		"免删保护期: 7天→5天",
		"断流分段间隔: 30分钟→45分钟",
		"关播等待时间: 30分钟→60分钟",
		"单次清理上限: 50个→30个",
		"避让时段: 0:00-0:00→2:00-8:00",
	}
	for _, sub := range expectedSubs {
		if !strings.Contains(diff, sub) {
			t.Errorf("DiffDTO 缺少期望内容 %q，实际输出: %s", sub, diff)
		}
	}
	if strings.Contains(diff, "安全模式") || strings.Contains(diff, "7200") {
		t.Errorf("DiffDTO 不应重复输出 SafeAgeMinutes 或 SafeMode，实际输出: %s", diff)
	}
}

func TestDiffStreamerRule(t *testing.T) {
	// 1. 新建开启自动剪辑与免删保护
	d1 := DiffStreamerRule(StreamerRule{}, StreamerRule{
		TrimEnabled: true,
		TrimStart:   "00:05:00",
		TrimEnd:     "",
		Protected:   true,
	}, false)
	if !strings.Contains(d1, "自动剪辑: 已启用 (00:05:00~结尾)") || !strings.Contains(d1, "免删保护: 已启用") {
		t.Errorf("新建规则变更摘要不符合预期: %s", d1)
	}

	// 2. 修改剪辑区间与关闭免删保护
	d2 := DiffStreamerRule(
		StreamerRule{TrimEnabled: true, TrimStart: "00:05:00", TrimEnd: "", Protected: true},
		StreamerRule{TrimEnabled: true, TrimStart: "00:10:00", TrimEnd: "01:30:00", Protected: false},
		true,
	)
	if !strings.Contains(d2, "剪辑区间: 00:05:00~结尾→00:10:00~01:30:00") || !strings.Contains(d2, "免删保护: 已停用") {
		t.Errorf("更新规则变更摘要不符合预期: %s", d2)
	}
}

func TestConfig_CredentialsPersistence(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.LogDir = dir
	cfg.ConfigFile = dir + "/config.json"
	cfg.Password = "secret-pass-123"
	cfg.SecretKey = "0123456789abcdef0123456789abcdef"

	if err := cfg.SaveCredential(); err != nil {
		t.Fatalf("SaveCredential 失败: %v", err)
	}

	reloaded := DefaultConfig()
	reloaded.LogDir = dir
	pwd, key := reloaded.LoadCredentials()
	if pwd != cfg.Password {
		t.Errorf("恢复密码不匹配: got %q, want %q", pwd, cfg.Password)
	}
	if key != cfg.SecretKey {
		t.Errorf("恢复 SecretKey 不匹配: got %q, want %q", key, cfg.SecretKey)
	}
	if reloaded.LoadCredential() != cfg.Password {
		t.Errorf("LoadCredential 兼容读取失败: got %q, want %q", reloaded.LoadCredential(), cfg.Password)
	}
}
