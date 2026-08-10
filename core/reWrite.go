package core

import (
	"strings"
	"unicode"
)

// cleanFileName 清理文件名：移除 emoji、零宽字符、控制字符等非文字内容，
// 保留纯文字（含中文、日文、韩文等），并清理多余空格和残留分隔符
func cleanFileName(s string) string {
	runes := []rune(s)
	result := make([]rune, 0, len(runes))
	for _, r := range runes {
		if shouldKeep(r) {
			result = append(result, r)
		}
	}
	cleaned := string(result)
	// 替换在特定文件系统（如 NTFS）上非法的字符，避免 GitHub Actions 上传工件时报错
	for _, invalidChar := range []string{"\"", ":", "<", ">", "|", "*", "?"} {
		cleaned = strings.ReplaceAll(cleaned, invalidChar, "")
	}
	// 将多个连续空白合并为单个空格
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	// 循环清理，直到文件名不再变化（因为某些清理可能产生新的需清理内容）
	for {
		prev := cleaned
		cleaned = strings.TrimSpace(cleaned)
		cleaned = strings.TrimLeft(cleaned, " -_")
		cleaned = strings.TrimSpace(cleaned)
		cleaned = strings.TrimRight(cleaned, " -_")
		cleaned = strings.TrimSpace(cleaned)
		// 清理空括号
		cleaned = strings.ReplaceAll(cleaned, "()", "")
		cleaned = strings.ReplaceAll(cleaned, "[]", "")
		cleaned = strings.ReplaceAll(cleaned, "{}", "")
		// 清理连续多个分隔符，如 " - - " → " - "
		for strings.Contains(cleaned, " - - ") {
			cleaned = strings.ReplaceAll(cleaned, " - - ", " - ")
		}
		// 去除尾部残留的 "-- 文本" 模式（emoji 移除后留下的分隔符 + 碎片）
		if idx := strings.LastIndex(cleaned, " -- "); idx != -1 {
			cleaned = strings.TrimSpace(cleaned[:idx])
		}
		if cleaned == prev {
			break
		}
	}
	return cleaned
}

// shouldKeep 判断一个字符是否应该保留在文件名中
func shouldKeep(r rune) bool {
	// 丢弃零宽字符、软连字符、替换字符等不可见或无意义字符
	switch r {
	case 0x200B, 0x200C, 0x200D, 0x2060, 0xFEFF: // 零宽空格、零宽非连接符、零宽连接符、词连接符、BOM
		return false
	case 0x200E, 0x200F: // 左右至右标记
		return false
	case 0x202A, 0x202B, 0x202C, 0x202D, 0x202E: // 双向文本控制字符
		return false
	case 0x00AD: // 软连字符 (SHY)
		return false
	case 0xFFFD: // 替换字符
		return false
	case 0xFFFE, 0xFFFF: // 非字符
		return false
	}
	// 保留可见 ASCII 字符（字母、数字、标点、符号），丢弃控制字符
	if r <= 0x7F {
		return r >= 0x20 && r <= 0x7E // 只保留可见 ASCII (0x20-0x7E)，丢弃控制字符和 DEL
	}
	// 保留所有 Unicode 字母（含中文、日文、韩文、西里尔、阿拉伯文等）
	if unicode.IsLetter(r) {
		return true
	}
	// 保留所有 Unicode 数字
	if unicode.IsNumber(r) {
		return true
	}
	// 保留组合标记（重音符号等）
	if unicode.IsMark(r) {
		return true
	}
	// 保留常见 Unicode 标点和货币符号
	switch {
	case unicode.Is(unicode.Sc, r): // € £ ¥ ₿ 等货币符号
		return true
	case unicode.IsPunct(r): // 所有 Unicode 通用标点（含 «» ‚ „ ‹ › ‽ 等）
		return true
	case r == '\u00A0': // 不间断空格
		return true
	}
	// 其余全部丢弃：emoji、Symbol、Separator、Cn 等
	return false
}
