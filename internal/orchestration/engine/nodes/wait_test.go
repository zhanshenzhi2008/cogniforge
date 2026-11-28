package nodes

import (
	"encoding/json"
	"testing"

	"cogniforge/internal/orchestration/engine/core"
)

func TestWaitNodeExecutor_ResolvesMessageTemplate(t *testing.T) {
	ex := &WaitNodeExecutor{}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"order_id": "ORD-9"})
	ctx.SetVariable("who", "柜员")
	cfg, _ := json.Marshal(WaitNodeConfig{
		Reason:  "approval",
		Message: "请{{who}}审批订单 {{order_id}}",
	})
	out, err := ex.Execute(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := out.(map[string]any)
	if m["message"] != "请柜员审批订单 ORD-9" {
		t.Fatalf("message=%v", m["message"])
	}
	if m["reason"] != "approval" {
		t.Fatalf("reason=%v", m["reason"])
	}
	if m["__wait__"] != true {
		t.Fatalf("out=%v", out)
	}
	if got, ok := ctx.GetVariable("last_wait_reason"); !ok || got != "approval" {
		t.Fatalf("last_wait_reason=%v ok=%v", got, ok)
	}
	if got, ok := ctx.GetVariable("last_wait_message"); !ok || got != "请柜员审批订单 ORD-9" {
		t.Fatalf("last_wait_message=%v ok=%v", got, ok)
	}
	if got, ok := ctx.GetVariable("last_wait_timeout_sec"); !ok || got != 0 {
		t.Fatalf("last_wait_timeout_sec=%v ok=%v", got, ok)
	}
}

func TestWaitNodeExecutor_WritesTimeoutVar(t *testing.T) {
	ex := &WaitNodeExecutor{}
	ctx := core.NewExecutionContext("wf", "exec", nil)
	cfg, _ := json.Marshal(map[string]any{"reason": "external", "timeout_sec": 120})
	out, err := ex.Execute(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := out.(map[string]any)
	if m["timeout_sec"] != 120 {
		t.Fatalf("timeout_sec=%v", m["timeout_sec"])
	}
	if got, ok := ctx.GetVariable("last_wait_timeout_sec"); !ok || got != 120 {
		t.Fatalf("last_wait_timeout_sec=%v ok=%v", got, ok)
	}
	if got, ok := ctx.GetVariable("last_wait_reason"); !ok || got != "external" {
		t.Fatalf("last_wait_reason=%v ok=%v", got, ok)
	}
}

func TestWaitNodeExecutor_TimeoutTemplate(t *testing.T) {
	ex := &WaitNodeExecutor{}
	ctx := core.NewExecutionContext("wf", "exec", nil)
	ctx.SetVariable("ttl", "90")
	cfg, _ := json.Marshal(map[string]any{
		"reason":      "approval",
		"timeout_sec": "{{ttl}}",
	})
	out, err := ex.Execute(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := out.(map[string]any)
	if m["timeout_sec"] != 90 {
		t.Fatalf("timeout_sec=%v", m["timeout_sec"])
	}
}
