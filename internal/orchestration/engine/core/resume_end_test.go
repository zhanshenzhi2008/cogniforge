package core

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 模拟 wait → resume 后跑 llm → end，确保 end 能跑完且结果为 completed
func TestExecute_ResumeAfterWait_ReachesEnd(t *testing.T) {
	engine := NewEngine()
	engine.RegisterExecutor("start", &fakeNode{out: map[string]any{"started": true}})
	engine.RegisterExecutor("wait", &fakeNode{out: map[string]any{
		"__wait__": true,
		"reason":   "approval",
		"message":  "please approve",
	}})
	engine.RegisterExecutor("llm", &fakeNode{out: map[string]any{"text": "ok"}})
	engine.RegisterExecutor("end", &fakeNode{out: map[string]any{"finished": true, "output_keys": 2}})

	def := &WorkflowDefinition{
		Nodes: []NodeDefinition{
			{ID: "start", Type: "start", Name: "Start"},
			{ID: "wait", Type: "wait", Name: "Wait"},
			{ID: "llm", Type: "llm", Name: "Notify"},
			{ID: "end", Type: "end", Name: "End"},
		},
		Edges: []EdgeDefinition{
			{ID: "e1", Source: "start", Target: "wait"},
			{ID: "e2", Source: "wait", Target: "llm"},
			{ID: "e3", Source: "llm", Target: "end"},
		},
	}

	ctx1 := NewExecutionContext("wf1", "exec1", map[string]any{})
	r1 := engine.Execute(ctx1, def)
	require.Equal(t, "waiting", r1.Status)

	// 模拟 ApplyCheckpoint：已完成 start/wait 标 skipped
	ctx2 := NewExecutionContext("wf1", "exec1", map[string]any{})
	ctx2.SetNodeState("start", "skipped")
	ctx2.SetNodeState("wait", "skipped")
	if ctx1.Output != nil {
		ctx2.Output = ctx1.Output
	}
	r2 := engine.Execute(ctx2, def)
	require.Equal(t, "completed", r2.Status, "resume after wait must finish at end")
	assert.Equal(t, "skipped", ctx2.GetNodeState("start"))
	assert.Equal(t, "succeeded", ctx2.GetNodeState("llm"))
	assert.Equal(t, "succeeded", ctx2.GetNodeState("end"))
}

type fakeNode struct {
	out any
	err error
}

func (f *fakeNode) Execute(ctx *ExecutionContext, config json.RawMessage) (any, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.out, nil
}
