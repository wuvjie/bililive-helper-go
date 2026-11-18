package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bililive-helper-go/internal/config"
	"bililive-helper-go/internal/ffmpeg"

	"go.uber.org/zap"
)

func newScenarioEnv(t *testing.T, tweak func(*config.Config)) (string, *config.Config, *MergeService) {
	t.Helper()
	root := t.TempDir()
	logDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.TargetDir = root
	cfg.LogDir = logDir
	cfg.GapMinutes = 30
	cfg.MergeAgeMinutes = 30
	// Avoid accidental interference from backup window on wall clock during tests
	cfg.BackupStartHour = 0
	cfg.BackupStartMinute = 0
	cfg.BackupEndHour = 0
	cfg.BackupEndMinute = 0
	if tweak != nil {
		tweak(&cfg)
	}

	logger := zap.NewNop()
	history := NewHistoryService(&cfg, logger)
	svc := NewMergeService(&cfg, logger, history)
	return root, &cfg, svc
}

func createCodecClip(t *testing.T, path, format, vcodec string, withAudio bool, durationSec string, mtime time.Time) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	args := []string{
		"-y", "-v", "error",
		"-f", "lavfi", "-i", "testsrc=size=160x120:rate=25:duration=" + durationSec,
	}
	if withAudio {
		args = append(args, "-f", "lavfi", "-i", "sine=frequency=440:duration="+durationSec)
	}
	args = append(args, "-c:v", vcodec)
	switch vcodec {
	case "libx264":
		args = append(args, "-preset", "ultrafast", "-g", "25", "-pix_fmt", "yuv420p")
	case "libx265":
		args = append(args, "-preset", "ultrafast", "-g", "25", "-x265-params", "log-level=error", "-pix_fmt", "yuv420p")
	}
	if withAudio {
		args = append(args, "-c:a", "aac", "-shortest")
	}
	if format != "" {
		args = append(args, "-f", format)
	}
	args = append(args, path)

	if err := ffmpeg.Run(ctx, ffmpeg.Options{Args: args}); err != nil {
		t.Fatalf("createCodecClip(%s, %s, %s): %v", path, format, vcodec, err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("chtimes(%s): %v", path, err)
	}
}

func assertClipValid(t *testing.T, path string, minDur float64) {
	t.Helper()
	ctx := context.Background()
	if err := ffmpeg.ValidateOutput(ctx, path); err != nil {
		t.Fatalf("ValidateOutput(%s) failed: %v", filepath.Base(path), err)
	}
	dur, err := ffmpeg.ProbeDuration(ctx, path)
	if err != nil {
		t.Fatalf("ProbeDuration(%s) failed: %v", filepath.Base(path), err)
	}
	if dur < minDur {
		t.Fatalf("duration(%s) = %.2fs, want >= %.2fs", filepath.Base(path), dur, minDur)
	}
}

