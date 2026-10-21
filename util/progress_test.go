package util

import (
	"strings"
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{0, "0小时00分"},
		{5 * time.Minute, "0小时05分"},
		{90 * time.Minute, "1小时30分"},
		{6*time.Hour + time.Minute, "6小时01分"},
		{-time.Minute, "0小时00分"},
	}
	for _, c := range cases {
		if got := formatDuration(c.in); got != c.want {
			t.Errorf("formatDuration(%v) = %q, 期望 %q", c.in, got, c.want)
		}
	}
}

func TestTick(t *testing.T) {
	InitProgress(2)
	outputs := []string{Tick(), Tick()}

	for _, s := range outputs {
		for _, field := range []string{"已用", "预计还需", "距6小时限制"} {
			if !strings.Contains(s, field) {
				t.Errorf("Tick 输出缺少字段 %q: %s", field, s)
			}
		}
	}
	if done != 2 {
		t.Errorf("done = %d, 期望 2", done)
	}

	// 翻译完全部条数后，预计剩余应为 0
	if !strings.Contains(outputs[1], "预计还需 0小时00分") {
		t.Errorf("全部完成后预计剩余应为 0, 实际: %s", outputs[1])
	}
}
