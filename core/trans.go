package core

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Engine 翻译引擎类型。
type Engine int

const (
	// EngineTrans translate-shell（GitHub Actions 上使用）。
	EngineTrans Engine = iota
	// EngineLibre 本地自建 LibreTranslate。
	EngineLibre
)

// transInterval 使用 translate-shell 时每条字幕之间的等待时间。
// 逐条串行翻译并留出间隔，避免请求过于频繁触发 Google 限流，
// 限流时 trans 命令会返回空内容（DST 为空）。
const transInterval = 2 * time.Second

var (
	currentEngine = EngineTrans
	engineDesc    = "translate-shell (Google)"
	libreBaseURL  string
)

// SetupTranslator 依次探测候选地址池，选择第一个可用的本地 LibreTranslate。
// 全部不可用则回落到 translate-shell（即工作流场景）。
// 返回值表示是否选中了本地 API。
func SetupTranslator(libreCandidates []string) bool {
	for _, base := range libreCandidates {
		fmt.Printf("探测本地翻译服务: %s ... ", base)
		if probeLibre(base) {
			fmt.Println("可用")
			currentEngine = EngineLibre
			libreBaseURL = base
			engineDesc = fmt.Sprintf("本地 LibreTranslate (%s)", base)
			return true
		}
		fmt.Println("不可用")
	}

	currentEngine = EngineTrans
	libreBaseURL = ""
	engineDesc = "translate-shell (Google)"
	return false
}

// EngineName 返回当前翻译引擎的描述，用于启动日志。
func EngineName() string {
	return engineDesc
}

// IsLocalEngine 判断当前是否使用本地翻译 API。
func IsLocalEngine() bool {
	return currentEngine == EngineLibre
}

// TranslateInterval 返回每条字幕之间的等待时间。
// 本地 API 是自己搭建的，不存在限流问题，因此不需要等待。
func TranslateInterval() time.Duration {
	if currentEngine == EngineLibre {
		return 0
	}
	return transInterval
}

// Translate 使用当前选定的引擎将原文翻译成中文简体。
func Translate(src string) (string, error) {
	if currentEngine == EngineLibre {
		return translateLibre(libreBaseURL, src)
	}
	return translateShell(src)
}

// translateShell 使用 translate-shell 将原文翻译成中文简体
func translateShell(src string) (dst string, err error) {
	// 修剪空白
	src = strings.TrimSpace(src)
	if src == "" {
		return "", nil
	}

	// 构建 translate-shell 命令
	// 主命令: trans
	// 参数列表可以灵活增删
	args := []string{}
	args = append(args, "-brief")             // 简洁模式，只显示翻译结果
	args = append(args, "-e", "google")       // 使用 Google 翻译引擎
	args = append(args, "-source", "auto")    // 自动检测源语言
	args = append(args, "-target", "Chinese") // 目标语言为中文
	args = append(args, src)                  // 待翻译文本
	cmd := exec.Command("trans", args...)

	// 执行命令
	output, err := cmd.Output()
	if err != nil {
		// 尝试获取错误信息
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("翻译命令执行失败: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("翻译命令执行失败: %v", err)
	}

	// 简洁模式下，输出就是纯译文
	dst = strings.TrimSpace(string(output))

	return dst, nil
}
