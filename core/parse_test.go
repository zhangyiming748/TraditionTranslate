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

func TestMergeSubtitles(t *testing.T) {
	testFile := "/Users/zen/github/TraditionTranslate/srt/kmi016/01.srt"

	subtitles, err := ParseSubtitle(testFile)
	if err != nil {
		t.Fatalf("解析文件时发生错误: %v", err)
	}
	if got := len(subtitles); got != 4 {
		t.Fatalf("解析后条目数 = %d, 期望 4", got)
	}

	merged := MergeSubtitles(subtitles)
	if got := len(merged); got != 1 {
		t.Fatalf("合并后条目数 = %d, 期望 1", got)
	}

	want := Subtitle{
		Index:    116,
		Timeline: "00:18:35,320 --> 00:19:03,320",
		Content:  "私はあなたを愛しています。",
	}
	got := merged[0]
	if got.Index != want.Index || got.Timeline != want.Timeline || got.Content != want.Content {
		t.Fatalf("合并结果不匹配\ngot:  %+v\nwant: %+v", got, want)
	}
	fmt.Printf("合并结果: 序号=%d 时间轴=%s 内容=%s\n", got.Index, got.Timeline, got.Content)
}
