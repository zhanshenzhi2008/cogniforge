package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEstimateUsage(t *testing.T) {
	u := EstimateUsage([]ChatMessage{{Role: "user", Content: "hello world"}}, "hi")
	assert.Greater(t, u.TotalTokens, 0)
	assert.True(t, u.Estimated)
}

func TestSSEUsageScan(t *testing.T) {
	s := &sseUsageScan{}
	s.feed([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"你好\"}}]}\n"))
	s.feed([]byte("data: {\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":2,\"total_tokens\":5}}\n"))
	s.feed([]byte("data: [DONE]\n"))
	assert.Equal(t, "你好", s.text.String())
	assert.Equal(t, 5, s.usage.TotalTokens)
	assert.Equal(t, 3, s.usage.PromptTokens)
}
