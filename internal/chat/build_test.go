package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"cogniforge/internal/model"
)

// TestBuildFullMessages_OnlyCurrent 测试只有本轮消息（无历史、无摘要、无长期记忆）
func TestBuildFullMessages_OnlyCurrent(t *testing.T) {
	current := []ChatMessage{{Role: "user", Content: "你好"}}
	msgs := buildFullMessages("", nil, nil, current)
	assert.Len(t, msgs, 1)
	assert.Equal(t, "user", msgs[0].Role)
	assert.Equal(t, "你好", msgs[0].Content)
}

// TestBuildFullMessages_WithHistory 测试有历史消息
func TestBuildFullMessages_WithHistory(t *testing.T) {
	history := []model.ConversationMessage{
		{Role: "user", Content: "你好"},
		{Role: "assistant", Content: "你好！有什么可以帮你？"},
	}
	current := []ChatMessage{{Role: "user", Content: "帮我写代码"}}
	msgs := buildFullMessages("", nil, history, current)
	assert.Len(t, msgs, 3)
	assert.Equal(t, "user", msgs[0].Role)
	assert.Equal(t, "你好", msgs[0].Content)
	assert.Equal(t, "assistant", msgs[1].Role)
	assert.Equal(t, "帮我写代码", msgs[2].Content)
}

// TestBuildFullMessages_WithSummary 测试有摘要
func TestBuildFullMessages_WithSummary(t *testing.T) {
	current := []ChatMessage{{Role: "user", Content: "继续"}}
	history := []model.ConversationMessage{{Role: "user", Content: "之前的问题"}}
	msgs := buildFullMessages("用户想了解 Redis 缓存", nil, history, current)
	assert.Len(t, msgs, 3)
	assert.Equal(t, "system", msgs[0].Role)
	assert.Contains(t, msgs[0].Content, "Redis 缓存")
	assert.Equal(t, "user", msgs[1].Role) // history
	assert.Equal(t, "user", msgs[2].Role) // current
}

// TestBuildFullMessages_WithLongTermMemory 测试有长期记忆块（14.4 核心场景）
func TestBuildFullMessages_WithLongTermMemory(t *testing.T) {
	longTerm := []string{
		"用户叫小明，职业后端开发",
		"偏好：回复用中文，代码要可复制",
	}
	current := []ChatMessage{{Role: "user", Content: "帮我设计一个 API"}}
	msgs := buildFullMessages("", longTerm, nil, current)
	// 顺序：[长期记忆块] + [本轮]
	assert.Len(t, msgs, 2)
	assert.Equal(t, "system", msgs[0].Role)
	assert.Contains(t, msgs[0].Content, "笔记本")
	assert.Contains(t, msgs[0].Content, "小明")
	assert.Contains(t, msgs[0].Content, "后端开发")
	assert.Contains(t, msgs[0].Content, "中文")
}

// TestBuildFullMessages_WithAllThree 测试摘要+长期记忆+历史+本轮（全量）
func TestBuildFullMessages_WithAllThree(t *testing.T) {
	summary := "用户在做一个 Redis 缓存项目"
	longTerm := []string{"用户用 Go 语言"}
	history := []model.ConversationMessage{
		{Role: "user", Content: "Redis 怎么用"},
		{Role: "assistant", Content: "可以用 Redis 做缓存"},
	}
	current := []ChatMessage{{Role: "user", Content: "具体怎么配置"}}
	msgs := buildFullMessages(summary, longTerm, history, current)
	// 顺序：摘要 → 长期记忆 → 历史(2) → 本轮(1) = 5条
	assert.Len(t, msgs, 5)
	assert.Equal(t, "system", msgs[0].Role)
	assert.Contains(t, msgs[0].Content, "摘要")
	assert.Contains(t, msgs[0].Content, "Redis")
	assert.Equal(t, "system", msgs[1].Role)
	assert.Contains(t, msgs[1].Content, "笔记本")
	assert.Contains(t, msgs[1].Content, "Go")
	assert.Equal(t, "user", msgs[2].Role)
	assert.Equal(t, "assistant", msgs[3].Role)
	assert.Equal(t, "user", msgs[4].Role) // 当前 user
}

// TestBuildFullMessages_LongTermMultipleBlocks 测试多条长期记忆编号
func TestBuildFullMessages_LongTermMultipleBlocks(t *testing.T) {
	longTerm := []string{
		"记忆1",
		"记忆2",
		"记忆3",
	}
	current := []ChatMessage{{Role: "user", Content: "问"}}
	msgs := buildFullMessages("", longTerm, nil, current)
	systemContent := msgs[0].Content
	assert.Contains(t, systemContent, "1. 记忆1")
	assert.Contains(t, systemContent, "2. 记忆2")
	assert.Contains(t, systemContent, "3. 记忆3")
}

// TestBuildFullMessages_LongTermEmpty 测试空长期记忆不注入
func TestBuildFullMessages_LongTermEmpty(t *testing.T) {
	current := []ChatMessage{{Role: "user", Content: "你好"}}
	history := []model.ConversationMessage{{Role: "user", Content: "hi"}}
	msgs := buildFullMessages("", []string{}, history, current)
	// 摘要空 + 长期记忆空 + 历史1条 + 本轮1条 = 2
	assert.Len(t, msgs, 2)
	assert.NotContains(t, msgs[0].Content, "笔记本")
}
