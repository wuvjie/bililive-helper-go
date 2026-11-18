package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bililive-helper-go/internal/config"
	"bililive-helper-go/internal/ffmpeg"

	"go.uber.org/zap"
)

// Edge 1: 同名 .mp4 虽 >=10KB 但容器头损坏时，getVideoFiles 与 Run 不得误删完好的 .flv 源片，
// 而应自愈清理损坏 .mp4 并从 .flv 重新转换出完好成片。
func TestEdge01_CorruptedMP4WithIntactFLVSourceSelfHeal(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播E1")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	flvName := "[2026-06-01 10-00-00][主播E1][自愈转换]001.flv"
	mp4Name := "[2026-06-01 10-00-00][主播E1][自愈转换]001.mp4"
	createCodecClip(t, filepath.Join(dir, flvName), "flv", "libx264", true, "2", oldTime)
	// 创建 20KB 假损坏的同名 mp4
	os.WriteFile(filepath.Join(dir, mp4Name), make([]byte, 20*1024), 0644)
	os.Chtimes(filepath.Join(dir, mp4Name), oldTime, oldTime)

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 1 || res.Failed != 0 {
		t.Fatalf("expected Run to reconvert 001.flv -> 001.mp4, got res=%+v err=%v", res, err)
	}
	assertClipValid(t, filepath.Join(dir, mp4Name), 1.5)
	if _, err := os.Stat(filepath.Join(dir, flvName)); !os.IsNotExist(err) {
		t.Errorf("source flv %s should be removed after valid reconversion", flvName)
	}
}

// Edge 2: 多分片批次过滤掉损坏与空文件后仅剩 1 个有效文件时，退化为单文件转换（或裁剪）流程。
func TestEdge02_BatchDegradesToSingleFileAfterCorruptFilter(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播E2")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	corrupt1 := filepath.Join(dir, "[2026-06-01 10-00-00][主播E2][降级单转换]001.flv")
	os.WriteFile(corrupt1, make([]byte, 15*1024), 0644)
	os.Chtimes(corrupt1, oldTime, oldTime)

	good2 := filepath.Join(dir, "[2026-06-01 10-02-00][主播E2][降级单转换]002.flv")
	createCodecClip(t, good2, "flv", "libx264", true, "2", oldTime)

	empty3 := filepath.Join(dir, "[2026-06-01 10-04-00][主播E2][降级单转换]003.flv")
	os.WriteFile(empty3, nil, 0644)
	os.Chtimes(empty3, oldTime, oldTime)

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 1 || res.Failed != 0 {
		t.Fatalf("expected degraded single-file convert Done=1, got res=%+v err=%v", res, err)
	}
	assertClipValid(t, filepath.Join(dir, "[2026-06-01 10-02-00][主播E2][降级单转换]002.mp4"), 1.5)
}

// Edge 3: 批次内所有分片均损坏或为空时，优雅跳过不报错不生成脏文件。
func TestEdge03_AllSegmentsCorruptOrEmptyGracefulSkip(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播E3")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	for _, name := range []string{
		"[2026-06-01 10-00-00][主播E3][全坏]001.flv",
		"[2026-06-01 10-02-00][主播E3][全坏]002.flv",
	} {
		p := filepath.Join(dir, name)
		os.WriteFile(p, []byte("corrupt-data"), 0644)
		os.Chtimes(p, oldTime, oldTime)
	}

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 0 || res.Failed != 0 {
		t.Fatalf("expected graceful skip (Done=0 Failed=0), got res=%+v err=%v", res, err)
	}
}

// Edge 4: 3 级合并管线全部失败时（例如内容全坏的伪造文件），必须保留全部源文件并清理任何临时产物。
func TestEdge04_AllMergeTiersFailPreservesSourceFiles(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播E4")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	f1 := "[2026-06-01 10-00-00][主播E4][必败]001.flv"
	f2 := "[2026-06-01 10-05-00][主播E4][必败]002.flv"
	for _, fn := range []string{f1, f2} {
		p := filepath.Join(dir, fn)
		os.WriteFile(p, make([]byte, 20*1024), 0644)
		os.Chtimes(p, oldTime, oldTime)
	}

	opLog, _ := NewOpLogger(filepath.Join(svc.config.LogDir, "merge_log"), "merge")
	defer opLog.Close()

	_, ok := svc.doMerge(context.Background(), []string{f1, f2}, dir, func(string) {}, opLog)
	if ok {
		t.Fatal("expected doMerge to return ok=false when all 3 merge tiers fail")
	}

	// 两个源分片必须完好保留，不得误删；且无残留 -合并版.mp4 与 .merge_tmp_*
	for _, fn := range []string{f1, f2} {
		if _, err := os.Stat(filepath.Join(dir, fn)); err != nil {
			t.Errorf("source %s must be preserved on merge failure: %v", fn, err)
		}
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), "-合并版") || strings.HasPrefix(e.Name(), ".merge_tmp_") {
			t.Errorf("unexpected leftover output/temp %s after failure", e.Name())
		}
	}
}

