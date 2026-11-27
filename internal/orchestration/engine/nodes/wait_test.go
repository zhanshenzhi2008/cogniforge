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
}
