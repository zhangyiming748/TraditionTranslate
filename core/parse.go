package core

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Subtitle represents a single subtitle entry with index, timeline, and content.
type Subtitle struct {
	Index    int    // 字幕序号
	Timeline string // 时间轴，格式如 00:00:00,000 --> 00:00:05,000
	Content  string //  原文内容
	Zhcn     string // 译文内容
}

// ParseSubtitle parses a standard SRT subtitle file and returns a slice of Subtitle.
// The path parameter is the file path of the subtitle file.
func ParseSubtitle(path string) ([]Subtitle, error) {
	// 读取字幕文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	return parseContent(content)
}

// parseContent 内部解析字幕内容的辅助函数
func parseContent(content string) ([]Subtitle, error) {
	// 处理开头的空白行：找到第一个非空内容的位置
	content = strings.TrimSpace(content)
	if content == "" {
		return []Subtitle{}, nil
	}

	// 按一个或多个空行分割字幕块（处理多种情况：双换行、多个换行等）
	blocks := regexp.MustCompile(`\n{2,}`).Split(content, -1)

	// 匹配序号、时间轴和内容的正则表达式
	indexRegex := regexp.MustCompile(`^\s*(\d+)\s*$`)
	timelineRegex := regexp.MustCompile(`^\s*\d{2}:\d{2}:\d{2},\d{3}\s*-->\s*\d{2}:\d{2}:\d{2},\d{3}\s*$`)

	var subtitles []Subtitle

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		lines := strings.Split(block, "\n")
		if len(lines) < 2 {
			continue
		}

		// 提取序号
		var index int
		if matches := indexRegex.FindStringSubmatch(lines[0]); len(matches) > 1 {
			index, _ = strconv.Atoi(matches[1])
		} else {
			continue
		}

		// 提取时间轴
		timeline := strings.TrimSpace(lines[1])
		if !timelineRegex.MatchString(timeline) {
			continue
		}

		// 提取字幕内容（合并所有剩余行）
		var contentLines []string
		for i := 2; i < len(lines); i++ {
			contentLines = append(contentLines, strings.TrimSpace(lines[i]))
		}
		content := strings.Join(contentLines, "\n")

		subtitles = append(subtitles, Subtitle{
			Index:    index,
			Timeline: timeline,
			Content:  content,
		})
	}

	return subtitles, nil
}
