//go:build windows

package utils

import (
	"syscall"
	"unsafe"
)

// DiskUsage 保存磁盘使用信息。
type DiskUsage struct {
	Total   uint64
	Used    uint64
	Free    uint64
	UsedPct float64
}

// GetDiskUsage 获取指定路径所在磁盘的使用情况（Windows 实现）。
func GetDiskUsage(path string) (*DiskUsage, error) {
	h := syscall.MustLoadDLL("kernel32.dll")
	defer h.Release()
	f := h.MustFindProc("GetDiskFreeSpaceExW")

	var freeBytes, totalBytes, totalFreeBytes int64
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	ret, _, err := f.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytes)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if ret == 0 {
		return nil, err
	}

	used := uint64(totalBytes - freeBytes)
	var pct float64
	if totalBytes > 0 {
		pct = float64(used) / float64(totalBytes) * 100
	}
	return &DiskUsage{
		Total:   uint64(totalBytes),
		Used:    used,
		Free:    uint64(freeBytes),
		UsedPct: pct,
	}, nil
}
