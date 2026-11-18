// companion.go 提供录制伴生文件（B 站 XML 弹幕、封面图、元数据）的对齐合并与级联清理能力。
package service

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"bililive-helper-go/internal/ffmpeg"
	"bililive-helper-go/internal/utils"
)

var (
	danmakuDTagRegex     = regexp.MustCompile(`(<d\b[^>]*\bp=")([0-9]+(?:\.[0-9]+)?)(,[^"]*")`)
	danmakuTsTagRegex    = regexp.MustCompile(`(<(?:sc|gift|guard)\b[^>]*\bts=")([0-9]+(?:\.[0-9]+)?)(")`)
	danmakuEntryStartReg = regexp.MustCompile(`<(?:d|sc|gift|guard)\b`)

	danmakuNodeMatchers = []struct {
		fullRe *regexp.Regexp
		attrRe *regexp.Regexp
	}{
		{regexp.MustCompile(`(?s)<d\b[^>]*\bp="([0-9]+(?:\.[0-9]+)?)(,[^"]*"[^>]*>.*?</d>)`), danmakuDTagRegex},
		{regexp.MustCompile(`(?s)<sc\b[^>]*\bts="([0-9]+(?:\.[0-9]+)?)"[^>]*>.*?</sc>`), danmakuTsTagRegex},
		{regexp.MustCompile(`(?s)<gift\b[^>]*\bts="([0-9]+(?:\.[0-9]+)?)"[^>]*(?:/>|>.*?</gift>)`), danmakuTsTagRegex},
		{regexp.MustCompile(`(?s)<guard\b[^>]*\bts="([0-9]+(?:\.[0-9]+)?)"[^>]*(?:/>|>.*?</guard>)`), danmakuTsTagRegex},
	}
)

type xmlSegment struct {
	content   string
	offsetSec float64
}

// shiftDanmakuTimestamps 将 XML 弹幕内容中所有弹幕/醒目留言/礼物/舰长节点的时间戳整体加上 offsetSec 秒。
func shiftDanmakuTimestamps(content string, offsetSec float64) string {
	if offsetSec <= 0 {
		return content
	}
	shiftAttr := func(s string, re *regexp.Regexp) string {
		return re.ReplaceAllStringFunc(s, func(match string) string {
			sub := re.FindStringSubmatch(match)
			if len(sub) != 4 {
				return match
			}
			val, err := strconv.ParseFloat(sub[2], 64)
			if err != nil {
				return match
			}
			return sub[1] + strconv.FormatFloat(val+offsetSec, 'f', 3, 64) + sub[3]
		})
	}
	content = shiftAttr(content, danmakuDTagRegex)
	return shiftAttr(content, danmakuTsTagRegex)
}

// mergeDanmakuXMLs 将多个分片 XML 弹幕按累计时间偏移合并为一个完整 XML 文档。
func mergeDanmakuXMLs(segments []xmlSegment) string {
	if len(segments) == 0 {
		return ""
	}
	if len(segments) == 1 && segments[0].offsetSec <= 0 {
		return segments[0].content
	}

	first := segments[0].content
	closeIdx := strings.LastIndex(first, "</i>")
	if closeIdx < 0 {
		return shiftDanmakuTimestamps(first, segments[0].offsetSec)
	}

	var out bytes.Buffer
	firstBody := shiftDanmakuTimestamps(first[:closeIdx], segments[0].offsetSec)
	out.WriteString(strings.TrimRight(firstBody, "\r\n"))

	for _, seg := range segments[1:] {
		c := seg.content
		segClose := strings.LastIndex(c, "</i>")
		if segClose < 0 {
			segClose = len(c)
		}
		loc := danmakuEntryStartReg.FindStringIndex(c[:segClose])
		if loc == nil {
			continue
		}
		entries := strings.Trim(shiftDanmakuTimestamps(c[loc[0]:segClose], seg.offsetSec), "\r\n")
		if entries == "" {
			continue
		}
		out.WriteByte('\n')
		out.WriteString(entries)
	}

	out.WriteString("\n</i>\n")
	return out.String()
}

