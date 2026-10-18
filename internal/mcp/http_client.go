package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

const (
	mcpProtocolVersion = "2025-06-18"
	mcpClientName      = "cogniforge"
	mcpClientVersion   = "1.0.0"
)

// StreamableHTTPClient MCP Streamable HTTP 客户端（initialize / tools/list / tools/call）
type StreamableHTTPClient struct {
	endpoint   string
	authHeader string
	http       *http.Client
	sessionID  string
	protoVer   string
	nextID     atomic.Int64
}

// NewStreamableHTTPClient 创建客户端。endpoint 为用户填写的 MCP URL（单端点）。
func NewStreamableHTTPClient(endpoint, authHeader string) *StreamableHTTPClient {
	return &StreamableHTTPClient{
		endpoint:   strings.TrimRight(strings.TrimSpace(endpoint), "/"),
		authHeader: normalizeAuthHeader(authHeader),
		http:       &http.Client{Timeout: 45 * time.Second},
		protoVer:   mcpProtocolVersion,
	}
}

func normalizeAuthHeader(h string) string {
	h = strings.TrimSpace(h)
	if h == "" {
		return ""
	}
	lower := strings.ToLower(h)
	if strings.HasPrefix(lower, "bearer ") || strings.HasPrefix(lower, "basic ") || strings.Contains(h, ":") {
		return h
	}
	return "Bearer " + h
}

type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type toolsListResult struct {
	Tools []mcpToolInfo `json:"tools"`
}

type toolsCallResult struct {
	Content []map[string]interface{} `json:"content"`
	IsError bool                     `json:"isError"`
}

// Initialize 握手并发送 initialized 通知
func (c *StreamableHTTPClient) Initialize(ctx context.Context) error {
	params := map[string]interface{}{
		"protocolVersion": mcpProtocolVersion,
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]string{
			"name":    mcpClientName,
			"version": mcpClientVersion,
		},
	}
	raw, err := c.request(ctx, "initialize", params, true)
	if err != nil {
		return err
	}
	var initRes struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	if err := json.Unmarshal(raw, &initRes); err == nil && initRes.ProtocolVersion != "" {
		c.protoVer = initRes.ProtocolVersion
	}
	// notifications/initialized（无 id）
	_ = c.notify(ctx, "notifications/initialized", map[string]interface{}{})
	return nil
}

// ListTools 发现远程工具
func (c *StreamableHTTPClient) ListTools(ctx context.Context) ([]ToolDefinition, error) {
	if c.sessionID == "" {
		if err := c.Initialize(ctx); err != nil {
			return nil, err
		}
	}
	raw, err := c.request(ctx, "tools/list", map[string]interface{}{}, true)
	if err != nil {
		return nil, err
	}
	var res toolsListResult
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("解析 tools/list: %w", err)
	}
	out := make([]ToolDefinition, 0, len(res.Tools))
	for _, t := range res.Tools {
		schema := t.InputSchema
		if schema == nil {
			schema = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
		}
		out = append(out, ToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: schema,
			ServerType:  "http",
		})
	}
	return out, nil
}

// CallTool 调用远程工具，返回可给模型看的文本/JSON
func (c *StreamableHTTPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (map[string]interface{}, error) {
	if c.sessionID == "" {
		if err := c.Initialize(ctx); err != nil {
			return nil, err
		}
	}
	if args == nil {
		args = map[string]interface{}{}
	}
	raw, err := c.request(ctx, "tools/call", map[string]interface{}{
		"name":      name,
		"arguments": args,
	}, true)
	if err != nil {
		return nil, err
	}
	var res toolsCallResult
	if err := json.Unmarshal(raw, &res); err != nil {
		// 非标准结果原样返回
		var any map[string]interface{}
		if err2 := json.Unmarshal(raw, &any); err2 == nil {
			return any, nil
		}
		return map[string]interface{}{"raw": string(raw)}, nil
	}
	textParts := make([]string, 0, len(res.Content))
	for _, block := range res.Content {
		if typ, _ := block["type"].(string); typ == "text" {
			if s, ok := block["text"].(string); ok {
				textParts = append(textParts, s)
			}
		}
	}
	out := map[string]interface{}{
		"content":  res.Content,
		"is_error": res.IsError,
		"status":   "ok",
	}
	if len(textParts) > 0 {
		out["text"] = strings.Join(textParts, "\n")
	}
	if res.IsError {
		out["status"] = "error"
	}
	return out, nil
}

func (c *StreamableHTTPClient) notify(ctx context.Context, method string, params interface{}) error {
	body, _ := json.Marshal(jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.setHeaders(req, false)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (c *StreamableHTTPClient) request(ctx context.Context, method string, params interface{}, captureSession bool) (json.RawMessage, error) {
	id := c.nextID.Add(1)
	body, _ := json.Marshal(jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req, true)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MCP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if captureSession {
		if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
			c.sessionID = sid
		}
	}

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("读取 MCP 响应失败: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("MCP HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 400))
	}

	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	var rpc jsonRPCResponse
	switch {
	case strings.Contains(ct, "text/event-stream"):
		raw, err := parseSSEJSONRPC(respBody, id)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &rpc); err != nil {
			return nil, fmt.Errorf("解析 SSE JSON-RPC: %w", err)
		}
	default:
		if err := json.Unmarshal(respBody, &rpc); err != nil {
			return nil, fmt.Errorf("解析 JSON-RPC: %w", err)
		}
	}
	if rpc.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", rpc.Error.Code, rpc.Error.Message)
	}
	if len(rpc.Result) == 0 {
		return nil, fmt.Errorf("MCP 响应无 result")
	}
	return rpc.Result, nil
}

