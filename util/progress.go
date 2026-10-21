// Package util 提供工具函数：翻译进度与耗时预估。
package util

import (
	"fmt"
	"time"
)

// ActionsTimeLimit GitHub Actions 单个 job 的超时上限为 6 小时。
const ActionsTimeLimit = 6 * time.Hour

// AbortThreshold 距 Actions 超时的安全余量。
// 剩余时间小于该阈值时触发 abort，留出时间让当前文件写完、
// 程序优雅退出、GitHub Actions 打包上传工件。
const AbortThreshold = 10 * time.Minute

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

// Remaining 返回距 Actions 超时的剩余时间。
func Remaining() time.Duration {
	return max(ActionsTimeLimit-time.Since(startTime), 0)
}

// ShouldAbort 检查剩余时间是否小于安全余量。
// 用于在每条字幕翻译前判断：是继续翻还是放弃当前文件的剩余部分，
// 把已完成的文件留给后续工件上传。
func ShouldAbort() bool {
	return Remaining() < AbortThreshold
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
