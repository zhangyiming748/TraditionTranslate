package core

import (
	"TraditionTranslate/util"
	"fmt"
	"os"
	"strings"
	"time"
)

// TranslateInterval 每条字幕翻译之间的间隔。
// 逐条串行翻译并留出间隔，避免请求过于频繁触发 Google 限流，
// 限流时 trans 命令会返回空内容（DST 为空）。
const TranslateInterval = 2 * time.Second

// Core 核心功能：解析字幕文件，翻译原文内容为中文，生成新的字幕文件。
// 返回值表示是否因接近 6 小时超时而提前 abort，
// 调用方据此决定是否停止处理后续文件。
func Core(inputfile string) bool {
	// 1. 解析字幕文件
	subtitles, err := ParseSubtitle(inputfile)
	if err != nil {
		fmt.Printf("解析字幕文件失败: %v\n", err)
		return false
	}

	fmt.Printf("成功解析 %d 条字幕\n", len(subtitles))

	// 第一步：清理 Whisper 幻觉产生的连续重复字符/字符组。
	for i := range subtitles {
		subtitles[i].Content = DedupRepeats(subtitles[i].Content)
	}

	// 第二步：合并内容相同的相邻字幕条目。
	beforeMerge := len(subtitles)
	subtitles = MergeSubtitles(subtitles)
	fmt.Printf("合并相同字幕: %d -> %d 条\n", beforeMerge, len(subtitles))

	// 3. 串行翻译每条字幕。
	// 每条翻译前检查剩余时间，接近超时则 abort，
	// 把已翻译的部分先写出去（partial save），不白丢。
	aborted := false
	for i := range subtitles {
		// 接近超时阈值，abort 当前文件的剩余字幕
		if util.ShouldAbort() {
			printAbortSummary(inputfile, subtitles, i)
			aborted = true
			break
		}

		if subtitles[i].Content == "" {
			continue
		}

		zhcn, transErr := Translate(subtitles[i].Content)
		if transErr != nil {
			fmt.Printf("第 %d 条字幕翻译失败: %v\n", subtitles[i].Index, transErr)
		} else {
			subtitles[i].Zhcn = zhcn
		}

		logContent := strings.ReplaceAll(subtitles[i].Content, "\n", " ")
		logZhcn := strings.ReplaceAll(subtitles[i].Zhcn, "\n", " ")
		fmt.Printf("%s [%d] 原文: %s | 译文: %s\n", util.Tick(), subtitles[i].Index, logContent, logZhcn)

		// 每条之间留出间隔，最后一条不用等
		if i < len(subtitles)-1 {
			time.Sleep(TranslateInterval)
		}
	}

	// 4. 写入文件：无论完整完成还是中途 abort，都把当前字幕切片写出去。
	// abort 时，未翻译的字幕 Zhcn 为空，原文仍保留，下次跑可以补翻。
	writeSubtitles(inputfile, subtitles, aborted)

	return aborted
}

// writeSubtitles 原子写入字幕到 SRT 文件。
// 先写到临时文件再 rename，避免留下半截原始文件。
// partial=true 时打印提示说明这是不完整的翻译。
func writeSubtitles(outputFile string, subtitles []Subtitle, partial bool) {
	tmpFile := outputFile + ".tmp"
	os.Remove(tmpFile)

	outFile, err := os.Create(tmpFile)
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		return
	}

	for i, sub := range subtitles {
		fmt.Fprintf(outFile, "%d\n", i+1)
		fmt.Fprintf(outFile, "%s\n", sub.Timeline)
		fmt.Fprintf(outFile, "%s\n", sub.Content)
		fmt.Fprintf(outFile, "%s\n", sub.Zhcn)
		fmt.Fprintf(outFile, "\n")
	}

	if err := outFile.Close(); err != nil {
		fmt.Printf("关闭输出文件失败: %v\n", err)
		os.Remove(tmpFile)
		return
	}
	if err := os.Rename(tmpFile, outputFile); err != nil {
		fmt.Printf("重命名输出文件失败: %v\n", err)
		os.Remove(tmpFile)
		return
	}

	if partial {
		fmt.Printf("⚠️ 已保存部分翻译（剩余字幕未译）: %s\n", outputFile)
	} else {
		fmt.Printf("翻译完成！已替换原始文件: %s\n", outputFile)
	}
}

// CountTranslatable 统计一个字幕文件中实际需要翻译的条数。
func CountTranslatable(inputfile string) int {
	subtitles, err := ParseSubtitle(inputfile)
	if err != nil {
		return 0
	}
	for i := range subtitles {
		subtitles[i].Content = DedupRepeats(subtitles[i].Content)
	}
	subtitles = MergeSubtitles(subtitles)

	count := 0
	for _, s := range subtitles {
		if s.Content != "" {
			count++
		}
	}
	return count
}

// printAbortSummary 在触发超时保护时打印醒目日志，
// 重复三遍带分割线，确保在 Actions 日志里一眼能看到。
// index 为 abort 时正要处理（还没翻译）的字幕下标。
func printAbortSummary(file string, subtitles []Subtitle, index int) {
	// 收集已翻译成功的字幕（Zhcn 非空）
	type doneItem struct {
		idx  int
		orig string
		zhcn string
	}
	var done []doneItem
	for j := 0; j < index; j++ {
		if subtitles[j].Zhcn != "" {
			done = append(done, doneItem{
				idx:  j + 1, // 输出序号从 1 开始
				orig: strings.ReplaceAll(subtitles[j].Content, "\n", " "),
				zhcn: strings.ReplaceAll(subtitles[j].Zhcn, "\n", " "),
			})
		}
	}
	remaining := util.Remaining()

	// 把同一份 summary 用分割线包起来打印三遍，重复确保显眼
	for pass := 1; pass <= 3; pass++ {
		fmt.Println()
		fmt.Println("============================================================")
		fmt.Printf("  ⚠️  超时保护触发！(第 %d/3 次)\n", pass)
		fmt.Println("============================================================")
		fmt.Printf("  剩余时间:    %v\n", remaining)
		fmt.Printf("  当前文件:    %s\n", file)
		fmt.Printf("  文件进度:    第 %d / %d 条 (即将翻译)\n", index+1, len(subtitles))
		fmt.Printf("  已译成功:    %d 条\n", len(done))
		fmt.Printf("  待翻/未完成:  %d 条\n", len(subtitles)-index)
		if len(done) > 0 {
			fmt.Println("  ---------- 已翻译成功的字幕 ----------")
			for _, d := range done {
				fmt.Printf("    [%d] 原文: %s\n", d.idx, d.orig)
				fmt.Printf("         译文: %s\n", d.zhcn)
			}
			fmt.Println("  --------------------------------------")
		} else {
			fmt.Println("  （当前文件还没有翻译成功任何字幕）")
		}
		fmt.Println("============================================================")
		fmt.Println()
	}
}
