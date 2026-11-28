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
// 空的 search_mode / top_k<=0 / min_score<=0 / rerank 未传 → 跟知识库默认
type SearchNodeConfig struct {
	KnowledgeBaseID      string          `json:"knowledge_base_id"`
	Query                string          `json:"query,omitempty"`
	TopK                 json.RawMessage `json:"top_k,omitempty"`     // 数字或 "{{变量}}"；≤0 跟库默认
	MinScore             json.RawMessage `json:"min_score,omitempty"` // 数字或 "{{变量}}"；≤0 跟库默认
	SearchMode           string          `json:"search_mode,omitempty"`
	Rerank               *bool           `json:"rerank,omitempty"`
	RerankProviderID     string          `json:"rerank_provider_id,omitempty"`
	// HybridSemanticWeight：数字或 "{{变量}}"；空/null 跟库默认（0 也是合法覆盖）
	HybridSemanticWeight json.RawMessage `json:"hybrid_semantic_weight,omitempty"`
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
	userID := ctx.UserID
	if userID == "" {
		return nil, fmt.Errorf("Search node: missing user_id on execution context")
	}

	kbID := strings.TrimSpace(resolveTemplate(cfg.KnowledgeBaseID, ctx))
	if kbID == "" {
		return nil, fmt.Errorf("Search node: knowledge_base_id is required")
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

	topK := resolveFlexibleInt(cfg.TopK, ctx)
	minScore := resolveFlexibleFloat(cfg.MinScore, ctx)
	hybridWeight := resolveOptionalFlexibleFloat(cfg.HybridSemanticWeight, ctx)
	rerankProviderID := strings.TrimSpace(resolveTemplate(cfg.RerankProviderID, ctx))
	searchMode := strings.ToLower(strings.TrimSpace(resolveTemplate(cfg.SearchMode, ctx)))

	started := time.Now()
	bg, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := e.Knowledge.SearchByUser(bg, userID, kbID, &knowledge.SearchRequest{
		Query:                query,
		TopK:                 topK,
		MinScore:             minScore,
		SearchMode:           searchMode,
		Rerank:               cfg.Rerank,
		RerankProviderID:     rerankProviderID,
		HybridSemanticWeight: hybridWeight,
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
	ctx.SetVariable("last_search_kb_id", kbID)
	ctx.SetVariable("last_search_rerank_provider_id", rerankProviderID)
	ctx.SetVariable("last_search_mode", searchMode)

	latency := time.Since(started).Milliseconds()
	if ctx.Logger != nil {
		ctx.Logger.Logf(ctx.NodeID, "info", "Search ok kb=%s hits=%d latency=%dms", kbID, len(results), latency)
	}

	return map[string]any{
		"query":             query,
		"knowledge_base_id": kbID,
		"results":           results,
		"total":             len(results),
		"context":           contextText,
		"latency_ms":        latency,
		"duration_ms":       resp.Duration,
	}, nil
}
