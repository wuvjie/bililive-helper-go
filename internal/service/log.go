// log.go 提供任务日志文件的写入和轮转功能。
// 日志按天自动轮转，保留 30 天，使用互斥锁防止并发写入交错。
package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bililive-helper/internal/utils"

	"go.uber.org/zap"
)

// logMu 保护日志文件写入，防止并发 goroutine 的输出交错。
var logMu sync.Mutex

// logToFile 向任务日志文件追加带时间戳的一行日志。
// 日志路径为 {logDir}/{task}_log/{task}_videos.log，午夜自动轮转，保留 30 天。
func logToFile(logDir, task, message string, logger *zap.Logger) {
	dir := filepath.Join(logDir, task+"_log")
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Warn("创建日志目录失败", zap.String("dir", dir), zap.Error(err))
		return
	}
	logFile := filepath.Join(dir, task+"_videos.log")

	logMu.Lock()
	defer logMu.Unlock()

	utils.RotateLogAndPrune(logFile, task+"_videos.log", 30)

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Warn("写入日志失败", zap.Error(err))
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}
