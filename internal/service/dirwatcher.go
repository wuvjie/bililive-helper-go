// dirwatcher.go 提供轻量级目录变化监听。
// 通过轮询各主播目录的修改时间检测录制结束，触发合并任务。
// 比 fsnotify 更可靠（Docker/网络文件系统下 fsnotify 不稳定）。
package service

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"bililive-helper-go/internal/config"

	"go.uber.org/zap"
)

// DirWatcher 监听录制目录变化，检测到主播目录 N 分钟无变化时触发回调。
type DirWatcher struct {
	cfg      *config.Config
	logger   *zap.Logger
	onChange func(streamer string) // 检测到录制结束时的回调

	mu        sync.Mutex
	lastSeen  map[string]time.Time // 主播目录 → 最后修改时间
	interval  time.Duration        // 轮询间隔（默认 1 分钟）
	debounce  time.Duration        // 防抖时间（默认 5 分钟）
	stopCh    chan struct{}
}

// NewDirWatcher 创建目录监听器。
// interval: 轮询间隔；debounce: 判定录制结束的无变化等待时间。
func NewDirWatcher(cfg *config.Config, logger *zap.Logger, interval, debounce time.Duration, onChange func(streamer string)) *DirWatcher {
	if interval <= 0 {
		interval = time.Minute
	}
	if debounce <= 0 {
		debounce = 5 * time.Minute
	}
	return &DirWatcher{
		cfg:      cfg,
		logger:   logger,
		onChange: onChange,
		lastSeen: make(map[string]time.Time),
		interval: interval,
		debounce: debounce,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动目录监听循环。应在独立 goroutine 中调用。
func (w *DirWatcher) Start() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.check()
		}
	}
}

// Stop 停止目录监听。
func (w *DirWatcher) Stop() {
	close(w.stopCh)
}

// check 扫描所有主播目录，检测修改时间变化。
func (w *DirWatcher) check() {
	snap := w.cfg.Snapshot()
	root := snap.TargetDir

	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}

	now := time.Now()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(root, entry.Name())
		mtime := w.getLatestMtime(dirPath)
		if mtime.IsZero() {
			continue
		}

		w.mu.Lock()
		prev, exists := w.lastSeen[entry.Name()]
		w.lastSeen[entry.Name()] = mtime
		w.mu.Unlock()

		// 首次发现或时间有变化 → 记录但不触发
		if !exists || mtime.After(prev) {
			continue
		}

		// 时间无变化且超过防抖阈值 → 认为录制结束
		if now.Sub(mtime) >= w.debounce {
			w.mu.Lock()
			// 清除记录，避免重复触发
			delete(w.lastSeen, entry.Name())
			w.mu.Unlock()

			w.logger.Info("检测到录制结束",
				zap.String("streamer", entry.Name()),
				zap.Duration("idle", now.Sub(mtime)),
			)
			if w.onChange != nil {
				w.onChange(entry.Name())
			}
		}
	}
}

// getLatestMtime 获取目录中所有视频文件的最新修改时间。
func (w *DirWatcher) getLatestMtime(dirPath string) time.Time {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return time.Time{}
	}

	var latest time.Time
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		info, err := f.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}
	return latest
}