// Edge 5: 主播裁剪边界场景（视频短于 TrimStart、Start=0 或 Start>0 时总时长不足 TrimEnd、以及反向起止时间安全回退保留原片）。
func TestEdge05_TrimRuleEdgeCases(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, func(c *config.Config) {
		c.StreamerRules = map[string]config.StreamerRule{
			"S_ShortStart":          {TrimEnabled: true, TrimStart: "00:00:10"},
			"S_ShortEnd":            {TrimEnabled: true, TrimStart: "00:00:00", TrimEnd: "00:00:10"},
			"S_NonZeroShortEnd":     {TrimEnabled: true, TrimStart: "00:00:02", TrimEnd: "00:00:10"},
			"S_NonZeroExactEnd":     {TrimEnabled: true, TrimStart: "00:00:02", TrimEnd: "00:00:05"},
			"S_BadRange":            {TrimEnabled: true, TrimStart: "00:00:04", TrimEnd: "00:00:02"},
		}
	})
	oldTime := time.Now().Add(-2 * time.Hour)

	cases := []struct {
		streamer string
		dur      string
		wantOk   bool
	}{
		{"S_ShortStart", "2", true},
		{"S_ShortEnd", "2", true},
		{"S_NonZeroShortEnd", "5", true}, // 5s <= 10s (即便 start=2s < 5s，仍不足截止点 10s，绝不裁片头)
		{"S_NonZeroExactEnd", "5", true}, // 5s <= 5s (恰好等于截止点，同样整段保留)
		{"S_BadRange", "5", false},
	}

	for _, tc := range cases {
		dir := filepath.Join(root, tc.streamer)
		os.MkdirAll(dir, 0755)
		mp4 := filepath.Join(dir, "[2026-06-01 10-00-00]["+tc.streamer+"][边界].mp4")
		createCodecClip(t, mp4, "mp4", "libx264", true, tc.dur, oldTime)
		origInfo, _ := os.Stat(mp4)

		got := svc.applyStreamerTrim(context.Background(), tc.streamer, mp4, oldTime, func(string) {}, nil)
		if got != tc.wantOk {
			t.Errorf("%s: applyStreamerTrim = %v, want %v", tc.streamer, got, tc.wantOk)
		}
		assertClipValid(t, mp4, 1.5)
		newInfo, _ := os.Stat(mp4)
		if newInfo.Size() != origInfo.Size() {
			t.Errorf("%s: expected original video bytes preserved (%d vs %d)", tc.streamer, newInfo.Size(), origInfo.Size())
		}
		if ffmpeg.IsVideoTrimmed(context.Background(), mp4) {
			t.Errorf("%s: skipped/invalid video must not carry bh_trimmed metadata", tc.streamer)
		}
	}
}

// Edge 6: 特殊字符文件名（包含单引号 '、空格、&、中文标点）合并与 concat 转义测试。
func TestEdge06_SpecialCharactersInFilenames(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播_Special")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	f1 := "[2026-06-01 10-00-00][主播][Let's Play & Chill！]001.flv"
	f2 := "[2026-06-01 10-05-00][主播][Let's Play & Chill！]002.flv"
	createCodecClip(t, filepath.Join(dir, f1), "flv", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, f2), "flv", "libx264", true, "2", oldTime)

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 1 || res.Failed != 0 {
		t.Fatalf("Run with special characters failed: res=%+v err=%v", res, err)
	}
	assertClipValid(t, filepath.Join(dir, "[2026-06-01 10-00-00][主播][Let's Play & Chill！]-合并版.mp4"), 3.5)
}