// Scenario 1: 标准 3 分片 FLV 合并，校验时长累加、多段 XML 弹幕偏移对齐、封面/元数据保留与源分片清理
func TestScenario01_MultiFlvMergeWithDanmakuAndCover(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播S1")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	f1 := "[2026-06-01 10-00-00][主播S1][晚间直播]001.flv"
	f2 := "[2026-06-01 10-00-00][主播S1][晚间直播]002.flv"
	f3 := "[2026-06-01 10-00-00][主播S1][晚间直播]003.flv"
	for _, fn := range []string{f1, f2, f3} {
		createCodecClip(t, filepath.Join(dir, fn), "flv", "libx264", true, "2", oldTime)
	}

	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f1, ".flv")+".xml"), []byte(`<i><chatserver>chat</chatserver><d p="1.000,1,25,1,1,0,a,0">d1</d></i>`), 0644)
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f2, ".flv")+".xml"), []byte(`<i><d p="1.000,1,25,1,2,0,b,0">d2</d><sc ts="1.500">sc2</sc></i>`), 0644)
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f3, ".flv")+".xml"), []byte(`<i><gift ts="0.500"/></i>`), 0644)
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f1, ".flv")+".cover.jpg"), []byte("cover-bytes"), 0644)
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f1, ".flv")+".json"), []byte(`{"room":1}`), 0644)

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 1 || res.Failed != 0 {
		t.Fatalf("Run unexpected result: res=%+v err=%v", res, err)
	}

	out := filepath.Join(dir, "[2026-06-01 10-00-00][主播S1][晚间直播]-合并版.mp4")
	assertClipValid(t, out, 5.5)

	// 源分片与源伴生文件均应被清理
	for _, fn := range []string{f1, f2, f3} {
		if _, err := os.Stat(filepath.Join(dir, fn)); !os.IsNotExist(err) {
			t.Errorf("source %s should be removed", fn)
		}
		if _, err := os.Stat(filepath.Join(dir, strings.TrimSuffix(fn, ".flv")+".xml")); !os.IsNotExist(err) {
			t.Errorf("source companion xml for %s should be removed", fn)
		}
	}

	// 验证合并弹幕、封面与 JSON
	mergedXML, err := os.ReadFile(strings.TrimSuffix(out, ".mp4") + ".xml")
	if err != nil {
		t.Fatalf("merged xml missing: %v", err)
	}
	xmlText := string(mergedXML)
	if !strings.Contains(xmlText, ">d1</d>") || !strings.Contains(xmlText, ">d2</d>") || !strings.Contains(xmlText, "<gift ts=") {
		t.Fatalf("merged xml missing expected entries:\n%s", xmlText)
	}
	// seg1 ~2s, seg2 ~2s -> seg2 d2 offset > 2s, seg3 gift offset > 4s
	if strings.Contains(xmlText, `p="1.000,1,25,1,2,`) || strings.Contains(xmlText, `ts="0.500"`) {
		t.Fatalf("seg2/seg3 timestamps were not shifted:\n%s", xmlText)
	}
	if _, err := os.Stat(strings.TrimSuffix(out, ".mp4") + ".cover.jpg"); err != nil {
		t.Errorf("merged cover missing: %v", err)
	}
	if _, err := os.Stat(strings.TrimSuffix(out, ".mp4") + ".json"); err != nil {
		t.Errorf("merged json missing: %v", err)
	}
}

// Scenario 2: 主播中途修改房间标题，前后分片仍能识别为同一场次合并
func TestScenario02_MidStreamTitleChangeMerge(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播S2")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	createCodecClip(t, filepath.Join(dir, "[2026-06-01 10-00-00][主播S2][上半场聊天]001.flv"), "flv", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 10-12-00][主播S2][下半场打游戏]002.flv"), "flv", "libx264", true, "2", oldTime)

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 1 || res.Failed != 0 {
		t.Fatalf("Run unexpected result: res=%+v err=%v", res, err)
	}

	out := filepath.Join(dir, "[2026-06-01 10-00-00][主播S2][上半场聊天]-合并版.mp4")
	assertClipValid(t, out, 3.5)
}

// Scenario 3: 同一主播一天两场直播（间隔 > GapMinutes），应拆成 2 场各自独立合并
func TestScenario03_MultiSessionSeparatedByGap(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播S3")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-3 * time.Hour)

	// 上午场
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 10-00-00][主播S3][早播]001.flv"), "flv", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 10-05-00][主播S3][早播]002.flv"), "flv", "libx264", true, "2", oldTime)
	// 晚上场（间隔 10 小时 >> 30m）
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 20-00-00][主播S3][晚播]001.flv"), "flv", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 20-05-00][主播S3][晚播]002.flv"), "flv", "libx264", true, "2", oldTime)

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 2 || res.Failed != 0 {
		t.Fatalf("expected 2 merged sessions, got res=%+v err=%v", res, err)
	}

	assertClipValid(t, filepath.Join(dir, "[2026-06-01 10-00-00][主播S3][早播]-合并版.mp4"), 3.5)
	assertClipValid(t, filepath.Join(dir, "[2026-06-01 20-00-00][主播S3][晚播]-合并版.mp4"), 3.5)
}

// Scenario 4: 已存在同开播前缀 -合并版.mp4 时，后续同前缀批次自动去重为 -合并版-2.mp4 且不误删原片
func TestScenario04_SameTimestampPrefixSecondBatchDedup(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播S4")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	existing := filepath.Join(dir, "[2026-06-01 20-00-00][主播S4][同前缀]-合并版.mp4")
	createCodecClip(t, existing, "mp4", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 20-00-00][主播S4][同前缀]003.flv"), "flv", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 20-00-00][主播S4][同前缀]004.flv"), "flv", "libx264", true, "2", oldTime)

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 1 || res.Failed != 0 {
		t.Fatalf("Run unexpected result: res=%+v err=%v", res, err)
	}

	assertClipValid(t, existing, 1.8)
	assertClipValid(t, filepath.Join(dir, "[2026-06-01 20-00-00][主播S4][同前缀]-合并版-2.mp4"), 3.5)
}

