package service

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OpLogger 封装单次操作的日志生命周期。
// 每次合并/清理操作创建一个独立的 OpLogger，写入独立的日志文件。
type OpLogger struct {
	mu     sync.Mutex
	file   *os.File
	logID  string
	closed bool
}

// NewOpLogger 创建操作日志器。logDir 为日志目录（如 merge_log/），taskType 为操作类型（merge/clean）。
// logID 格式: {taskType}_{YYYYMMDD}_{HHMMSS}_{4位hex}，兼顾时间排序与唯一性。
// 创建失败时返回 error，调用方可降级为仅 SSE 输出。
func NewOpLogger(logDir, taskType string) (*OpLogger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}
	logID := fmt.Sprintf("%s_%s_%04x", taskType, time.Now().Format("20060102_150405"), rand.Intn(0xFFFF))
	path := filepath.Join(logDir, "op_"+logID+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("创建日志文件失败: %w", err)
	}
	return &OpLogger{file: f, logID: logID}, nil
}

// Log 写入一行带时间戳的日志。并发安全。
func (l *OpLogger) Log(msg string) {
	if l == nil || l.closed {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.file, "[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

// ProgressFunc 返回一个 ProgressFunc，同时写入日志文件和调用 SSE 回调。
// 调用方只需传入原始的 onProgress 回调，返回值替代 onProgress 使用。
func (l *OpLogger) ProgressFunc(sseCallback func(string)) func(string) {
	return func(msg string) {
		l.Log(msg)
		if sseCallback != nil {
			sseCallback(msg)
		}
	}
}

// LogID 返回操作日志 ID，用于关联 HistoryRecord。
func (l *OpLogger) LogID() string {
	if l == nil {
		return ""
	}
	return l.logID
}

// Close 关闭日志文件。重复调用安全。
func (l *OpLogger) Close() {
	if l == nil || l.closed {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closed = true
	l.file.Close()
}
