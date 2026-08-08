package core

import (
	"fmt"
	"os/exec"
	"strings"
)

// Translate 使用 translate-shell 将原文翻译成中文简体
func Translate(src string) (dst string, err error) {
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
	args = append(args, "-source", "auto")      // 自动检测源语言
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