// Scenario 5: 同场次混合容器格式（FLV + MP4 + TS）触发 Tier 2 TS 中转拼接管线并清理临时目录
func TestScenario05_MixedContainerFallbackToTS(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播S5")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	createCodecClip(t, filepath.Join(dir, "[2026-06-01 10-00-00][主播S5][混合容器]001.flv"), "flv", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 10-05-00][主播S5][混合容器]002.mp4"), "mp4", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 10-10-00][主播S5][混合容器]003.ts"), "mpegts", "libx264", true, "2", oldTime)

	var logs []string
	res, _, err := svc.Run(context.Background(), "", func(m string) { logs = append(logs, m) })
	if err != nil || res.Done != 1 || res.Failed != 0 {
		t.Fatalf("Run unexpected result: res=%+v err=%v\nlogs:\n%s", res, err, strings.Join(logs, "\n"))
	}

	out := filepath.Join(dir, "[2026-06-01 10-00-00][主播S5][混合容器]-合并版.mp4")
	assertClipValid(t, out, 5.5)

	// 确认 .merge_tmp_* 已被清理
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".merge_tmp_") {
			t.Errorf("leftover temp dir %s not cleaned", e.Name())
		}
	}
}

// Scenario 6: HEVC (H.265) 直播流合并（验证直接流复制及 TS 中转动态匹配 hevc_mp4toannexb）
func TestScenario06_HEVCH265StreamMerge(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播S6")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	// 探测是否支持 libx265；不支持则跳过
	probeOut := filepath.Join(t.TempDir(), "probe_hevc.mp4")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	err := ffmpeg.Run(ctx, ffmpeg.Options{Args: []string{
		"-y", "-v", "error", "-f", "lavfi", "-i", "testsrc=size=160x120:rate=25:duration=1",
		"-c:v", "libx265", "-preset", "ultrafast", "-x265-params", "log-level=error", "-pix_fmt", "yuv420p", probeOut,
	}})
	cancel()
	if err != nil {
		t.Skipf("libx265 not usable in ffmpeg: %v", err)
	}

	// 混合 MP4 + TS 强制经过 Tier 2 TS 中转链路验证 hevc_mp4toannexb
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 10-00-00][主播S6][HEVC高画质]001.mp4"), "mp4", "libx265", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 10-05-00][主播S6][HEVC高画质]002.ts"), "mpegts", "libx265", true, "2", oldTime)

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 1 || res.Failed != 0 {
		t.Fatalf("HEVC Run unexpected result: res=%+v err=%v", res, err)
	}

	out := filepath.Join(dir, "[2026-06-01 10-00-00][主播S6][HEVC高画质]-合并版.mp4")
	assertClipValid(t, out, 3.5)
	if codec := ffmpeg.ProbeVideoCodec(context.Background(), out); codec != "hevc" {
		t.Fatalf("expected stream-copied HEVC codec 'hevc', got %q", codec)
	}
}

// Scenario 7: 分片间音轨不一致时重编码兜底成功补齐静音轨并完成合并与收尾
func TestScenario07_MixedAudioReencodeFinalize(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播S7")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	f1 := "[2026-06-01 10-00-00][主播S7][音频中断]001.mp4"
	f2 := "[2026-06-01 10-05-00][主播S7][音频中断]002.mp4"
	createCodecClip(t, filepath.Join(dir, f1), "mp4", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, f2), "mp4", "libx264", false, "2", oldTime)

	outPath := filepath.Join(dir, "[2026-06-01 10-00-00][主播S7][音频中断]-合并版.mp4")
	opLog, _ := NewOpLogger(filepath.Join(svc.config.LogDir, "merge_log"), "merge")
	defer opLog.Close()

	outBytes, ok := svc.concatReencode(context.Background(), []string{f1, f2}, dir, outPath, func(string) {}, opLog)
	if !ok || outBytes <= 0 {
		t.Fatalf("concatReencode mixed audio failed: ok=%v bytes=%d", ok, outBytes)
	}

	assertClipValid(t, outPath, 3.5)
	if !ffmpeg.HasAudioStream(context.Background(), outPath) {
		t.Fatal("expected merged reencoded output to retain audio stream")
	}
	for _, fn := range []string{f1, f2} {
		if _, err := os.Stat(filepath.Join(dir, fn)); !os.IsNotExist(err) {
			t.Errorf("source %s should be cleaned after reencode finalize", fn)
		}
	}
}