func (c *StreamableHTTPClient) setHeaders(req *http.Request, expectBody bool) {
	if expectBody {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
	} else {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
	}
	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}
	if c.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", c.sessionID)
	}
	if c.protoVer != "" {
		req.Header.Set("MCP-Protocol-Version", c.protoVer)
	}
}

// parseSSEJSONRPC 从 SSE 流中取出匹配 id 的 JSON-RPC 响应
func parseSSEJSONRPC(body []byte, wantID int64) (json.RawMessage, error) {
	sc := bufio.NewScanner(bytes.NewReader(body))
	sc.Buffer(make([]byte, 0, 64*1024), 2<<20)
	var dataBuf strings.Builder
	flush := func() (json.RawMessage, bool, error) {
		if dataBuf.Len() == 0 {
			return nil, false, nil
		}
		payload := strings.TrimSpace(dataBuf.String())
		dataBuf.Reset()
		if payload == "" || payload == "[DONE]" {
			return nil, false, nil
		}
		var rpc jsonRPCResponse
		if err := json.Unmarshal([]byte(payload), &rpc); err != nil {
			return nil, false, nil
		}
		if rpc.Error != nil || len(rpc.Result) > 0 {
			if idMatches(rpc.ID, wantID) || len(rpc.ID) == 0 {
				raw, _ := json.Marshal(rpc)
				return raw, true, nil
			}
		}
		return nil, false, nil
	}
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			if raw, ok, err := flush(); err != nil {
				return nil, err
			} else if ok {
				return raw, nil
			}
			continue
		}
		if strings.HasPrefix(line, "data:") {
			if dataBuf.Len() > 0 {
				dataBuf.WriteByte('\n')
			}
			dataBuf.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if raw, ok, err := flush(); err != nil {
		return nil, err
	} else if ok {
		return raw, nil
	}
	return nil, fmt.Errorf("SSE 流中未找到 JSON-RPC 响应")
}

func idMatches(raw json.RawMessage, want int64) bool {
	if len(raw) == 0 {
		return false
	}
	var n int64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n == want
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s == fmt.Sprintf("%d", want)
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// CallHTTPClient 优先走标准 MCP Streamable HTTP；失败再回退旧网关格式
func (e *BuiltInExecutor) CallHTTPClient(url, authHeader string, toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	client := NewStreamableHTTPClient(url, authHeader)
	result, err := client.CallTool(ctx, toolName, args)
	if err == nil {
		return result, nil
	}
	mcpErr := err

	// 回退：旧简化网关 POST {url}/tools/{name}/call
	legacy, legErr := e.callLegacyHTTPGateway(ctx, url, authHeader, toolName, args)
	if legErr == nil {
		return legacy, nil
	}
	return nil, fmt.Errorf("MCP 调用失败: %v; 网关回退: %v", mcpErr, legErr)
}

func (e *BuiltInExecutor) callLegacyHTTPGateway(ctx context.Context, url, authHeader, toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	body, _ := json.Marshal(map[string]interface{}{"arguments": args})
	endpoint := strings.TrimRight(url, "/") + "/tools/" + toolName + "/call"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if h := normalizeAuthHeader(authHeader); h != "" {
		req.Header.Set("Authorization", h)
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 300))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DiscoverHTTPTools 对远程 MCP 做 tools/list（供 Registry / API 使用）
func DiscoverHTTPTools(ctx context.Context, serverID, url, authHeader string) ([]ToolDefinition, error) {
	client := NewStreamableHTTPClient(url, authHeader)
	tools, err := client.ListTools(ctx)
	if err != nil {
		return nil, err
	}
	for i := range tools {
		tools[i].ServerID = serverID
		tools[i].ServerType = "http"
	}
	return tools, nil
}
