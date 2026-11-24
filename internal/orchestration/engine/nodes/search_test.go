package nodes

import (
	"context"
	"encoding/json"
	"testing"

	"cogniforge/internal/capability/knowledge"
	"cogniforge/internal/orchestration/engine/core"
)

type mockKBSearcher struct {
	lastQuery string
	lastKB    string
}

func (m *mockKBSearcher) SearchByUser(_ context.Context, _, kbID string, req *knowledge.SearchRequest) (*knowledge.SearchResponse, error) {
	m.lastKB = kbID
	m.lastQuery = req.Query
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
	cfg, _ := json.Marshal(SearchNodeConfig{
		KnowledgeBaseID: "kb-1",
		Query:           "{{query}}",
		TopK:            3,
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
	if v, _ := ctx.GetVariable("last_search_total"); v != 1 {
		t.Fatalf("last_search_total=%v", v)
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
