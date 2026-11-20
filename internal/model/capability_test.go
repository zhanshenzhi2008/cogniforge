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
	assert.True(t, HasCapability(CapWebFetch, CapWebFetch))
	assert.Equal(t, CapWebFetch, JoinCapabilities([]string{CapWebFetch}))
	assert.Equal(t, "chat,web_fetch", JoinCapabilities([]string{CapWebFetch, CapChat}))
	assert.False(t, HasCapability(CapWebFetch, CapChat))
	assert.True(t, HasCapability(CapRerank, CapRerank))
	assert.Equal(t, CapRerank, JoinCapabilities([]string{CapRerank}))
	assert.Equal(t, "chat,rerank", JoinCapabilities([]string{CapRerank, CapChat}))
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

func TestRerankModelName(t *testing.T) {
	p := &AIProvider{DefaultModel: "deepseek-chat", Capabilities: CapChat}
	assert.Empty(t, RerankModelName(p))
	p.Capabilities = CapRerank
	p.DefaultModel = "jina-reranker-v2-base-multilingual"
	assert.Equal(t, "jina-reranker-v2-base-multilingual", RerankModelName(p))
	p.RerankModel = "jina-reranker-v2-base-multilingual"
	p.Capabilities = CapChat + "," + CapRerank
	assert.Equal(t, "jina-reranker-v2-base-multilingual", RerankModelName(p))
}
