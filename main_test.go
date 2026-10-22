package main

import (
	"TraditionTranslate/util"
	"os"
	"strings"
	"testing"
)

func TestWriteReport(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)

	util.InitProgress(10)
	order := []string{"srt/kmi001", "srt/kmi016"}
	folders := map[string]*folderStat{
		"srt/kmi001": {total: 2, done: 2},
		"srt/kmi016": {total: 2, done: 1, partial: 1},
	}

	writeReport(order, folders)

	data, readErr := os.ReadFile(ReportFile)
	if readErr != nil {
		t.Fatalf("读取报告失败: %v", readErr)
	}
	text := string(data)
	for _, want := range []string{
		"视频(文件夹)总数: 2 个，成功: 1 个，未完成: 1 个",
		"✅ 成功      srt/kmi001",
		"❌ 未完成    srt/kmi016  (完成 1/2，部分翻译 1，跳过 0)",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("报告缺少内容 %q，实际:\n%s", want, text)
		}
	}
}
