package chat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractToolCalls_FromMessage(t *testing.T) {
	raw := `{
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": null,
				"tool_calls": [{
					"id": "call_1",
					"type": "function",
					"function": {
						"name": "web_search_search",
						"arguments": "{\"query\":\"北京天气\"}"
					}
				}]
			},
			"finish_reason": "tool_calls"
		}]
	}`
	var resp ChatResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))

	calls := extractToolCalls(&resp)
	require.Len(t, calls, 1)
	assert.Equal(t, "call_1", calls[0].ID)
	assert.Equal(t, "web_search_search", calls[0].Function.Name)
	assert.Contains(t, calls[0].Function.Arguments, "北京天气")
}

func TestExtractToolCalls_Empty(t *testing.T) {
	resp := &ChatResponse{Choices: []ChatChoice{{
		Message: ChatMessage{Role: "assistant", Content: "你好"},
	}}}
	assert.Empty(t, extractToolCalls(resp))
}
