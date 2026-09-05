package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BuiltInExecutor 内置工具执行器
type BuiltInExecutor struct {
	client *http.Client
}

func NewBuiltInExecutor() *BuiltInExecutor {
	return &BuiltInExecutor{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Execute 执行内置工具
func (e *BuiltInExecutor) Execute(toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	switch toolName {
	case "fetch_get_page":
		return e.fetchGetPage(args)
	case "web_search_search":
		return e.webSearch(args)
	case "code_execute_python":
		return e.codeExecute("python", args)
	case "code_execute_go":
		return e.codeExecute("go", args)
	case "code_execute_javascript":
		return e.codeExecute("javascript", args)
	default:
		return nil, fmt.Errorf("未知工具: %s", toolName)
	}
}

// fetchGetPage 获取网页内容
func (e *BuiltInExecutor) fetchGetPage(args map[string]interface{}) (map[string]interface{}, error) {
	url, ok := args["url"].(string)
	if !ok || url == "" {
		return nil, fmt.Errorf("url 参数必填")
	}
	maxLength := 5000
	if ml, ok := args["max_length"].(float64); ok {
		maxLength = int(ml)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; CogniForge-MCP/1.0)")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // 最多 1MB
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	content := string(body)
	if len(content) > maxLength {
		content = content[:maxLength] + "..."
	}

	return map[string]interface{}{
		"url":     url,
		"content": content,
		"status":  "ok",
	}, nil
}

// webSearch 网络搜索（占位：实际需要 Bing/Google API Key）
func (e *BuiltInExecutor) webSearch(args map[string]interface{}) (map[string]interface{}, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query 参数必填")
	}
	maxResults := 5
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// 占位返回：提示用户需要配置搜索 API
	return map[string]interface{}{
		"query": query,
		"results": []map[string]string{
			{
				"title":       "搜索功能待配置",
				"url":         "https://www.bing.com",
				"description": "请在环境变量中配置 BING_API_KEY 或 GOOGLE_API_KEY 以启用搜索功能",
			},
		},
		"total":     1,
		"max_result": maxResults,
		"status":    "placeholder",
	}, nil
}

// codeExecute 代码执行（占位：沙箱容器后期实现）
func (e *BuiltInExecutor) codeExecute(language string, args map[string]interface{}) (map[string]interface{}, error) {
	code, ok := args["code"].(string)
	if !ok || code == "" {
		return nil, fmt.Errorf("code 参数必填")
	}
	timeout := 30
	if t, ok := args["timeout"].(float64); ok {
		timeout = int(t)
	}

	// 占位返回：提示代码执行需后期沙箱支持
	return map[string]interface{}{
		"language": language,
		"output":  fmt.Sprintf("[沙箱执行占位] 语言: %s, 超时: %ds\n\n代码:\n%s\n\n请部署代码执行沙箱服务后启用此功能", language, timeout, code),
		"status":  "placeholder",
	}, nil
}

// CallHTTPClient 调用 HTTP MCP Server
func (e *BuiltInExecutor) CallHTTPClient(url, authHeader string, toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	body, _ := json.Marshal(map[string]interface{}{
		"arguments": args,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url+"/tools/"+toolName+"/call", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP Server 返回错误 %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return result, nil
}
