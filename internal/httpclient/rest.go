package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cogniforge/internal/trace"
)

// Client 是封装了 trace 自动注入的通用 HTTP Client。
// 通过 RoundTripper 拦截所有请求，自动从 context 中提取 trace ID 注入 header，
// 同时提供 JSON 序列化和响应解析的便捷方法，所有调用点无需关心 trace 传递。
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClient 创建一个 HTTP Client，所有请求自动携带 trace ID。
//
//	cli := httpclient.NewClient("http://python-service:8086")
//	resp, err := cli.Post(ctx, "/api/rag/process", payload, &result)
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Transport: &trace.Transport{},
			Timeout:   60 * time.Second,
		},
	}
}

// SetTimeout 设置 HTTP Client 的超时时间。
func (c *Client) SetTimeout(timeout time.Duration) {
	c.client.Timeout = timeout
}

// doRequest 是统一的请求发送方法，从 context 提取 trace ID。
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// X-Trace-ID 由 trace.Transport 自动从 context 注入，无需手动处理

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// checkResponse 检查响应状态码，非 200 返回 error。
func checkResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
}

// Get 发送 GET 请求，将响应 body 解析到 result。
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return err
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// Post 发送 POST 请求，将响应 body 解析到 result。
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return err
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// Put 发送 PUT 请求，将响应 body 解析到 result。
func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return err
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// Delete 发送 DELETE 请求，将响应 body 解析到 result。
func (c *Client) Delete(ctx context.Context, path string, body, result interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return err
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// Do 发送任意 HTTP 请求，返回原始响应（调用方自行解析 body）。
func (c *Client) Do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, method, path, body)
}

// RawClient 返回底层的 *http.Client，可用于不需要走 doRequest 的场景（如健康检查）。
func (c *Client) RawClient() *http.Client {
	return c.client
}

// BaseURL 返回 client 的 base URL。
func (c *Client) BaseURL() string {
	return c.baseURL
}

// HTTPClient 返回底层的标准 *http.Client。
func (c *Client) HTTPClient() *http.Client {
	return c.client
}
