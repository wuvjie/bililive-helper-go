//go:build !windows

package handler

import (
	"os/exec"
	"syscall"
)

// setProcessGroup 将子进程放入独立进程组，以便杀死时不影响父进程。
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
