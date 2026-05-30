package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bililive-helper/internal/utils"

	"go.uber.org/zap"
)

// logToFile appends a timestamped line to {logDir}/{task}_log/{task}_videos.log,
// rotating the file at midnight and pruning archives older than 30 days.
func logToFile(logDir, task, message string, logger *zap.Logger) {
	dir := filepath.Join(logDir, task+"_log")
	os.MkdirAll(dir, 0755)
	logFile := filepath.Join(dir, task+"_videos.log")
	utils.RotateLogAndPrune(logFile, task+"_videos.log", 30)

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Warn("写入日志失败", zap.Error(err))
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}
