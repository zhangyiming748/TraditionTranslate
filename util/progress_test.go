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
	InitProgress(2, true)
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

func TestTickWithoutTimeLimit(t *testing.T) {
	InitProgress(1, false)
	output := Tick()

	if strings.Contains(output, "距6小时限制") {
		t.Errorf("本地运行不应显示 6 小时限制: %s", output)
	}
	if !strings.Contains(output, "已用") || !strings.Contains(output, "预计还需") {
		t.Errorf("Tick 输出缺少已用/预计字段: %s", output)
	}
	if ShouldAbort() {
		t.Error("本地运行（未启用时间限制）不应触发超时 abort")
	}
	if TimeLimitEnabled() {
		t.Error("InitProgress(_, false) 后 TimeLimitEnabled 应返回 false")
	}
}

func TestShouldAbortWithTimeLimit(t *testing.T) {
	InitProgress(1, true)
	// 刚开始时剩余时间充足，不应 abort
	if ShouldAbort() {
		t.Error("刚开始时不应触发超时 abort")
	}

	// 把起点回拨到接近 6 小时上限，应触发 abort
	startTime = time.Now().Add(-(ActionsTimeLimit - AbortThreshold + time.Minute))
	defer func() { startTime = time.Now() }()
	if !ShouldAbort() {
		t.Error("剩余时间小于安全余量时应触发 abort")
	}
}
