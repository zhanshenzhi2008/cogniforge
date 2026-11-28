package core

import (
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingNode struct {
	ran *int32
	out any
}

func (c *countingNode) Execute(ctx *ExecutionContext, config json.RawMessage) (any, error) {
	atomic.AddInt32(c.ran, 1)
	if c.out != nil {
		return c.out, nil
	}
	return map[string]any{"node": ctx.NodeID}, nil
}

func TestChosenConditionHandle(t *testing.T) {
	assert.Equal(t, "s-top", chosenConditionHandle(map[string]any{"result": true}))
	assert.Equal(t, "s-bottom", chosenConditionHandle(map[string]any{"result": false}))
	assert.Equal(t, "s-bottom", chosenConditionHandle(map[string]any{"branch": "default"}))
	assert.Equal(t, "s-top", chosenConditionHandle(map[string]any{"branch": "true"}))
}

func TestExecute_ConditionExclusive_TrueBranch(t *testing.T) {
	var ranTrue, ranFalse int32
	engine := NewEngine()
	engine.RegisterExecutor("start", &fakeNode{out: map[string]any{"started": true}})
	engine.RegisterExecutor("condition", &fakeNode{out: map[string]any{"result": true, "branch": "true"}})
	engine.RegisterExecutor("llm", &countingNode{ran: &ranTrue})
	engine.RegisterExecutor("code", &countingNode{ran: &ranFalse})
	engine.RegisterExecutor("end", &fakeNode{out: map[string]any{"finished": true}})

	def := &WorkflowDefinition{
		Nodes: []NodeDefinition{
			{ID: "start", Type: "start", Name: "Start"},
			{ID: "cond", Type: "condition", Name: "Cond"},
			{ID: "yes", Type: "llm", Name: "Yes"},
			{ID: "no", Type: "code", Name: "No"},
			{ID: "end", Type: "end", Name: "End"},
		},
		Edges: []EdgeDefinition{
			{ID: "e1", Source: "start", Target: "cond"},
			{ID: "e2", Source: "cond", Target: "yes", SourceHandle: "s-top"},
			{ID: "e3", Source: "cond", Target: "no", SourceHandle: "s-bottom"},
			{ID: "e4", Source: "yes", Target: "end"},
			{ID: "e5", Source: "no", Target: "end"},
		},
	}

	ctx := NewExecutionContext("wf-cond", "exec-t", nil)
	res := engine.Execute(ctx, def)
	require.Equal(t, "completed", res.Status)
	assert.Equal(t, int32(1), atomic.LoadInt32(&ranTrue))
	assert.Equal(t, int32(0), atomic.LoadInt32(&ranFalse))
	assert.Equal(t, "succeeded", ctx.GetNodeState("yes"))
	assert.Equal(t, "skipped", ctx.GetNodeState("no"))
	assert.Equal(t, "succeeded", ctx.GetNodeState("end"))
}

func TestExecute_ConditionExclusive_FalseBranch(t *testing.T) {
	var ranTrue, ranFalse int32
	engine := NewEngine()
	engine.RegisterExecutor("start", &fakeNode{out: map[string]any{"started": true}})
	engine.RegisterExecutor("condition", &fakeNode{out: map[string]any{"result": false, "branch": "default"}})
	engine.RegisterExecutor("llm", &countingNode{ran: &ranTrue})
	engine.RegisterExecutor("code", &countingNode{ran: &ranFalse})
	engine.RegisterExecutor("end", &fakeNode{out: map[string]any{"finished": true}})

	def := &WorkflowDefinition{
		Nodes: []NodeDefinition{
			{ID: "start", Type: "start", Name: "Start"},
			{ID: "cond", Type: "condition", Name: "Cond"},
			{ID: "yes", Type: "llm", Name: "Yes"},
			{ID: "no", Type: "code", Name: "No"},
			{ID: "end", Type: "end", Name: "End"},
		},
		Edges: []EdgeDefinition{
			{ID: "e1", Source: "start", Target: "cond"},
			{ID: "e2", Source: "cond", Target: "yes", SourceHandle: "s-top"},
			{ID: "e3", Source: "cond", Target: "no", SourceHandle: "s-bottom"},
			{ID: "e4", Source: "yes", Target: "end"},
			{ID: "e5", Source: "no", Target: "end"},
		},
	}

	ctx := NewExecutionContext("wf-cond", "exec-f", nil)
	res := engine.Execute(ctx, def)
	require.Equal(t, "completed", res.Status)
	assert.Equal(t, int32(0), atomic.LoadInt32(&ranTrue))
	assert.Equal(t, int32(1), atomic.LoadInt32(&ranFalse))
	assert.Equal(t, "skipped", ctx.GetNodeState("yes"))
	assert.Equal(t, "succeeded", ctx.GetNodeState("no"))
	assert.Equal(t, "succeeded", ctx.GetNodeState("end"))
}

func TestExecute_Condition_NoHandles_BothBranches(t *testing.T) {
	var ranA, ranB int32
	engine := NewEngine()
	engine.RegisterExecutor("start", &fakeNode{out: map[string]any{"started": true}})
	engine.RegisterExecutor("condition", &fakeNode{out: map[string]any{"result": true}})
	engine.RegisterExecutor("llm", &countingNode{ran: &ranA})
	engine.RegisterExecutor("code", &countingNode{ran: &ranB})
	engine.RegisterExecutor("end", &fakeNode{out: map[string]any{"finished": true}})

	def := &WorkflowDefinition{
		Nodes: []NodeDefinition{
			{ID: "start", Type: "start", Name: "Start"},
			{ID: "cond", Type: "condition", Name: "Cond"},
			{ID: "a", Type: "llm", Name: "A"},
			{ID: "b", Type: "code", Name: "B"},
			{ID: "end", Type: "end", Name: "End"},
		},
		Edges: []EdgeDefinition{
			{ID: "e1", Source: "start", Target: "cond"},
			{ID: "e2", Source: "cond", Target: "a"},
			{ID: "e3", Source: "cond", Target: "b"},
			{ID: "e4", Source: "a", Target: "end"},
			{ID: "e5", Source: "b", Target: "end"},
		},
	}

	ctx := NewExecutionContext("wf-cond", "exec-both", nil)
	res := engine.Execute(ctx, def)
	require.Equal(t, "completed", res.Status)
	assert.Equal(t, int32(1), atomic.LoadInt32(&ranA))
	assert.Equal(t, int32(1), atomic.LoadInt32(&ranB))
}

func TestParseDefinition_SourceHandleCamelCase(t *testing.T) {
	eng := NewEngine()
	definition := `{
		"nodes": [
			{"id": "c", "type": "condition", "name": "C", "config": {}},
			{"id": "a", "type": "llm", "name": "A", "config": {}},
			{"id": "b", "type": "end", "name": "B", "config": {}}
		],
		"edges": [
			{"id": "e1", "source": "c", "target": "a", "sourceHandle": "s-top"},
			{"id": "e2", "source": "c", "target": "b", "source_handle": "s-bottom"}
		]
	}`
	def, err := eng.ParseDefinition(definition)
	require.NoError(t, err)
	require.Len(t, def.Edges, 2)
	assert.Equal(t, "s-top", def.Edges[0].SourceHandle)
	assert.Equal(t, "s-bottom", def.Edges[1].SourceHandle)
}
