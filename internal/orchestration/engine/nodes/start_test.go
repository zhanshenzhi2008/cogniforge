package nodes

import (
	"encoding/json"
	"testing"

	"cogniforge/internal/orchestration/engine/core"
)

func TestStartNodeAppliesDefaultsAndRequires(t *testing.T) {
	ex := &StartNodeExecutor{}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{})
	cfg, _ := json.Marshal(StartNodeConfig{Variables: []StartVariable{
		{Name: "query", Type: "string", Required: true, Default: "你好"},
		{Name: "top_k", Type: "number", Default: 3},
	}})
	out, err := ex.Execute(ctx, cfg)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	m := out.(map[string]any)
	vars := m["variables"].(map[string]any)
	if vars["query"] != "你好" {
		t.Fatalf("query=%v", vars["query"])
	}
	if ctx.Input["query"] != "你好" {
		t.Fatalf("input query=%v", ctx.Input["query"])
	}
}

func TestStartNodeRequiredMissing(t *testing.T) {
	ex := &StartNodeExecutor{}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{})
	cfg, _ := json.Marshal(StartNodeConfig{Variables: []StartVariable{
		{Name: "query", Type: "string", Required: true},
	}})
	_, err := ex.Execute(ctx, cfg)
	if err == nil {
		t.Fatal("expected required error")
	}
}

func TestStartNodeKeepsProvidedInput(t *testing.T) {
	ex := &StartNodeExecutor{}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"query": "退款"})
	cfg, _ := json.Marshal(StartNodeConfig{Variables: []StartVariable{
		{Name: "query", Type: "string", Required: true, Default: "默认"},
	}})
	_, err := ex.Execute(ctx, cfg)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if ctx.Input["query"] != "退款" {
		t.Fatalf("got %v", ctx.Input["query"])
	}
}
