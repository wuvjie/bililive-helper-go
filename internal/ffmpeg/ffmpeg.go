package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

const DefaultTimeout = 2 * time.Hour

// Options configures an ffmpeg command execution.
type Options struct {
	Args     []string
	Timeout  time.Duration    // 0 means DefaultTimeout
	OnStdout func(line string) // called for each stdout line (for progress parsing)
}

// Run executes ffmpeg with the given options. It handles:
// - Process group creation (Setpgid) for clean termination
// - Context cancellation (kills entire process group)
// - Timeout (kills entire process group)
// - Optional stdout line callback for progress parsing
func Run(ctx context.Context, opts Options) error {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	cmd := exec.Command("ffmpeg", opts.Args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if opts.OnStdout != nil {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("创建 stdout pipe 失败: %w", err)
		}
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				opts.OnStdout(scanner.Text())
			}
		}()
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 ffmpeg 失败: %w", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		killProcessGroup(cmd)
		<-done
		return ctx.Err()
	case <-time.After(timeout):
		killProcessGroup(cmd)
		<-done
		return fmt.Errorf("ffmpeg 超时（%v）", timeout)
	}
}

// killProcessGroup sends SIGKILL to the entire process group.
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
