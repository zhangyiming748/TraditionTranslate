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
// 文件级别的限流：连续处理多个文件时请求过于密集，
// 容易触发 Google 翻译限流，导致 trans 返回空内容。
const FileInterval = 10 * time.Second

func main() {
	// 构建 SRT 文件夹的路径（相对于当前工作目录，即仓库根目录）
	srtDir := "srt"

	// 查找所有字幕文件
	srts := finder.FindAllFiles(srtDir)

	// 先统计所有文件实际需要翻译的字幕总条数，初始化进度追踪。
	// 总数用于每条日志中"预计还要多少时间"的估算。
	total := 0
	for _, srtpath := range srts {
		if strings.HasSuffix(srtpath, "_zhs.srt") {
			continue
		}
		total += core.CountTranslatable(srtpath)
	}
	util.InitProgress(total)

	for i, srtpath := range srts {
		fmt.Printf("处理文件: %s\n", srtpath)
		if strings.HasSuffix(srtpath, "_zhs.srt") {
			continue
		}
		core.Core(srtpath)

		// 文件之间留出间隔，最后一个文件不用等
		if i < len(srts)-1 {
			fmt.Printf("等待 %v 后处理下一个文件...\n", FileInterval)
			time.Sleep(FileInterval)
		}
	}
}
