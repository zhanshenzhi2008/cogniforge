package chat

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/knowledge"
)

type mockRAGSearcher struct {
	results map[string]*knowledge.SearchResponse
	errs    map[string]error
	docKB   map[string]string
}

func (m *mockRAGSearcher) SearchByUser(_ context.Context, _, kbID string, _ *knowledge.SearchRequest) (*knowledge.SearchResponse, error) {
	if err, ok := m.errs[kbID]; ok {
		return nil, err
	}
	if r, ok := m.results[kbID]; ok {
		return r, nil
	}
	return &knowledge.SearchResponse{}, nil
}

func (m *mockRAGSearcher) ResolveKBIDsForDocuments(_ context.Context, _ string, docIDs []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, id := range docIDs {
		kb := m.docKB[id]
		if kb == "" {
			continue
		}
		if _, ok := seen[kb]; ok {
			continue
		}
		seen[kb] = struct{}{}
		out = append(out, kb)
	}
	return out
}

func TestBuildRAGSystemPrompt(t *testing.T) {
	assert.Empty(t, buildRAGSystemPrompt(nil))
	got := buildRAGSystemPrompt([]string{"[manual.pdf] 退货 7 天"})
	assert.Contains(t, got, "知识库参考")
	assert.Contains(t, got, "退货 7 天")
	assert.Contains(t, got, "不要编造")
}

func TestRetrieveRAGFragments_DegradeOnError(t *testing.T) {
	m := &mockRAGSearcher{
		errs: map[string]error{"kb-bad": errors.New("down")},
		results: map[string]*knowledge.SearchResponse{
			"kb-ok": {
				Results: []knowledge.SearchResult{
					{DocumentName: "a.pdf", Content: "hello"},
				},
			},
		},
	}
	frags := retrieveRAGFragments(context.Background(), m, "u1", []string{"kb-bad", "kb-ok"}, "退货")
	require.Len(t, frags, 1)
	assert.Contains(t, frags[0], "hello")
}

func TestResolveRAGKnowledgeBaseIDs(t *testing.T) {
	m := &mockRAGSearcher{docKB: map[string]string{"doc-1": "kb-from-doc"}}
	req := &ChatRequest{
		RAGKnowledgeBaseIDs: []string{"kb-req"},
		Attachments:         []MessageAttachment{{ID: "doc-1"}},
	}
	ids := resolveRAGKnowledgeBaseIDs(context.Background(), m, "u1", req, []string{"kb-conv"})
	assert.Equal(t, []string{"kb-req", "kb-conv", "kb-from-doc"}, ids)
}

func TestBuildFullMessages_WithRAG(t *testing.T) {
	msgs := buildFullMessages("", nil, []string{"[doc] 片段内容"}, nil, []ChatMessage{
		{Role: "user", Content: "问一下"},
	})
	require.Len(t, msgs, 2)
	assert.Equal(t, "system", msgs[0].Role)
	assert.Contains(t, msgs[0].Content, "片段内容")
	assert.Equal(t, "user", msgs[1].Role)
}