// Edge 7: 弹幕边界场景（第 1 段无 XML、第 3 段空 XML、第 2/4 段含弹幕需累计前序分片时长）。
func TestEdge07_DanmakuMissingFirstSegmentAndEmptyXML(t *testing.T) {
	dir := t.TempDir()
	f1 := "[2026-06-01 10-00-00][s][t]001.flv"
	f2 := "[2026-06-01 10-10-00][s][t]002.flv" // +600s
	f3 := "[2026-06-01 10-20-00][s][t]003.flv" // +1200s
	f4 := "[2026-06-01 10-30-00][s][t]004.flv" // +1800s
	for _, fn := range []string{f1, f2, f3, f4} {
		os.WriteFile(filepath.Join(dir, fn), []byte("stub"), 0644)
	}
	// f1 无 xml；f2 有 xml；f3 空 xml；f4 有 xml
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f2, ".flv")+".xml"), []byte(`<i><d p="10.000,1,25,1,1,0,a,0">seg2</d></i>`), 0644)
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f3, ".flv")+".xml"), []byte(`<i><source>e-r</source></i>`), 0644)
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f4, ".flv")+".xml"), []byte(`<i><d p="5.500,1,25,1,2,0,b,0">seg4</d></i>`), 0644)

	out := filepath.Join(dir, "[2026-06-01 10-00-00][s][t]-合并版.mp4")
	alignCompanionDanmaku(dir, []string{f1, f2, f3, f4}, out)

	data, err := os.ReadFile(strings.TrimSuffix(out, ".mp4") + ".xml")
	if err != nil {
		t.Fatalf("merged xml missing: %v", err)
	}
	text := string(data)
	// f2 应在 600 + 10 = 610.000s；f4 应在 1800 + 5.5 = 1805.500s
	if !strings.Contains(text, `p="610.000,`) || !strings.Contains(text, `p="1805.500,`) {
		t.Fatalf("unexpected danmaku offsets in merged xml:\n%s", text)
	}
	if !strings.HasSuffix(strings.TrimSpace(text), "</i>") {
		t.Fatalf("merged xml missing closing </i> tag:\n%s", text)
	}
}

// Edge 8: 主播级锁并发冲突拦截 + CleanService 自动规避正在合并的主播。
func TestEdge08_ConcurrentStreamerLockAndCleanAvoidance(t *testing.T) {
	root, cfg, mergeSvc := newScenarioEnv(t, func(c *config.Config) {
		c.MinKeepPerStreamer = 1
		c.SafeAgeMinutes = 0
	})
	oldTime := time.Now().Add(-48 * time.Hour)

	lockedDir := filepath.Join(root, "主播Locked")
	unlockedDir := filepath.Join(root, "主播Unlocked")
	os.MkdirAll(lockedDir, 0755)
	os.MkdirAll(unlockedDir, 0755)

	for _, name := range []string{"a.mp4", "b.mp4"} {
		p1 := filepath.Join(lockedDir, name)
		p2 := filepath.Join(unlockedDir, name)
		os.WriteFile(p1, make([]byte, 2000), 0644)
		os.WriteFile(p2, make([]byte, 2000), 0644)
		os.Chtimes(p1, oldTime, oldTime)
		os.Chtimes(p2, oldTime, oldTime)
	}

	ok, sl := mergeSvc.tryLockStreamer("主播Locked")
	if !ok {
		t.Fatal("failed to acquire initial streamer lock")
	}
	if !mergeSvc.IsStreamerLocked("主播Locked") {
		t.Fatal("IsStreamerLocked should be true while locked")
	}

	// 并发手动合并同一主播应立即报冲突错误
	if _, err := mergeSvc.ManualMerge(context.Background(), "主播Locked", []string{"a.mp4", "b.mp4"}, nil); err == nil || !strings.Contains(err.Error(), "正在执行中") {
		t.Errorf("expected lock conflict error from ManualMerge, got: %v", err)
	}

	// CleanService 必须跳过被锁定的主播，仅收集未锁定主播的候选文件
	cleanSvc := NewCleanService(cfg, zap.NewNop(), NewHistoryService(cfg, zap.NewNop()), mergeSvc)
	candidates, _ := cleanSvc.collectCandidates(root, "", *cfg)
	for _, c := range candidates {
		if strings.Contains(c.Path, "主播Locked") {
			t.Errorf("CleanService collected candidate from locked streamer: %s", c.Path)
		}
	}
	if len(candidates) != 1 || !strings.Contains(candidates[0].Path, "主播Unlocked") {
		t.Errorf("expected 1 candidate from 主播Unlocked, got: %+v", candidates)
	}

	mergeSvc.unlockStreamer(sl)
	if mergeSvc.IsStreamerLocked("主播Locked") {
		t.Error("IsStreamerLocked should be false after unlock")
	}
}

