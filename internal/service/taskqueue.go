// taskqueue.go 提供带并发控制的任务队列。
// 限制同时运行的任务数量，避免多个 ffmpeg 进程同时打满磁盘 I/O。
package service

import (
	"context"
	"fmt"
	"sync/atomic"

	"go.uber.org/zap"
)

// TaskQueue 是带并发限制的任务队列。
type TaskQueue struct {
	sem     chan struct{}       // 并发信号量
	queue   chan *taskRequest   // 等待队列
	running atomic.Int32        // 当前运行中的任务数
	logger  *zap.Logger
}

type taskRequest struct {
	runner   TaskRunner
	streamer string
	progress ProgressFunc
	resultCh chan taskResult
}

type taskResult struct {
	result *TaskResult
	err    error
}

// NewTaskQueue 创建任务队列。maxConcurrency 为同时运行的最大任务数。
func NewTaskQueue(maxConcurrency int, logger *zap.Logger) *TaskQueue {
	if maxConcurrency < 1 {
		maxConcurrency = 1
	}
	return &TaskQueue{
		sem:    make(chan struct{}, maxConcurrency),
		queue:  make(chan *taskRequest, 100), // 缓冲队列，最多 100 个等待任务
		logger: logger,
	}
}

// Submit 提交任务到队列，阻塞等待直到任务完成并返回结果。
// 如果队列已满（100 个等待任务），返回错误。
func (q *TaskQueue) Submit(ctx context.Context, runner TaskRunner, streamer string, progress ProgressFunc) (*TaskResult, error) {
	req := &taskRequest{
		runner:   runner,
		streamer: streamer,
		progress: progress,
		resultCh: make(chan taskResult, 1),
	}

	select {
	case q.queue <- req:
		// 已入队，等待结果
		select {
		case res := <-req.resultCh:
			return res.result, res.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, fmt.Errorf("任务队列已满（100 个等待任务），请稍后重试")
	}
}

// Run 启动队列消费循环。应在独立 goroutine 中调用。
// 阻塞直到 ctx 取消。
func (q *TaskQueue) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-q.queue:
			// 获取信号量（限制并发）
			select {
			case q.sem <- struct{}{}:
				q.running.Add(1)
				go q.execute(ctx, req)
			case <-ctx.Done():
				req.resultCh <- taskResult{err: ctx.Err()}
				return
			}
		}
	}
}

// execute 执行单个任务并在完成后释放信号量。
func (q *TaskQueue) execute(ctx context.Context, req *taskRequest) {
	defer func() {
		<-q.sem
		q.running.Add(-1)
	}()

	result, err := req.runner.Run(ctx, req.streamer, req.progress)
	req.resultCh <- taskResult{result: result, err: err}
}

// RunningCount 返回当前正在执行的任务数。
func (q *TaskQueue) RunningCount() int {
	return int(q.running.Load())
}
