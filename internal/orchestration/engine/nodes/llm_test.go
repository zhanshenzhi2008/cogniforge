package nodes

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"cogniforge/internal/orchestration/engine/core"
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

func TestDefaultLLMUserPrompt_FromInput(t *testing.T) {
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{
		"order_id": "ORD-1",
		"amount":   99,
	})
	out := defaultLLMUserPrompt(ctx)
	require.Contains(t, out, "ORD-1")
	require.Contains(t, out, "input")
}

func TestResolveLLMUserPrompt_PrefersLastAgentContent(t *testing.T) {
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"order_id": "ORD-1"})
	ctx.SetVariable("last_agent_content", "上游摘要")
	out := resolveLLMUserPrompt("", ctx)
	require.Equal(t, "上游摘要", out)

	out = resolveLLMUserPrompt("固定 {{order_id}}", ctx)
	require.Equal(t, "固定 ORD-1", out)
}

func TestResolveLLMModelTemplate(t *testing.T) {
	ctx := core.NewExecutionContext("wf", "exec", nil)
	ctx.SetVariable("model_id", "deepseek-chat")
	require.Equal(t, "deepseek-chat", strings.TrimSpace(resolveTemplate("{{model_id}}", ctx)))
	require.Equal(t, "fixed-model", strings.TrimSpace(resolveTemplate("fixed-model", ctx)))
}

func TestLLMNodeConfig_TemperatureMaxTokensTemplate(t *testing.T) {
	ctx := core.NewExecutionContext("wf", "exec", nil)
	ctx.SetVariable("temp", "0.7")
	ctx.SetVariable("cap", "512")
	raw := []byte(`{"model":"m","prompt":"hi","temperature":"{{temp}}","max_tokens":"{{cap}}"}`)
	var cfg LLMNodeConfig
	require.NoError(t, json.Unmarshal(raw, &cfg))
	require.InDelta(t, 0.7, resolveFlexibleFloat(cfg.Temperature, ctx), 1e-9)
	require.Equal(t, 512, resolveFlexibleInt(cfg.MaxTokens, ctx))

	rawNum := []byte(`{"model":"m","prompt":"hi","temperature":0.3,"max_tokens":128}`)
	require.NoError(t, json.Unmarshal(rawNum, &cfg))
	require.InDelta(t, 0.3, resolveFlexibleFloat(cfg.Temperature, ctx), 1e-9)
	require.Equal(t, 128, resolveFlexibleInt(cfg.MaxTokens, ctx))
}

func TestLLMNodeExecutor_SystemPromptOnly_UsesFallbackUser(t *testing.T) {
	e := &LLMNodeExecutor{}
	cfg, _ := json.Marshal(LLMNodeConfig{SystemPrompt: "你是助手", Model: "m"})
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"q": "hi"}).WithUserID("u1")
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
