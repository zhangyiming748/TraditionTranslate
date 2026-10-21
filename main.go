package main

import (
	"TraditionTranslate/core"
	"TraditionTranslate/util"
	"fmt"
	"strings"
	"time"

	"github.com/zhangyiming748/finder"
)

// FileInterval 每处理完一个字幕文件后的等待时间。
const FileInterval = 10 * time.Second

// 文件处理状态
const (
	statusDone    = "✅ 完成翻译"
	statusPartial = "⚠️  部分翻译（超时中断）"
	statusSkipped = "⏭️  跳过（超时保护）"
)

// fileResult 记录单个文件的处理状态，用于退出时汇总。
type fileResult struct {
	path   string
	status string
}

func main() {
	srtDir := "srt"

	srts := finder.FindAllFiles(srtDir)

	// 先统计所有文件实际需要翻译的字幕总条数，初始化进度追踪。
	total := 0
	for _, srtpath := range srts {
		if strings.HasSuffix(srtpath, "_zhs.srt") {
			continue
		}
		total += core.CountTranslatable(srtpath)
	}
	util.InitProgress(total)

	// 追踪每个文件的处理结果，退出时打印全局汇总
	var results []fileResult

	for i, srtpath := range srts {
		if strings.HasSuffix(srtpath, "_zhs.srt") {
			continue
		}

		fmt.Printf("处理文件: %s\n", srtpath)

		// 每个文件开始前也检查一次，避免在上一个文件的 FileInterval 期间就已经超时
		if util.ShouldAbort() {
			results = append(results, fileResult{path: srtpath, status: statusSkipped})
			fmt.Printf("⏸️ 剩余时间不足 %v，跳过该文件。\n", util.AbortThreshold)
			for _, rest := range srts[i+1:] {
				if !strings.HasSuffix(rest, "_zhs.srt") {
					results = append(results, fileResult{path: rest, status: statusSkipped})
				}
			}
			break
		}

		// Core 返回 true 表示因超时提前 abort（当前文件也只翻了一部分）
		if core.Core(srtpath) {
			results = append(results, fileResult{path: srtpath, status: statusPartial})
			// 剩余文件全部标为跳过
			for _, rest := range srts[i+1:] {
				if !strings.HasSuffix(rest, "_zhs.srt") {
					results = append(results, fileResult{path: rest, status: statusSkipped})
				}
			}
			break
		}
		results = append(results, fileResult{path: srtpath, status: statusDone})

		// 文件之间留出间隔，最后一个文件不用等
		if i < len(srts)-1 {
			if util.ShouldAbort() {
				fmt.Printf("⏸️ 剩余时间不足 %v，跳过剩余文件。\n", util.AbortThreshold)
				for _, rest := range srts[i+1:] {
					if !strings.HasSuffix(rest, "_zhs.srt") {
						results = append(results, fileResult{path: rest, status: statusSkipped})
					}
				}
				break
			}
			fmt.Printf("等待 %v 后处理下一个文件...\n", FileInterval)
			time.Sleep(FileInterval)
		}
	}

	// 程序退出前打印全局汇总（三遍带分割线，醒目）
	printGlobalSummary(results)
}

// printGlobalSummary 打印所有文件的处理状态汇总。
func printGlobalSummary(results []fileResult) {
	remaining := util.Remaining()
	doneCount, partialCount, skippedCount := 0, 0, 0
	for _, r := range results {
		switch r.status {
		case statusDone:
			doneCount++
		case statusPartial:
			partialCount++
		case statusSkipped:
			skippedCount++
		}
	}

	for pass := 1; pass <= 3; pass++ {
		fmt.Println()
		fmt.Println("============================================================")
		fmt.Printf("  📋  全局处理汇总 (第 %d/3 次)\n", pass)
		fmt.Println("============================================================")
		fmt.Printf("  剩余时间:  %v\n", remaining)
		fmt.Printf("  完成翻译:  %d 个文件\n", doneCount)
		fmt.Printf("  部分翻译:  %d 个文件\n", partialCount)
		fmt.Printf("  跳过未译:  %d 个文件\n", skippedCount)
		fmt.Println("  ------------------------------------------------------------")
		for _, r := range results {
			fmt.Printf("  %s  %s\n", r.status, r.path)
		}
		fmt.Println("============================================================")
		fmt.Println()
	}
}
