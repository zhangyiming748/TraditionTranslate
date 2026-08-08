package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"github.com/zhangyiming748/finder"
	"TraditionTranslate/core"
)

func main() {
	// 获取当前工作目录（代码仓库根目录）
	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)

	// 构建 SRT 文件夹的相对路径
	srtDir := filepath.Join(baseDir, "srt")

	// 查找所有字幕文件
	srts := finder.FindAllFiles(srtDir)

	for _, srtpath := range srts {
		fmt.Printf("处理文件: %s\n", srtpath)
		if strings.HasSuffix(srtpath,"_zhs.srt") {
			continue
		}
		core.Core(srtpath)
	}
}
