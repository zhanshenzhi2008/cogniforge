package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCatalogModels_DeepSeekV4(t *testing.T) {
	assert.Equal(t, []string{"deepseek-chat", "deepseek-reasoner", "deepseek-v4-flash", "deepseek-v4-pro"}, CatalogModels(string(ProviderDeepSeek)))
	assert.Empty(t, CatalogModels(string(ProviderOpenAI)))
}

func TestVendorFromBaseURL(t *testing.T) {
	assert.Equal(t, "zhipu", VendorFromBaseURL("https://open.bigmodel.cn/api/paas/v4"))
	assert.Equal(t, "openai", VendorFromBaseURL("https://api.openai.com/v1"))
	assert.Equal(t, "groq", VendorFromBaseURL("https://api.groq.com/openai/v1"))
	assert.Equal(t, "aicore", VendorFromBaseURL("https://api.xty.app/v1"))
	assert.Equal(t, "", VendorFromBaseURL("https://my-proxy.example/v1"))
	assert.Equal(t, "", VendorFromBaseURL("  "))
}

func TestCatalogModels_ZhipuFreeFirst(t *testing.T) {
	got := CatalogModels(string(ProviderZhipu))
	assert.Equal(t, "glm-4.7-flash", got[0])
	assert.Contains(t, got, "glm-4.5-air")
	assert.Contains(t, got, "glm-4.6v-flash")
}
