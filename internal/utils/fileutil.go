package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func IsMergedFile(name string) bool {
	return strings.Contains(name, "-合并版")
}

func IsVideoFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".mp4" || ext == ".flv" || ext == ".ts"
}

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
	return true
}

// SafeUnlink attempts to delete a file with retries to handle temporary file locks
// (e.g. from Plex/Jellyfin media scanners). Returns nil if file doesn't exist.
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

func MakeOutputName(firstFile string) string {
	ext := filepath.Ext(firstFile)
	stem := strings.TrimSuffix(firstFile, ext)
	idx := strings.LastIndex(stem, "]")
	if idx > 0 {
		return stem[:idx+1] + "-合并版" + ext
	}
	return stem + "-合并版" + ext
}

func MakeMP4Name(flvName string) string {
	return strings.TrimSuffix(flvName, filepath.Ext(flvName)) + ".mp4"
}
