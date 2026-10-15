package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBuiltInTool_ByFunctionName(t *testing.T) {
	r := NewRegistry()
	tool := r.GetBuiltInTool("web_search_search")
	require.NotNil(t, tool)
	assert.Equal(t, "web_search", tool.ServerID)
	assert.Equal(t, "web_search_search", tool.Name)
}

func TestGetBuiltInTool_ByRegistryKey(t *testing.T) {
	r := NewRegistry()
	tool := r.GetBuiltInTool("web_search:search")
	require.NotNil(t, tool)
	assert.Equal(t, "web_search_search", tool.Name)
}

func TestGetToolsForAgent_FiltersByServerID(t *testing.T) {
	r := NewRegistry()
	tools := r.GetToolsForAgent([]string{"web_search"})
	require.Len(t, tools, 1)
	assert.Equal(t, "web_search_search", tools[0].Name)

	none := r.GetToolsForAgent([]string{"nonexistent"})
	assert.Empty(t, none)
}
