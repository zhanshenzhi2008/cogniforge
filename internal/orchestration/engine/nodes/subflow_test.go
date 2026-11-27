package nodes

import (
	"encoding/json"
	"testing"

	"cogniforge/internal/orchestration/engine/core"
)

func TestSubflowNodeExecutor_RequiresWorkflowID(t *testing.T) {
	ex := &SubflowNodeExecutor{Run: func(parent *core.ExecutionContext, workflowID string, input map[string]any) (map[string]any, error) {
		t.Fatal("should not run")
		return nil, nil
	}}
	ctx := core.NewExecutionContext("wf-parent", "exec-1", map[string]any{"a": 1}).WithUserID("u1")
	_, err := ex.Execute(ctx, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error for missing workflow_id")
	}
}

func TestSubflowNodeExecutor_PassesVariables(t *testing.T) {
	var gotID string
	var gotIn map[string]any
	ex := &SubflowNodeExecutor{Run: func(parent *core.ExecutionContext, workflowID string, input map[string]any) (map[string]any, error) {
		gotID = workflowID
		gotIn = input
		return map[string]any{"ok": true}, nil
	}}
	ctx := core.NewExecutionContext("wf-parent", "exec-1", map[string]any{"q": "hi"}).WithUserID("u1")
	ctx.SetVariable("x", 42)
	out, err := ex.Execute(ctx, json.RawMessage(`{"workflow_id":"wf-child"}`))
	if err != nil {
		t.Fatal(err)
	}
	if gotID != "wf-child" {
		t.Fatalf("workflow_id=%s", gotID)
	}
	if gotIn["x"] != 42 || gotIn["q"] != "hi" {
		t.Fatalf("input=%v", gotIn)
	}
	m, ok := out.(map[string]any)
	if !ok || m["status"] != "completed" {
		t.Fatalf("out=%v", out)
	}
}

func TestSubflowNodeExecutor_CustomInputTemplates(t *testing.T) {
	var gotIn map[string]any
	ex := &SubflowNodeExecutor{Run: func(parent *core.ExecutionContext, workflowID string, input map[string]any) (map[string]any, error) {
		gotIn = input
		return map[string]any{}, nil
	}}
	ctx := core.NewExecutionContext("wf-parent", "exec-1", nil).WithUserID("u1")
	ctx.SetVariable("name", "Ada")
	_, err := ex.Execute(ctx, json.RawMessage(`{"workflow_id":"wf-child","input":{"hello":"Hi {{name}}"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if gotIn["hello"] != "Hi Ada" {
		t.Fatalf("input=%v", gotIn)
	}
}
