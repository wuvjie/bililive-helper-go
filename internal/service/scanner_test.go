package service

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"bililive-helper-go/internal/config"
	"bililive-helper-go/internal/ffmpeg"

	"go.uber.org/zap"
)

func TestIsFileBeingWritten_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.flv")
	got := isFileBeingWritten(context.Background(), path, 10*time.Millisecond)
	if !got {
		t.Error("expected true for nonexistent file, got false")
	}
}

func TestIsFileBeingWritten_StableFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stable.flv")
	if err := os.WriteFile(path, make([]byte, 4096), 0644); err != nil {
		t.Fatal(err)
	}
	got := isFileBeingWritten(context.Background(), path, 10*time.Millisecond)
	if got {
		t.Error("expected false for stable file, got true")
	}
}

func TestIsFileBeingWritten_SizeChanging(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "growing.flv")
	if err := os.WriteFile(path, make([]byte, 1024), 0644); err != nil {
		t.Fatal(err)
	}
	// Start a goroutine that appends data to simulate active writing.
	done := make(chan struct{})
	go func() {
		defer close(done)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()
		for i := 0; i < 5; i++ {
			time.Sleep(5 * time.Millisecond)
			f.Write(make([]byte, 512))
		}
	}()

	got := isFileBeingWritten(context.Background(), path, 50*time.Millisecond)
	<-done
	if !got {
		t.Error("expected true for changing file, got false")
	}
}

func TestIsFileSizeStable_StableFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stable.flv")
	if err := os.WriteFile(path, make([]byte, 8192), 0644); err != nil {
		t.Fatal(err)
	}
	got := isFileSizeStable(context.Background(), path, 10*time.Millisecond)
	if !got {
		t.Error("expected true for stable file, got false")
	}
}

func TestIsFileSizeStable_ChangingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "growing.flv")
	if err := os.WriteFile(path, make([]byte, 1024), 0644); err != nil {
		t.Fatal(err)
	}
	// Start a goroutine that appends data to simulate active writing.
	done := make(chan struct{})
	go func() {
		defer close(done)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()
		for i := 0; i < 5; i++ {
			time.Sleep(5 * time.Millisecond)
			f.Write(make([]byte, 512))
		}
	}()

	got := isFileSizeStable(context.Background(), path, 50*time.Millisecond)
	<-done
	if got {
		t.Error("expected false for changing file, got true")
	}
}

func TestIsFileSizeStable_NonexistentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.flv")
	got := isFileSizeStable(context.Background(), path, 10*time.Millisecond)
	if got {
		t.Error("expected false for nonexistent file, got true")
	}
}

func TestStreamerGroupKey(t *testing.T) {
	if got := streamerGroupKey("[主播A][下半场换标题]"); got != "[主播A]" {
		t.Errorf("got %q, want %q", got, "[主播A]")
	}
	if got := streamerGroupKey("raw|custom.flv"); got != "raw|custom.flv" {
		t.Errorf("got %q, want %q", got, "raw|custom.flv")
	}
}

func TestStabilityWaitDuration(t *testing.T) {
	oldMtime := time.Now().Add(-10 * time.Minute)
	if got := stabilityWaitDuration(oldMtime); got != 500*time.Millisecond {
		t.Errorf("old mtime wait = %v, want 500ms", got)
	}
	recentMtime := time.Now().Add(-1 * time.Minute)
	if got := stabilityWaitDuration(recentMtime); got != 5*time.Second {
		t.Errorf("recent mtime wait = %v, want 5s", got)
	}
}

func requireFFmpegTools(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg binary not installed")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe binary not installed")
	}
}

func createValidTestClip(t *testing.T, path, format string, durationSec string, mtime time.Time) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	args := []string{
		"-y", "-v", "error",
		"-f", "lavfi", "-i", "testsrc=size=160x120:rate=25:duration=" + durationSec,
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + durationSec,
		"-c:v", "libx264", "-preset", "ultrafast", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-shortest",
		"-f", format, path,
	}
	if err := ffmpeg.Run(ctx, ffmpeg.Options{Args: args}); err != nil {
		t.Fatalf("createValidTestClip(%s): %v", path, err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("chtimes(%s): %v", path, err)
	}
}

