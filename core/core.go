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

// Core 核心功能：解析字幕文件，翻译原文内容为中文，生成新的字幕文件
func Core(inputfile string) {
	/*
		TODO:
		这里首先使用函数把一个字幕解析成每一组字幕的结构体。
		然后使用函数把每一组字幕的结构体中的正文内容翻译成中文。
		补充到结构体的最后一个值就是中文字幕的值。
		重新把每一组字幕的四个值（序号、时间轴、原文内容和译文内容）写成新的字幕，中间有一个换行，然后继续写下一组，直到写完。
		这个新的结构体是写到新的字幕文件里
		原字幕文件不变，但是在扩展名之前加一个_zhs 标记，表示这是中文翻译的字幕文件。
	*/

	// 1. 定义输入文件路径
	// 使用参数 inputfile

	// 2. 解析字幕文件
	subtitles, err := ParseSubtitle(inputfile)
	if err != nil {
		fmt.Printf("解析字幕文件失败: %v\n", err)
		return
	}

	fmt.Printf("成功解析 %d 条字幕\n", len(subtitles))

	// 第一步：清理 Whisper 幻觉产生的连续重复字符/字符组（单字→双字组→三字组）。
	// 必须在合并之前做，因为去重可能让原本不同的条目变得相同，
	// 从而让后续合并能再多合并几条。
	for i := range subtitles {
		subtitles[i].Content = DedupRepeats(subtitles[i].Content)
	}

	// 第二步：合并内容相同的相邻字幕条目，减少重复内容的翻译请求次数。
	// 合并后取该组第一条的开始时间轴与最后一条的结束时间轴，文字不变。
	beforeMerge := len(subtitles)
	subtitles = MergeSubtitles(subtitles)
	fmt.Printf("合并相同字幕: %d -> %d 条\n", beforeMerge, len(subtitles))

	// 3. 串行翻译每条字幕的原文内容并填充 Zhcn 字段。
	// 文件变多时并发请求容易触发 Google 限流（trans 返回空内容），
	// 因此逐条翻译，每条之间留出 TranslateInterval 间隔。
	for i := range subtitles {
		if subtitles[i].Content == "" {
			continue
		}

		zhcn, transErr := Translate(subtitles[i].Content)
		if transErr != nil {
			fmt.Printf("第 %d 条字幕翻译失败: %v\n", subtitles[i].Index, transErr)
		} else {
			subtitles[i].Zhcn = zhcn
		}
		// 日志严格单行：原文/译文内部可能含换行（多行字幕、trans 折行），
		// 替换为空格，保证一条字幕只占一行日志
		logContent := strings.ReplaceAll(subtitles[i].Content, "\n", " ")
		logZhcn := strings.ReplaceAll(subtitles[i].Zhcn, "\n", " ")
		fmt.Printf("%s [%d] 原文: %s | 译文: %s\n", util.Tick(), subtitles[i].Index, logContent, logZhcn)

		// 每条之间留出间隔，最后一条不用等
		if i < len(subtitles)-1 {
			time.Sleep(TranslateInterval)
		}
	}

	// 4. 输出文件直接替换原始文件（文件名保持一致，不加尾部标记）
	outputFile := inputfile

	// 5. 原子写入：先写到临时文件，全部写完且关闭成功后再 rename 成最终文件。
	// 这样即使程序在写入中途被强杀（如 Actions 6 小时超时），
	// 也不会留下半截的原始文件被误判为"已翻译完成"。
	// 同时清理上一次运行可能残留的 .tmp 文件。
	tmpFile := outputFile + ".tmp"
	os.Remove(tmpFile) // 清理可能残留的临时文件

	outFile, err := os.Create(tmpFile)
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		return
	}

	// 6. 写入字幕格式（序号、时间轴、原文、译文），直接替换原始文件
	// 序号用循环计数器从 1 开始连续编号，合并后可能减少了条目数，
	// 不再保留原文件中不连续的序号。
	for i, sub := range subtitles {
		fmt.Fprintf(outFile, "%d\n", i+1)
		fmt.Fprintf(outFile, "%s\n", sub.Timeline)
		fmt.Fprintf(outFile, "%s\n", sub.Content)
		fmt.Fprintf(outFile, "%s\n", sub.Zhcn)
		fmt.Fprintf(outFile, "\n") // 字幕块之间用换行分隔
	}

	// 关闭文件后再 rename，确保数据已落盘。
	// 只要 rename 没成功，原始文件就不会被覆盖，绝无半截文件。
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

	fmt.Printf("翻译完成！已替换原始文件: %s\n", outputFile)
}

// CountTranslatable 统计一个字幕文件中实际需要翻译的条数。
// 统计口径与 Core 的预处理流程完全一致（解析 → 去重 → 合并 → 非空内容），
// 用于 main 在处理前汇总所有文件的总条数，供进度与耗时预估使用。
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
