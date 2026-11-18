package chat

import (
	"fmt"
	"strings"

	"cogniforge/internal/capability/memory"
	"cogniforge/internal/model"
)

// ToWindowMessages ChatMessage → memory 窗口消息
func ToWindowMessages(in []ChatMessage) []memory.WindowMessage {
	out := make([]memory.WindowMessage, len(in))
	for i, m := range in {
		out[i] = memory.WindowMessage{Role: m.Role, Content: m.Content}
	}
	return out
}

// FromWindowMessages memory 窗口消息 → ChatMessage
func FromWindowMessages(in []memory.WindowMessage) []ChatMessage {
	out := make([]ChatMessage, len(in))
	for i, m := range in {
		out[i] = ChatMessage{Role: m.Role, Content: m.Content}
	}
	return out
}

// BuildFullMessages 组装完整 messages：[摘要?] + [长期记忆] + [RAG?] + [历史] + [本轮]
func BuildFullMessages(summary string, longTermBlocks, ragBlocks []string, history []model.ConversationMessage, current []ChatMessage) []ChatMessage {
	var msgs []ChatMessage
	if summary != "" {
		msgs = append(msgs, ChatMessage{
			Role:    "system",
			Content: "【对话摘要】" + summary,
		})
	}
	if len(longTermBlocks) > 0 {
		var b strings.Builder
		b.WriteString("【你的笔记本】以下是你之前记录的重要信息，回答时可参考（但不要复述）：\n")
		for i, block := range longTermBlocks {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, block))
		}
		msgs = append(msgs, ChatMessage{Role: "system", Content: b.String()})
	}
	if ragPrompt := buildRAGSystemPrompt(ragBlocks); ragPrompt != "" {
		msgs = append(msgs, ChatMessage{Role: "system", Content: ragPrompt})
	}
	for _, m := range history {
		msgs = append(msgs, ChatMessage{Role: m.Role, Content: m.Content})
	}
	for _, m := range current {
		msgs = append(msgs, m)
	}
	return msgs
}

func toStringContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var b strings.Builder
		for _, part := range v {
			if m, ok := part.(map[string]any); ok {
				if t, ok := m["text"].(string); ok {
					b.WriteString(t)
				}
			}
		}
		return b.String()
	default:
		return ""
	}
}

// 包内测试兼容旧名
func buildFullMessages(summary string, longTermBlocks, ragBlocks []string, history []model.ConversationMessage, current []ChatMessage) []ChatMessage {
	return BuildFullMessages(summary, longTermBlocks, ragBlocks, history, current)
}

// ToConversationMessages 转会话存储格式
func ToConversationMessages(msgs []ChatMessage) []model.ConversationMessage {
	out := make([]model.ConversationMessage, len(msgs))
	for i, m := range msgs {
		out[i] = model.ConversationMessage{
			Role:    m.Role,
			Content: toStringContent(m.Content),
		}
	}
	return out
}
