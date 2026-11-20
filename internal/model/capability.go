package model

import "strings"

const (
	CapChat      = "chat"
	CapEmbedding = "embedding"
	// CapWebFetch 网页抓取类服务（Jina Reader / Firecrawl），不进对话 modelcfg
	CapWebFetch = "web_fetch"
	// CapRerank 重排序模型（知识库检索 Rerank），不进对话 modelcfg
	CapRerank = "rerank"
)

var knownCapabilities = []string{CapChat, CapEmbedding, CapWebFetch, CapRerank}

func isKnownCapability(p string) bool {
	for _, want := range knownCapabilities {
		if p == want {
			return true
		}
	}
	return false
}

func SplitCapabilities(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{CapChat}
	}
	seen := map[string]struct{}{}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.ToLower(strings.TrimSpace(p))
		if !isKnownCapability(p) {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	if len(out) == 0 {
		return []string{CapChat}
	}
	return out
}

func JoinCapabilities(caps []string) string {
	parts := SplitCapabilities(strings.Join(caps, ","))
	ordered := make([]string, 0, len(knownCapabilities))
	for _, want := range knownCapabilities {
		for _, p := range parts {
			if p == want {
				ordered = append(ordered, p)
			}
		}
	}
	return strings.Join(ordered, ",")
}

func HasCapability(raw, want string) bool {
	for _, c := range SplitCapabilities(raw) {
		if c == want {
			return true
		}
	}
	return false
}

func EmbeddingModelName(p *AIProvider) string {
	if p == nil {
		return ""
	}
	if m := strings.TrimSpace(p.EmbeddingModel); m != "" {
		return m
	}
	if HasCapability(p.Capabilities, CapEmbedding) && !HasCapability(p.Capabilities, CapChat) {
		return strings.TrimSpace(p.DefaultModel)
	}
	return ""
}

// RerankModelName 解析重排序模型名：优先 rerank_model；纯 rerank 供应商可回落到 default_model
func RerankModelName(p *AIProvider) string {
	if p == nil {
		return ""
	}
	if m := strings.TrimSpace(p.RerankModel); m != "" {
		return m
	}
	if HasCapability(p.Capabilities, CapRerank) && !HasCapability(p.Capabilities, CapChat) {
		return strings.TrimSpace(p.DefaultModel)
	}
	return ""
}
