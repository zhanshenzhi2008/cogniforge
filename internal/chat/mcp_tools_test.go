package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/mcp"
)

func TestInjectMCPTools_WebSearch(t *testing.T) {
	reg := mcp.NewRegistry()
	req := &ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "北京天气"}}}
	InjectMCPTools(req, reg, []string{"web_search"})
	require.Len(t, req.Tools, 1)
	assert.Equal(t, "function", req.Tools[0].Type)
	assert.Equal(t, "web_search_search", req.Tools[0].Function.Name)
	assert.NotEmpty(t, req.Tools[0].Function.Description)
	require.NotEmpty(t, req.Messages)
	assert.Equal(t, "system", req.Messages[0].Role)
	assert.Contains(t, req.Messages[0].Content.(string), "web_search_search")
}

func TestInjectMCPTools_NilSafe(t *testing.T) {
	reg := mcp.NewRegistry()
	InjectMCPTools(nil, reg, []string{"web_search"})
	req := &ChatRequest{}
	InjectMCPTools(req, nil, []string{"web_search"})
	assert.Empty(t, req.Tools)
	InjectMCPTools(req, reg, nil)
	assert.Empty(t, req.Tools)
}

func TestInjectMCPTools_Appends(t *testing.T) {
	reg := mcp.NewRegistry()
	req := &ChatRequest{
		Tools: []ToolCallSpec{{
			Type:     "function",
			Function: ToolFunctionSpec{Name: "existing"},
		}},
	}
	InjectMCPTools(req, reg, []string{"fetch"})
	require.Len(t, req.Tools, 2)
	assert.Equal(t, "existing", req.Tools[0].Function.Name)
	assert.Equal(t, "fetch_get_page", req.Tools[1].Function.Name)
}

func TestNewMCPToolExecutor_UnknownTool(t *testing.T) {
	ex := NewMCPToolExecutor(mcp.NewRegistry())
	_, err := ex.Execute(ToolCall{Function: ToolCallFunction{Name: "no_such_tool"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tool")
}
