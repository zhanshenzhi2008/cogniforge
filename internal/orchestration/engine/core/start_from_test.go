package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyStartFrom_SkipsAncestorsKeepsStart(t *testing.T) {
	def := &WorkflowDefinition{
		Nodes: []NodeDefinition{
			{ID: "start", Type: "start", Name: "Start"},
			{ID: "a", Type: "llm", Name: "A"},
			{ID: "b", Type: "llm", Name: "B"},
			{ID: "end", Type: "end", Name: "End"},
		},
		Edges: []EdgeDefinition{
			{Source: "start", Target: "a"},
			{Source: "a", Target: "b"},
			{Source: "b", Target: "end"},
		},
	}
	ctx := NewExecutionContext("wf", "ex", map[string]any{})
	skipped, err := ApplyStartFrom(ctx, def, "b")
	require.NoError(t, err)
	assert.Equal(t, "", ctx.GetNodeState("start"), "start must still run")
	assert.Equal(t, "skipped", ctx.GetNodeState("a"))
	assert.Equal(t, "", ctx.GetNodeState("b"))
	assert.Equal(t, "", ctx.GetNodeState("end"))
	require.Len(t, skipped, 1)
	assert.Equal(t, "a", skipped[0].ID)
}

func TestApplyStartFrom_MissingNode(t *testing.T) {
	def := &WorkflowDefinition{
		Nodes: []NodeDefinition{{ID: "start", Type: "start"}},
	}
	ctx := NewExecutionContext("wf", "ex", nil)
	_, err := ApplyStartFrom(ctx, def, "nope")
	require.Error(t, err)
}

func TestApplyStartFrom_EmptyIsNoop(t *testing.T) {
	def := &WorkflowDefinition{Nodes: []NodeDefinition{{ID: "start", Type: "start"}}}
	ctx := NewExecutionContext("wf", "ex", nil)
	skipped, err := ApplyStartFrom(ctx, def, "")
	require.NoError(t, err)
	assert.Nil(t, skipped)
}
