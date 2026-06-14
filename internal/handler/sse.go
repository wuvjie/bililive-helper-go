// sse.go 提供 Server-Sent Events 流式传输辅助函数。
package handler

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// runSSE 同步执行 fn 并通过 Server-Sent Events 流式传输进度消息。
// 进度更新会合并 — 每次 tick/notify 只发送最新消息，避免消息积压。
func (h *Handler) runSSE(c *gin.Context, task string, fn func(ctx context.Context, onProgress func(string)) string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	var latest atomic.Value
	notify := make(chan struct{}, 1)
	onProgress := func(msg string) {
		latest.Store(msg)
		select {
		case notify <- struct{}{}:
		default:
		}
	}

	done := make(chan string, 1)
	ctx := c.Request.Context()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Sprintf("❌ 内部错误: %v", r)
			}
		}()
		done <- fn(ctx, onProgress)
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastSent string
	for {
		select {
		case <-ctx.Done():
			return
		case <-notify:
			if v := latest.Load(); v != nil {
				if msg := v.(string); msg != "" && msg != lastSent {
					msg = strings.ReplaceAll(msg, "\n", "\ndata: ")
					fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
					c.Writer.Flush()
					lastSent = msg
				}
			}
		case <-ticker.C:
			if v := latest.Load(); v != nil {
				if msg := v.(string); msg != "" && msg != lastSent {
					msg = strings.ReplaceAll(msg, "\n", "\ndata: ")
					fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
					c.Writer.Flush()
					lastSent = msg
				}
			}
		case result := <-done:
			result = strings.ReplaceAll(result, "\n", "\ndata: ")
			fmt.Fprintf(c.Writer, "data: %s\n\n", result)
			fmt.Fprintf(c.Writer, "data: [END]\n\n")
			c.Writer.Flush()
			return
		}
	}
}
