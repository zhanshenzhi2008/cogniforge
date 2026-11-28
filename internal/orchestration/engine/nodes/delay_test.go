package nodes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"cogniforge/internal/orchestration/engine/core"
)

func TestResolveDelayDuration_NumberAndTemplate(t *testing.T) {
	ctx := core.NewExecutionContext("wf1", "exec1", nil)
	ctx.SetVariable("wait_sec", "3")

	assert.Equal(t, 2, resolveDelayDuration(json.RawMessage(`2`), ctx))
	assert.Equal(t, 3, resolveDelayDuration(json.RawMessage(`"{{wait_sec}}"`), ctx))
	assert.Equal(t, 1, resolveDelayDuration(json.RawMessage(`"{{missing}}"`), ctx))
	assert.Equal(t, 1, resolveDelayDuration(nil, ctx))
	assert.Equal(t, 0, resolveFlexibleInt(nil, ctx))
	assert.Equal(t, 3, resolveFlexibleInt(json.RawMessage(`"{{wait_sec}}"`), ctx))
}

func TestDelayNode_WritesLastDelayVars(t *testing.T) {
	executor := &DelayNodeExecutor{}
	config := `{"duration": 1, "unit": "seconds"}`
	ctx := core.NewExecutionContext("wf1", "exec1", nil)

	output, err := executor.Execute(ctx, json.RawMessage(config))
	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, true, result["delayed"])
	assert.Equal(t, 1, result["duration"])
	assert.Equal(t, int64(1000), result["actual_ms"])

	gotMs, ok := ctx.GetVariable("last_delay_ms")
	assert.True(t, ok)
	assert.Equal(t, int64(1000), gotMs)
	gotDur, ok := ctx.GetVariable("last_delay_duration")
	assert.True(t, ok)
	assert.Equal(t, 1, gotDur)
	gotUnit, ok := ctx.GetVariable("last_delay_unit")
	assert.True(t, ok)
	assert.Equal(t, "seconds", gotUnit)
}
