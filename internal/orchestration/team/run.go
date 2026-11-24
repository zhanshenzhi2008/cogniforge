package team

import (
	"context"
	"fmt"
	"strings"

	"cogniforge/internal/model"
	"cogniforge/internal/runtime/chat"
)

// RunTask 任务型小队执行：单轮 user task，不落 chat_conversations（正式任务走工作流 checkpoint）
func (d *ChatDeps) RunTask(ctx context.Context, userID string, t *model.AgentTeam, task, goal string, prior *RunState) (*ChatResult, error) {
	task = strings.TrimSpace(task)
	if task == "" {
		return nil, fmt.Errorf("task 不能为空")
	}
	messages := []chat.ChatMessage{{Role: "user", Content: task}}
	if strings.TrimSpace(goal) == "" {
		goal = task
	}
	return d.RunChat(ctx, userID, t, messages, goal, prior)
}
