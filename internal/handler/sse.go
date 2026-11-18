// sse.go 提供 Server-Sent Events 流式传输辅助函数。
package handler

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// runSSE 同步执行 fn 并通过 Server-Sent Events 流式传输进度消息。
// 使用有界缓冲队列按序推送日志并过滤连续重复行，确保结束摘要明细不被覆盖吞行；
// 客户端断开时自动取消 fn 的 context，避免 goroutine 泄漏。
func (h *Handler) runSSE(c *gin.Context, task string, fn func(ctx context.Context, onProgress func(string)) string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	msgCh := make(chan string, 256)
	var prodMu sync.Mutex
	var lastQueued string

	onProgress := func(msg string) {
		if msg == "" {
			return
		}
		prodMu.Lock()
		if msg == lastQueued {
			prodMu.Unlock()
			return
		}
		lastQueued = msg
		prodMu.Unlock()

		select {
		case msgCh <- msg:
		default:
			// 队列满时丢弃最旧一条消息，保证最新状态与摘要不丢失
			select {
			case <-msgCh:
			default:
			}
			select {
			case msgCh <- msg:
			default:
			}
		}
	}

	// 创建可取消的 context，客户端断开时自动取消 fn
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	done := make(chan string, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Sprintf("❌ 内部错误: %v", r)
			}
		}()
		done <- fn(ctx, onProgress)
	}()

	var lastRawSent string
	sendMsg := func(rawMsg string) {
		if rawMsg == "" || rawMsg == lastRawSent {
			return
		}
		lastRawSent = rawMsg
		encoded := strings.ReplaceAll(rawMsg, "\n", "\ndata: ")
		fmt.Fprintf(c.Writer, "data: %s\n\n", encoded)
		c.Writer.Flush()
	}

	for {
		select {
		case <-ctx.Done():
			cancel()
			return
		case msg := <-msgCh:
			sendMsg(msg)
		case result := <-done:
			for {
				select {
				case msg := <-msgCh:
					sendMsg(msg)
				default:
					goto sendFinal
				}
			}
		sendFinal:
			sendMsg(result)
			fmt.Fprintf(c.Writer, "data: [END]\n\n")
			c.Writer.Flush()
			return
		}
	}
}
