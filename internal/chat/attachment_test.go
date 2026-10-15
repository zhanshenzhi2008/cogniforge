package chat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/model"
)

func TestMessageAttachmentJSONRoundTrip(t *testing.T) {
	req := ChatRequest{
		Model: "deepseek-chat",
		Messages: []ChatMessage{
			{Role: "user", Content: "这份说明书怎么说？"},
		},
		Attachments: []MessageAttachment{
			{
				ID:         "doc-1",
				Name:       "manual.pdf",
				FileType:   "pdf",
				Status:     "completed",
				ChunkCount: 12,
			},
		},
		RAGKnowledgeBaseIDs: []string{"kb-a", "kb-b"},
	}

	raw, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded ChatRequest
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Len(t, decoded.Attachments, 1)
	assert.Equal(t, "doc-1", decoded.Attachments[0].ID)
	assert.Equal(t, "manual.pdf", decoded.Attachments[0].Name)
	assert.Equal(t, "pdf", decoded.Attachments[0].FileType)
	assert.Equal(t, "completed", decoded.Attachments[0].Status)
	assert.Equal(t, 12, decoded.Attachments[0].ChunkCount)
	assert.Equal(t, []string{"kb-a", "kb-b"}, decoded.RAGKnowledgeBaseIDs)

	var wire map[string]any
	require.NoError(t, json.Unmarshal(raw, &wire))
	assert.Contains(t, wire, "attachments")
	assert.Contains(t, wire, "rag_kb_ids")
}

func TestConversationMessageAttachmentsJSON(t *testing.T) {
	msg := model.ConversationMessage{
		ID:      "m1",
		Role:    "user",
		Content: "看附件",
		Attachments: []model.MessageAttachment{
			{ID: "d1", Name: "a.txt", FileType: "txt", Status: "pending"},
		},
	}
	raw, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded model.ConversationMessage
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Len(t, decoded.Attachments, 1)
	assert.Equal(t, "d1", decoded.Attachments[0].ID)
	assert.Equal(t, "a.txt", decoded.Attachments[0].Name)
}
