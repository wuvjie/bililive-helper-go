package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"bililive-helper-go/internal/config"

	"go.uber.org/zap"
)

func TestCollectCandidates_RespectsMinKeep(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	cfg.MinKeepPerStreamer = 3
	cfg.SafeAgeMinutes = 0
	cfg.SafeMode = "hours"
	cfg.WhitelistKeywords = nil

	// 创建主播目录，只有 2 个视频文件（少于 minKeep=3）
	streamerDir := filepath.Join(dir, "streamer_a")
	os.MkdirAll(streamerDir, 0755)
	os.WriteFile(filepath.Join(streamerDir, "old.mp4"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(streamerDir, "new.mp4"), make([]byte, 1000), 0644)

	svc := &CleanService{
		config: &cfg,
		logger: zap.NewNop(),
	}

	candidates, _ := svc.collectCandidates(dir, "", cfg)
	// 只有 2 个文件，少于 minKeep=3，不应有任何候选
	if len(candidates) != 0 {
		t.Errorf("期望 0 个候选（文件数 < minKeep），得到 %d", len(candidates))
	}
}

func TestCollectCandidates_FiltersWhitelist(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	cfg.MinKeepPerStreamer = 1
	cfg.SafeAgeMinutes = 0
	cfg.SafeMode = "hours"
	cfg.WhitelistKeywords = []string{"留存"}

	streamerDir := filepath.Join(dir, "streamer_b")
	os.MkdirAll(streamerDir, 0755)
	// 创建 3 个文件，1 个在白名单中
	os.WriteFile(filepath.Join(streamerDir, "2026-01-01-留存.mp4"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(streamerDir, "2026-01-02.mp4"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(streamerDir, "2026-01-03.mp4"), make([]byte, 1000), 0644)

	svc := &CleanService{
		config: &cfg,
		logger: zap.NewNop(),
	}

	candidates, _ := svc.collectCandidates(dir, "", cfg)
	// 3 个文件 - 1 个保底 - 1 个白名单 = 1 个候选
	if len(candidates) != 1 {
		t.Errorf("期望 1 个候选（白名单过滤），得到 %d", len(candidates))
	}
}

func TestCollectCandidates_FiltersStreamerWhitelist(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	cfg.MinKeepPerStreamer = 1
	cfg.SafeAgeMinutes = 0
	cfg.SafeMode = "hours"
	cfg.WhitelistKeywords = []string{"高能"}

	streamerDir := filepath.Join(dir, "高能主播")
	os.MkdirAll(streamerDir, 0755)
	os.WriteFile(filepath.Join(streamerDir, "2026-01-01.mp4"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(streamerDir, "2026-01-02.mp4"), make([]byte, 1000), 0644)

	svc := &CleanService{
		config: &cfg,
		logger: zap.NewNop(),
	}

	candidates, _ := svc.collectCandidates(dir, "", cfg)
	// 主播名包含白名单关键词，所有文件都被跳过
	if len(candidates) != 0 {
		t.Errorf("期望 0 个候选（主播名在白名单），得到 %d", len(candidates))
	}
}

func TestCollectCandidates_SafeAgeFilter(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	cfg.MinKeepPerStreamer = 1
	cfg.SafeAgeMinutes = 60 // 1 小时安全期
	cfg.SafeMode = "hours"
	cfg.WhitelistKeywords = nil

	streamerDir := filepath.Join(dir, "streamer_c")
	os.MkdirAll(streamerDir, 0755)

	// 创建一个新文件（在安全期内）和一个旧文件
	newFile := filepath.Join(streamerDir, "new.mp4")
	oldFile := filepath.Join(streamerDir, "old.mp4")
	os.WriteFile(newFile, make([]byte, 1000), 0644)
	os.WriteFile(oldFile, make([]byte, 1000), 0644)
	// 把 old.mp4 的修改时间设为 2 小时前
	oldTime := time.Now().Add(-2 * time.Hour)
	os.Chtimes(oldFile, oldTime, oldTime)

	svc := &CleanService{
		config: &cfg,
		logger: zap.NewNop(),
	}

	candidates, _ := svc.collectCandidates(dir, "", cfg)
	// 2 个文件 - 1 个保底 = 1 个候选，但新文件在安全期内被过滤
	// 所以只有 old.mp4 是候选
	if len(candidates) != 1 {
		t.Errorf("期望 1 个候选（安全期过滤），得到 %d", len(candidates))
	}
	if len(candidates) > 0 && filepath.Base(candidates[0].Path) != "old.mp4" {
		t.Errorf("期望候选文件是 old.mp4，得到 %s", filepath.Base(candidates[0].Path))
	}
}

func TestCollectStreamerCandidates_ReturnsValue(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	cfg.MinKeepPerStreamer = 1
	cfg.SafeAgeMinutes = 0
	cfg.SafeMode = "hours"
	cfg.WhitelistKeywords = nil

	streamerDir := filepath.Join(dir, "streamer_d")
	os.MkdirAll(streamerDir, 0755)
	os.WriteFile(filepath.Join(streamerDir, "video1.mp4"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(streamerDir, "video2.mp4"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(streamerDir, "video3.mp4"), make([]byte, 1000), 0644)

	svc := &CleanService{
		config: &cfg,
		logger: zap.NewNop(),
	}

	// 验证返回值模式（非指针模式）
	result := svc.collectStreamerCandidates(streamerDir, "streamer_d", cfg)
	// 3 个文件 - 1 个保底 = 2 个候选
	if len(result) != 2 {
		t.Errorf("期望 2 个候选，得到 %d", len(result))
	}
}
