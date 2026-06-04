//go:build !windows

package utils

import (
	"syscall"
)

// DiskUsage 保存磁盘使用信息。
type DiskUsage struct {
	Total   uint64
	Used    uint64
	Free    uint64
	UsedPct float64
}

// GetDiskUsage 返回指定路径所在磁盘的总空间、已用空间、可用空间和使用率百分比。
// 使用 statfs 系统调用 — 适用于 Linux、macOS 和大多数 Unix 系统。
func GetDiskUsage(path string) (*DiskUsage, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, err
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	var usedPct float64
	if total > 0 {
		usedPct = float64(used) / float64(total) * 100
	}
	return &DiskUsage{Total: total, Used: used, Free: free, UsedPct: usedPct}, nil
}
