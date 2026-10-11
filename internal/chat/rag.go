package chat

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"cogniforge/internal/knowledge"
)

const (
	ragDefaultTopK     = 3
	ragDefaultMinScore = 0.5
	ragMaxFragments    = 12
)

// RAGSearcher Chat 侧检索接口（便于单测 mock）
type RAGSearcher interface {
	SearchByUser(ctx context.Context, userID, kbID string, req *knowledge.SearchRequest) (*knowledge.SearchResponse, error)
	ResolveKBIDsForDocuments(ctx context.Context, userID string, docIDs []string) []string
}

func lastUserQuery(msgs []ChatMessage) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			return strings.TrimSpace(toStringContent(msgs[i].Content))
		}
	}
	return ""
}

func mergeUniqueIDs(lists ...[]string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, list := range lists {
		for _, id := range list {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

func buildRAGSystemPrompt(fragments []string) string {
	if len(fragments) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## 知识库参考\n")
	sb.WriteString("以下是相关文档片段，请基于这些信息回答：\n\n")
	for i, f := range fragments {
		sb.WriteString(fmt.Sprintf("【片段 %d】\n%s\n\n", i+1, f))
	}
	sb.WriteString("如果上述信息无法回答问题，请如实说明，不要编造。")
	return sb.String()
}

// retrieveRAGFragments 对多个知识库做 Top-K 检索；失败降级为空，不阻断对话
func retrieveRAGFragments(ctx context.Context, searcher RAGSearcher, userID string, kbIDs []string, query string) []string {
	if searcher == nil || userID == "" || query == "" || len(kbIDs) == 0 {
		return nil
	}
	var fragments []string
	for _, kbID := range kbIDs {
		resp, err := searcher.SearchByUser(ctx, userID, kbID, &knowledge.SearchRequest{
			Query:    query,
			TopK:     ragDefaultTopK,
			MinScore: ragDefaultMinScore,
		})
		if err != nil {
			slog.Warn("RAG 检索失败，降级跳过", "kb_id", kbID, "error", err)
			continue
		}
		for _, r := range resp.Results {
			label := r.DocumentName
			if label == "" {
				label = kbID
			}
			fragments = append(fragments, fmt.Sprintf("[%s] %s", label, r.Content))
			if len(fragments) >= ragMaxFragments {
				return fragments
			}
		}
	}
	return fragments
}

func resolveRAGKnowledgeBaseIDs(ctx context.Context, searcher RAGSearcher, userID string, req *ChatRequest, convKBIDs []string) []string {
	ids := mergeUniqueIDs(req.RAGKnowledgeBaseIDs, convKBIDs)
	if searcher != nil && len(req.Attachments) > 0 {
		docIDs := make([]string, 0, len(req.Attachments))
		for _, a := range req.Attachments {
			if id := strings.TrimSpace(a.ID); id != "" {
				docIDs = append(docIDs, id)
			}
		}
		ids = mergeUniqueIDs(ids, searcher.ResolveKBIDsForDocuments(ctx, userID, docIDs))
	}
	return ids
}
