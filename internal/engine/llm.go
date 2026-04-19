package engine

import (
	"encoding/json"
	"fmt"
	"time"
)

type LLMNodeConfig struct {
	Model        string            `json:"model"`
	Prompt       string            `json:"prompt"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
	Temperature  float64           `json:"temperature,omitempty"`
	MaxTokens    int               `json:"max_tokens,omitempty"`
	Variables    map[string]string `json:"variables,omitempty"`
}

type LLMNodeExecutor struct{}

func (e *LLMNodeExecutor) Execute(ctx *ExecutionContext, config json.RawMessage) (any, error) {
	var cfg LLMNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid LLM config: %w", err)
	}

	prompt := e.resolveVariables(cfg.Prompt, ctx)

	ctx.Logger.Logf("llm", "info", "Calling LLM: model=%s, prompt_length=%d", cfg.Model, len(prompt))

	result := map[string]any{
		"model":       cfg.Model,
		"prompt":      prompt,
		"response":    fmt.Sprintf("[Simulated LLM response for: %s]", truncateString(prompt, 50)),
		"tokens_used": len(prompt) / 4,
		"latency_ms":  150,
	}

	time.Sleep(100 * time.Millisecond)

	return result, nil
}

func (e *LLMNodeExecutor) resolveVariables(template string, ctx *ExecutionContext) string {
	for key, val := range ctx.Variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		template = replaceAll(template, placeholder, fmt.Sprintf("%v", val))
	}
	return template
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func replaceAll(s, old, new string) string {
	result := s
	for {
		idx := -1
		for i := 0; i <= len(result)-len(old); i++ {
			if result[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx == -1 {
			break
		}
		result = result[:idx] + new + result[idx+len(old):]
	}
	return result
}
