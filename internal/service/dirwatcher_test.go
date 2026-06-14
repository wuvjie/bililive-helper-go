package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"bililive-helper/internal/config"

	"go.uber.org/zap"
)

func TestDirWatcher_DetectsIdleStreamer(t *testing.T) {
	dir := t.TempDir()
	streamerDir := filepath.Join(dir, "test_streamer")
	os.MkdirAll(streamerDir, 0755)
	os.WriteFile(filepath.Join(streamerDir, "video.mp4"), []byte("data"), 0644)

	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	logger := zap.NewNop()

	triggered := make(chan string, 1)
	w := NewDirWatcher(&cfg, logger, 100*time.Millisecond, 200*time.Millisecond, func(streamer string) {
		triggered <- streamer
	})

	go w.Start()
	defer w.Stop()

	// 等待足够长的时间让 DirWatcher 检测到空闲
	select {
	case name := <-triggered:
		if name != "test_streamer" {
			t.Errorf("期望触发 test_streamer，得到 %s", name)
		}
	case <-time.After(3 * time.Second):
		t.Error("DirWatcher 未在预期时间内触发")
	}
}

func TestDirWatcher_NoFalsePositive(t *testing.T) {
	dir := t.TempDir()
	streamerDir := filepath.Join(dir, "active_streamer")
	os.MkdirAll(streamerDir, 0755)

	cfg := config.DefaultConfig()
	cfg.TargetDir = dir
	logger := zap.NewNop()

	triggered := make(chan string, 1)
	w := NewDirWatcher(&cfg, logger, 50*time.Millisecond, 1*time.Second, func(streamer string) {
		triggered <- streamer
	})

	go w.Start()
	defer w.Stop()

	// 在防抖时间内持续更新文件
	for i := 0; i < 10; i++ {
		os.WriteFile(filepath.Join(streamerDir, "recording.flv"), make([]byte, 100+i*10), 0644)
		time.Sleep(80 * time.Millisecond)
	}

	// 短时间内不应触发
	select {
	case <-triggered:
		t.Error("不应在文件持续更新时触发")
	case <-time.After(300 * time.Millisecond):
		// 预期：不触发
	}
}
