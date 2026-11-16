// Package webfetch 知识库网页入库：Jina Reader / Firecrawl / 简易 GET。
package webfetch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"

	"cogniforge/internal/crypto"
	"cogniforge/internal/model"
)

const (
	EngineAuto      = "auto"
	EngineJina      = "jina"
	EngineFirecrawl = "firecrawl"
	EngineBasic     = "basic"

	maxBody = 50 * 1024 * 1024
)

// Result 抓取结果（优先 Markdown，便于入库）
type Result struct {
	Engine   string // 实际使用的引擎
	FileType string // md | html
	NameHint string
	Content  []byte
}

// Fetch 按 engine 抓取；auto 时优先 Firecrawl → Jina → 简易 GET
func Fetch(ctx context.Context, db *gorm.DB, userID, rawURL, engine string) (*Result, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("网页地址无效，请使用 http:// 或 https:// 开头")
	}
	target := parsed.String()
	engine = strings.ToLower(strings.TrimSpace(engine))
	if engine == "" {
		engine = EngineAuto
	}

	switch engine {
	case EngineJina:
		return fetchJina(ctx, db, userID, target)
	case EngineFirecrawl:
		return fetchFirecrawl(ctx, db, userID, target)
	case EngineBasic:
		return fetchBasic(ctx, target)
	case EngineAuto:
		if r, err := fetchFirecrawl(ctx, db, userID, target); err == nil {
			return r, nil
		}
		if r, err := fetchJina(ctx, db, userID, target); err == nil {
			return r, nil
		}
		return fetchBasic(ctx, target)
	default:
		return nil, fmt.Errorf("不支持的抓取引擎：%s（可用 auto / jina / firecrawl / basic）", engine)
	}
}

func fetchJina(ctx context.Context, db *gorm.DB, userID, target string) (*Result, error) {
	p, apiKey, err := resolveProvider(db, userID, string(model.ProviderJina))
	if err != nil {
		return nil, err
	}
	base := strings.TrimRight(strings.TrimSpace(p.BaseURL), "/")
	if base == "" {
		base = "https://r.jina.ai"
	}
	reqURL := base + "/" + target

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Jina 请求失败")
	}
	req.Header.Set("Accept", "text/markdown")
	req.Header.Set("User-Agent", "CogniForgeKnowledgeBot/1.0")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	body, err := doRead(req, 45*time.Second)
	if err != nil {
		return nil, fmt.Errorf("Jina Reader 抓取失败: %w", err)
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil, fmt.Errorf("Jina Reader 返回空内容")
	}
	return &Result{Engine: EngineJina, FileType: "md", Content: body}, nil
}

func fetchFirecrawl(ctx context.Context, db *gorm.DB, userID, target string) (*Result, error) {
	p, apiKey, err := resolveProvider(db, userID, string(model.ProviderFirecrawl))
	if err != nil {
		return nil, err
	}
	if apiKey == "" {
		return nil, fmt.Errorf("请先在「模型」页为 Firecrawl 填写 API Key 并启用")
	}
	base := strings.TrimRight(strings.TrimSpace(p.BaseURL), "/")
	if base == "" {
		base = "https://api.firecrawl.dev"
	}
	endpoint := base + "/v1/scrape"
	payload, _ := json.Marshal(map[string]interface{}{
		"url":     target,
		"formats": []string{"markdown"},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("创建 Firecrawl 请求失败")
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CogniForgeKnowledgeBot/1.0")

	body, err := doRead(req, 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("Firecrawl 抓取失败: %w", err)
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Markdown string `json:"markdown"`
			Content  string `json:"content"`
		} `json:"data"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Firecrawl 响应无法解析")
	}
	if !resp.Success && resp.Error != "" {
		return nil, fmt.Errorf("Firecrawl: %s", resp.Error)
	}
	md := strings.TrimSpace(resp.Data.Markdown)
	if md == "" {
		md = strings.TrimSpace(resp.Data.Content)
	}
	if md == "" {
		return nil, fmt.Errorf("Firecrawl 返回空内容")
	}
	return &Result{Engine: EngineFirecrawl, FileType: "md", Content: []byte(md)}, nil
}

func fetchBasic(ctx context.Context, target string) (*Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败")
	}
	req.Header.Set("User-Agent", "CogniForgeKnowledgeBot/1.0")
	body, err := doRead(req, 20*time.Second)
	if err != nil {
		return nil, fmt.Errorf("简易抓取失败: %w", err)
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil, fmt.Errorf("网页没有可读内容")
	}
	return &Result{Engine: EngineBasic, FileType: "html", Content: body}, nil
}

func doRead(req *http.Request, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		msg := strings.TrimSpace(string(limited))
		if msg == "" {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}
	limited := io.LimitReader(resp.Body, maxBody+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(body) > maxBody {
		return nil, fmt.Errorf("内容过大，超过 %dMB", maxBody/(1024*1024))
	}
	return body, nil
}

// resolveProvider 取用户可见且启用的 web_fetch 供应商，并解密 Key
func resolveProvider(db *gorm.DB, userID, vendor string) (*model.AIProvider, string, error) {
	if db == nil {
		return nil, "", fmt.Errorf("数据库未就绪")
	}
	var rows []model.AIProvider
	q := db.Where("provider = ? AND is_enabled = ? AND deleted_at IS NULL", vendor, true).
		Order("priority DESC, updated_at DESC")
	if err := q.Find(&rows).Error; err != nil {
		return nil, "", fmt.Errorf("查询 %s 配置失败", vendor)
	}
	var picked *model.AIProvider
	for i := range rows {
		p := &rows[i]
		// 可见：公开，或本人私有（厂商键已是 jina/firecrawl）
		if p.Scope == model.ProviderScopePublic || p.Scope == "" || p.OwnerUserID == userID {
			picked = p
			break
		}
	}
	if picked == nil {
		return nil, "", fmt.Errorf("未配置或未启用 %s：请到「模型」页添加 API Key 并勾选「网页抓取」", vendor)
	}
	apiKey := strings.TrimSpace(picked.APIKey)
	if apiKey != "" {
		dec, err := crypto.Decrypt(picked.APIKey)
		if err != nil {
			return nil, "", fmt.Errorf("%s API Key 解密失败", vendor)
		}
		apiKey = strings.TrimSpace(dec)
	}
	return picked, apiKey, nil
}
