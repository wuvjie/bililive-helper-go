package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestCollectCandidates_RespectsStreamerRuleProtected(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	cfg.MinKeepPerStreamer = 1
	cfg.SafeAgeMinutes = 0
	cfg.SafeMode = "hours"
	cfg.StreamerRules = map[string]config.StreamerRule{
		"vip_streamer": {
			Protected: true,
		},
	}

	streamerDir := filepath.Join(dir, "vip_streamer")
	os.MkdirAll(streamerDir, 0755)
	os.WriteFile(filepath.Join(streamerDir, "video1.mp4"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(streamerDir, "video2.mp4"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(streamerDir, "video3.mp4"), make([]byte, 1000), 0644)

	svc := &CleanService{
		config: &cfg,
		logger: zap.NewNop(),
	}

	result := svc.collectStreamerCandidates(streamerDir, "vip_streamer", cfg)
	if len(result) != 0 {
		t.Errorf("免删保护的主播应返回 0 个候选，得到 %d", len(result))
	}
}

func TestCollectCandidates_RespectsStreamerRuleMinKeep(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	cfg.MinKeepPerStreamer = 1 // 全局默认只保留 1 个
	cfg.SafeAgeMinutes = 0
	cfg.SafeMode = "hours"
	cfg.StreamerRules = map[string]config.StreamerRule{
		"streamer_custom_keep": {
			MinKeep: 3, // 个性化设置保底 3 个
		},
	}

	streamerDir := filepath.Join(dir, "streamer_custom_keep")
	os.MkdirAll(streamerDir, 0755)
	os.WriteFile(filepath.Join(streamerDir, "video1.mp4"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(streamerDir, "video2.mp4"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(streamerDir, "video3.mp4"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(streamerDir, "video4.mp4"), make([]byte, 1000), 0644)

	svc := &CleanService{
		config: &cfg,
		logger: zap.NewNop(),
	}

	result := svc.collectStreamerCandidates(streamerDir, "streamer_custom_keep", cfg)
	// 4 个文件 - 3 个保底 = 1 个候选
	if len(result) != 1 {
		t.Errorf("期望自定义保底后产生 1 个候选，得到 %d", len(result))
	}

	candidates, perStreamer := svc.collectCandidates(dir, "", cfg)
	if len(candidates) != 1 {
		t.Errorf("collectCandidates 期望 1 个候选，得到 %d", len(candidates))
	}
	stats := perStreamer["streamer_custom_keep"]
	if stats.total != 4 || stats.candidate != 1 || stats.minKeep != 3 || stats.skipped != 0 || stats.protected {
		t.Errorf("streamer_custom_keep 统计异常: %+v", stats)
	}
}

func TestCollectCandidates_PendingRawDoesNotDisplaceCompletedMinKeep(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	cfg.MinKeepPerStreamer = 2
	cfg.MergeAgeMinutes = 30
	cfg.SafeAgeMinutes = 0
	cfg.SafeMode = "hours"
	cfg.WhitelistKeywords = nil

	streamerDir := filepath.Join(dir, "streamer_mixed")
	os.MkdirAll(streamerDir, 0755)

	oldTime := time.Now().Add(-72 * time.Hour)
	for _, name := range []string{
		"[2026-05-01 20-00-00][s][t1]-合并版.mp4",
		"[2026-05-02 20-00-00][s][t2]-合并版.mp4",
	} {
		p := filepath.Join(streamerDir, name)
		os.WriteFile(p, make([]byte, 2000), 0644)
		os.Chtimes(p, oldTime, oldTime)
	}

	// 新下播的 4 个待合并 .flv 分片（距今 5 分钟，处于 30 分钟等待期内）
	recentTime := time.Now().Add(-5 * time.Minute)
	for _, name := range []string{
		"[2026-05-05 20-00-00][s][t3]001.flv",
		"[2026-05-05 20-00-00][s][t3]002.flv",
		"[2026-05-05 20-00-00][s][t3]003.flv",
		"[2026-05-05 20-00-00][s][t3]004.flv",
	} {
		p := filepath.Join(streamerDir, name)
		os.WriteFile(p, make([]byte, 2000), 0644)
		os.Chtimes(p, recentTime, recentTime)
	}

	svc := &CleanService{config: &cfg, logger: zap.NewNop()}
	candidates, _ := svc.collectCandidates(dir, "", cfg)
	if len(candidates) != 0 {
		t.Fatalf("未合并的等待期内分片不应挤占已完成 .mp4 的 minKeep 名额，期望 0 候选，得到 %d: %+v", len(candidates), candidates)
	}
}

func TestCollectCandidates_StaleRawSegmentsKeepCompletedMinKeepProtected(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	cfg.MinKeepPerStreamer = 2
	cfg.MergeAgeMinutes = 30
	cfg.SafeAgeMinutes = 60
	cfg.SafeMode = "hours"
	cfg.WhitelistKeywords = nil

	streamerDir := filepath.Join(dir, "streamer_stale_raw")
	os.MkdirAll(streamerDir, 0755)

	oldMP4 := time.Now().Add(-96 * time.Hour)
	for _, name := range []string{
		"[2026-05-01 20-00-00][s][t1]-合并版.mp4",
		"[2026-05-02 20-00-00][s][t2]-合并版.mp4",
	} {
		p := filepath.Join(streamerDir, name)
		os.WriteFile(p, make([]byte, 2000), 0644)
		os.Chtimes(p, oldMP4, oldMP4)
	}

	// 超出等待期的遗留未合并 .flv（48 小时前）不应顶替并导致已完成成片被删
	staleRaw := time.Now().Add(-48 * time.Hour)
	for _, name := range []string{
		"[2026-05-03 20-00-00][s][stale]001.flv",
		"[2026-05-03 20-00-00][s][stale]002.flv",
	} {
		p := filepath.Join(streamerDir, name)
		os.WriteFile(p, make([]byte, 2000), 0644)
		os.Chtimes(p, staleRaw, staleRaw)
	}

	svc := &CleanService{config: &cfg, logger: zap.NewNop()}
	candidates, _ := svc.collectCandidates(dir, "", cfg)
	if len(candidates) != 2 {
		t.Fatalf("期望仅 2 个遗留 .flv 成为候选，得到 %d: %+v", len(candidates), candidates)
	}
	for _, c := range candidates {
		if filepath.Ext(c.Name) != ".flv" {
			t.Errorf("已完成成片 %s 不应先于未合并遗留分片被淘汰", c.Name)
		}
	}
}

func TestCleanService_DeleteFiles_MaxDeleteCompanionAndActiveSkip(t *testing.T) {
	dir := t.TempDir()
	streamerDir := filepath.Join(dir, "streamer_del")
	os.MkdirAll(streamerDir, 0755)

	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	cfg.MaxDeletePerRun = 2

	svc := &CleanService{config: &cfg, logger: zap.NewNop()}

	// 创建一个正在持续写入的文件（应被 1 秒双快照跳过不删）
	growingPath := filepath.Join(streamerDir, "growing.mp4")
	os.WriteFile(growingPath, make([]byte, 1024), 0644)
	doneWriting := make(chan struct{})
	go func() {
		defer close(doneWriting)
		f, err := os.OpenFile(growingPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()
		for i := 0; i < 12; i++ {
			time.Sleep(120 * time.Millisecond)
			_, _ = f.Write(make([]byte, 256))
		}
	}()

	// 创建 3 个静态旧文件及其伴生文件（MaxDeletePerRun=2 应仅删前 2 个并级联清伴生）
	static1 := filepath.Join(streamerDir, "old1.mp4")
	static2 := filepath.Join(streamerDir, "old2.mp4")
	static3 := filepath.Join(streamerDir, "old3.mp4")
	for _, p := range []string{static1, static2, static3} {
		os.WriteFile(p, make([]byte, 2048), 0644)
		os.WriteFile(strings.TrimSuffix(p, ".mp4")+".xml", []byte("<i/>"), 0644)
		os.WriteFile(strings.TrimSuffix(p, ".mp4")+".cover.jpg", []byte("jpg"), 0644)
	}

	candidates := []candidateFile{
		{Path: growingPath, Name: "growing.mp4", Size: 1024, Mtime: 100},
		{Path: static1, Name: "old1.mp4", Size: 2048, Mtime: 200},
		{Path: static2, Name: "old2.mp4", Size: 2048, Mtime: 300},
		{Path: static3, Name: "old3.mp4", Size: 2048, Mtime: 400},
	}

	deleted, freed, truncated := svc.deleteFiles(context.Background(), candidates, 0, cfg, func(string) {})
	<-doneWriting

	if deleted != 2 || freed != 4096 || !truncated {
		t.Fatalf("unexpected deleteFiles summary: deleted=%d freed=%d truncated=%v", deleted, freed, truncated)
	}
	if _, err := os.Stat(growingPath); err != nil {
		t.Errorf("actively writing file should NOT be deleted: %v", err)
	}
	for _, p := range []string{static1, static2} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("expected %s to be deleted", p)
		}
		if _, err := os.Stat(strings.TrimSuffix(p, ".mp4") + ".xml"); !os.IsNotExist(err) {
			t.Errorf("expected companion xml of %s to be cascade deleted", p)
		}
		if _, err := os.Stat(strings.TrimSuffix(p, ".mp4") + ".cover.jpg"); !os.IsNotExist(err) {
			t.Errorf("expected companion cover of %s to be cascade deleted", p)
		}
	}
	if _, err := os.Stat(static3); err != nil {
		t.Errorf("expected old3.mp4 to remain untouched after hitting MaxDeletePerRun: %v", err)
	}
	if _, err := os.Stat(strings.TrimSuffix(static3, ".mp4") + ".xml"); err != nil {
		t.Errorf("expected companion xml of old3.mp4 to remain untouched: %v", err)
	}
}