// Edge 9: 避让静默时段（BackupWindow）生效时，Merge / ManualMerge / Clean 均应拒绝执行且不留空日志。
func TestEdge09_BackupWindowBlocksMergeAndClean(t *testing.T) {
	root, cfg, mergeSvc := newScenarioEnv(t, func(c *config.Config) {
		now := time.Now()
		start := now.Add(-1 * time.Hour)
		end := now.Add(1 * time.Hour)
		c.BackupStartHour = start.Hour()
		c.BackupStartMinute = start.Minute()
		c.BackupEndHour = end.Hour()
		c.BackupEndMinute = end.Minute()
	})
	cleanSvc := NewCleanService(cfg, zap.NewNop(), NewHistoryService(cfg, zap.NewNop()), mergeSvc)

	if _, _, err := mergeSvc.Run(context.Background(), "", nil); err == nil || !strings.Contains(err.Error(), "避让时段") {
		t.Errorf("expected backup window error from merge.Run, got: %v", err)
	}
	os.MkdirAll(filepath.Join(root, "主播BW"), 0755)
	if _, err := mergeSvc.ManualMerge(context.Background(), "主播BW", []string{"a.flv", "b.flv"}, nil); err == nil || !strings.Contains(err.Error(), "静默时段") {
		t.Errorf("expected backup window error from ManualMerge, got: %v", err)
	}
	if _, _, err := cleanSvc.Run(context.Background(), "", nil); err == nil || !strings.Contains(err.Error(), "避让时段") {
		t.Errorf("expected backup window error from clean.Run, got: %v", err)
	}
}

// Edge 10: ManualMerge 非法输入校验（路径穿越、全合并版文件）及部分文件缺失降级合并成功。
func TestEdge10_ManualMergeValidationAndPartialMissing(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播E10")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	// 1. 路径穿越应被拒绝
	if _, err := svc.ManualMerge(context.Background(), "主播E10", []string{"../evil.flv", "b.flv"}, nil); err == nil || !strings.Contains(err.Error(), "非法文件名") {
		t.Errorf("expected path traversal rejection, got: %v", err)
	}

	// 2. 全为 -合并版 应被拒绝
	m1 := "[2026-06-01 10-00-00][s][t]-合并版.mp4"
	m2 := "[2026-06-01 11-00-00][s][t]-合并版.mp4"
	os.WriteFile(filepath.Join(dir, m1), []byte("m1"), 0644)
	os.WriteFile(filepath.Join(dir, m2), []byte("m2"), 0644)
	if _, err := svc.ManualMerge(context.Background(), "主播E10", []string{m1, m2}, nil); err == nil || !strings.Contains(err.Error(), "全部是合并版") {
		t.Errorf("expected all-merged rejection, got: %v", err)
	}

	// 3. 3 个文件中 1 个缺失、2 个有效：使用剩余 2 个降级合并成功
	f1 := "[2026-06-02 10-00-00][主播E10][部分缺失]001.flv"
	f2 := "[2026-06-02 10-05-00][主播E10][部分缺失]002.flv"
	createCodecClip(t, filepath.Join(dir, f1), "flv", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, f2), "flv", "libx264", true, "2", oldTime)

	logID, err := svc.ManualMerge(context.Background(), "主播E10", []string{"missing.flv", f2, f1}, func(string) {})
	if err != nil || logID == "" {
		t.Fatalf("expected ManualMerge with 1 missing + 2 valid files to succeed, got logID=%q err=%v", logID, err)
	}
	assertClipValid(t, filepath.Join(dir, "[2026-06-02 10-00-00][主播E10][部分缺失]-合并版.mp4"), 3.5)
}

func TestEdgeHelpers_CompileSanity(t *testing.T) {
	_ = ffmpeg.DefaultTimeout
}