// estimateSegmentDurationSec 估算第 i 个分片的时长（秒）：优先使用 ffprobe，不可用时回退到与下一分片的文件名时间差。
func estimateSegmentDurationSec(folder string, files []string, idx int) float64 {
	path := filepath.Join(folder, files[idx])
	if dur, err := ffmpeg.ProbeDuration(context.Background(), path); err == nil && dur > 0 {
		return dur
	}
	if idx+1 < len(files) {
		_, dtCur, okCur := utils.ParseFilename(files[idx])
		_, dtNext, okNext := utils.ParseFilename(files[idx+1])
		if okCur && okNext && dtNext.After(dtCur) {
			return dtNext.Sub(dtCur).Seconds()
		}
	}
	return 0
}

// alignCompanionDanmaku 合并所有分片的 XML 弹幕文件（按分片时长偏移时间轴），并将封面图等伴生文件对齐到合并版输出名。
func alignCompanionDanmaku(folder string, files []string, concatOutputPath string) {
	mergedBase := strings.TrimSuffix(concatOutputPath, filepath.Ext(concatOutputPath))
	mergedXml := mergedBase + ".xml"

	if _, err := os.Stat(mergedXml); os.IsNotExist(err) {
		var segments []xmlSegment
		cumOffset := 0.0
		for i, f := range files {
			fBase := strings.TrimSuffix(f, filepath.Ext(f))
			srcXml := filepath.Join(folder, fBase+".xml")
			if data, err := os.ReadFile(srcXml); err == nil && len(data) > 0 {
				segments = append(segments, xmlSegment{
					content:   string(data),
					offsetSec: cumOffset,
				})
			}
			cumOffset += estimateSegmentDurationSec(folder, files, i)
		}
		if len(segments) > 0 {
			mergedContent := mergeDanmakuXMLs(segments)
			_ = os.WriteFile(mergedXml, []byte(mergedContent), 0644)
		}
	}

	// 对齐其他伴生文件（封面图、元数据等），防止源分片清理时一并丢失
	for _, cExt := range []string{".cover.jpg", ".json", ".nfo"} {
		target := mergedBase + cExt
		if _, err := os.Stat(target); err == nil {
			continue
		}
		for _, f := range files {
			fBase := strings.TrimSuffix(f, filepath.Ext(f))
			src := filepath.Join(folder, fBase+cExt)
			if data, err := os.ReadFile(src); err == nil {
				_ = os.WriteFile(target, data, 0644)
				break
			}
		}
	}
}

// CleanCompanionFiles 级联清理与视频同名的伴生文件（XML 弹幕、封面图、元数据等）。
func CleanCompanionFiles(videoPath string) {
	ext := filepath.Ext(videoPath)
	base := strings.TrimSuffix(videoPath, ext)
	for _, cExt := range []string{".xml", ".cover.jpg", ".json", ".nfo"} {
		cPath := base + cExt
		if _, err := os.Stat(cPath); err == nil {
			_ = utils.SafeUnlink(cPath)
		}
	}
}

// trimCompanionDanmaku 当视频按 [startSec, endSec] 裁剪后，同步裁剪并平移同名 XML 弹幕时间轴：
// 剔除位于 [startSec, endSec] 之外的弹幕/礼物节点，并将保留节点时间戳减去 startSec 秒。
func trimCompanionDanmaku(videoPath string, startSec, endSec float64) {
	if startSec <= 0 && endSec <= 0 {
		return
	}
	xmlPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".xml"
	raw, err := os.ReadFile(xmlPath)
	if err != nil || len(raw) == 0 {
		return
	}

	filterAndShift := func(s string, fullRe, attrRe *regexp.Regexp) string {
		return fullRe.ReplaceAllStringFunc(s, func(node string) string {
			m := attrRe.FindStringSubmatch(node)
			if len(m) != 4 {
				return node
			}
			tSec, parseErr := strconv.ParseFloat(m[2], 64)
			if parseErr != nil {
				return node
			}
			if (startSec > 0 && tSec < startSec) || (endSec > 0 && tSec > endSec) {
				return ""
			}
			if startSec > 0 {
				shifted := strconv.FormatFloat(tSec-startSec, 'f', 3, 64)
				return attrRe.ReplaceAllString(node, "${1}"+shifted+"${3}")
			}
			return node
		})
	}

	content := string(raw)
	for _, matcher := range danmakuNodeMatchers {
		content = filterAndShift(content, matcher.fullRe, matcher.attrRe)
	}

	// 清理移除节点后产生的空行
	lines := strings.Split(content, "\n")
	compact := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			compact = append(compact, line)
		}
	}
	_ = os.WriteFile(xmlPath, []byte(strings.Join(compact, "\n")+"\n"), 0644)
}
