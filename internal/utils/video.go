package utils

import (
	"regexp"
	"time"
)

var filenameRegex = regexp.MustCompile(`(?i)^\[(\d{4}-\d{2}-\d{2}) (\d{2}-\d{2}-\d{2})\](\[.+?\]\[.+?\]).*\.(mp4|flv|ts)$`)

// ParseFilename 从 bililive-go 格式的文件名中解析录制时间和主播标识。
// 期望格式：[YYYY-MM-DD HH-MM-SS][streamer_id][title].ext
// 返回主播标识方括号段、解析后的时间和是否成功。
func ParseFilename(name string) (string, time.Time, bool) {
	matches := filenameRegex.FindStringSubmatch(name)
	if matches == nil {
		return "", time.Time{}, false
	}
	dt, err := time.ParseInLocation("2006-01-02 15-04-05", matches[1]+" "+matches[2], time.Local)
	if err != nil {
		return "", time.Time{}, false
	}
	return matches[3], dt, true
}
