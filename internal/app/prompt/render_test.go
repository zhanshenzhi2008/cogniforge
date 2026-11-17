package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/model"
)

func TestExtractVarNames(t *testing.T) {
	names := ExtractVarNames("你好 {{brand}}，我是 {{ brand }} 助手，再见 {{user}}")
	assert.Equal(t, []string{"brand", "user"}, names)
}

func TestRenderContent_DefaultsAndRequired(t *testing.T) {
	defs := model.PromptVarDefs{
		{Name: "brand", Default: "CogniForge", Required: true},
		{Name: "tone", Default: "专业"},
	}
	out, err := RenderContent("品牌={{brand}} 语气={{tone}}", nil, defs, true)
	require.NoError(t, err)
	assert.Equal(t, "品牌=CogniForge 语气=专业", out)

	_, err = RenderContent("品牌={{brand}}", map[string]string{}, model.PromptVarDefs{{Name: "brand", Required: true}}, true)
	require.Error(t, err)

	out, err = RenderContent("品牌={{brand}}", map[string]string{"brand": "X"}, defs, true)
	require.NoError(t, err)
	assert.Equal(t, "品牌=X", out)
}
