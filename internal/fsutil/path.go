package fsutil

import "strings"

// ValidatePath 校验文件名或路径安全性，防止路径穿越攻击。
// 拒绝空名、"."、".."、包含路径分隔符、管道符或空字节的输入。
func ValidatePath(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	if strings.Contains(name, "\x00") {
		return false
	}
	if strings.Contains(name, "|") {
		return false
	}
	return true
}
