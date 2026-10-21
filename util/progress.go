// Package util 提供工具函数：翻译进度与耗时预估。
package util

import (
	"fmt"
	"time"
)

// ActionsTimeLimit GitHub Actions 单个 job 的超时上限为 6 小时，
// 用于计算"距离上限还剩多少时间"。
const ActionsTimeLimit = 6 * time.Hour

var (
	startTime time.Time // 进度追踪的起点（InitProgress 时记录）
	total     int       // 需要翻译的字幕总条数
	done      int       // 已翻译完成的条数
)

// InitProgress 初始化进度追踪。
// totalCount 为所有文件中需要翻译的字幕总条数（去重合并后、非空内容）。
func InitProgress(totalCount int) {
	startTime = time.Now()
	total = totalCount
	done = 0
}

// Tick 每翻译完一条字幕后调用一次，返回当条日志的时间信息前缀。
// 包含：已用时长、按当前平均速度预估的剩余时长、距 6 小时限制的剩余时长。
// 未调用 InitProgress 时（如测试直接调用 Core），total 为 0，预估显示为 0。
func Tick() string {
	done++
	elapsed := time.Since(startTime)

	// 预计剩余 = 已用时间 / 已完成条数 * 剩余条数（按平均速度线性估算）
	var eta time.Duration
	if total > done {
		eta = elapsed / time.Duration(done) * time.Duration(total-done)
	}

	// 距 6 小时限制的剩余时间，超时后按 0 显示
	remaining := max(ActionsTimeLimit-elapsed, 0)

	return fmt.Sprintf("[已用 %s | 预计还需 %s | 距6小时限制 %s]",
		formatDuration(elapsed), formatDuration(eta), formatDuration(remaining))
}

// formatDuration 把时长格式化为 "X小时YY分"。
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%d小时%02d分", h, m)
}
