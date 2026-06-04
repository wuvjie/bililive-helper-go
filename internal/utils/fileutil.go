package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// IsMergedFile 检查文件名是否表示已合并的文件（包含 "-合并版" 标记）。
func IsMergedFile(name string) bool {
	return strings.Contains(name, "-合并版")
}

// IsVideoFile 检查文件扩展名是否为支持的视频格式（.mp4、.flv、.ts）。
func IsVideoFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".mp4" || ext == ".flv" || ext == ".ts"
}

// ValidateFilename 校验文件名安全性，防止路径穿越攻击。
// 拒绝空名、"."、".."、包含路径分隔符或空字节的文件名。
func ValidateFilename(filename string) bool {
	if filename == "" || filename == "." || filename == ".." {
		return false
	}
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return false
	}
	if strings.Contains(filename, "\x00") {
		return false
	}
	if strings.Contains(filename, "|") {
		return false
	}
	return true
}

// SafeUnlink 尝试删除文件，带重试机制以处理临时文件锁定（如 Plex/Jellyfin 媒体扫描器）。
// 文件不存在时返回 nil。最多重试 3 次，每次间隔 500ms。
func SafeUnlink(path string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = os.Remove(path)
		if err == nil || os.IsNotExist(err) {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("删除文件失败 %s: %w", path, err)
}

// MakeOutputName 根据批次中第一个文件生成合并后的输出文件名。
// 标准格式：[YYYY-MM-DD HH-MM-SS][streamer][title]-合并版.ext
// 若文件名不含 "]" 则回退为 stem + "-合并版" + ext。
func MakeOutputName(firstFile string) string {
	ext := filepath.Ext(firstFile)
	stem := strings.TrimSuffix(firstFile, ext)
	// 去掉已有的 "-合并版" 后缀，避免双后缀
	stem = strings.TrimSuffix(stem, "-合并版")
	idx := strings.LastIndex(stem, "]")
	if idx > 0 {
		return stem[:idx+1] + "-合并版" + ext
	}
	return stem + "-合并版" + ext
}

// MakeMP4Name 将文件扩展名替换为 .mp4。
func MakeMP4Name(flvName string) string {
	return strings.TrimSuffix(flvName, filepath.Ext(flvName)) + ".mp4"
}
