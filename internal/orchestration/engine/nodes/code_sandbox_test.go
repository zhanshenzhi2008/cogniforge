package nodes

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"cogniforge/internal/orchestration/engine/core"
)

func TestCodeNode_PythonSandbox(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err2 := exec.LookPath("python"); err2 != nil {
			t.Skip("python not installed")
		}
	}
	ex := &CodeNodeExecutor{}
	ctx := core.NewExecutionContext("wf", "ex1", map[string]any{"name": "Ada"}).WithUserID("u1")
	ctx.SetVariable("n", 2)
	cfg, _ := json.Marshal(map[string]any{
		"language": "python",
		"code":     "print(variables['n'] + 1)\nprint(input['name'])",
		"timeout":  10,
	})
	out, err := ex.Execute(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("out=%T", out)
	}
	text, _ := m["output"].(string)
	if !strings.Contains(text, "3") || !strings.Contains(text, "Ada") {
		t.Fatalf("output=%q", text)
	}
}

func TestWrapCodeWithContext_Python(t *testing.T) {
	ctx := core.NewExecutionContext("wf", "ex1", map[string]any{"q": "hi"})
	src, err := wrapCodeWithContext("python", "print(1)", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(src, "_cf_context") || !strings.Contains(src, "print(1)") {
		t.Fatalf("src=%s", src)
	}
}
