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

// MergeSubtitles 合并内容相同的相邻字幕条目。
// 合并后：序号取该组第一条的序号，开始时间用第一条的开始时间，
// 结束时间用最后一条的结束时间，文字内容不变。
// 这是翻译前的预处理步骤，可减少重复内容造成的翻译请求次数。
// 注意：Content 为空的条目不参与合并，保持原样，避免去重后空条目
// 被合并导致 SRT 结构变化（条目数减少）。
func MergeSubtitles(subs []Subtitle) []Subtitle {
	if len(subs) == 0 {
		return subs
	}

	var result []Subtitle
	current := subs[0]

	for i := 1; i < len(subs); i++ {
		// 空内容条目不参与合并，遇到就封口，保持 SRT 结构
		if current.Content == "" || subs[i].Content == "" {
			result = append(result, current)
			current = subs[i]
			continue
		}
		if subs[i].Content == current.Content {
			// 内容相同：合并时间轴（取 current 的开始，subs[i] 的结束）
			current.Timeline = mergeTimeline(current.Timeline, subs[i].Timeline)
		} else {
			result = append(result, current)
			current = subs[i]
		}
	}
	result = append(result, current)

	return result
}

// mergeTimeline 合并两条时间轴：使用第一条的开始时间和第二条的结束时间。
// 输入格式: "00:18:35,320 --> 00:18:37,320"
func mergeTimeline(first, second string) string {
	parts1 := strings.Split(first, "-->")
	parts2 := strings.Split(second, "-->")
	if len(parts1) != 2 || len(parts2) != 2 {
		// 时间轴格式异常，回退使用第一条
		return first
	}
	start := strings.TrimSpace(parts1[0])
	end := strings.TrimSpace(parts2[1])
	return start + " --> " + end
}

// DedupRepeats 清理 Whisper 幻觉产生的连续重复字符/字符组。
// 从短到长扫描所有可能的分组长度（1 字、2 字、3 字……直到 len/2），
// 任何长度的连续重复组都压缩成一组。
// 使用递归：一次扫描后如果产生了变化，说明可能出现新的相邻重复对
// （例如 ああいいああいい 先经单字去重变 あいあい，两边新对上了，
// 再递归进入双字组去重变 あい），所以递归调用自身直到完全收敛。
func DedupRepeats(content string) string {
	if content == "" {
		return content
	}

	runes := []rune(content)
	for n := 1; n*2 <= len(runes); n++ {
		collapsed := collapseGroupN(runes, n)
		if len(collapsed) != len(runes) {
			// 有变化 → 可能产生新的相邻重复对，递归继续去重
			return DedupRepeats(string(collapsed))
		}
	}
	return string(runes)
}

// collapseGroupN 把连续重复 n 个字的组压缩成一组。
// 例如 n=1: ああああ → あ，n=2: あいあいあい → あい，
// n=4: すみませんすみませんすみません → すみません
func collapseGroupN(runes []rune, n int) []rune {
	if len(runes) < n*2 {
		return runes
	}

	var result []rune
	i := 0
	for i < len(runes) {
		// 检查从 i 开始是否有连续 n 个字的重复组（至少 2 次）
		if i+n*2 <= len(runes) {
			pattern := string(runes[i : i+n])
			count := 1
			j := i + n
			for j+n <= len(runes) && string(runes[j:j+n]) == pattern {
				count++
				j += n
			}
			if count >= 2 {
				result = append(result, runes[i:i+n]...)
				i = j
				continue
			}
		}
		result = append(result, runes[i])
		i++
	}
	return result
}
