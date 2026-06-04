//go:build windows

package utils

import "fmt"

// DiskUsage 保存磁盘使用信息。
type DiskUsage struct {
	Total   uint64
	Used    uint64
	Free    uint64
	UsedPct float64
}

// GetDiskUsage 在 Windows 上未实现。本项目仅面向 Linux/Docker 部署。
func GetDiskUsage(path string) (*DiskUsage, error) {
	return nil, fmt.Errorf("GetDiskUsage 不支持 Windows 平台")
}
