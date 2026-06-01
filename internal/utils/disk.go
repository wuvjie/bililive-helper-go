package utils

import (
	"syscall"
)

type DiskUsage struct {
	Total   uint64
	Used    uint64
	Free    uint64
	UsedPct float64
}

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