func TestIntegration_ScanTasks_TitleChangeDuplicateSingleTSRawAndTrimMP4(t *testing.T) {
	requireFFmpegTools(t)
	root := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = root
	cfg.GapMinutes = 30
	cfg.MergeAgeMinutes = 30
	cfg.StreamerRules = map[string]config.StreamerRule{
		"streamer_mp4_trim": {
			TrimEnabled: true,
			TrimStart:   "00:00:01",
		},
	}

	svc := &MergeService{config: &cfg, logger: zap.NewNop()}
	oldTime := time.Now().Add(-2 * time.Hour)

	// 1. 主播中途改标题：两段间隔 10 分钟（<= 30m），应聚合并生成同一个 mergeTask
	sTitle := filepath.Join(root, "streamer_title")
	os.MkdirAll(sTitle, 0755)
	createValidTestClip(t, filepath.Join(sTitle, "[2026-06-01 10-00-00][主播A][杂谈]001.flv"), "flv", "2", oldTime)
	createValidTestClip(t, filepath.Join(sTitle, "[2026-06-01 10-10-00][主播A][换标题打游戏]002.flv"), "flv", "2", oldTime)

	// 2. 同开播前缀已存在 -合并版.mp4，新批次分片绝不能被误删且应进入 mergeTask
	sDup := filepath.Join(root, "streamer_dup")
	os.MkdirAll(sDup, 0755)
	existingMerged := filepath.Join(sDup, "[2026-06-01 10-00-00][主播B][同标题]-合并版.mp4")
	seg4 := filepath.Join(sDup, "[2026-06-01 10-00-00][主播B][同标题]004.flv")
	seg5 := filepath.Join(sDup, "[2026-06-01 10-00-00][主播B][同标题]005.flv")
	createValidTestClip(t, existingMerged, "mp4", "2", oldTime)
	createValidTestClip(t, seg4, "flv", "2", oldTime)
	createValidTestClip(t, seg5, "flv", "2", oldTime)

	// 3. 单个 .ts 原片与非标准命名 .flv 应加入 convertTasks
	sConvert := filepath.Join(root, "streamer_convert")
	os.MkdirAll(sConvert, 0755)
	createValidTestClip(t, filepath.Join(sConvert, "[2026-06-01 12-00-00][主播C][TS场]001.ts"), "mpegts", "2", oldTime)
	createValidTestClip(t, filepath.Join(sConvert, "custom_recording.flv"), "flv", "2", oldTime)

	// 4. 开启裁剪规则的单分片 .mp4 应加入裁剪队列（FlvPath == Mp4Path）
	sTrim := filepath.Join(root, "streamer_mp4_trim")
	os.MkdirAll(sTrim, 0755)
	trimMP4 := filepath.Join(sTrim, "[2026-06-01 14-00-00][主播D][开播裁剪].mp4")
	createValidTestClip(t, trimMP4, "mp4", "3", oldTime)

	tasks, convertTasks, _ := svc.scanTasks(context.Background(), root, "", cfg)

	// 验证 streamer_title 与 streamer_dup 各自产生 1 个 mergeTask，且 seg4/seg5 未被误删
	if _, err := os.Stat(seg4); err != nil {
		t.Fatalf("seg4 was unexpectedly deleted: %v", err)
	}
	if _, err := os.Stat(seg5); err != nil {
		t.Fatalf("seg5 was unexpectedly deleted: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 mergeTasks (title-change + same-prefix), got %d: %+v", len(tasks), tasks)
	}

	// 验证 single TS + non-standard FLV + single MP4 trim 共 3 个 convertTasks
	if len(convertTasks) != 3 {
		t.Fatalf("expected 3 convertTasks (ts, raw flv, trim mp4), got %d: %+v", len(convertTasks), convertTasks)
	}
	hasTrimTask := false
	for _, ct := range convertTasks {
		if ct.FlvPath == trimMP4 && ct.Mp4Path == trimMP4 {
			hasTrimTask = true
		}
	}
	if !hasTrimTask {
		t.Fatalf("expected single MP4 trim task for %s, got %+v", trimMP4, convertTasks)
	}
}

func TestIntegration_CheckSingleMP4Trim_FingerprintCache(t *testing.T) {
	requireFFmpegTools(t)
	root := t.TempDir()
	folder := filepath.Join(root, "主播指纹")
	os.MkdirAll(folder, 0755)
	oldTime := time.Now().Add(-2 * time.Hour)
	mp4Path := filepath.Join(folder, "[2026-06-01 10-00-00][主播指纹][短视频].mp4")
	createValidTestClip(t, mp4Path, "mp4", "3", oldTime)

	cfg := config.DefaultConfig()
	svc := &MergeService{config: &cfg, logger: zap.NewNop()}

	// 1. 起始裁剪 5s > 视频时长 3s -> 无需裁剪且写入 checkedMP4s 缓存
	ruleLong := config.StreamerRule{TrimEnabled: true, TrimStart: "00:00:05"}
	if _, _, _, ok := svc.checkSingleMP4Trim(context.Background(), folder, "主播指纹", "[主播指纹]", mp4Path, ruleLong, 30); ok {
		t.Fatal("expected short video not queued for trim when start > duration")
	}
	if _, cached := svc.checkedMP4s.Load(mp4Path); !cached {
		t.Fatal("expected checkedMP4s cache entry to be populated")
	}

	// 2. 相同指纹下，即使传入已取消 context 也不再调用 ffprobe（直接缓存命中返回 false）
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, _, ok := svc.checkSingleMP4Trim(cancelledCtx, folder, "主播指纹", "[主播指纹]", mp4Path, ruleLong, 30); ok {
		t.Fatal("expected cached fingerprint to skip trim without probing")
	}

	// 3. 当主播修改裁剪规则后，ruleSig 变化应使缓存失效并重新入队裁剪
	ruleShort := config.StreamerRule{TrimEnabled: true, TrimStart: "00:00:01"}
	if _, _, _, ok := svc.checkSingleMP4Trim(context.Background(), folder, "主播指纹", "[主播指纹]", mp4Path, ruleShort, 30); !ok {
		t.Fatal("expected changed rule to invalidate fingerprint cache and enqueue trim")
	}
}
