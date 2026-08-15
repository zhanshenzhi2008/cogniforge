package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPayload_StreamFlag(t *testing.T) {
	s := NewChatService(nil)
	temp := 0.7
	maxTokens := 2048
	req := &ChatRequest{
		Model:       "deepseek-chat",
		Messages:    []ChatMessage{{Role: "user", Content: "hello"}},
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}

	streamPayload := s.buildPayload(req, true)
	assert.Equal(t, true, streamPayload["stream"])
	assert.Equal(t, "deepseek-chat", streamPayload["model"])
	assert.Equal(t, 0.7, streamPayload["temperature"])
	assert.Equal(t, 2048, streamPayload["max_tokens"])

	syncPayload := s.buildPayload(req, false)
	assert.Equal(t, false, syncPayload["stream"])
}

func TestAIChatCompletionsURL(t *testing.T) {
	s := NewChatService(nil)
	require.Equal(t, "https://api.deepseek.com/v1/chat/completions", s.aiChatCompletionsURL("https://api.deepseek.com"))
	require.Equal(t, "https://api.deepseek.com/v1/chat/completions", s.aiChatCompletionsURL("https://api.deepseek.com/v1"))
	require.Equal(t, "https://api.deepseek.com/v1/chat/completions", s.aiChatCompletionsURL("https://api.deepseek.com/v1/"))
}
