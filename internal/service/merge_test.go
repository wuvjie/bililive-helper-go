package service

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"bililive-helper-go/internal/config"
	"bililive-helper-go/internal/ffmpeg"
	"bililive-helper-go/internal/utils"

	"go.uber.org/zap"
)

func TestSortByFilename_DifferentDates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "dates sorted ascending",
			input: []string{
				"[2026-06-04 00-19-00][user1][title1]001.mp4",
				"[2026-06-01 10-00-00][user1][title1]001.mp4",
				"[2026-06-03 21-06-44][user1][title1]001.mp4",
			},
			expected: []string{
				"[2026-06-01 10-00-00][user1][title1]001.mp4",
				"[2026-06-03 21-06-44][user1][title1]001.mp4",
				"[2026-06-04 00-19-00][user1][title1]001.mp4",
			},
		},
		{
			name: "different years",
			input: []string{
				"[2027-01-01 00-00-00][a][b]001.mp4",
				"[2025-12-31 23-59-59][a][b]001.mp4",
				"[2026-06-15 12-00-00][a][b]001.mp4",
			},
			expected: []string{
				"[2025-12-31 23-59-59][a][b]001.mp4",
				"[2026-06-15 12-00-00][a][b]001.mp4",
				"[2027-01-01 00-00-00][a][b]001.mp4",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortByFilename(tt.input)
			if !reflect.DeepEqual(tt.input, tt.expected) {
				t.Errorf("SortByFilename():\n  got  %v\n  want %v", tt.input, tt.expected)
			}
		})
	}
}

func TestSortByFilename_SameDateDifferentSequence(t *testing.T) {
	input := []string{
		"[2026-06-04 00-19-00][user1][title1]003.mp4",
		"[2026-06-04 00-19-00][user1][title1]001.mp4",
		"[2026-06-04 00-19-00][user1][title1]002.mp4",
	}
	expected := []string{
		"[2026-06-04 00-19-00][user1][title1]001.mp4",
		"[2026-06-04 00-19-00][user1][title1]002.mp4",
		"[2026-06-04 00-19-00][user1][title1]003.mp4",
	}
	SortByFilename(input)
	if !reflect.DeepEqual(input, expected) {
		t.Errorf("same-date sort:\n  got  %v\n  want %v", input, expected)
	}
}

func TestSortByFilename_MixedParsedAndUnparsed(t *testing.T) {
	input := []string{
		"randomfile.mp4",
		"[2026-06-04 00-19-00][user1][title1]001.mp4",
		"another_unknown.flv",
		"[2026-06-01 10-00-00][user1][title1]001.mp4",
	}
	expected := []string{
		"another_unknown.flv",
		"randomfile.mp4",
		"[2026-06-01 10-00-00][user1][title1]001.mp4",
		"[2026-06-04 00-19-00][user1][title1]001.mp4",
	}
	SortByFilename(input)
	if !reflect.DeepEqual(input, expected) {
		t.Errorf("mixed sort:\n  got  %v\n  want %v", input, expected)
	}
}

func TestSortByFilename_MergedAndOriginal(t *testing.T) {
	input := []string{
		"[2026-06-03 21-06-44][user1][title1]-合并版.mp4",
		"[2026-06-04 00-19-00][user1][title1]001.mp4",
		"[2026-06-03 21-06-44][user1][title1]001.mp4",
	}
	expected := []string{
		"[2026-06-03 21-06-44][user1][title1]-合并版.mp4",
		"[2026-06-03 21-06-44][user1][title1]001.mp4",
		"[2026-06-04 00-19-00][user1][title1]001.mp4",
	}
	SortByFilename(input)
	if !reflect.DeepEqual(input, expected) {
		t.Errorf("merged+original sort:\n  got  %v\n  want %v", input, expected)
	}
}

func TestSortByFilename_EmptySlice(t *testing.T) {
	input := []string{}
	SortByFilename(input)
	if len(input) != 0 {
		t.Errorf("empty slice: got len %d, want 0", len(input))
	}
}

func TestSortByFilename_SingleElement(t *testing.T) {
	input := []string{"[2026-06-04 00-19-00][user1][title1]001.mp4"}
	expected := []string{"[2026-06-04 00-19-00][user1][title1]001.mp4"}
	SortByFilename(input)
	if !reflect.DeepEqual(input, expected) {
		t.Errorf("single element:\n  got  %v\n  want %v", input, expected)
	}
}

