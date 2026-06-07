// notify.go 提供 Webhook 通知功能。
// 任务完成时通过 HTTP POST 发送通知到配置的 WEBHOOK_URL。
package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

// 包级 HTTP 客户端用于 Webhook 通知 — 复用连接池。
var webhookClient = &http.Client{
	Timeout:   10 * time.Second,
	Transport: &http.Transport{Proxy: nil},
}

// NotifyWebhook 向配置的 WEBHOOK_URL 发送消息（fire-and-forget）。
// 期望接收 JSON POST 端点，请求体格式为 {"text": "...", "timestamp": "..."}。
// 使用 nil proxy 以适配无代理环境（如 Docker）。
func NotifyWebhook(message string) {
	url := os.Getenv("WEBHOOK_URL")
	if url == "" {
		return
	}
	go func() {
		payload, err := json.Marshal(map[string]string{
			"text":      "[Bililive Helper] " + message,
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		})
		if err != nil {
			return
		}
		resp, err := webhookClient.Post(url, "application/json", bytes.NewReader(payload))
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			// Webhook 返回错误状态码，静默忽略（fire-and-forget）
		}
	}()
}
