package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPNodeConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    any               `json:"body,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
}

type HTTPNodeExecutor struct {
	client *http.Client
}

func NewHTTPNodeExecutor() *HTTPNodeExecutor {
	return &HTTPNodeExecutor{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *HTTPNodeExecutor) Execute(ctx *ExecutionContext, config json.RawMessage) (any, error) {
	var cfg HTTPNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid HTTP config: %w", err)
	}

	if cfg.Method == "" {
		cfg.Method = "GET"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30
	}

	var bodyReader io.Reader
	if cfg.Body != nil {
		body, err := json.Marshal(cfg.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(cfg.Method, cfg.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, val := range cfg.Headers {
		req.Header.Set(key, val)
	}

	ctx.Logger.Logf(ctx.NodeID, "info", "Making HTTP request: %s %s", cfg.Method, cfg.URL)

	startTime := time.Now()
	resp, err := e.client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		ctx.Logger.Logf(ctx.NodeID, "error", "HTTP request failed: %v", err)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	result := map[string]any{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        string(respBody),
		"duration_ms": duration.Milliseconds(),
	}

	ctx.Logger.Logf(ctx.NodeID, "info", "HTTP response: status=%d, duration=%dms", resp.StatusCode, duration.Milliseconds())

	return result, nil
}
