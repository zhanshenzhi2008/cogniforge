package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jzf21/tavily-client/tavily"
)

// TavilyClient Tavily 搜索客户端
// 支持两种模式：
// 1. API Key 模式：使用 TAVILY_API_KEY 环境变量，每月 1000 次免费额度
// 2. Keyless 模式：无 API Key，有频率限制，适合测试/开发
type TavilyClient struct {
	client  *tavily.Client
	baseURL string // 自定义 API 地址（可选，用于代理）
	keyless bool
}

// NewTavilyClient 创建 Tavily 客户端
func NewTavilyClient() *TavilyClient {
	apiKey := os.Getenv("TAVILY_API_KEY")

	if apiKey != "" {
		client := tavily.NewClient(apiKey)
		slog.Info("Tavily client: using API key mode")
		return &TavilyClient{client: client, keyless: false}
	}

	slog.Info("Tavily client: using keyless mode (rate-limited)")
	return &TavilyClient{client: nil, keyless: true}
}

// SearchResult 搜索结果
type SearchResult struct {
	Query      string             `json:"query"`
	Results    []SearchResultItem `json:"results"`
	Total      int                `json:"total"`
	MaxResults int                `json:"max_results"`
	Status     string             `json:"status"` // "ok" | "error" | "keyless_limited"
}

// SearchResultItem 单条搜索结果
type SearchResultItem struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// Search 执行网络搜索
func (t *TavilyClient) Search(ctx context.Context, query string, maxResults int) (*SearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if t.keyless {
		return t.searchKeyless(ctx, query, maxResults)
	}
	return t.searchWithKey(ctx, query, maxResults)
}

// searchWithKey 使用 API Key 搜索
func (t *TavilyClient) searchWithKey(ctx context.Context, query string, maxResults int) (*SearchResult, error) {
	resp, err := t.client.SearchWithOptions(&tavily.SearchRequest{
		Query:      query,
		MaxResults: maxResults,
	})
	if err != nil {
		slog.Error("Tavily search failed", "error", err, "query", query)
		return nil, err
	}

	results := make([]SearchResultItem, 0, len(resp.Results))
	for _, r := range resp.Results {
		results = append(results, SearchResultItem{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Content,
		})
	}

	return &SearchResult{
		Query:      query,
		Results:    results,
		Total:      len(results),
		MaxResults: maxResults,
		Status:     "ok",
	}, nil
}

// searchKeyless 使用 keyless 模式搜索
// 参考: https://docs.tavily.com/documentation/keyless
func (t *TavilyClient) searchKeyless(ctx context.Context, query string, maxResults int) (*SearchResult, error) {
	body := map[string]interface{}{
		"query":       query,
		"max_results": maxResults,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := "https://api.tavily.com/search"
	if t.baseURL != "" {
		url = t.baseURL + "/search"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tavily-Access-Mode", "keyless")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Tavily keyless search failed", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("Tavily keyless search error", "status", resp.StatusCode, "body", string(respBody))
		return nil, err
	}

	var rawResp struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, err
	}

	results := make([]SearchResultItem, 0, len(rawResp.Results))
	for _, r := range rawResp.Results {
		results = append(results, SearchResultItem{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Content,
		})
	}

	return &SearchResult{
		Query:      query,
		Results:    results,
		Total:      len(results),
		MaxResults: maxResults,
		Status:     "ok",
	}, nil
}
