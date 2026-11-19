package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// LibreAPIBaseURLs 本地 LibreTranslate 候选地址池。
// 按顺序逐个探测，使用第一个可用的地址；
// 全部探测失败才判定为在工作流中运行，回落到 translate-shell。
// 换网络环境或换机器时，把该环境下的地址追加到这里即可。
var LibreAPIBaseURLs = []string{
	"http://192.168.110.131:9002",
	"http://192.168.110.151:9002",
}

const (
	// libreTargetLang 目标语言代码，/languages 返回的支持列表中使用 zh-Hans。
	libreTargetLang = "zh-Hans"
	// probeTimeout 探测本地服务是否可用的超时时间，需足够短避免拖慢启动。
	probeTimeout = 2 * time.Second
	// libreTimeout 单条字幕翻译请求的超时时间。
	libreTimeout = 60 * time.Second
)

// libreTranslateRequest LibreTranslate /translate 请求体。
type libreTranslateRequest struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
	Format string `json:"format"`
}

// libreTranslateResponse /translate 响应体。
// detectedLanguage 只是语言检测结果，不影响译文，但需要兼容不同版本：
// confidence 有时是整数（100）有时是浮点（90.0），因此用 json.Number
// 原样保留 JSON 字面量，避免整型/浮点类型来回变动导致解析失败。
// 译文只取 translatedText。
type libreTranslateResponse struct {
	DetectedLanguage struct {
		Confidence json.Number `json:"confidence"`
		Language   string      `json:"language"`
	} `json:"detectedLanguage"`
	TranslatedText string `json:"translatedText"`
	Error          string `json:"error"`
}

// probeLibre 通过 GET /health 判断本地 LibreTranslate 是否可用。
// 使用 HTTP 探测而不是 ICMP ping：ping 依赖 ICMP 放行，
// 在部分网络环境下会误判为不通。
func probeLibre(baseURL string) bool {
	client := &http.Client{Timeout: probeTimeout}
	resp, err := client.Get(strings.TrimRight(baseURL, "/") + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode == http.StatusOK
}

// translateLibre 调用本地 LibreTranslate 翻译单条文本。
func translateLibre(baseURL, src string) (string, error) {
	payload, err := json.Marshal(libreTranslateRequest{
		Q:      src,
		Source: "auto",
		Target: libreTargetLang,
		Format: "text",
	})
	if err != nil {
		return "", fmt.Errorf("构造 LibreTranslate 请求失败: %w", err)
	}

	url := strings.TrimRight(baseURL, "/") + "/translate"
	client := &http.Client{Timeout: libreTimeout}
	resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("请求 LibreTranslate 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 LibreTranslate 响应失败: %w", err)
	}

	var result libreTranslateResponse
	// 使用 Decoder + UseNumber 统一解码：所有数字以 json.Number 保留原始字面量，
	// 不提前落成 float64，confidence 是整数还是浮点都能安全解析。
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return "", fmt.Errorf("解析 LibreTranslate 响应失败: %w (body: %s)", err, strings.TrimSpace(string(body)))
	}
	if resp.StatusCode != http.StatusOK {
		if result.Error != "" {
			return "", fmt.Errorf("LibreTranslate 返回错误(%d): %s", resp.StatusCode, result.Error)
		}
		return "", fmt.Errorf("LibreTranslate 返回错误(%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if result.Error != "" {
		return "", fmt.Errorf("LibreTranslate 返回错误: %s", result.Error)
	}

	dst := strings.TrimSpace(result.TranslatedText)
	if dst == "" {
		return "", fmt.Errorf("LibreTranslate 返回空译文")
	}
	return dst, nil
}
