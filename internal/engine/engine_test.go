package engine_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"cogniforge/internal/engine"
)

func TestTopologicalSort_Basic(t *testing.T) {
	eng := engine.NewEngine()

	def := &engine.WorkflowDefinition{
		Nodes: []engine.NodeDefinition{
			{ID: "1", Type: "start", Name: "Start"},
			{ID: "2", Type: "llm", Name: "LLM"},
			{ID: "3", Type: "end", Name: "End"},
		},
		Edges: []engine.EdgeDefinition{
			{ID: "e1", Source: "1", Target: "2"},
			{ID: "e2", Source: "2", Target: "3"},
		},
	}

	nodes, err := eng.TopologicalSort(def)
	assert.NoError(t, err)
	assert.Len(t, nodes, 3)

	// Verify order: 1 -> 2 -> 3
	assert.Equal(t, "1", nodes[0].ID)
	assert.Equal(t, "2", nodes[1].ID)
	assert.Equal(t, "3", nodes[2].ID)
}

func TestTopologicalSort_MultipleRoots(t *testing.T) {
	eng := engine.NewEngine()

	def := &engine.WorkflowDefinition{
		Nodes: []engine.NodeDefinition{
			{ID: "1", Type: "start", Name: "Start1"},
			{ID: "2", Type: "start", Name: "Start2"},
			{ID: "3", Type: "end", Name: "End"},
		},
		Edges: []engine.EdgeDefinition{
			{ID: "e1", Source: "1", Target: "3"},
			{ID: "e2", Source: "2", Target: "3"},
		},
	}

	nodes, err := eng.TopologicalSort(def)
	assert.NoError(t, err)
	assert.Len(t, nodes, 3)

	// Both 1 and 2 should come before 3
	lastIdx := -1
	for i, node := range nodes {
		if node.ID == "3" {
			lastIdx = i
			break
		}
	}
	assert.Greater(t, lastIdx, 0)
}

func TestTopologicalSort_ComplexGraph(t *testing.T) {
	eng := engine.NewEngine()

	def := &engine.WorkflowDefinition{
		Nodes: []engine.NodeDefinition{
			{ID: "start", Type: "start", Name: "Start"},
			{ID: "cond", Type: "condition", Name: "Condition"},
			{ID: "task1", Type: "llm", Name: "Task 1"},
			{ID: "task2", Type: "llm", Name: "Task 2"},
			{ID: "end", Type: "end", Name: "End"},
		},
		Edges: []engine.EdgeDefinition{
			{ID: "e1", Source: "start", Target: "cond"},
			{ID: "e2", Source: "cond", Target: "task1"},
			{ID: "e3", Source: "cond", Target: "task2"},
			{ID: "e4", Source: "task1", Target: "end"},
			{ID: "e5", Source: "task2", Target: "end"},
		},
	}

	nodes, err := eng.TopologicalSort(def)
	assert.NoError(t, err)
	assert.Len(t, nodes, 5)

	// Start must be first
	assert.Equal(t, "start", nodes[0].ID)
	// Cond must come after start
	for i, node := range nodes {
		if node.ID == "cond" {
			assert.Greater(t, i, 0)
			assert.Less(t, i, 4)
		}
	}
}

