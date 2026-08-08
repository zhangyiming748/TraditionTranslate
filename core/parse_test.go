package core

import (
	"fmt"
	"testing"
)

func TestParseSubtitle(t *testing.T) {
	testFile := "/Users/zen/github/TraditionTranslate/srt/suto新体操教室_116283119045179.srt"

	subtitles, err := ParseSubtitle(testFile)
	if err != nil {
		t.Fatalf("解析文件时发生错误: %v", err)
	}

	for _, sub := range subtitles {
		fmt.Printf("序号: %d, 时间轴: %s, 内容: %s\n", sub.Index, sub.Timeline, sub.Content)
	}
}
