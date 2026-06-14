package fsutil

import (
	"fmt"
	"os"
	"time"
)

// SafeUnlink 尝试删除文件，带重试机制以处理临时文件锁定（如 Plex/Jellyfin 媒体扫描器）。
// 文件不存在时返回 nil。最多重试 3 次，每次间隔 500ms。
func SafeUnlink(path string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = os.Remove(path)
		if err == nil || os.IsNotExist(err) {
			return nil
		}
		if i < 2 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	return fmt.Errorf("删除文件失败 %s: %w", path, err)
}
