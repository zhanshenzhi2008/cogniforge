package engine_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"cogniforge/internal/engine"
)

func TestConditionNode_EqualOperator(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}

	config := `{
		"conditions": [
			{
				"field": "status",
				"operator": "==",
				"value": "active",
				"branch": "active_branch"
			}
		]
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{"status": "active"})
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, true, result["result"])
	assert.Equal(t, "active_branch", result["branch"])
}

func TestConditionNode_NotEqual(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}

	config := `{
		"conditions": [
			{
				"field": "status",
				"operator": "!=",
				"value": "inactive",
				"branch": "not_inactive"
			}
		]
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{"status": "active"})
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, true, result["result"])
}

func TestConditionNode_GreaterThan(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}

	tests := []struct {
		name     string
		config   string
		expected bool
	}{
		{"7 > 5 is true", `{"conditions":[{"field":"count","operator":">","value":5,"branch":"test"}]}`, true},
		{"7 > 10 is false", `{"conditions":[{"field":"count","operator":">","value":10,"branch":"test"}]}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{"count": float64(7)})
			output, _ := executor.Execute(ctx, json.RawMessage(tt.config))
			result := output.(map[string]any)
			assert.Equal(t, tt.expected, result["result"])
		})
	}
}

func TestConditionNode_LessThan(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}

	tests := []struct {
		name     string
		config   string
		expected bool
	}{
		{"7 < 10 is true", `{"conditions":[{"field":"count","operator":"<","value":10,"branch":"test"}]}`, true},
		{"7 < 5 is false", `{"conditions":[{"field":"count","operator":"<","value":5,"branch":"test"}]}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{"count": float64(7)})
			output, _ := executor.Execute(ctx, json.RawMessage(tt.config))
			result := output.(map[string]any)
			assert.Equal(t, tt.expected, result["result"])
		})
	}
}

func TestConditionNode_Contains(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}

	config := `{
		"conditions": [
			{
				"field": "message",
				"operator": "contains",
				"value": "error",
				"branch": "error_branch"
			}
		]
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{"message": "This is an error message"})
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, true, result["result"])
	assert.Equal(t, "error_branch", result["branch"])
}

func TestConditionNode_StartsWith(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}

	tests := []struct {
		name     string
		typeVal  string
		expected bool
	}{
		{"starts with admin", "admin_user", true},
		{"does not start with admin", "user_admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := `{"conditions":[{"field":"type","operator":"starts_with","value":"admin","branch":"admin_branch"}]}`
			ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{"type": tt.typeVal})
			output, _ := executor.Execute(ctx, json.RawMessage(config))
			result := output.(map[string]any)
			assert.Equal(t, tt.expected, result["result"])
		})
	}
}

func TestConditionNode_IsEmpty(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}
	config := `{"conditions":[{"field":"name","operator":"is_empty","value":"","branch":"empty"}]}`

	tests := []struct {
		name     string
		input    map[string]any
		expected bool
	}{
		{"nil value", map[string]any{"name": nil}, true},
		{"empty string", map[string]any{"name": ""}, true},
		{"non-empty string", map[string]any{"name": "John"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewExecutionContext("wf1", "exec1", tt.input)
			output, _ := executor.Execute(ctx, json.RawMessage(config))
			result := output.(map[string]any)
			assert.Equal(t, tt.expected, result["result"])
		})
	}
}

func TestConditionNode_IsNotEmpty(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}
	config := `{"conditions":[{"field":"email","operator":"is_not_empty","value":"","branch":"has_email"}]}`

	tests := []struct {
		name     string
		input    map[string]any
		expected bool
	}{
		{"has value", map[string]any{"email": "test@example.com"}, true},
		{"empty string", map[string]any{"email": ""}, false},
		{"nil value", map[string]any{"email": nil}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewExecutionContext("wf1", "exec1", tt.input)
			output, _ := executor.Execute(ctx, json.RawMessage(config))
			result := output.(map[string]any)
			assert.Equal(t, tt.expected, result["result"])
		})
	}
}

func TestConditionNode_MatchesRegex(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}
	config := `{"conditions":[{"field":"email","operator":"matches","value":"^[a-z]+@[a-z]+\\.com$","branch":"valid_email"}]}`

	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid email", "test@example.com", true},
		{"invalid domain", "test@example.org", false},
		{"invalid format", "not-an-email", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{"email": tt.email})
			output, _ := executor.Execute(ctx, json.RawMessage(config))
			result := output.(map[string]any)
			assert.Equal(t, tt.expected, result["result"])
		})
	}
}

func TestConditionNode_NoConditions(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}

	config := `{"conditions": []}`

	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, true, result["result"])
	assert.Equal(t, "default", result["branch"])
}

func TestConditionNode_FirstMatchingWins(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}

	config := `{
		"conditions": [
			{"field": "level", "operator": "==", "value": "debug", "branch": "debug"},
			{"field": "level", "operator": "==", "value": "info", "branch": "info"},
			{"field": "level", "operator": "==", "value": "error", "branch": "error"}
		]
	}`

	ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{"level": "info"})
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, "info", result["branch"])
}

func TestConditionNode_InvalidConfig(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}

	config := `invalid json`

	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	_, err := executor.Execute(ctx, json.RawMessage(config))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid condition config")
}

func TestConditionNode_VariablePriority(t *testing.T) {
	executor := &engine.ConditionNodeExecutor{}
	config := `{"conditions":[{"field":"status","operator":"==","value":"active","branch":"active"}]}`

	// Test that variables set via SetVariable are used
	ctx := engine.NewExecutionContext("wf1", "exec1", nil)
	ctx.SetVariable("status", "active")
	output, err := executor.Execute(ctx, json.RawMessage(config))

	assert.NoError(t, err)
	result := output.(map[string]any)
	assert.Equal(t, true, result["result"])
}
