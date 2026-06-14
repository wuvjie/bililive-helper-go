package fsutil

import (
	"os"
	"path/filepath"
)

// StreamerDir 表示一个主播的录制目录。
type StreamerDir struct {
	Name  string        // 主播目录名
	Path  string        // 完整路径
	Files []os.DirEntry // 目录下的所有文件
}

// ScanStreamerDirs 扫描录制根目录，返回所有主播子目录及其文件列表。
// 跳过空目录和非目录条目。用于替代散落各处的 os.ReadDir + 遍历循环。
func ScanStreamerDirs(root string) ([]StreamerDir, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var dirs []StreamerDir
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(root, entry.Name())
		files, err := os.ReadDir(dirPath)
		if err != nil {
			continue // 跳过无法读取的目录
		}

		dirs = append(dirs, StreamerDir{
			Name:  entry.Name(),
			Path:  dirPath,
			Files: files,
		})
	}

	return dirs, nil
}
