package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// MaxConcurrent 同时进行的翻译请求数上限。
// 每条字幕都会启动一个 trans 子进程并向 Google 发送请求，
// 该值过大会触发 Google 限流（HTTP 429），过小则速度提升不明显。
// 你可以根据测试结果自行调整这个常量。
const MaxConcurrent = 2

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

	// 3. 并发翻译每条字幕的原文内容并填充 Zhcn 字段
	// 使用信号量（容量为 MaxConcurrent）限制同时在途的翻译请求数，
	// 避免触发 Google 限流；用结果切片按原序号保存译文以保证输出顺序。
	sem := make(chan struct{}, MaxConcurrent)
	results := make([]string, len(subtitles))
	errs := make([]error, len(subtitles))
	var wg sync.WaitGroup

	for i := range subtitles {
		if subtitles[i].Content == "" {
			continue
		}

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// 获取信号量：达到 MaxConcurrent 时阻塞，实现并发限流
			sem <- struct{}{}
			defer func() { <-sem }()

			zhcn, err := Translate(subtitles[idx].Content)
			if err != nil {
				errs[idx] = err
				return
			}
			results[idx] = zhcn
		}(i)
	}
	wg.Wait()

	// 按原顺序回填结果，保证输出字幕序号正确
	for i := range subtitles {
		if errs[i] != nil {
			fmt.Printf("第 %d 条字幕翻译失败: %v\n", subtitles[i].Index, errs[i])
			continue
		}
		subtitles[i].Zhcn = results[i]
		fmt.Printf("[%d] 原文: %s\n   译文: %s\n\n", subtitles[i].Index, subtitles[i].Content, subtitles[i].Zhcn)
	}

	// 4. 生成输出文件路径（在扩展名之前添加 _zhs，并移除文件名中的 emoji 字符）
	dir := filepath.Dir(inputfile)
	base := filepath.Base(inputfile)
	ext := filepath.Ext(base)
	nameWithoutExt := base[:len(base)-len(ext)]
	nameWithoutExt = cleanFileName(nameWithoutExt)
	outputFile := filepath.Join(dir, nameWithoutExt+"_zhs"+ext)

	// 5. 原子写入：先写到临时文件，全部写完且关闭成功后再 rename 成最终文件。
	// 这样即使程序在写入中途被强杀（如 Actions 6 小时超时），
	// 也不会留下半截的 _zhs.srt 被误判为"已翻译完成"。
	// 同时清理上一次运行可能残留的 .tmp 文件。
	tmpFile := outputFile + ".tmp"
	os.Remove(tmpFile) // 清理可能残留的临时文件

	outFile, err := os.Create(tmpFile)
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		return
	}

	// 6. 写入新的字幕格式（序号、时间轴、原文、空一行、译文）
	for _, sub := range subtitles {
		fmt.Fprintf(outFile, "%d\n", sub.Index)
		fmt.Fprintf(outFile, "%s\n", sub.Timeline)
		fmt.Fprintf(outFile, "%s\n", sub.Content)
		fmt.Fprintf(outFile, "%s\n", sub.Zhcn)
		fmt.Fprintf(outFile, "\n") // 字幕块之间用换行分隔
	}

	// 关闭文件后再 rename，确保数据已落盘。
	// 只要 rename 没成功，_zhs.srt 就不会出现，绝无半截文件。
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

	fmt.Printf("翻译完成！输出文件: %s\n", outputFile)
}
