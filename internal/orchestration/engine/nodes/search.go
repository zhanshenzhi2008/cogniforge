package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cogniforge/internal/capability/knowledge"
	"cogniforge/internal/orchestration/engine/core"
)

// SearchNodeConfig 知识库检索节点（画布 type=search）
type SearchNodeConfig struct {
	KnowledgeBaseID string  `json:"knowledge_base_id"`
	Query           string  `json:"query,omitempty"`
	TopK            int     `json:"top_k,omitempty"`
	MinScore        float64 `json:"min_score,omitempty"`
	SearchMode      string  `json:"search_mode,omitempty"`
}

// KnowledgeSearcher 与 Chat RAG 共用的检索能力
type KnowledgeSearcher interface {
	SearchByUser(ctx context.Context, userID, kbID string, req *knowledge.SearchRequest) (*knowledge.SearchResponse, error)
}

// SearchNodeExecutor 按配置知识库做向量检索，结果写入上下文供下游 LLM/Agent 使用
type SearchNodeExecutor struct {
	Knowledge KnowledgeSearcher
}

func (e *SearchNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	var cfg SearchNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid search config: %w", err)
	}
	if e == nil || e.Knowledge == nil {
		return nil, fmt.Errorf("Search node: KnowledgeService not configured")
	}
	if cfg.KnowledgeBaseID == "" {
		return nil, fmt.Errorf("Search node: knowledge_base_id is required")
	}
	userID := ctx.UserID
	if userID == "" {
		return nil, fmt.Errorf("Search node: missing user_id on execution context")
	}

	query := resolveTemplate(cfg.Query, ctx)
	if query == "" {
		for _, key := range []string{"query", "message", "prompt", "input"} {
			if v, ok := ctx.Input[key]; ok {
				query = strings.TrimSpace(fmt.Sprint(v))
				if query != "" {
					break
				}
			}
		}
	}
	if query == "" {
		if v, ok := ctx.GetVariable("last_agent_content"); ok {
			query = strings.TrimSpace(fmt.Sprint(v))
		}
	}
	if query == "" {
		return nil, fmt.Errorf("Search node: empty query (set config.query or input.query)")
	}

	started := time.Now()
	bg, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := e.Knowledge.SearchByUser(bg, userID, cfg.KnowledgeBaseID, &knowledge.SearchRequest{
		Query:      query,
		TopK:       cfg.TopK,
		MinScore:   cfg.MinScore,
		SearchMode: cfg.SearchMode,
	})
	if err != nil {
		return nil, fmt.Errorf("Search node: %w", err)
	}

	results := make([]map[string]any, 0, len(resp.Results))
	var parts []string
	for i, r := range resp.Results {
		results = append(results, map[string]any{
			"document_id":   r.DocumentID,
			"document_name": r.DocumentName,
			"chunk_id":      r.ChunkID,
			"content":       r.Content,
			"score":         r.Score,
		})
		name := r.DocumentName
		if name == "" {
			name = r.DocumentID
		}
		parts = append(parts, fmt.Sprintf("[%d] (%s, score=%.3f)\n%s", i+1, name, r.Score, r.Content))
	}
	contextText := strings.Join(parts, "\n\n")
	ctx.SetVariable("last_search_query", query)
	ctx.SetVariable("last_search_context", contextText)
	ctx.SetVariable("last_search_total", len(results))

	latency := time.Since(started).Milliseconds()
	if ctx.Logger != nil {
		ctx.Logger.Logf(ctx.NodeID, "info", "Search ok kb=%s hits=%d latency=%dms", cfg.KnowledgeBaseID, len(results), latency)
	}

	return map[string]any{
		"query":               query,
		"knowledge_base_id":   cfg.KnowledgeBaseID,
		"results":             results,
		"total":               len(results),
		"context":             contextText,
		"latency_ms":          latency,
		"duration_ms":         resp.Duration,
	}, nil
}
