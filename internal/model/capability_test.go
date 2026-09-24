package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasCapability(t *testing.T) {
	assert.True(t, HasCapability("", CapChat))
	assert.False(t, HasCapability("", CapEmbedding))
	assert.True(t, HasCapability("chat,embedding", CapEmbedding))
	assert.False(t, HasCapability("chat", CapEmbedding))
	assert.Equal(t, "chat,embedding", JoinCapabilities([]string{"embedding", "chat", "chat"}))
}

func TestEmbeddingModelName(t *testing.T) {
	p := &AIProvider{DefaultModel: "deepseek-chat", Capabilities: CapChat}
	assert.Empty(t, EmbeddingModelName(p))
	p.Capabilities = CapEmbedding
	p.DefaultModel = "BAAI/bge-m3"
	assert.Equal(t, "BAAI/bge-m3", EmbeddingModelName(p))
	p.EmbeddingModel = "text-embedding-3-small"
	p.Capabilities = CapChat + "," + CapEmbedding
	assert.Equal(t, "text-embedding-3-small", EmbeddingModelName(p))
}
