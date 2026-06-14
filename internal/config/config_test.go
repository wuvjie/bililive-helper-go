package config

import (
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

// validConfig 返回一个能通过 Validate 的合法配置
func validConfig() *Config {
	return &Config{
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
