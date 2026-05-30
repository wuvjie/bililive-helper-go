package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

func NotifyWebhook(message string) {
	url := os.Getenv("WEBHOOK_URL")
	if url == "" {
		return
	}
	go func() {
		client := &http.Client{
			Timeout:   10 * time.Second,
			Transport: &http.Transport{Proxy: nil},
		}
		payload, err := json.Marshal(map[string]string{
			"text": "[Bililive Helper] " + message,
		})
		if err != nil {
			return
		}
		resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
		if err == nil {
			resp.Body.Close()
		}
	}()
}
