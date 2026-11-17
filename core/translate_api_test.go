package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProbeLibre(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ok.Close()

	if !probeLibre(ok.URL) {
		t.Error("/health 返回 200 时应判定为可用")
	}

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer bad.Close()

	if probeLibre(bad.URL) {
		t.Error("/health 返回非 200 时应判定为不可用")
	}
}

func TestTranslateLibre(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		confidence string // json.Number 原样保留的字面量
	}{
		{"整数 confidence", `{"detectedLanguage":{"confidence":100,"language":"ja"},"translatedText":"  你好  "}`, "100"},
		{"浮点 confidence", `{"detectedLanguage":{"confidence":90.0,"language":"ja"},"translatedText":"  你好  "}`, "90.0"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got libreTranslateRequest
			var resp libreTranslateResponse
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/translate" || r.Method != http.MethodPost {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Write([]byte(c.body))
			}))
			defer srv.Close()

			dst, err := translateLibre(srv.URL, "こんにちは")
			if err != nil {
				t.Fatalf("翻译失败: %v", err)
			}
			if dst != "你好" {
				t.Errorf("译文应为 translatedText 字段的修剪内容，实际 %q", dst)
			}
			if got.Q != "こんにちは" || got.Source != "auto" || got.Target != "zh-Hans" || got.Format != "text" {
				t.Errorf("请求参数不符合预期: %+v", got)
			}

			// 单独验证 json.Number 能接住两种形态，且字面量不被改写
			dec := json.NewDecoder(strings.NewReader(c.body))
			dec.UseNumber()
			if err := dec.Decode(&resp); err != nil {
				t.Fatalf("响应解析失败: %v", err)
			}
			if string(resp.DetectedLanguage.Confidence) != c.confidence {
				t.Errorf("confidence 应原样保留为 %q，实际 %q", c.confidence, string(resp.DetectedLanguage.Confidence))
			}
		})
	}
}

func TestSetupTranslator(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ok.Close()

	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer dead.Close()

	// 保存并恢复引擎状态，避免影响其他测试
	oldEngine, oldDesc, oldURL := currentEngine, engineDesc, libreBaseURL
	defer func() {
		currentEngine, engineDesc, libreBaseURL = oldEngine, oldDesc, oldURL
	}()

	// 第一个地址不可用、第二个可用时应选中第二个
	if !SetupTranslator([]string{dead.URL, ok.URL}) {
		t.Fatal("地址池中有一个可用地址时应选中本地 API")
	}
	if !IsLocalEngine() || libreBaseURL != ok.URL {
		t.Errorf("应使用可用地址 %s，实际引擎=%v 地址=%s", ok.URL, IsLocalEngine(), libreBaseURL)
	}

	// 全部不可用时应回落到 translate-shell
	if SetupTranslator([]string{dead.URL}) {
		t.Fatal("地址池全部不可用时应回落到 translate-shell")
	}
	if IsLocalEngine() || libreBaseURL != "" {
		t.Errorf("回落时不应保留本地地址，实际引擎=%v 地址=%q", IsLocalEngine(), libreBaseURL)
	}
}

func TestTranslateLibreError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid target language"}`))
	}))
	defer srv.Close()

	if _, err := translateLibre(srv.URL, "hello"); err == nil {
		t.Fatal("非 200 响应应返回错误")
	} else if !strings.Contains(err.Error(), "Invalid target language") {
		t.Errorf("错误信息应包含服务端返回内容，实际: %v", err)
	}
}
