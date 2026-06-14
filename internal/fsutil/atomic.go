// Package fsutil 提供崩溃安全的文件操作原语。
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWrite 将数据写入指定路径并 fsync，确保数据刷入持久存储。
// 调用方负责后续的 os.Rename 操作。写入或 fsync 失败时自动清理临时文件。
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	return f.Close()
}

// AtomicSave 原子保存数据到文件：自动创建父目录 → 写入 .tmp → fsync → rename。
// 崩溃时要么保留旧文件，要么已完整写入新文件，不会出现半写状态。
func AtomicSave(path string, data []byte, perm os.FileMode) error {
	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	tmp := path + ".tmp"
	if err := AtomicWrite(tmp, data, perm); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("原子替换文件失败: %w", err)
	}

	return nil
}
