package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// ContainsAny 检查字符串 s 是否包含任意一个关键词（不区分大小写）。
// 用于白名单匹配（文件名和主播名）。
func ContainsAny(s string, keywords []string) bool {
	s = strings.ToLower(s)
	for _, kw := range keywords {
		if strings.Contains(s, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// RandomHex 生成密码学安全的随机十六进制字符串。
// 参数 n 为字节数，返回 2n 个十六进制字符。
func RandomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
