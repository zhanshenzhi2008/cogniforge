package nodes

import (
	"encoding/json"
	"fmt"
	"testing"

	"cogniforge/internal/orchestration/engine/core"
)

func TestCodeNodeExecutor_ExpressionWritesLastCodeOutput(t *testing.T) {
	ex := &CodeNodeExecutor{}
	ctx := core.NewExecutionContext("wf", "exec", nil)
	ctx.SetVariable("score", 42)

	cfg, _ := json.Marshal(map[string]any{
		"language": "expression",
		"code":     "score",
	})
	out, err := ex.Execute(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if out != 42 {
		t.Fatalf("out=%v (%T)", out, out)
	}
	got, ok := ctx.GetVariable("last_code_output")
	if !ok || fmt.Sprint(got) != "42" {
		t.Fatalf("last_code_output=%v ok=%v", got, ok)
	}
}
