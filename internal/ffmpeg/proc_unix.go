//go:build !windows

package ffmpeg

import (
	"os/exec"
	"syscall"
)

// setProcessGroup 将子进程放入独立进程组，以便杀死时不影响父进程。
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killProcessGroup 向整个进程组发送 SIGKILL（负 PID）。
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
