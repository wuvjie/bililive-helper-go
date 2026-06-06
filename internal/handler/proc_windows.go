//go:build windows

package handler

import "os/exec"

// setProcessGroup 在 Windows 上为 no-op，因为 Windows 不支持 Unix 风格的进程组。
func setProcessGroup(cmd *exec.Cmd) {
	// Windows 不支持 Setpgid，使用默认行为
}
