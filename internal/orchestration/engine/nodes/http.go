package nodes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cogniforge/internal/orchestration/engine/core"
)

type HTTPNodeConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        any               `json:"body,omitempty"`
	Timeout     json.RawMessage   `json:"timeout,omitempty"` // 数字或 "{{变量}}"；空/0→30
	FailOnError bool              `json:"fail_on_error,omitempty"` // true 时 status>=400 使节点失败
}

type HTTPNodeExecutor struct {
	client *http.Client
}

func NewHTTPNodeExecutor() *HTTPNodeExecutor {
	return &HTTPNodeExecutor{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *HTTPNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	var cfg HTTPNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid HTTP config: %w", err)
	}

	if cfg.Method == "" {
		cfg.Method = "GET"
	}
	timeoutSec := resolveFlexibleInt(cfg.Timeout, ctx)
	if timeoutSec <= 0 {
		timeoutSec = 30
	}

	url := resolveTemplate(cfg.URL, ctx)
	if url == "" {
		return nil, fmt.Errorf("HTTP node: url is required")
	}

	var bodyReader io.Reader
	if cfg.Body != nil {
		resolved := resolveAnyTemplates(cfg.Body, ctx)
		switch b := resolved.(type) {
		case string:
			bodyReader = bytes.NewBufferString(b)
		default:
			body, err := json.Marshal(b)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			bodyReader = bytes.NewBuffer(body)
		}
	}

	req, err := http.NewRequest(cfg.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, val := range cfg.Headers {
		req.Header.Set(resolveTemplate(key, ctx), resolveTemplate(val, ctx))
	}

	client := e.client
	if client == nil {
		client = &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	} else {
		client = &http.Client{
			Timeout:   time.Duration(timeoutSec) * time.Second,
			Transport: e.client.Transport,
		}
	}

	ctx.Logger.Logf(ctx.NodeID, "info", "Making HTTP request: %s %s", cfg.Method, url)

	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		ctx.Logger.Logf(ctx.NodeID, "error", "HTTP request failed: %v", err)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	bodyStr := string(respBody)
	result := map[string]any{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        bodyStr,
		"duration_ms": duration.Milliseconds(),
	}

	// 供下游 {{last_http_body}} / {{last_http_status}} 引用（与 last_agent_content 同套路）
	ctx.SetVariable("last_http_status", resp.StatusCode)
	ctx.SetVariable("last_http_body", bodyStr)

	ctx.Logger.Logf(ctx.NodeID, "info", "HTTP response: status=%d, duration=%dms", resp.StatusCode, duration.Milliseconds())

	if cfg.FailOnError && resp.StatusCode >= 400 {
		snippet := bodyStr
		if len(snippet) > 200 {
			snippet = snippet[:200] + "…"
		}
		return result, fmt.Errorf("HTTP status %d: %s", resp.StatusCode, snippet)
	}

	return result, nil
}

// resolveAnyTemplates 递归替换 map/slice/string 中的 {{var}}
func resolveAnyTemplates(v any, ctx *core.ExecutionContext) any {
	switch t := v.(type) {
	case string:
		return resolveTemplate(t, ctx)
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			out[k] = resolveAnyTemplates(vv, ctx)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, vv := range t {
			out[i] = resolveAnyTemplates(vv, ctx)
		}
		return out
	default:
		return v
	}
}
