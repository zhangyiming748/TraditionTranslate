package main

import (
	"TraditionTranslate/core"
	"TraditionTranslate/util"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhangyiming748/finder"
)

// FileInterval 每处理完一个字幕文件后的等待时间。
const FileInterval = 10 * time.Second

// ReportFile 翻译报告文件名。
// 按子文件夹（即一个视频）汇总全部片段是否完整翻译成功，
// 与 SRT 一起上传到 Actions 工件。
const ReportFile = "translate-report.txt"

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

// folderStat 一个子文件夹（一个视频）的翻译状态统计。
type folderStat struct {
	total   int // 该文件夹下待翻译的 SRT 文件数
	done    int // 完整翻译成功的文件数
	partial int // 超时中断、只翻译了一部分的文件数
	skipped int // 超时保护下未开始翻译的文件数
}

// complete 判断该文件夹（视频）是否全部片段翻译成功。
func (s *folderStat) complete() bool {
	return s.total > 0 && s.done == s.total
}

// isProcessable 判断路径是否为需要处理的字幕文件（.srt，且不是已翻译的 _zhs.srt）。
// finder.FindAllFiles 会把根目录本身也放进结果，需要在这里过滤掉非字幕文件。
func isProcessable(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".srt") && !strings.HasSuffix(lower, "_zhs.srt")
}

func main() {
	srtDir := "srt"

	srts := finder.FindAllFiles(srtDir)

	// 按子文件夹分组，并统计需要翻译的字幕总条数。
	// 一个子文件夹对应一个视频，只有其中所有 SRT 都翻译成功才算该视频成功。
	folderOrder := []string{}
	folders := map[string]*folderStat{}
	total := 0

	for _, srtpath := range srts {
		if !isProcessable(srtpath) {
			continue
		}
		total += core.CountTranslatable(srtpath)
		dir := filepath.ToSlash(filepath.Dir(srtpath))
		if _, ok := folders[dir]; !ok {
			folders[dir] = &folderStat{}
			folderOrder = append(folderOrder, dir)
		}
		folders[dir].total++
	}
	util.InitProgress(total)

	// 启动时先写一份报告，所有视频标记为未完成。
	// 之后每处理完一个文件就刷新，即使程序被强杀也能保留已有结果。
	writeReport(folderOrder, folders)

	var results []fileResult

	// setStatus 记录单个文件的处理结果并更新文件夹统计，不写报告。
	setStatus := func(path, status string) {
		results = append(results, fileResult{path: path, status: status})

		if s, ok := folders[filepath.ToSlash(filepath.Dir(path))]; ok {
			switch status {
			case statusDone:
				s.done++
			case statusPartial:
				s.partial++
			case statusSkipped:
				s.skipped++
			}
		}
	}

	// mark 记录单个文件的处理结果，并立即刷新报告。
	mark := func(path, status string) {
		setStatus(path, status)
		writeReport(folderOrder, folders)
	}

	// skipRest 把从 from 开始的所有未处理文件标记为跳过，
	// 批量更新后只写一次报告，避免超时后频繁刷盘。
	skipRest := func(from int) {
		for _, rest := range srts[from:] {
			if isProcessable(rest) {
				setStatus(rest, statusSkipped)
			}
		}
		writeReport(folderOrder, folders)
	}

	for i, srtpath := range srts {
		if !isProcessable(srtpath) {
			continue
		}

		fmt.Printf("处理文件: %s\n", srtpath)

		// 每个文件开始前也检查一次，避免在上一个文件的 FileInterval 期间就已经超时
		if util.ShouldAbort() {
			fmt.Printf("⏸️ 剩余时间不足 %v，跳过剩余文件。\n", util.AbortThreshold)
			skipRest(i)
			break
		}

		// Core 返回 true 表示因超时提前 abort（当前文件也只翻了一部分）
		if core.Core(srtpath) {
			mark(srtpath, statusPartial)
			skipRest(i + 1)
			break
		}
		mark(srtpath, statusDone)

		// 文件之间留出间隔，最后一个文件不用等
		if i < len(srts)-1 {
			if util.ShouldAbort() {
				fmt.Printf("⏸️ 剩余时间不足 %v，跳过剩余文件。\n", util.AbortThreshold)
				skipRest(i + 1)
				break
			}
			fmt.Printf("等待 %v 后处理下一个文件...\n", FileInterval)
			time.Sleep(FileInterval)
		}
	}

	// 程序退出前打印全局汇总（三遍带分割线，醒目）
	printGlobalSummary(results)

	fmt.Printf("\n翻译报告已写入: %s\n", ReportFile)
}

// writeReport 生成按视频（子文件夹）汇总的翻译报告。
// 一个文件夹内所有 SRT 都完整翻译成功记为“成功”，否则记为“未完成”。
func writeReport(order []string, folders map[string]*folderStat) {
	var b strings.Builder
	b.WriteString("翻译报告\n")
	fmt.Fprintf(&b, "生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "距6小时限制剩余: %v\n\n", util.Remaining().Truncate(time.Second))

	success, incomplete := 0, 0
	for _, dir := range order {
		if folders[dir].complete() {
			success++
		} else {
			incomplete++
		}
	}
	fmt.Fprintf(&b, "视频(文件夹)总数: %d 个，成功: %d 个，未完成: %d 个\n\n", len(order), success, incomplete)

	for _, dir := range order {
		s := folders[dir]
		if s.complete() {
			fmt.Fprintf(&b, "✅ 成功      %s\n", dir)
		} else {
			fmt.Fprintf(&b, "❌ 未完成    %s  (完成 %d/%d，部分翻译 %d，跳过 %d)\n",
				dir, s.done, s.total, s.partial, s.skipped)
		}
	}

	if err := os.WriteFile(ReportFile, []byte(b.String()), 0644); err != nil {
		fmt.Printf("写翻译报告失败: %v\n", err)
	}
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
