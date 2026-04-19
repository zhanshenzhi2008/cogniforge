package engine

import (
	"encoding/json"
	"fmt"
	"time"
)

type AgentNodeConfig struct {
	AgentID     string            `json:"agent_id"`
	Prompt      string            `json:"prompt,omitempty"`
	Tools       []string          `json:"tools,omitempty"`
	MemoryTurns int               `json:"memory_turns,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
}

type AgentNodeExecutor struct{}

func (e *AgentNodeExecutor) Execute(ctx *ExecutionContext, config json.RawMessage) (any, error) {
	var cfg AgentNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid agent config: %w", err)
	}

	prompt := e.resolveVariables(cfg.Prompt, ctx)

	ctx.Logger.Logf(ctx.NodeID, "info", "Calling Agent: agent_id=%s, prompt_length=%d", cfg.AgentID, len(prompt))

	result := map[string]any{
		"agent_id":     cfg.AgentID,
		"prompt":       prompt,
		"response":     fmt.Sprintf("[Simulated Agent response for: %s]", truncateString(prompt, 50)),
		"tools_used":   cfg.Tools,
		"memory_turns": cfg.MemoryTurns,
		"latency_ms":   200,
	}

	time.Sleep(150 * time.Millisecond)

	return result, nil
}

func (e *AgentNodeExecutor) resolveVariables(template string, ctx *ExecutionContext) string {
	for key, val := range ctx.Variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		template = replaceAll(template, placeholder, fmt.Sprintf("%v", val))
	}
	return template
}
