package service

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"bililive-helper/internal/utils"
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