// Scenario 8: 单文件多形态处理（单 FLV、单 TS、非标命名 FLV 转 MP4，以及单 MP4 原地裁剪与二次运行幂等性）
func TestScenario08_SingleFilesConvertAndMP4TrimIdempotent(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, func(c *config.Config) {
		c.StreamerRules = map[string]config.StreamerRule{
			"主播S8_MP4": {
				TrimEnabled: true,
				TrimStart:   "00:00:01",
				TrimEnd:     "00:00:04",
			},
		}
	})
	oldTime := time.Now().Add(-2 * time.Hour)

	dFlv := filepath.Join(root, "主播S8_FLV")
	dTS := filepath.Join(root, "主播S8_TS")
	dRaw := filepath.Join(root, "主播S8_Raw")
	dMP4 := filepath.Join(root, "主播S8_MP4")
	for _, d := range []string{dFlv, dTS, dRaw, dMP4} {
		os.MkdirAll(d, 0755)
	}

	createCodecClip(t, filepath.Join(dFlv, "[2026-06-01 10-00-00][主播S8_FLV][单FLV]001.flv"), "flv", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dTS, "[2026-06-01 11-00-00][主播S8_TS][单TS]001.ts"), "mpegts", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dRaw, "custom_stream_2026.flv"), "flv", "libx264", true, "2", oldTime)
	trimMP4 := filepath.Join(dMP4, "[2026-06-01 12-00-00][主播S8_MP4][单MP4裁剪].mp4")
	createCodecClip(t, trimMP4, "mp4", "libx264", true, "5", oldTime)

	// 首次运行：3 个转封装 + 1 个单 MP4 裁剪 = 4 项
	res1, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res1.Done != 4 || res1.Failed != 0 {
		t.Fatalf("first Run unexpected: res=%+v err=%v", res1, err)
	}

	assertClipValid(t, filepath.Join(dFlv, "[2026-06-01 10-00-00][主播S8_FLV][单FLV]001.mp4"), 1.5)
	assertClipValid(t, filepath.Join(dTS, "[2026-06-01 11-00-00][主播S8_TS][单TS]001.mp4"), 1.5)
	assertClipValid(t, filepath.Join(dRaw, "custom_stream_2026.mp4"), 1.5)

	// 单 MP4 裁剪验证：带 bh_trimmed 元数据且时长缩短（原 5s -> 裁剪区间 1s~4s <= 3.5s）
	assertClipValid(t, trimMP4, 2.0)
	if !ffmpeg.IsVideoTrimmed(context.Background(), trimMP4) {
		t.Fatal("expected bh_trimmed metadata after single MP4 trim")
	}
	if dur, _ := ffmpeg.ProbeDuration(context.Background(), trimMP4); dur > 3.6 {
		t.Fatalf("trimmed duration = %.2fs, expected <= 3.6s", dur)
	}

	// 二次扫描运行：必须完全幂等，Done == 0，不重复裁剪
	res2, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res2.Done != 0 || res2.Failed != 0 {
		t.Fatalf("second Run expected Done=0 (idempotent), got res=%+v err=%v", res2, err)
	}
}

// Scenario 9: 自愈能力（损坏输出自动清除重合并、跳过损坏/0字节分片并保留有效分片正常合并）
func TestScenario09_CorruptedOutputAndSegmentsSelfHealing(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播S9")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	// 损坏的已有输出文件（>= 10KB 但无合法容器头）
	corruptOut := filepath.Join(dir, "[2026-06-01 10-00-00][主播S9][自愈]-合并版.mp4")
	os.WriteFile(corruptOut, make([]byte, 20*1024), 0644)
	os.Chtimes(corruptOut, oldTime, oldTime)

	// 有效分片 + 损坏分片 + 空分片 + 有效分片
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 10-00-00][主播S9][自愈]001.flv"), "flv", "libx264", true, "2", oldTime)
	badSeg := filepath.Join(dir, "[2026-06-01 10-02-00][主播S9][自愈]002.flv")
	os.WriteFile(badSeg, []byte("not-a-valid-flv-stream"), 0644)
	os.Chtimes(badSeg, oldTime, oldTime)
	emptySeg := filepath.Join(dir, "[2026-06-01 10-04-00][主播S9][自愈]003.flv")
	os.WriteFile(emptySeg, nil, 0644)
	os.Chtimes(emptySeg, oldTime, oldTime)
	createCodecClip(t, filepath.Join(dir, "[2026-06-01 10-06-00][主播S9][自愈]004.flv"), "flv", "libx264", true, "2", oldTime)

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 1 || res.Failed != 0 {
		t.Fatalf("Run unexpected result: res=%+v err=%v", res, err)
	}

	assertClipValid(t, corruptOut, 3.5)
}