func TestTopologicalSort_CycleDetection(t *testing.T) {
	eng := engine.NewEngine()

	def := &engine.WorkflowDefinition{
		Nodes: []engine.NodeDefinition{
			{ID: "1", Type: "start", Name: "Start"},
			{ID: "2", Type: "llm", Name: "LLM"},
			{ID: "3", Type: "end", Name: "End"},
		},
		Edges: []engine.EdgeDefinition{
			{ID: "e1", Source: "1", Target: "2"},
			{ID: "e2", Source: "2", Target: "1"}, // Creates cycle
		},
	}

	nodes, err := eng.TopologicalSort(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle")
	assert.Nil(t, nodes)
}

func TestParseDefinition(t *testing.T) {
	eng := engine.NewEngine()

	definition := `{
		"nodes": [
			{"id": "n1", "type": "start", "name": "Start", "config": {}},
			{"id": "n2", "type": "llm", "name": "LLM Node", "config": {"model": "gpt-4"}},
			{"id": "n3", "type": "end", "name": "End", "config": {}}
		],
		"edges": [
			{"id": "e1", "source": "n1", "target": "n2"},
			{"id": "e2", "source": "n2", "target": "n3"}
		]
	}`

	def, err := eng.ParseDefinition(definition)
	assert.NoError(t, err)
	assert.Len(t, def.Nodes, 3)
	assert.Len(t, def.Edges, 2)

	// Verify node types
	nodeMap := make(map[string]string)
	for _, n := range def.Nodes {
		nodeMap[n.ID] = n.Type
	}
	assert.Equal(t, "start", nodeMap["n1"])
	assert.Equal(t, "llm", nodeMap["n2"])
	assert.Equal(t, "end", nodeMap["n3"])
}

func TestParseDefinition_InvalidJSON(t *testing.T) {
	eng := engine.NewEngine()

	_, err := eng.ParseDefinition("invalid json")
	assert.Error(t, err)
}

func TestExecutionContext_Variables(t *testing.T) {
	ctx := engine.NewExecutionContext("wf1", "exec1", map[string]any{"input": "value"})

	// Test SetVariable and GetVariable
	ctx.SetVariable("key1", "value1")
	val, ok := ctx.GetVariable("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	// Test non-existent variable
	_, ok = ctx.GetVariable("nonexistent")
	assert.False(t, ok)
}

func TestExecutionContext_NodeState(t *testing.T) {
	ctx := engine.NewExecutionContext("wf1", "exec1", nil)

	// Test SetNodeState and GetNodeState
	ctx.SetNodeState("node1", "completed")
	state := ctx.GetNodeState("node1")
	assert.Equal(t, "completed", state)

	// Test non-existent node state
	state = ctx.GetNodeState("nonexistent")
	assert.Equal(t, "", state)
}

func TestExecutionLogger(t *testing.T) {
	logger := engine.NewExecutionLogger("exec1")

	// Test Log
	logger.Log("node1", "info", "Test message")
	assert.Len(t, logger.Logs, 1)
	assert.Equal(t, "node1", logger.Logs[0].NodeID)
	assert.Equal(t, "info", logger.Logs[0].Level)
	assert.Equal(t, "Test message", logger.Logs[0].Message)

	// Test Logf
	logger.Logf("node2", "error", "Error: %s", "something went wrong")
	assert.Len(t, logger.Logs, 2)
	assert.Equal(t, "node2", logger.Logs[1].NodeID)
	assert.Contains(t, logger.Logs[1].Message, "Error: something went wrong")
}

func TestWorkflowDefinition_NodeTypes(t *testing.T) {
	def := &engine.WorkflowDefinition{
		Nodes: []engine.NodeDefinition{
			{ID: "start", Type: "start", Name: "Start", Config: json.RawMessage(`{}`)},
			{ID: "llm", Type: "llm", Name: "LLM", Config: json.RawMessage(`{"model":"gpt-4"}`)},
			{ID: "condition", Type: "condition", Name: "Condition", Config: json.RawMessage(`{"conditions":[]}`)},
			{ID: "loop", Type: "loop", Name: "Loop", Config: json.RawMessage(`{"type":"times","count":3}`)},
			{ID: "http", Type: "http", Name: "HTTP", Config: json.RawMessage(`{"url":"https://api.example.com"}`)},
			{ID: "end", Type: "end", Name: "End", Config: json.RawMessage(`{}`)},
		},
		Edges: []engine.EdgeDefinition{
			{ID: "e1", Source: "start", Target: "llm"},
			{ID: "e2", Source: "llm", Target: "condition"},
			{ID: "e3", Source: "condition", Target: "loop"},
			{ID: "e4", Source: "loop", Target: "http"},
			{ID: "e5", Source: "http", Target: "end"},
		},
	}

	assert.Len(t, def.Nodes, 6)
	assert.Len(t, def.Edges, 5)
}
