package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"bililive-helper-go/internal/config"
	"bililive-helper-go/internal/service"
	"bililive-helper-go/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

func TestScanAllStreamersAt_TitleChangeSameDayMP4Pending(t *testing.T) {
	root := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = root
	h := &Handler{config: &cfg}

	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, loc)
	dir := filepath.Join(root, "主播改标题")
	os.MkdirAll(dir, 0755)
	for _, fn := range []string{
		"[2026-06-01 08-00-00][1008][上半场聊天].mp4",
		"[2026-06-01 09-00-00][1008][下半场游戏].mp4",
	} {
		p := filepath.Join(dir, fn)
		os.WriteFile(p, make([]byte, 1024), 0644)
		os.Chtimes(p, now.Add(-30*time.Minute), now.Add(-30*time.Minute))
	}

	_, pendingCount, pendingSize := h.scanAllStreamersAt(root, now)
	if pendingCount != 2 || pendingSize != 2048 {
		t.Fatalf("同日改标题多分片应均计入 pending: got count=%d size=%d, want 2/2048", pendingCount, pendingSize)
	}
}

func TestScanAllStreamers_CacheAndInvalidate(t *testing.T) {
	root := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = root
	h := &Handler{config: &cfg}

	dir := filepath.Join(root, "主播缓存")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "video1.mp4"), make([]byte, 1024), 0644)

	streamers, pend, _ := h.scanAllStreamers(root)
	if len(streamers) != 1 || pend != 1 {
		t.Fatalf("initial scan got streamers=%d pend=%d, want 1/1", len(streamers), pend)
	}

	// TTL 内写入新文件应命中缓存
	os.WriteFile(filepath.Join(dir, "video2.mp4"), make([]byte, 1024), 0644)
	_, pendCached, _ := h.scanAllStreamers(root)
	if pendCached != 1 {
		t.Fatalf("expected cached pend=1 within TTL, got %d", pendCached)
	}

	// 使缓存失效后应立即反映最新文件数量
	h.InvalidateStreamerCache()
	_, pendAfterInvalidate, _ := h.scanAllStreamers(root)
	if pendAfterInvalidate != 2 {
		t.Fatalf("expected pend=2 after InvalidateStreamerCache, got %d", pendAfterInvalidate)
	}
}

func TestScanAllStreamersAt_SameDayMultiStandaloneAndCrossMidnight(t *testing.T) {
	root := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = root
	h := &Handler{config: &cfg}

	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 6, 2, 12, 0, 0, 0, loc)

	// 1. 同日早晚两场单分片转换完成的 MP4（10:00 与 20:00，相距 10h > 3h）-> 均为独立成片，pendingCount 必须为 0
	s1 := filepath.Join(root, "同日两场单分片")
	os.MkdirAll(s1, 0755)
	for _, fn := range []string{
		"[2026-06-01 10-00-00][2001][早场]001.mp4",
		"[2026-06-01 20-00-00][2001][晚场]001.mp4",
	} {
		p := filepath.Join(s1, fn)
		os.WriteFile(p, make([]byte, 2048), 0644)
		os.Chtimes(p, now.Add(-10*time.Hour), now.Add(-10*time.Hour))
	}

	_, pend1, _ := h.scanAllStreamersAt(root, now)
	if pend1 != 0 {
		t.Fatalf("同日早晚两场单分片 MP4 不应虚报待合并: got pendingCount=%d, want 0", pend1)
	}

	// 2. 跨午夜分片直录 MP4（23:40 与次日 00:15，相距 35m <= 3h）-> 属于同一未合并场次，pendingCount 必须为 2
	s2 := filepath.Join(root, "跨午夜连续分片")
	os.MkdirAll(s2, 0755)
	for _, fn := range []string{
		"[2026-06-01 23-40-00][2002][熬夜]001.mp4",
		"[2026-06-02 00-15-00][2002][熬夜]002.mp4",
	} {
		p := filepath.Join(s2, fn)
		os.WriteFile(p, make([]byte, 1024), 0644)
		os.Chtimes(p, now.Add(-10*time.Hour), now.Add(-10*time.Hour))
	}

	_, pend2, pendSize2 := h.scanAllStreamersAt(root, now)
	if pend2 != 2 || pendSize2 != 2048 {
		t.Fatalf("跨午夜连续 MP4 分片应均计入 pending: got count=%d size=%d, want 2/2048", pend2, pendSize2)
	}
}

func TestScanAllStreamersAt_TrimmedCopyNotMergedAndNotLive(t *testing.T) {
	root := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = root
	h := &Handler{config: &cfg}

	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 6, 2, 12, 0, 0, 0, loc)

	sDir := filepath.Join(root, "剪辑副本主播")
	if err := os.MkdirAll(sDir, 0755); err != nil {
		t.Fatal(err)
	}
	orig := filepath.Join(sDir, "[2026-06-01 10-00-00][3001][直播标题]001.mp4")
	trimmed := filepath.Join(sDir, "[2026-06-01 10-00-00][3001][直播标题]001-剪辑版.mp4")
	_ = os.WriteFile(orig, make([]byte, 2048), 0644)
	_ = os.WriteFile(trimmed, make([]byte, 1024), 0644)
	_ = os.Chtimes(orig, now.Add(-12*time.Hour), now.Add(-12*time.Hour))
	// 模拟刚刚生成的剪辑副本，不得误触发主播在线绿点或待合并计数
	_ = os.Chtimes(trimmed, now.Add(-30*time.Second), now.Add(-30*time.Second))

	streamers, pend, _ := h.scanAllStreamersAt(root, now)
	if pend != 0 {
		t.Fatalf("原片与 -剪辑版.mp4 同时间戳并存时 pendingCount 必须为 0，实际=%d", pend)
	}
	if len(streamers) != 1 {
		t.Fatalf("expected 1 streamer, got %d", len(streamers))
	}
	if isLive, _ := streamers[0]["is_live"].(bool); isLive {
		t.Fatalf("生成 -剪辑版.mp4 不应触发 is_live=true")
	}
}

