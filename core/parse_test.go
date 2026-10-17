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

func TestDedupRepeats(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"单字长重复", "ああああああああああああああああああああああああああああああああああああ", "あ"},
		{"双字组重复", "あいあいあいあいあいあいあいあい", "あい"},
		{"三字组重复", "あいうあいうあいうあいう", "あいう"},
		{"单字+双字混合", "ああいいああいい", "あい"},
		{"正常文本不动", "私はあなたを愛しています。", "私はあなたを愛しています。"},
		{"重复后跟正常文本", "ああああすみません", "あすみません"},
		{"双字组后跟正常文本", "あいあいあいすみません", "あいすみません"},
		{"空串", "", ""},
		{"单字无重复", "あ", "あ"},
		{"两字无重复", "あい", "あい"},
		{"四字组也去重", "すみませんすみませんすみません", "すみません"},
		{"句中重复正常", "いいええええええええええ", "いえ"},
		{"递归-单字去重后产生双字组重复", "あああいいいあああいいい", "あい"},
		{"递归-双字去重后产生三字组重复", "あいあいあいうあいあいあいう", "あいう"},
		{"递归-三层收敛", "ああいいああいいああいいああいい", "あい"},
		{"多行内容保留换行", "あああ\nいいい\nあああ\nいいい", "あ\nい\nあ\nい"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := DedupRepeats(c.in)
			if got != c.want {
				t.Errorf("DedupRepeats(%q) = %q, 期望 %q", c.in, got, c.want)
			}
		})
	}
}

func TestMergeSubtitles_EmptyContentGuard(t *testing.T) {
	// 空内容条目不参与合并，保持 SRT 结构（条目数不变）
	subs := []Subtitle{
		{Index: 1, Timeline: "00:00:01,000 --> 00:00:02,000", Content: ""},
		{Index: 2, Timeline: "00:00:02,000 --> 00:00:03,000", Content: ""},
		{Index: 3, Timeline: "00:00:03,000 --> 00:00:04,000", Content: "あ"},
		{Index: 4, Timeline: "00:00:04,000 --> 00:00:05,000", Content: "あ"},
	}

	merged := MergeSubtitles(subs)

	// 两个空条目不合并，两个 "あ" 合并，所以 4 → 3
	if got := len(merged); got != 3 {
		t.Fatalf("空内容保护: 合并后条目数 = %d, 期望 3", got)
	}
	if merged[0].Content != "" || merged[1].Content != "" {
		t.Errorf("空条目应保持空内容, 实际: [%d]=%q, [%d]=%q", 0, merged[0].Content, 1, merged[1].Content)
	}
	if merged[2].Timeline != "00:00:03,000 --> 00:00:05,000" {
		t.Errorf("非空合并时间轴不对, 实际: %s", merged[2].Timeline)
	}
}
