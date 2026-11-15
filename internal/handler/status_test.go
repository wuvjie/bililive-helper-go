package handler

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"bililive-helper-go/internal/config"
	"bililive-helper-go/internal/utils"
)

func TestScanAllStreamersAt_LiveDetection(t *testing.T) {
	root := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = root
	h := &Handler{config: &cfg}

	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, loc)

	writeVideo := func(streamer, filename string, size int, modTime time.Time) {
		t.Helper()
		dir := filepath.Join(root, streamer)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, make([]byte, size), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("chtimes %s: %v", path, err)
		}
	}

	// Case 1: 开播时间在 2 小时前（文件名时间），但未合并分片 30 秒前仍在持续写入 -> is_live=true 且 mtime 更新为写入时间
	liveName := "[2026-06-01 10-00-00][1001][午间直播].flv"
	liveModTime := now.Add(-30 * time.Second)
	writeVideo("主播A", liveName, 4096, liveModTime)

	// Case 2: 已合并成品虽然 30 秒前刚生成，但属于 -合并版 -> is_live=false
	mergedName := "[2026-06-01 10-00-00][1002][昨日回放]-合并版.mp4"
	writeVideo("主播B", mergedName, 2048, now.Add(-30*time.Second))

	// Case 3: 普通未合并分片距今 10 分钟未写入 -> 超出 5 分钟活跃窗口，is_live=false，mtime 回退到文件名开播时间
	idleName := "[2026-06-01 10-00-00][1003][已下播].flv"
	idleModTime := now.Add(-10 * time.Minute)
	writeVideo("主播C", idleName, 1024, idleModTime)
	_, parsedIdleTime, ok := utils.ParseFilename(idleName)
	if !ok {
		t.Fatalf("expected ParseFilename(%q) to succeed", idleName)
	}

	// Case 4: 同目录既有刚生成的合并成品又有超过 5 分钟未写入的旧分片 -> 不得被合并成品误判为正在直播
	writeVideo("主播D", "[2026-06-01 09-00-00][1004][旧分片].flv", 512, now.Add(-20*time.Minute))
	writeVideo("主播D", "[2026-06-01 08-00-00][1004][新成片]-合并版.mp4", 512, now.Add(-10*time.Second))

	streamers, pendingCount, pendingSize := h.scanAllStreamersAt(root, now)
	if len(streamers) != 4 {
		t.Fatalf("expected 4 streamers, got %d", len(streamers))
	}
	// 主播A(4096) + 主播C(1024) + 主播D旧分片(512) 共 3 个待处理文件
	if pendingCount != 3 {
		t.Fatalf("expected pendingCount=3, got %d", pendingCount)
	}
	if pendingSize != 4096+1024+512 {
		t.Fatalf("expected pendingSize=%d, got %d", 4096+1024+512, pendingSize)
	}

	byName := make(map[string]map[string]any, len(streamers))
	for _, s := range streamers {
		name, _ := s["name"].(string)
		byName[name] = s
	}

	if gotLive, _ := byName["主播A"]["is_live"].(bool); !gotLive {
		t.Errorf("主播A is_live = false, want true")
	}
	if gotMtime, _ := byName["主播A"]["mtime"].(int64); gotMtime != liveModTime.Unix() {
		t.Errorf("主播A mtime = %d, want modTime.Unix() %d", gotMtime, liveModTime.Unix())
	}

	if gotLive, _ := byName["主播B"]["is_live"].(bool); gotLive {
		t.Errorf("主播B is_live = true for merged file, want false")
	}

	if gotLive, _ := byName["主播C"]["is_live"].(bool); gotLive {
		t.Errorf("主播C is_live = true after 10m idle, want false")
	}
	if gotMtime, _ := byName["主播C"]["mtime"].(int64); gotMtime != parsedIdleTime.Unix() {
		t.Errorf("主播C mtime = %d, want parsed filename time %d", gotMtime, parsedIdleTime.Unix())
	}

	if gotLive, _ := byName["主播D"]["is_live"].(bool); gotLive {
		t.Errorf("主播D is_live = true when only merged file is recent, want false")
	}
}

func TestScanAllStreamersAt_StandaloneMP4(t *testing.T) {
	root := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = root
	h := &Handler{config: &cfg}

	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, loc)

	writeVideo := func(streamer, filename string, size int, modTime time.Time) {
		t.Helper()
		dir := filepath.Join(root, streamer)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, make([]byte, size), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("chtimes %s: %v", path, err)
		}
	}

	// 单文件 FLV 转出的已完成 .mp4（超出 5 分钟活跃窗口、同场次仅 1 个文件）-> 计为成片，不算待合并分片
	writeVideo("主播E", "[2026-06-01 09-00-00][1005][单场回放].mp4", 4096, now.Add(-15*time.Minute))

	// 正在写入的 .mp4（在 5 分钟活跃窗口内）-> 仍处于录制/待处理状态且 is_live=true
	writeVideo("主播F", "[2026-06-01 11-50-00][1006][直播直录].mp4", 2048, now.Add(-1*time.Minute))

	// 同日同场次有多个 .mp4 分片（即使已停止写入 > 5 分钟）-> 仍需合并，均计入 pending
	writeVideo("主播G", "[2026-06-01 08-00-00][1007][多分片].mp4", 1024, now.Add(-30*time.Minute))
	writeVideo("主播G", "[2026-06-01 08-30-00][1007][多分片].mp4", 1024, now.Add(-20*time.Minute))

	streamers, pendingCount, pendingSize := h.scanAllStreamersAt(root, now)
	if len(streamers) != 3 {
		t.Fatalf("expected 3 streamers, got %d", len(streamers))
	}
	if pendingCount != 3 {
		t.Fatalf("expected pendingCount=3 (1 live mp4 + 2 multipart mp4), got %d", pendingCount)
	}
	if pendingSize != 2048+1024+1024 {
		t.Fatalf("expected pendingSize=%d, got %d", 2048+1024+1024, pendingSize)
	}

	byName := make(map[string]map[string]any, len(streamers))
	for _, s := range streamers {
		name, _ := s["name"].(string)
		byName[name] = s
	}
	if gotLive, _ := byName["主播E"]["is_live"].(bool); gotLive {
		t.Errorf("主播E is_live = true for idle standalone mp4, want false")
	}
	if gotLive, _ := byName["主播F"]["is_live"].(bool); !gotLive {
		t.Errorf("主播F is_live = false for active mp4, want true")
	}
}
