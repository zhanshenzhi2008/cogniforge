package core

import (
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type slowNode struct {
	delay   time.Duration
	running *int32
	max     *int32
}

func (s *slowNode) Execute(ctx *ExecutionContext, config json.RawMessage) (any, error) {
	cur := atomic.AddInt32(s.running, 1)
	for {
		old := atomic.LoadInt32(s.max)
		if cur <= old || atomic.CompareAndSwapInt32(s.max, old, cur) {
			break
		}
	}
	time.Sleep(s.delay)
	atomic.AddInt32(s.running, -1)
	return map[string]any{"ok": true, "node": ctx.NodeID}, nil
}

func TestExecute_ParallelWave_IndependentDelays(t *testing.T) {
	var running, max int32
	slow := &slowNode{delay: 80 * time.Millisecond, running: &running, max: &max}

	engine := NewEngine()
	engine.RegisterExecutor("start", &fakeNode{out: map[string]any{"started": true}})
	engine.RegisterExecutor("delay", slow)
	engine.RegisterExecutor("end", &fakeNode{out: map[string]any{"finished": true}})

	def := &WorkflowDefinition{
		Nodes: []NodeDefinition{
			{ID: "start", Type: "start", Name: "Start"},
			{ID: "a", Type: "delay", Name: "A"},
			{ID: "b", Type: "delay", Name: "B"},
			{ID: "end", Type: "end", Name: "End"},
		},
		Edges: []EdgeDefinition{
			{ID: "e1", Source: "start", Target: "a"},
			{ID: "e2", Source: "start", Target: "b"},
			{ID: "e3", Source: "a", Target: "end"},
			{ID: "e4", Source: "b", Target: "end"},
		},
	}

	ctx := NewExecutionContext("wf-par", "exec-par", nil)
	start := time.Now()
	res := engine.Execute(ctx, def)
	elapsed := time.Since(start)

	require.Equal(t, "completed", res.Status)
	assert.Equal(t, "succeeded", ctx.GetNodeState("a"))
	assert.Equal(t, "succeeded", ctx.GetNodeState("b"))
	assert.Equal(t, "succeeded", ctx.GetNodeState("end"))
	// 串行约 160ms+；并行应明显小于 150ms 且峰值并发≥2
	assert.Less(t, elapsed, 150*time.Millisecond, "independent delay nodes should overlap")
	assert.GreaterOrEqual(t, atomic.LoadInt32(&max), int32(2))
}
