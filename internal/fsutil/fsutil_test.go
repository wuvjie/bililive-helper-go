package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicSave_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "test.json")
	data := []byte(`{"key":"value"}`)

	if err := AtomicSave(path, data, 0644); err != nil {
		t.Fatalf("AtomicSave 失败: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("内容不匹配:\n  got  %q\n  want %q", got, data)
	}

	// 确认 .tmp 文件已清理
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error(".tmp 文件应该已被清理")
	}
}

func TestAtomicSave_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "file.txt")

	if err := AtomicSave(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("AtomicSave 应自动创建父目录: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("内容不匹配: got %q, want %q", got, "hello")
	}
}

func TestAtomicSave_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	// 第一次写入
	if err := AtomicSave(path, []byte("old"), 0644); err != nil {
		t.Fatalf("第一次 AtomicSave 失败: %v", err)
	}

	// 第二次写入覆盖
	if err := AtomicSave(path, []byte("new"), 0644); err != nil {
		t.Fatalf("第二次 AtomicSave 失败: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("内容不匹配: got %q, want %q", got, "new")
	}
}

func TestSafeUnlink_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("data"), 0644)

	if err := SafeUnlink(path); err != nil {
		t.Fatalf("SafeUnlink 失败: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("文件应该已被删除")
	}
}

func TestSafeUnlink_NonExistent(t *testing.T) {
	// 不存在的文件应该返回 nil（不是错误）
	if err := SafeUnlink("/nonexistent/path/file.txt"); err != nil {
		t.Errorf("删除不存在的文件应返回 nil，得到: %v", err)
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"正常文件名", "test.mp4", true},
		{"中文文件名", "录像-合并版.mp4", true},
		{"带空格", "my file.txt", true},
		{"空字符串", "", false},
		{"单点", ".", false},
		{"双点", "..", false},
		{"包含 ..", "../etc/passwd", false},
		{"包含斜杠", "dir/file", false},
		{"包含反斜杠", "dir\\file", false},
		{"包含空字节", "file\x00.txt", false},
		{"包含管道符", "file|name", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePath(tt.input)
			if got != tt.want {
				t.Errorf("ValidatePath(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestScanStreamerDirs(t *testing.T) {
	dir := t.TempDir()

	// 创建主播目录结构
	os.MkdirAll(filepath.Join(dir, "streamer_a"), 0755)
	os.MkdirAll(filepath.Join(dir, "streamer_b"), 0755)
	os.WriteFile(filepath.Join(dir, "streamer_a", "video1.mp4"), []byte("v1"), 0644)
	os.WriteFile(filepath.Join(dir, "streamer_a", "video2.mp4"), []byte("v2"), 0644)
	os.WriteFile(filepath.Join(dir, "streamer_b", "video3.flv"), []byte("v3"), 0644)
	// 根目录下的文件应被忽略
	os.WriteFile(filepath.Join(dir, "root_file.txt"), []byte("ignored"), 0644)

	dirs, err := ScanStreamerDirs(dir)
	if err != nil {
		t.Fatalf("ScanStreamerDirs 失败: %v", err)
	}

	if len(dirs) != 2 {
		t.Fatalf("期望 2 个主播目录，得到 %d", len(dirs))
	}

	// 验证目录内容（不保证顺序）
	dirMap := make(map[string]StreamerDir)
	for _, d := range dirs {
		dirMap[d.Name] = d
	}

	sa, ok := dirMap["streamer_a"]
	if !ok {
		t.Fatal("缺少 streamer_a")
	}
	if len(sa.Files) != 2 {
		t.Errorf("streamer_a 应有 2 个文件，得到 %d", len(sa.Files))
	}

	sb, ok := dirMap["streamer_b"]
	if !ok {
		t.Fatal("缺少 streamer_b")
	}
	if len(sb.Files) != 1 {
		t.Errorf("streamer_b 应有 1 个文件，得到 %d", len(sb.Files))
	}
}

func TestScanStreamerDirs_EmptyRoot(t *testing.T) {
	dir := t.TempDir()

	dirs, err := ScanStreamerDirs(dir)
	if err != nil {
		t.Fatalf("ScanStreamerDirs 失败: %v", err)
	}
	if len(dirs) != 0 {
		t.Errorf("空目录应返回 0 个结果，得到 %d", len(dirs))
	}
}

func TestScanStreamerDirs_NonExistentRoot(t *testing.T) {
	_, err := ScanStreamerDirs("/nonexistent/path")
	if err == nil {
		t.Error("不存在的路径应返回错误")
	}
}
