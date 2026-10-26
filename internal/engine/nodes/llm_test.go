package nodes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"cogniforge/internal/engine/core"
)

func TestResolveTemplate_InputAndVariables(t *testing.T) {
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{
		"brand": "CogniForge",
	})
	ctx.SetVariable("tone", "friendly")
	out := resolveTemplate("Hello {{brand}}, be {{tone}}", ctx)
	require.Equal(t, "Hello CogniForge, be friendly", out)
}

func TestLLMNodeExecutor_RequiresChat(t *testing.T) {
	e := &LLMNodeExecutor{}
	cfg, _ := json.Marshal(LLMNodeConfig{Prompt: "hi"})
	ctx := core.NewExecutionContext("wf", "exec", nil).WithUserID("u1")
	_, err := e.Execute(ctx, cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ChatService not configured")
}

func TestAgentNodeExecutor_RequiresAgentID(t *testing.T) {
	e := &AgentNodeExecutor{Chat: nil, DB: nil}
	cfg, _ := json.Marshal(AgentNodeConfig{})
	ctx := core.NewExecutionContext("wf", "exec", nil).WithUserID("u1")
	_, err := e.Execute(ctx, cfg)
	require.Error(t, err)
}
