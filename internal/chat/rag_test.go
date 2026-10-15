package chat

import (
	"context"
	"errors"
	"strings"
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
	frags := retrieveRAGFragments(context.Background(), m, "u1", []string{"kb-bad", "kb-ok"}, "退货", RAGOptions{TopK: 3, MinScore: 0.5, MaxTokens: 2000})
	require.Len(t, frags, 1)
	assert.Contains(t, frags[0], "hello")
}

func TestResolveRAGOptions_RequestOverrides(t *testing.T) {
	SetRAGDefaults(3, 0.5, 2000)
	topK := 7
	score := 0.8
	opt := resolveRAGOptions(&ChatRequest{RAGTopK: &topK, RAGMinScore: &score})
	assert.Equal(t, 7, opt.TopK)
	assert.InDelta(t, 0.8, opt.MinScore, 1e-9)
	assert.Equal(t, 2000, opt.MaxTokens)
}

func TestRetrieveRAGFragments_RespectsTopKBudget(t *testing.T) {
	m := &mockRAGSearcher{
		results: map[string]*knowledge.SearchResponse{
			"kb": {
				Results: []knowledge.SearchResult{
					{DocumentName: "a", Content: strings.Repeat("字", 40)},
					{DocumentName: "b", Content: strings.Repeat("字", 40)},
				},
			},
		},
	}
	// ~10 tokens per frag; budget 12 → only first fragment kept after first fills
	frags := retrieveRAGFragments(context.Background(), m, "u1", []string{"kb"}, "q", RAGOptions{TopK: 5, MinScore: 0.1, MaxTokens: 12})
	require.Len(t, frags, 1)
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

// TestStage13_E2E_AttachmentToInjectedSystem 13.6：附件 → 解析 KB → 检索 → 注入 system（不依赖实机 PDF）
func TestStage13_E2E_AttachmentToInjectedSystem(t *testing.T) {
	m := &mockRAGSearcher{
		docKB: map[string]string{"doc-manual": "kb-play"},
		results: map[string]*knowledge.SearchResponse{
			"kb-play": {
				Results: []knowledge.SearchResult{
					{DocumentName: "产品说明书.pdf", Content: "退货政策：自购买日起 7 天内可退"},
				},
			},
		},
	}
	req := &ChatRequest{
		Messages:    []ChatMessage{{Role: "user", Content: "退货几天？"}},
		Attachments: []MessageAttachment{{ID: "doc-manual", Name: "产品说明书.pdf", FileType: "pdf"}},
	}
	kbIDs := resolveRAGKnowledgeBaseIDs(context.Background(), m, "u1", req, nil)
	require.Equal(t, []string{"kb-play"}, kbIDs)

	query := lastUserQuery(req.Messages)
	frags := retrieveRAGFragments(context.Background(), m, "u1", kbIDs, query, resolveRAGOptions(req))
	require.NotEmpty(t, frags)
	assert.Contains(t, frags[0], "7 天")

	msgs := buildFullMessages("", nil, frags, nil, req.Messages)
	require.GreaterOrEqual(t, len(msgs), 2)
	assert.Equal(t, "system", msgs[0].Role)
	assert.Contains(t, msgs[0].Content, "知识库参考")
	assert.Contains(t, msgs[0].Content, "7 天")
	assert.Equal(t, "user", msgs[len(msgs)-1].Role)
	assert.Equal(t, "退货几天？", msgs[len(msgs)-1].Content)
}

// TestStage13_E2E_DegradeKeepsUserTurn 13.6：检索全失败时仍只保留用户轮次，不注入空 RAG
func TestStage13_E2E_DegradeKeepsUserTurn(t *testing.T) {
	m := &mockRAGSearcher{
		errs: map[string]error{"kb-x": errors.New("rag down")},
	}
	req := &ChatRequest{
		Messages:            []ChatMessage{{Role: "user", Content: "你好"}},
		RAGKnowledgeBaseIDs: []string{"kb-x"},
	}
	kbIDs := resolveRAGKnowledgeBaseIDs(context.Background(), m, "u1", req, nil)
	frags := retrieveRAGFragments(context.Background(), m, "u1", kbIDs, lastUserQuery(req.Messages), resolveRAGOptions(req))
	assert.Empty(t, frags)
	msgs := buildFullMessages("", nil, frags, nil, req.Messages)
	require.Len(t, msgs, 1)
	assert.Equal(t, "user", msgs[0].Role)
}

func TestInjectRAGSystemIfNeeded_AfterAgentSystem(t *testing.T) {
	m := &mockRAGSearcher{
		results: map[string]*knowledge.SearchResponse{
			"kb1": {Results: []knowledge.SearchResult{{DocumentName: "a.pdf", Content: "片段X"}}},
		},
	}
	req := &ChatRequest{
		Messages:            []ChatMessage{{Role: "user", Content: "问"}},
		RAGKnowledgeBaseIDs: []string{"kb1"},
	}
	base := []ChatMessage{
		{Role: "system", Content: "You are agent"},
		{Role: "user", Content: "问"},
	}
	out := InjectRAGSystemIfNeeded(context.Background(), m, "u1", req, nil, base)
	require.Len(t, out, 3)
	assert.Equal(t, "You are agent", out[0].Content)
	assert.Contains(t, out[1].Content, "片段X")
	assert.Equal(t, "user", out[2].Role)
}
