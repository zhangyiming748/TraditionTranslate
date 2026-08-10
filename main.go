package main

import (
	"TraditionTranslate/core"
	"fmt"
	"strings"

	"github.com/zhangyiming748/finder"
)

func main() {
	// 构建 SRT 文件夹的路径（相对于当前工作目录，即仓库根目录）
	srtDir := "srt"

	// 查找所有字幕文件
	srts := finder.FindAllFiles(srtDir)

	for _, srtpath := range srts {
		fmt.Printf("处理文件: %s\n", srtpath)
		if strings.HasSuffix(srtpath, "_zhs.srt") {
			continue
		}
		core.Core(srtpath)
	}
}
