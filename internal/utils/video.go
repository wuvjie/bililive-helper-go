package utils

import (
	"regexp"
	"strings"
	"time"
)

var filenameRegex = regexp.MustCompile(`(?i)^\[(\d{4}-\d{2}-\d{2}) (\d{2}-\d{2}-\d{2})\](\[.+?\]\[.+?\]).*\.(mp4|flv|ts)$`)

var defaultLoc *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		defaultLoc = loc
	} else {
		defaultLoc = time.FixedZone("CST", 8*3600)
	}
}

// ParseFilename 从 bililive-go 格式的文件名中解析录制时间和主播标识。
// 期望格式：[YYYY-MM-DD HH-MM-SS][streamer_id][title].ext
// 返回主播标识方括号段、解析后的时间和是否成功。
func ParseFilename(name string) (string, time.Time, bool) {
	matches := filenameRegex.FindStringSubmatch(name)
	if matches == nil {
		return "", time.Time{}, false
	}
	// 优先使用当前系统本地时区；若系统为 UTC（如未指定 TZ 环境变量的容器环境），兜底按 Asia/Shanghai 解析
	loc := time.Local
	if loc == time.UTC || loc.String() == "UTC" || loc.String() == "" {
		loc = defaultLoc
	}
	dt, err := time.ParseInLocation("2006-01-02 15-04-05", matches[1]+" "+matches[2], loc)
	if err != nil {
		return "", time.Time{}, false
	}
	return matches[3], dt, true
}

// StreamerKeyFromKey 从 ParseFilename 返回的 key（格式 "[主播名][房间标题]"）中提取第一个方括号段 "[主播名]"。
// 用于按主播维度而非标题维度聚类同一场次直播，防止主播中途修改标题导致分片无法合并。
// 若不含 "]" 则原样返回 key。
func StreamerKeyFromKey(key string) string {
	if idx := strings.Index(key, "]"); idx >= 0 {
		return key[:idx+1]
	}
	return key
}