func TestSortByFilename_NilSlice(t *testing.T) {
	// Should not panic
	SortByFilename(nil)
}

// --- checkFileAvailability tests ---

func TestCheckFileAvailability_AllAccessible(t *testing.T) {
	dir := t.TempDir()
	// Create two files with some content so isFileBeingWritten detects stability.
	f1 := filepath.Join(dir, "a.flv")
	f2 := filepath.Join(dir, "b.flv")
	if err := os.WriteFile(f1, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := checkFileAvailability(context.Background(), dir, []string{"a.flv", "b.flv"}); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestCheckFileAvailability_MissingFile(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "exists.flv")
	if err := os.WriteFile(f1, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	err := checkFileAvailability(context.Background(), dir, []string{"exists.flv", "missing.flv"})
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestCheckFileAvailability_EmptyFileList(t *testing.T) {
	dir := t.TempDir()
	// Empty list means no files to check — should pass.
	if err := checkFileAvailability(context.Background(), dir, []string{}); err != nil {
		t.Errorf("expected nil for empty list, got %v", err)
	}
}

func TestCheckFileAvailability_DirectoryInsteadOfFile(t *testing.T) {
	dir := t.TempDir()
	// Create a subdirectory with the expected name.
	sub := filepath.Join(dir, "sub.flv")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	// os.Stat succeeds for directories, so the function will not error
	// on the Stat check. However, the file name is a directory, and
	// checkFileAvailability does not distinguish files from directories
	// at the Stat level — this documents current behaviour.
	err := checkFileAvailability(context.Background(), dir, []string{"sub.flv"})
	// The current implementation treats the directory as accessible (nil),
	// unless isFileBeingWritten flags it. Document this behaviour.
	if err != nil {
		t.Logf("directory check returned error (acceptable): %v", err)
	} else {
		t.Log("directory treated as accessible — current implementation does not distinguish")
	}
}

func TestCheckFileAvailability_PathTraversalRejected(t *testing.T) {
	dir := t.TempDir()
	// Create a file outside the directory
	outside := filepath.Join(dir, "..", "outside.flv")
	if err := os.WriteFile(outside, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outside)

	// Attempt path traversal
	err := checkFileAvailability(context.Background(), dir, []string{"../outside.flv"})
	if err == nil {
		t.Error("expected error for path traversal, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "路径穿越") {
		t.Errorf("expected path traversal error, got: %v", err)
	}
}

func TestCheckFileAvailability_FileAlreadyMerged(t *testing.T) {
	dir := t.TempDir()
	// Source file does not exist, but its merged output does
	firstFile := "[2026-06-01 10-00-00][streamer][title]001.flv"
	mergedName := utils.MakeOutputName(firstFile)
	mergedPath := filepath.Join(dir, mergedName)
	if err := os.WriteFile(mergedPath, make([]byte, 4096), 0644); err != nil {
		t.Fatal(err)
	}

	err := checkFileAvailability(context.Background(), dir, []string{firstFile})
	if err == nil {
		t.Error("expected 'already merged' error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "文件已合并") {
		t.Errorf("expected 'already merged' error, got: %v", err)
	}
}

// --- classifyMergeFailure tests ---

func TestClassifyMergeFailure_SourceNotExist(t *testing.T) {
	dir := t.TempDir()
	firstFile := "[2026-06-01 10-00-00][streamer][title]001.flv"
	// Source file does not exist; output file also does not exist.
	got := classifyMergeFailure(dir, firstFile)
	if got != "源文件不存在" {
		t.Errorf("got %q, want %q", got, "源文件不存在")
	}
}

func TestClassifyMergeFailure_SourceTooSmall(t *testing.T) {
	dir := t.TempDir()
	firstFile := "[2026-06-01 10-00-00][streamer][title]001.flv"
	// Create a tiny source file (< 10240 bytes).
	path := filepath.Join(dir, firstFile)
	if err := os.WriteFile(path, make([]byte, 500), 0644); err != nil {
		t.Fatal(err)
	}
	got := classifyMergeFailure(dir, firstFile)
	if got == "" {
		t.Error("expected non-empty reason for small source file")
	}
	t.Logf("small source reason: %s", got)
}

func TestClassifyMergeFailure_OutputTooSmall(t *testing.T) {
	dir := t.TempDir()
	firstFile := "[2026-06-01 10-00-00][streamer][title]001.flv"
	// Create a normal-size source file.
	srcPath := filepath.Join(dir, firstFile)
	if err := os.WriteFile(srcPath, make([]byte, 20000), 0644); err != nil {
		t.Fatal(err)
	}
	// Create output file that is too small.
	outputName := utils.MakeOutputName(firstFile)
	outputPath := filepath.Join(dir, outputName)
	if err := os.WriteFile(outputPath, make([]byte, 100), 0644); err != nil {
		t.Fatal(err)
	}
	got := classifyMergeFailure(dir, firstFile)
	if got == "" {
		t.Error("expected non-empty reason for small output file")
	}
	t.Logf("small output reason: %s", got)
}

func TestClassifyMergeFailure_OutputNormalSize(t *testing.T) {
	dir := t.TempDir()
	firstFile := "[2026-06-01 10-00-00][streamer][title]001.flv"
	// Create a normal-size source file.
	srcPath := filepath.Join(dir, firstFile)
	if err := os.WriteFile(srcPath, make([]byte, 20000), 0644); err != nil {
		t.Fatal(err)
	}
	// Create output file with normal size.
	outputName := utils.MakeOutputName(firstFile)
	outputPath := filepath.Join(dir, outputName)
	if err := os.WriteFile(outputPath, make([]byte, 20000), 0644); err != nil {
		t.Fatal(err)
	}
	got := classifyMergeFailure(dir, firstFile)
	if got == "" {
		t.Error("expected non-empty reason for output validation failure")
	}
	t.Logf("normal-size output reason: %s", got)
}

func TestBuildDetail_SeparateConvertAndMergeSizes(t *testing.T) {
	const mib = 1024 * 1024
	const gib = 1024 * mib

	t.Run("convert_only", func(t *testing.T) {
		convertBytes := int64(256 * mib)
		got := buildDetail(2, 0, 0, nil, convertBytes, 0)
		want := "转换 2 个 FLV (" + utils.FormatSize(convertBytes) + ")"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("merge_only", func(t *testing.T) {
		mergeBytes := int64(3 * gib)
		got := buildDetail(0, 1, 0, nil, 0, mergeBytes)
		want := "合并 1 场次 (" + utils.FormatSize(mergeBytes) + ")"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("convert_and_merge_with_failure", func(t *testing.T) {
		convertBytes := int64(128 * mib)
		mergeBytes := int64(2 * gib)
		reasons := map[string]int{"输出校验失败": 1}
		got := buildDetail(1, 2, 1, reasons, convertBytes, mergeBytes)
		want := "转换 1 个 FLV (" + utils.FormatSize(convertBytes) + "), 合并 2 场次 (" + utils.FormatSize(mergeBytes) + "), 失败 1 项(输出校验失败 ×1)"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		if got := buildDetail(0, 0, 0, nil, 0, 0); got != "无待合并文件" {
			t.Fatalf("got %q, want %q", got, "无待合并文件")
		}
	})
}

func TestResolveUniqueMergeOutputName(t *testing.T) {
	dir := t.TempDir()
	first := "[2026-06-01 10-00-00][user1][title1]004.flv"

	if got, want := resolveUniqueMergeOutputName(dir, first), "[2026-06-01 10-00-00][user1][title1]-合并版.mp4"; got != want {
		t.Fatalf("initial output name = %q, want %q", got, want)
	}

	existing := filepath.Join(dir, "[2026-06-01 10-00-00][user1][title1]-合并版.mp4")
	if err := os.WriteFile(existing, []byte("mp4"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, want := resolveUniqueMergeOutputName(dir, first), "[2026-06-01 10-00-00][user1][title1]-合并版-2.mp4"; got != want {
		t.Fatalf("deduped output name = %q, want %q", got, want)
	}
}

func TestAlignCompanionDanmaku_MultiXMLAndCover(t *testing.T) {
	dir := t.TempDir()
	f1 := "[2026-06-01 10-00-00][user1][t1]001.flv"
	f2 := "[2026-06-01 10-30-00][user1][t2]002.flv"
	out := filepath.Join(dir, "[2026-06-01 10-00-00][user1][t1]-合并版.mp4")

	// Dummy media files so timestamp fallback (30min = 1800s) triggers cleanly even without ffprobe
	os.WriteFile(filepath.Join(dir, f1), []byte("flv1"), 0644)
	os.WriteFile(filepath.Join(dir, f2), []byte("flv2"), 0644)
	os.WriteFile(out, []byte("mp4"), 0644)

	xml1 := `<?xml version="1.0" encoding="utf-8"?>
<i>
  <chatserver>chat.bilibili.com</chatserver>
  <d p="12.500,1,25,16777215,1700000012,0,abc,0">part1</d>
  <sc ts="25.000" uid="1">sc1</sc>
</i>`
	xml2 := `<?xml version="1.0" encoding="utf-8"?>
<i>
  <chatserver>chat.bilibili.com</chatserver>
  <d p="5.000,1,25,16777215,1700001805,0,def,0">part2</d>
  <gift ts="10.250" uid="2"/>
</i>`
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f1, ".flv")+".xml"), []byte(xml1), 0644)
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f2, ".flv")+".xml"), []byte(xml2), 0644)
	os.WriteFile(filepath.Join(dir, strings.TrimSuffix(f1, ".flv")+".cover.jpg"), []byte("coverdata"), 0644)

	alignCompanionDanmaku(dir, []string{f1, f2}, out)

	mergedXmlPath := filepath.Join(dir, "[2026-06-01 10-00-00][user1][t1]-合并版.xml")
	data, err := os.ReadFile(mergedXmlPath)
	if err != nil {
		t.Fatalf("read merged xml: %v", err)
	}
	content := string(data)
	for _, wantSub := range []string{
		`p="12.500,1,`,
		`ts="25.000"`,
		`p="1805.000,1,`,
		`ts="1810.250"`,
		">part1</d>",
		">part2</d>",
	} {
		if !strings.Contains(content, wantSub) {
			t.Errorf("merged xml missing %q\ncontent:\n%s", wantSub, content)
		}
	}

	mergedCover := filepath.Join(dir, "[2026-06-01 10-00-00][user1][t1]-合并版.cover.jpg")
	if cdata, err := os.ReadFile(mergedCover); err != nil || string(cdata) != "coverdata" {
		t.Fatalf("expected merged cover image to be preserved, got err=%v", err)
	}
}

func TestCleanupTempFiles_PreservesStandaloneTS(t *testing.T) {
	root := t.TempDir()
	streamerDir := filepath.Join(root, "主播TS")
	if err := os.MkdirAll(streamerDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Normal standalone .ts recording (> 10KB) MUST NOT be deleted
	validTS := filepath.Join(streamerDir, "[2026-06-01 10-00-00][主播TS][正常录像]001.ts")
	os.WriteFile(validTS, make([]byte, 20*1024), 0644)

	// Temp files/dirs that MUST be cleaned
	os.MkdirAll(filepath.Join(streamerDir, ".merge_tmp_123"), 0755)
	os.WriteFile(filepath.Join(streamerDir, ".concat_123.txt"), []byte("file 'a'\n"), 0644)
	os.WriteFile(filepath.Join(streamerDir, "partial.mp4.tmp"), []byte("tmp"), 0644)
	os.WriteFile(filepath.Join(streamerDir, "partial.mp4.tmp.ts"), []byte("tmp"), 0644)
	os.WriteFile(filepath.Join(streamerDir, "partial.mp4.trim.mp4"), []byte("tmp"), 0644)

	cfg := config.DefaultConfig()
	cfg.TargetDir = root
	svc := &MergeService{config: &cfg, logger: zap.NewNop()}

	cleaned := svc.CleanupTempFiles()
	if cleaned != 5 {
		t.Fatalf("cleaned = %d, want 5", cleaned)
	}
	if _, err := os.Stat(validTS); err != nil {
		t.Fatalf("standalone valid .ts recording was falsely deleted: %v", err)
	}
}

func TestIntegration_MergeDirectConcatDedupAndSingleTrim(t *testing.T) {
	requireFFmpegTools(t)
	ctx := context.Background()
	root := t.TempDir()
	streamerFolder := filepath.Join(root, "主播E2E")
	if err := os.MkdirAll(streamerFolder, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.TargetDir = root
	cfg.StreamerRules = map[string]config.StreamerRule{
		"主播E2E": {
			TrimEnabled: true,
			TrimStart:   "00:00:01",
		},
	}
	svc := &MergeService{config: &cfg, logger: zap.NewNop()}
	oldTime := time.Now().Add(-2 * time.Hour).Truncate(time.Second)

	// 1. 直接 concat demuxer 合并两段 .flv + 多段 XML + 封面图
	f1 := "[2026-06-01 10-00-00][主播E2E][场1]001.flv"
	f2 := "[2026-06-01 10-05-00][主播E2E][场1改标题]002.flv"
	createValidTestClip(t, filepath.Join(streamerFolder, f1), "flv", "2", oldTime)
	createValidTestClip(t, filepath.Join(streamerFolder, f2), "flv", "2", oldTime)
	os.WriteFile(filepath.Join(streamerFolder, strings.TrimSuffix(f1, ".flv")+".xml"), []byte(`<i><d p="1.0,1,25,16777215,1,0,a,0">one</d></i>`), 0644)
	os.WriteFile(filepath.Join(streamerFolder, strings.TrimSuffix(f2, ".flv")+".xml"), []byte(`<i><d p="1.0,1,25,16777215,1,0,b,0">two</d></i>`), 0644)
	os.WriteFile(filepath.Join(streamerFolder, strings.TrimSuffix(f1, ".flv")+".cover.jpg"), []byte("coverbytes"), 0644)

	noopProgress := func(string) {}
	if outBytes, ok := svc.doMerge(ctx, []string{f1, f2}, streamerFolder, noopProgress, nil); !ok || outBytes <= 0 {
		t.Fatalf("first doMerge failed: ok=%v bytes=%d", ok, outBytes)
	}
	merged1 := filepath.Join(streamerFolder, "[2026-06-01 10-00-00][主播E2E][场1]-合并版.mp4")
	if err := ffmpeg.ValidateOutput(ctx, merged1); err != nil {
		t.Fatalf("merged1 ValidateOutput failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(streamerFolder, f1)); !os.IsNotExist(err) {
		t.Errorf("expected %s to be unlinked after merge", f1)
	}
	if _, err := os.Stat(strings.TrimSuffix(merged1, ".mp4") + ".cover.jpg"); err != nil {
		t.Errorf("expected merged cover to exist: %v", err)
	}
	xmlData, err := os.ReadFile(strings.TrimSuffix(merged1, ".mp4") + ".xml")
	if err != nil || !strings.Contains(string(xmlData), ">one</d>") || !strings.Contains(string(xmlData), ">two</d>") {
		t.Fatalf("expected merged xml to contain both danmakus, got err=%v content=%s", err, string(xmlData))
	}

	// 2. 同前缀第二批次合并应去重输出为 -合并版-2.mp4，不得覆盖 merged1
	f3 := "[2026-06-01 10-00-00][主播E2E][场1]003.flv"
	f4 := "[2026-06-01 10-00-00][主播E2E][场1]004.flv"
	createValidTestClip(t, filepath.Join(streamerFolder, f3), "flv", "2", oldTime)
	createValidTestClip(t, filepath.Join(streamerFolder, f4), "flv", "2", oldTime)
	if outBytes, ok := svc.doMerge(ctx, []string{f3, f4}, streamerFolder, noopProgress, nil); !ok || outBytes <= 0 {
		t.Fatalf("second doMerge (dedup) failed: ok=%v bytes=%d", ok, outBytes)
	}
	merged2 := filepath.Join(streamerFolder, "[2026-06-01 10-00-00][主播E2E][场1]-合并版-2.mp4")
	if err := ffmpeg.ValidateOutput(ctx, merged2); err != nil {
		t.Fatalf("merged2 ValidateOutput failed: %v", err)
	}

	// 3. 单分片 .mp4 原地裁剪保留 mtime 且打上 bh_trimmed 元数据
	singleMP4 := filepath.Join(streamerFolder, "[2026-06-01 15-00-00][主播E2E][单场直录].mp4")
	createValidTestClip(t, singleMP4, "mp4", "3", oldTime)
	outBytes, ok := svc.convertFlvToMp4(ctx, singleMP4, singleMP4, noopProgress, nil)
	if !ok || outBytes <= 0 {
		t.Fatalf("single MP4 trim failed: ok=%v bytes=%d", ok, outBytes)
	}
	if !ffmpeg.IsVideoTrimmed(ctx, singleMP4) {
		t.Errorf("expected IsVideoTrimmed=true after single MP4 trim")
	}
	if fi, err := os.Stat(singleMP4); err != nil || fi.ModTime().Unix() != oldTime.Unix() {
		t.Errorf("expected single MP4 mtime preserved (%v), got %v (err=%v)", oldTime, fi.ModTime(), err)
	}
}

