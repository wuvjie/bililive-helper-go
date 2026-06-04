//go:build windows

package ffmpeg

import "os/exec"

// setProcessGroup 在 Windows 上为 no-op。
func setProcessGroup(cmd *exec.Cmd) {}

// killProcessGroup 在 Windows 上直接杀死进程（无进程组概念）。
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		cmd.Process.Kill()
	}
}
