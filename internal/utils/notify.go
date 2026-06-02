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
// 期望接收 JSON POST 端点，请求体格式为 {"text": "..."}。
// 使用 nil proxy 以适配无代理环境（如 Docker）。
func NotifyWebhook(message string) {
	url := os.Getenv("WEBHOOK_URL")
	if url == "" {
		return
	}
	go func() {
		payload, err := json.Marshal(map[string]string{
			"text": "[Bililive Helper] " + message,
		})
		if err != nil {
			return
		}
		resp, err := webhookClient.Post(url, "application/json", bytes.NewReader(payload))
		if err == nil {
			resp.Body.Close()
		}
	}()
}
