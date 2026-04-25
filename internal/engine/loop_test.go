package engine_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"cogniforge/internal/engine"
)

func TestLoopNode_TimesLoop(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	config := `{
		"type": "times",
		"count": 5,
		"max_iter": 100
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, "times", result["type"])
	assert.Equal(t, 5, result["total"])

	iterations := result["iterations"].([]map[string]any)
	assert.Len(t, iterations, 5)
}

func TestLoopNode_TimesLoop_MaxIterCapped(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	// Request 200 iterations but max is 100
	config := `{
		"type": "times",
		"count": 200,
		"max_iter": 100
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, 100, result["total"]) // Capped at max_iter

	iterations := result["iterations"].([]map[string]any)
	assert.Len(t, iterations, 100)
}

func TestLoopNode_DefaultMaxIter(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	// No max_iter specified, should default to 100
	config := `{
		"type": "times",
		"count": 50
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, 50, result["total"])
}

func TestLoopNode_ForEachLoop(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	config := `{
		"type": "for_each",
		"array_var": "items",
		"item_var": "item",
		"max_iter": 100
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{
		"items": []any{"a", "b", "c"},
	})
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, "for_each", result["type"])
	assert.Equal(t, 3, result["total"])

	iterations := result["iterations"].([]map[string]any)
	assert.Len(t, iterations, 3)
}

func TestLoopNode_ForEachLoop_EmptyArray(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	config := `{
		"type": "for_each",
		"array_var": "items",
		"max_iter": 100
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{
		"items": []any{},
	})
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, 0, result["total"])
}

func TestLoopNode_ForEachLoop_ArrayNotFound(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	config := `{
		"type": "for_each",
		"array_var": "nonexistent",
		"max_iter": 100
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, 0, result["total"]) // Empty iterations when variable not found
}

func TestLoopNode_ForEachLoop_MaxIterCapped(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	config := `{
		"type": "for_each",
		"array_var": "items",
		"max_iter": 2
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{
		"items": []any{1, 2, 3, 4, 5},
	})
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, 2, result["total"]) // Capped at max_iter

	iterations := result["iterations"].([]map[string]any)
	assert.Len(t, iterations, 2)
}

func TestLoopNode_WhileLoop(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	// While loop with condition loop.index < 3
	// Should run 3 times (index 0, 1, 2)
	config := `{
		"type": "while",
		"condition": "loop.index < 3",
		"max_iter": 10
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, "while", result["type"])
	assert.Equal(t, 3, result["total"])

	iterations := result["iterations"].([]map[string]any)
	assert.Len(t, iterations, 3)
}

func TestLoopNode_WhileLoop_MaxIterReached(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	// Condition is always true (loop.index >= 0), but should be capped by max_iter
	config := `{
		"type": "while",
		"condition": "loop.index >= 0",
		"max_iter": 5
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, 5, result["total"]) // Stopped at max_iter
}

func TestLoopNode_InvalidType(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	config := `{
		"type": "invalid",
		"count": 5
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	_, err := executor.Execute(ctx, json.RawMessage(config))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown loop type")
}

func TestLoopNode_InvalidConfig(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	config := `invalid json`

	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	_, err := executor.Execute(ctx, json.RawMessage(config))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid loop config")
}

func TestLoopNode_TimesLoop_IndexVariable(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	config := `{
		"type": "times",
		"count": 3,
		"max_iter": 100
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	_, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)

	// After execution, loop.index should be the last value
	idx, ok := ctx.GetVariable("loop.index")
	assert.True(t, ok)
	assert.Equal(t, 2, idx) // Last index (0-based)

	count, ok := ctx.GetVariable("loop.count")
	assert.True(t, ok)
	assert.Equal(t, 3, count)
}

func TestLoopNode_ForEachLoop_ItemVariable(t *testing.T) {
	executor := &engine.LoopNodeExecutor{}

	config := `{
		"type": "for_each",
		"array_var": "items",
		"item_var": "current_item",
		"key_var": "current_key",
		"max_iter": 100
	}`

	items := []any{"apple", "banana", "cherry"}
	ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{"items": items})
	_, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)

	// After execution, should have the last item
	item, ok := ctx.GetVariable("current_item")
	assert.True(t, ok)
	assert.Equal(t, "cherry", item)

	key, ok := ctx.GetVariable("current_key")
	assert.True(t, ok)
	assert.Equal(t, 2, key) // Last index
}