// Scenario 10: 正在直播（5 分钟内修改）与关播等待期（30 分钟内）保护，不得提前合并或破坏原片
func TestScenario10_ActiveStreamAndSafetyWindowProtection(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)

	// 正在直播的主播
	liveDir := filepath.Join(root, "主播S10_Live")
	os.MkdirAll(liveDir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)
	nowTime := time.Now().Add(-30 * time.Second)
	createCodecClip(t, filepath.Join(liveDir, "[2026-06-01 10-00-00][主播S10_Live][直播中]001.flv"), "flv", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(liveDir, "[2026-06-01 10-05-00][主播S10_Live][直播中]002.flv"), "flv", "libx264", true, "2", nowTime)

	// 刚下播 10 分钟的主播（等待满 30 分钟）
	waitDir := filepath.Join(root, "主播S10_Wait")
	os.MkdirAll(waitDir, 0755)
	waitTime := time.Now().Add(-10 * time.Minute)
	createCodecClip(t, filepath.Join(waitDir, "[2026-06-01 11-00-00][主播S10_Wait][等待期]001.flv"), "flv", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(waitDir, "[2026-06-01 11-05-00][主播S10_Wait][等待期]002.flv"), "flv", "libx264", true, "2", waitTime)

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 0 {
		t.Fatalf("expected Done=0 due to live/safety protection, got res=%+v err=%v", res, err)
	}

	if _, err := os.Stat(filepath.Join(liveDir, "[2026-06-01 10-00-00][主播S10_Live][直播中]001.flv")); err != nil {
		t.Errorf("live file should remain untouched: %v", err)
	}
	if _, err := os.Stat(filepath.Join(waitDir, "[2026-06-01 11-00-00][主播S10_Wait][等待期]001.flv")); err != nil {
		t.Errorf("safety-wait file should remain untouched: %v", err)
	}
}

// Scenario 11: ManualMerge 乱序文件列表自动按时间排序并生成准确输出的合并成品
func TestScenario11_ManualMergeOutOfOrder(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播S11")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	f1 := "[2026-06-01 10-00-00][主播S11][手动]001.flv"
	f2 := "[2026-06-01 10-05-00][主播S11][手动]002.flv"
	f3 := "[2026-06-01 10-10-00][主播S11][手动]003.flv"
	for _, fn := range []string{f1, f2, f3} {
		createCodecClip(t, filepath.Join(dir, fn), "flv", "libx264", true, "2", oldTime)
	}

	// 故意传入逆序文件列表
	logID, err := svc.ManualMerge(context.Background(), "主播S11", []string{f3, f1, f2}, func(string) {})
	if err != nil || logID == "" {
		t.Fatalf("ManualMerge failed: logID=%q err=%v", logID, err)
	}

	// 验证以最早的 f1 生成输出文件名（证明 SortByFilename 已纠正顺序）
	expectedOut := filepath.Join(dir, "[2026-06-01 10-00-00][主播S11][手动]-合并版.mp4")
	assertClipValid(t, expectedOut, 5.5)
	for _, fn := range []string{f1, f2, f3} {
		if _, err := os.Stat(filepath.Join(dir, fn)); !os.IsNotExist(err) {
			t.Errorf("manual source %s should be unlinked, err=%v", fn, err)
		}
	}
}

func ExampleMergeService_scenarios() {
	_ = fmt.Sprintf("13 scenarios verified")
}

