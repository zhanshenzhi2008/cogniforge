package nodes

import (
	"context"
	"encoding/json"
	"testing"

	"cogniforge/internal/capability/knowledge"
	"cogniforge/internal/orchestration/engine/core"
)

type mockKBSearcher struct {
	lastQuery            string
	lastKB               string
	lastTopK             int
	lastMinScore         float64
	lastHybridWeight     *float64
	lastRerankProviderID string
	lastSearchMode       string
}

func (m *mockKBSearcher) SearchByUser(_ context.Context, _, kbID string, req *knowledge.SearchRequest) (*knowledge.SearchResponse, error) {
	m.lastKB = kbID
	m.lastQuery = req.Query
	m.lastTopK = req.TopK
	m.lastMinScore = req.MinScore
	m.lastHybridWeight = req.HybridSemanticWeight
	m.lastRerankProviderID = req.RerankProviderID
	m.lastSearchMode = req.SearchMode
	return &knowledge.SearchResponse{
		Query: req.Query,
		Results: []knowledge.SearchResult{
			{DocumentID: "d1", DocumentName: "手册", ChunkID: "c1", Content: "退款需审批", Score: 0.91},
		},
		Total:    1,
		Duration: 12,
	}, nil
}

func TestSearchNodeExecutor_OK(t *testing.T) {
	mock := &mockKBSearcher{}
	ex := &SearchNodeExecutor{Knowledge: mock}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"query": "如何退款"})
	ctx.UserID = "u1"
	cfg, _ := json.Marshal(map[string]any{
		"knowledge_base_id": "kb-1",
		"query":             "{{query}}",
		"top_k":             3,
	})
	out, err := ex.Execute(ctx, cfg)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("unexpected out type %T", out)
	}
	if m["total"] != 1 {
		t.Fatalf("total=%v", m["total"])
	}
	if mock.lastKB != "kb-1" || mock.lastQuery == "" {
		t.Fatalf("searcher not called properly: kb=%s q=%q", mock.lastKB, mock.lastQuery)
	}
	if mock.lastTopK != 3 {
		t.Fatalf("lastTopK=%d", mock.lastTopK)
	}
	if v, _ := ctx.GetVariable("last_search_total"); v != 1 {
		t.Fatalf("last_search_total=%v", v)
	}
}

func TestSearchNodeExecutor_TopKTemplate(t *testing.T) {
	mock := &mockKBSearcher{}
	ex := &SearchNodeExecutor{Knowledge: mock}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"query": "hi"})
	ctx.UserID = "u1"
	ctx.SetVariable("k", "7")
	cfg, _ := json.Marshal(map[string]any{
		"knowledge_base_id": "kb-1",
		"query":             "hi",
		"top_k":             "{{k}}",
	})
	if _, err := ex.Execute(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if mock.lastTopK != 7 {
		t.Fatalf("lastTopK=%d", mock.lastTopK)
	}
}

func TestSearchNodeExecutor_MinScoreTemplate(t *testing.T) {
	mock := &mockKBSearcher{}
	ex := &SearchNodeExecutor{Knowledge: mock}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"query": "hi"})
	ctx.UserID = "u1"
	ctx.SetVariable("floor", "0.42")
	cfg, _ := json.Marshal(map[string]any{
		"knowledge_base_id": "kb-1",
		"query":             "hi",
		"min_score":         "{{floor}}",
	})
	if _, err := ex.Execute(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if mock.lastMinScore < 0.419 || mock.lastMinScore > 0.421 {
		t.Fatalf("lastMinScore=%v", mock.lastMinScore)
	}
}

func TestSearchNodeExecutor_KnowledgeBaseIDTemplate(t *testing.T) {
	mock := &mockKBSearcher{}
	ex := &SearchNodeExecutor{Knowledge: mock}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"query": "hi"})
	ctx.UserID = "u1"
	ctx.SetVariable("kb", "kb-dyn")
	cfg, _ := json.Marshal(map[string]any{
		"knowledge_base_id": "{{kb}}",
		"query":             "hi",
	})
	if _, err := ex.Execute(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if mock.lastKB != "kb-dyn" {
		t.Fatalf("lastKB=%s", mock.lastKB)
	}
	if v, _ := ctx.GetVariable("last_search_kb_id"); v != "kb-dyn" {
		t.Fatalf("last_search_kb_id=%v", v)
	}
}

func TestSearchNodeExecutor_HybridWeightTemplate(t *testing.T) {
	mock := &mockKBSearcher{}
	ex := &SearchNodeExecutor{Knowledge: mock}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"query": "hi"})
	ctx.UserID = "u1"
	ctx.SetVariable("sem", "0.65")
	cfg, _ := json.Marshal(map[string]any{
		"knowledge_base_id":      "kb-1",
		"query":                  "hi",
		"hybrid_semantic_weight": "{{sem}}",
	})
	if _, err := ex.Execute(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if mock.lastHybridWeight == nil || *mock.lastHybridWeight < 0.649 || *mock.lastHybridWeight > 0.651 {
		t.Fatalf("lastHybridWeight=%v", mock.lastHybridWeight)
	}
}

func TestSearchNodeExecutor_SearchModeTemplate(t *testing.T) {
	mock := &mockKBSearcher{}
	ex := &SearchNodeExecutor{Knowledge: mock}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"query": "hi"})
	ctx.UserID = "u1"
	ctx.SetVariable("mode", "Hybrid")
	cfg, _ := json.Marshal(map[string]any{
		"knowledge_base_id": "kb-1",
		"query":             "hi",
		"search_mode":       "{{mode}}",
	})
	if _, err := ex.Execute(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if mock.lastSearchMode != "hybrid" {
		t.Fatalf("lastSearchMode=%s", mock.lastSearchMode)
	}
	if v, _ := ctx.GetVariable("last_search_mode"); v != "hybrid" {
		t.Fatalf("last_search_mode=%v", v)
	}
}

func TestSearchNodeExecutor_RerankProviderIDTemplate(t *testing.T) {
	mock := &mockKBSearcher{}
	ex := &SearchNodeExecutor{Knowledge: mock}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"query": "hi"})
	ctx.UserID = "u1"
	ctx.SetVariable("rrp", "prov-rerank-9")
	on := true
	cfg, _ := json.Marshal(map[string]any{
		"knowledge_base_id":  "kb-1",
		"query":              "hi",
		"rerank":             on,
		"rerank_provider_id": "{{rrp}}",
	})
	if _, err := ex.Execute(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if mock.lastRerankProviderID != "prov-rerank-9" {
		t.Fatalf("lastRerankProviderID=%s", mock.lastRerankProviderID)
	}
	if v, _ := ctx.GetVariable("last_search_rerank_provider_id"); v != "prov-rerank-9" {
		t.Fatalf("last_search_rerank_provider_id=%v", v)
	}
}

func TestSearchNodeExecutor_RequiresKB(t *testing.T) {
	ex := &SearchNodeExecutor{Knowledge: &mockKBSearcher{}}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"query": "x"})
	ctx.UserID = "u1"
	_, err := ex.Execute(ctx, json.RawMessage(`{"query":"hi"}`))
	if err == nil {
		t.Fatal("expected error for missing knowledge_base_id")
	}
}
