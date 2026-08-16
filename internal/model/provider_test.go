package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCatalogModels_DeepSeekV4(t *testing.T) {
	assert.Equal(t, []string{"deepseek-chat", "deepseek-reasoner", "deepseek-v4-flash", "deepseek-v4-pro"}, CatalogModels(string(ProviderDeepSeek)))
	assert.Empty(t, CatalogModels(string(ProviderOpenAI)))
}
