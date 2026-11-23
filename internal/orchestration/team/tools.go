package team

import (
	"encoding/json"
	"fmt"
	"strings"

	"cogniforge/internal/runtime/chat"
)

const (
	ToolDelegate = "delegate_to_agent"
	ToolHandoff  = "handoff_to_agent"
)

// OrchestrationToolSpecs 主管可用的编排原语（不是 MCP）
func OrchestrationToolSpecs() []chat.ToolCallSpec {
	return []chat.ToolCallSpec{
		{
			Type: "function",
			Function: chat.ToolFunctionSpec{
				Name:        ToolDelegate,
				Description: "把子任务委派给小队中的工人 Agent，等待其完成后拿回结果（你仍是主管，继续汇总）。",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"agent_id": map[string]any{
							"type":        "string",
							"description": "工人 Agent ID（必须在小队成员列表中）",
						},
						"task": map[string]any{
							"type":        "string",
							"description": "交给工人的具体任务说明",
						},
					},
					"required": []string{"agent_id", "task"},
				},
			},
		},
		{
			Type: "function",
			Function: chat.ToolFunctionSpec{
				Name:        ToolHandoff,
				Description: "把对话权交给另一名小队成员（交接棒）。之后应由对方继续对用户发言。",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"agent_id": map[string]any{
							"type":        "string",
							"description": "接手的 Agent ID",
						},
						"reason": map[string]any{
							"type":        "string",
							"description": "交接原因",
						},
					},
					"required": []string{"agent_id"},
				},
			},
		},
	}
}

func parseToolArgs(raw string) (map[string]any, error) {
	args := map[string]any{}
	if strings.TrimSpace(raw) == "" {
		return args, nil
	}
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		return nil, fmt.Errorf("invalid tool arguments: %w", err)
	}
	if args == nil {
		args = map[string]any{}
	}
	return args, nil
}

func argString(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	v, ok := args[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}
