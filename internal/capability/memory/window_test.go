package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ===================== Config.Apply 测试 =====================

func makeMsgs(roles ...string) []WindowMessage {
	msgs := make([]WindowMessage, len(roles))
	for i, r := range roles {
		msgs[i] = WindowMessage{Role: r, Content: "content " + r}
	}
	return msgs
}

func TestApply_NoTrim_NilConfig(t *testing.T) {
	msgs := makeMsgs("user", "assistant", "user", "assistant")
	result := Config{MemoryTurns: 0, MaxInputTokens: 0}.Apply(msgs)
	assert.Equal(t, msgs, result)
}

func TestApply_NoTrim_WithinTurns(t *testing.T) {
	// 2 turns, 4 msgs -> 不裁
	msgs := makeMsgs("user", "assistant", "user", "assistant")
	result := Config{MemoryTurns: 2, MaxInputTokens: 0}.Apply(msgs)
	assert.Len(t, result, 4)
}

func TestApply_Trim_TurnsExceeded(t *testing.T) {
	// 10 msgs = 5 turns, memoryTurns=2 -> 只留最后 4
	msgs := makeMsgs(
		"user", "assistant",
		"user", "assistant",
		"user", "assistant",
		"user", "assistant",
		"user", "assistant",
	)
	result := Config{MemoryTurns: 2, MaxInputTokens: 0}.Apply(msgs)
	assert.Len(t, result, 4) // 2 turns = 4 msgs
	// 应该是最后 2 轮
	assert.Equal(t, "user", result[0].Role)
	assert.Equal(t, "assistant", result[1].Role)
	assert.Equal(t, "user", result[2].Role)
	assert.Equal(t, "assistant", result[3].Role)
}

func TestApply_Trim_TurnsOddUser(t *testing.T) {
	// 5 msgs: user assistant user assistant user
	// 2 turns -> 保留后 4 条（最后一轮 user 落单也保留）
	msgs := makeMsgs("user", "assistant", "user", "assistant", "user")
	result := Config{MemoryTurns: 2, MaxInputTokens: 0}.Apply(msgs)
	// 最后 2 轮 = assistant+user（最后 user 优先）+ user+assistant
	assert.Equal(t, "user", result[len(result)-1].Role) // 最后一个是 user
}

func TestApply_EmptyInput(t *testing.T) {
	result := Config{MemoryTurns: 2, MaxInputTokens: 0}.Apply([]WindowMessage{})
	assert.Empty(t, result)
}

func TestApply_EmptyContentFiltered(t *testing.T) {
	msgs := []WindowMessage{
		{Role: "user", Content: ""},
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: ""},
		{Role: "assistant", Content: "hi"},
	}
	result := Config{MemoryTurns: 10, MaxInputTokens: 0}.Apply(msgs)
	assert.Len(t, result, 2)
	assert.Equal(t, "hello", result[0].Content)
	assert.Equal(t, "hi", result[1].Content)
}

func TestApply_MaxInputTokens(t *testing.T) {
	// 用大量文本测试 token 预算裁剪
	msgs := make([]WindowMessage, 4)
	largeContent := ""
	for i := 0; i < 500; i++ {
		largeContent += "这是一段很长的中文内容用于测试token预算裁剪功能。 "
	}
	for i := 0; i < 4; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		msgs[i] = WindowMessage{Role: role, Content: largeContent}
	}

	// 限制很小 token 数，应该至少保留最后一条 user
	result := Config{MemoryTurns: 10, MaxInputTokens: 50}.Apply(msgs)
	assert.NotEmpty(t, result)
	// 最后一条 user 不应被删
	assert.Equal(t, "user", result[len(result)-1].Role)
}

// ===================== filterEmpty 测试 =====================

func TestFilterEmpty(t *testing.T) {
	tests := []struct {
		name     string
		content  any
		expected bool
	}{
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"nil", nil, true},
		{"empty array", []any{}, true},
		{"array with text", []any{map[string]any{"type": "text", "text": "hi"}}, false},
		{"array with image", []any{map[string]any{"type": "image_url", "image_url": map[string]any{"url": "x"}}}, false},
		{"array with empty text only", []any{map[string]any{"type": "text", "text": ""}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isEmptyContent(tt.content))
		})
	}
}

// ===================== trimToTurns 测试 =====================

func TestTrimToTurns_UnderLimit(t *testing.T) {
	msgs := makeMsgs("user", "assistant")
	result := trimToTurns(msgs, 5)
	assert.Len(t, result, 2)
}

func TestTrimToTurns_OddMessages(t *testing.T) {
	// user, assistant, user（3条）
	msgs := makeMsgs("user", "assistant", "user")
	result := trimToTurns(msgs, 2)
	// 2 turns = 4，但只有 3 条，应返回 3
	assert.Len(t, result, 3)
}

// ===================== tokenCount 测试 =====================

func TestTokenCount(t *testing.T) {
	msgs := []WindowMessage{
		{Role: "user", Content: "hello world"},
	}
	cnt := tokenCount(msgs)
	assert.Greater(t, cnt, 0)
}

func TestTokenCount_Empty(t *testing.T) {
	assert.Equal(t, 0, tokenCount([]WindowMessage{}))
}

// ===================== trimByToken 测试 =====================

func TestTrimByToken_NoLimit(t *testing.T) {
	msgs := makeMsgs("user", "assistant", "user", "assistant")
	result := trimByToken(msgs, 0)
	assert.Len(t, result, 4)
}

func TestTrimByToken_PreservesLastUser(t *testing.T) {
	// 3 msgs: user, assistant, user
	msgs := []WindowMessage{
		{Role: "user", Content: "user1"},
		{Role: "assistant", Content: "assistant1"},
		{Role: "user", Content: "user2"},
	}
	result := trimByToken(msgs, 1)
	assert.NotEmpty(t, result)
	assert.Equal(t, "user", result[len(result)-1].Role, "last user should be preserved")
}

func TestTrimByToken_VerySmallBudget(t *testing.T) {
	msgs := makeMsgs("user", "assistant", "user", "assistant")
	result := trimByToken(msgs, 5)
	// 应该只保留最后一个 user
	assert.NotEmpty(t, result)
	assert.Equal(t, "user", result[len(result)-1].Role)
}

// ===================== DefaultConfig 测试 =====================

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 10, cfg.MemoryTurns)
	assert.Equal(t, 0, cfg.MaxInputTokens)
}
