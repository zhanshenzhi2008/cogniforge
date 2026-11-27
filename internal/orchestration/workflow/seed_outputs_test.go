package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"cogniforge/internal/model"
	"cogniforge/internal/orchestration/engine"
)

func TestPickSeedOutputs_PrefersNonEmpty(t *testing.T) {
	runs := []model.WorkflowNodeRun{
		{NodeID: "a", NodeType: "llm", Output: model.JSONBMap{"text": "old"}},
		{NodeID: "a", NodeType: "llm", Output: model.JSONBMap{"text": "new"}},
		{NodeID: "b", NodeType: "code", Output: nil},
		{NodeID: "c", NodeType: "human_action", Output: model.JSONBMap{"ok": true}},
		{NodeID: "d", NodeType: "http", Output: model.JSONBMap{"status": 200}},
	}
	got := pickSeedOutputs(runs, []string{"a", "b", "c", "d"})
	assert.Equal(t, "new", got["a"]["text"])
	_, hasB := got["b"]
	assert.False(t, hasB)
	_, hasC := got["c"]
	assert.False(t, hasC)
	assert.Equal(t, 200, got["d"]["status"])
	assert.Len(t, got, 2)
}

func TestApplySeedOutputsToContext(t *testing.T) {
	ctx := engine.NewExecutionContext("wf", "ex", nil)
	n := applySeedOutputsToContext(ctx, map[string]map[string]any{
		"n1": {"v": 1},
	})
	assert.Equal(t, 1, n)
	v, ok := ctx.LookupOutput("n1")
	assert.True(t, ok)
	m := v.(map[string]any)
	assert.Equal(t, 1, m["v"])
}
