package chat

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"cogniforge/internal/knowledge"
)

const (
	ragFallbackTopK     = 3
	ragFallbackMinScore = 0.5
	ragFallbackMaxTok   = 2000
	ragMaxFragments     = 12
)

// RAGOptions 单次检索参数（请求体优先，否则服务端默认）
type RAGOptions struct {
	TopK      int
	MinScore  float64
	MaxTokens int
}

var ragDefaults = RAGOptions{
	TopK:      ragFallbackTopK,
	MinScore:  ragFallbackMinScore,
	MaxTokens: ragFallbackMaxTok,
}

// SetRAGDefaults 从配置注入默认 Top-K / min_score / 片段 token 预算（阶段十三 13.5）
func SetRAGDefaults(topK int, minScore float64, maxTokens int) {
	if topK > 0 {
		ragDefaults.TopK = topK
	}
	if minScore > 0 {
		ragDefaults.MinScore = minScore
	}
	if maxTokens > 0 {
		ragDefaults.MaxTokens = maxTokens
	}
}

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

func resolveRAGOptions(req *ChatRequest) RAGOptions {
	opt := ragDefaults
	if req != nil {
		if req.RAGTopK != nil && *req.RAGTopK > 0 {
			opt.TopK = *req.RAGTopK
		}
		if req.RAGMinScore != nil && *req.RAGMinScore > 0 {
			opt.MinScore = *req.RAGMinScore
		}
	}
	if opt.TopK < 1 {
		opt.TopK = ragFallbackTopK
	}
	if opt.TopK > 20 {
		opt.TopK = 20
	}
	if opt.MinScore <= 0 {
		opt.MinScore = ragFallbackMinScore
	}
	if opt.MinScore > 1 {
		opt.MinScore = 1
	}
	if opt.MaxTokens <= 0 {
		opt.MaxTokens = ragFallbackMaxTok
	}
	return opt
}

// approxTokens 粗估 token（按 rune/4，与前端 tokenCount 一致）
func approxTokens(s string) int {
	n := utf8.RuneCountInString(s)
	if n <= 0 {
		return 0
	}
	return (n + 3) / 4
}

// retrieveRAGFragments 对多个知识库做 Top-K 检索；失败降级为空，不阻断对话
func retrieveRAGFragments(ctx context.Context, searcher RAGSearcher, userID string, kbIDs []string, query string, opts RAGOptions) []string {
	if searcher == nil || userID == "" || query == "" || len(kbIDs) == 0 {
		return nil
	}
	if opts.TopK <= 0 {
		opts = resolveRAGOptions(nil)
	}
	var fragments []string
	usedTokens := 0
	for _, kbID := range kbIDs {
		resp, err := searcher.SearchByUser(ctx, userID, kbID, &knowledge.SearchRequest{
			Query:    query,
			TopK:     opts.TopK,
			MinScore: opts.MinScore,
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
			frag := fmt.Sprintf("[%s] %s", label, r.Content)
			cost := approxTokens(frag)
			if opts.MaxTokens > 0 && usedTokens+cost > opts.MaxTokens && len(fragments) > 0 {
				return fragments
			}
			fragments = append(fragments, frag)
			usedTokens += cost
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

// InjectRAGSystemIfNeeded 供 Agent / Chat 共用：检索后在首批 system 之后插入「知识库参考」。
// 无 KB、无命中或检索失败时原样返回 messages（降级）。
func InjectRAGSystemIfNeeded(ctx context.Context, searcher RAGSearcher, userID string, req *ChatRequest, convKBIDs []string, messages []ChatMessage) []ChatMessage {
	if searcher == nil || req == nil || userID == "" || len(messages) == 0 {
		return messages
	}
	kbIDs := resolveRAGKnowledgeBaseIDs(ctx, searcher, userID, req, convKBIDs)
	query := lastUserQuery(req.Messages)
	if query == "" {
		query = lastUserQuery(messages)
	}
	if len(kbIDs) == 0 || query == "" {
		return messages
	}
	frags := retrieveRAGFragments(ctx, searcher, userID, kbIDs, query, resolveRAGOptions(req))
	prompt := buildRAGSystemPrompt(frags)
	if prompt == "" {
		return messages
	}
	insertAt := 0
	for insertAt < len(messages) && messages[insertAt].Role == "system" {
		insertAt++
	}
	out := make([]ChatMessage, 0, len(messages)+1)
	out = append(out, messages[:insertAt]...)
	out = append(out, ChatMessage{Role: "system", Content: prompt})
	out = append(out, messages[insertAt:]...)
	return out
}