func TestHandler_ProbeAndManualTrimEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TargetDir = root
	cfg.LogDir = filepath.Join(t.TempDir(), "logs")
	logger := zap.NewNop()
	history := service.NewHistoryService(&cfg, logger)
	merge := service.NewMergeService(&cfg, logger, history)
	h := &Handler{config: &cfg, logger: logger, merge: merge, history: history}

	r := gin.New()
	r.GET("/api/streamers/:name/files/probe", h.ProbeStreamerFile)
	r.POST("/api/streamers/:name/files/trim", h.ManualTrimStreamerFile)

	streamer := "测试接口主播"
	sDir := filepath.Join(root, streamer)
	if err := os.MkdirAll(sDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 1. 路径穿越被拒绝 (400)
	req := httptest.NewRequest(http.MethodGet, "/api/streamers/"+url.PathEscape(streamer)+"/files/probe?file=../bad.mp4", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on path traversal probe, got %d (%s)", w.Code, w.Body.String())
	}

	// 2. 文件不存在返回 404
	req = httptest.NewRequest(http.MethodGet, "/api/streamers/"+url.PathEscape(streamer)+"/files/probe?file=missing.mp4", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on missing file probe, got %d (%s)", w.Code, w.Body.String())
	}

	// 3. 若环境具备 ffmpeg/ffprobe，端到端测试 probe + trim API
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return
	}

	srcName := "[2026-11-18 12-00-00][测试接口主播][接口回放]001.mp4"
	srcPath := filepath.Join(sDir, srcName)
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "testsrc=duration=6:size=160x120:rate=10",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=6",
		"-c:v", "libx264", "-preset", "ultrafast", "-c:a", "aac", srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create fixture failed: %v\n%s", err, string(out))
	}
	oldTime := time.Now().Add(-10 * time.Minute)
	_ = os.Chtimes(srcPath, oldTime, oldTime)
	_ = os.WriteFile(filepath.Join(sDir, "[2026-11-18 12-00-00][测试接口主播][接口回放]001.xml"),
		[]byte(`<i><d p="2.500,1,25,16777215,1,0,u,0">hi</d></i>`), 0644)

	// Probe 校验
	probeURL := "/api/streamers/" + url.PathEscape(streamer) + "/files/probe?file=" + url.QueryEscape(srcName)
	req = httptest.NewRequest(http.MethodGet, probeURL, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("probe expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var probeResp struct {
		ProbeOK     bool    `json:"probe_ok"`
		DurationSec float64 `json:"duration_sec"`
		IsTrimmed   bool    `json:"is_trimmed"`
		HasDanmaku  bool    `json:"has_danmaku"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &probeResp); err != nil {
		t.Fatal(err)
	}
	if !probeResp.ProbeOK || probeResp.DurationSec < 5.5 || probeResp.IsTrimmed || !probeResp.HasDanmaku {
		t.Fatalf("unexpected probe response: %+v", probeResp)
	}

	// Trim POST 请求校验 (copy 模式)
	payload, _ := json.Marshal(map[string]string{
		"file":       srcName,
		"trim_start": "00:00:01",
		"trim_end":   "00:00:04",
		"mode":       "copy",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/streamers/"+url.PathEscape(streamer)+"/files/trim", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("trim expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var trimResp service.ManualTrimResult
	if err := json.Unmarshal(w.Body.Bytes(), &trimResp); err != nil {
		t.Fatal(err)
	}
	wantOut := "[2026-11-18 12-00-00][测试接口主播][接口回放]001-剪辑版.mp4"
	if trimResp.OutputFile != wantOut || !trimResp.DanmakuSynced || trimResp.LogID == "" {
		t.Fatalf("unexpected trim HTTP response: %+v", trimResp)
	}

	// 验证裁剪副本 Probe 返回 is_trimmed=true 且 is_trimmed_file=true
	probeCopyURL := "/api/streamers/" + url.PathEscape(streamer) + "/files/probe?file=" + url.QueryEscape(wantOut)
	req = httptest.NewRequest(http.MethodGet, probeCopyURL, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("probe copy expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var probeCopy struct {
		IsTrimmed     bool `json:"is_trimmed"`
		IsTrimmedFile bool `json:"is_trimmed_file"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &probeCopy)
	if !probeCopy.IsTrimmed || !probeCopy.IsTrimmedFile {
		t.Fatalf("expected trimmed metadata & filename flags true, got %+v", probeCopy)
	}
}