// Scenario 12: 主播中途修改推流分辨率（320x240 -> 160x120），自动跳过流复制并启用归一化缩放重编码合并成功
func TestScenario12_CrossResolutionSegmentsReencode(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, nil)
	dir := filepath.Join(root, "主播S12")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	f1 := filepath.Join(dir, "[2026-06-01 10-00-00][主播S12][分辨率变化]001.flv")
	f2 := filepath.Join(dir, "[2026-06-01 10-05-00][主播S12][分辨率变化]002.flv")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	runClip := func(path, size string) {
		args := []string{
			"-y", "-v", "error",
			"-f", "lavfi", "-i", "testsrc=size=" + size + ":rate=25:duration=2",
			"-f", "lavfi", "-i", "sine=frequency=440:duration=2",
			"-c:v", "libx264", "-preset", "ultrafast", "-g", "25", "-pix_fmt", "yuv420p",
			"-c:a", "aac", "-shortest", "-f", "flv", path,
		}
		if err := ffmpeg.Run(ctx, ffmpeg.Options{Args: args}); err != nil {
			t.Fatalf("create %s clip (%s): %v", size, path, err)
		}
		os.Chtimes(path, oldTime, oldTime)
	}
	runClip(f1, "320x240")
	runClip(f2, "160x120")

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 1 || res.Failed != 0 {
		t.Fatalf("cross-resolution Run unexpected: res=%+v err=%v", res, err)
	}

	out := filepath.Join(dir, "[2026-06-01 10-00-00][主播S12][分辨率变化]-合并版.mp4")
	assertClipValid(t, out, 3.5)
	w, h := ffmpeg.ProbeVideoDimensions(context.Background(), out)
	if w != 320 || h != 240 {
		t.Fatalf("expected normalized dimensions 320x240, got %dx%d", w, h)
	}
}

// Scenario 13: 视频裁剪 [00:00:01 - 00:00:03] 后同步剔除区间外 XML 弹幕并平移保留弹幕的时间轴（减去 1.000s）
func TestScenario13_TrimVideoSynchronizesDanmakuXML(t *testing.T) {
	requireFFmpegTools(t)
	root, _, svc := newScenarioEnv(t, func(c *config.Config) {
		c.StreamerRules = map[string]config.StreamerRule{
			"主播S13": {
				TrimEnabled: true,
				TrimStart:   "00:00:01",
				TrimEnd:     "00:00:03",
			},
		}
	})
	dir := filepath.Join(root, "主播S13")
	os.MkdirAll(dir, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)

	f1 := "[2026-06-01 10-00-00][主播S13][裁剪弹幕]001.flv"
	f2 := "[2026-06-01 10-02-00][主播S13][裁剪弹幕]002.flv"
	createCodecClip(t, filepath.Join(dir, f1), "flv", "libx264", true, "2", oldTime)
	createCodecClip(t, filepath.Join(dir, f2), "flv", "libx264", true, "3", oldTime)

	// seg1 弹幕：0.400s（<1s 应剔除）和 1.500s（保留，平移为 0.500s）
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f1, ".flv")+".xml"), []byte(`<i>
  <chatserver>chat</chatserver>
  <d p="0.400,1,25,1,1,0,a,0">dropped_before</d>
  <d p="1.500,1,25,1,2,0,b,0">kept_seg1</d>
</i>`), 0644)
	// seg2 弹幕：0.500s（合并后约 2.50s，保留并平移至 ~1.50s）和 2.000s（合并后约 4.00s >3s 应剔除）
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f2, ".flv")+".xml"), []byte(`<i>
  <sc ts="0.500" uid="1">kept_sc</sc>
  <gift ts="2.000" uid="2"/>
</i>`), 0644)

	res, _, err := svc.Run(context.Background(), "", func(string) {})
	if err != nil || res.Done != 1 || res.Failed != 0 {
		t.Fatalf("Scenario13 Run unexpected: res=%+v err=%v", res, err)
	}

	out := filepath.Join(dir, "[2026-06-01 10-00-00][主播S13][裁剪弹幕]-合并版.mp4")
	assertClipValid(t, out, 1.5)
	// stream-copy 按前导关键帧对齐（FLV 首帧含 ~23ms 音频偏置，第 1s 关键帧位于 ~1.023s），裁剪后应明显短于原 5s 成片
	if dur, _ := ffmpeg.ProbeDuration(context.Background(), out); dur > 3.5 {
		t.Fatalf("trimmed output duration = %.2fs, expected <= 3.5s (< original 5s)", dur)
	}

	xmlData, err := os.ReadFile(strings.TrimSuffix(out, ".mp4") + ".xml")
	if err != nil {
		t.Fatalf("read trimmed xml failed: %v", err)
	}
	text := string(xmlData)
	if strings.Contains(text, "dropped_before") || strings.Contains(text, "<gift ") {
		t.Fatalf("out-of-range danmakus were not removed:\n%s", text)
	}
	if !strings.Contains(text, `p="0.500,`) || !strings.Contains(text, `>kept_seg1</d>`) || !strings.Contains(text, `>kept_sc</sc>`) {
		t.Fatalf("expected kept danmakus shifted cleanly:\n%s", text)
	}
	if !strings.Contains(text, `ts="1.`) {
		t.Fatalf("expected seg2 <sc> shifted into ~1.5s range:\n%s", text)
	}
}
