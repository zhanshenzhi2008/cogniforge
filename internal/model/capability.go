package model

import "strings"

const (
	CapChat      = "chat"
	CapEmbedding = "embedding"
)

func SplitCapabilities(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{CapChat}
	}
	seen := map[string]struct{}{}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != CapChat && p != CapEmbedding {
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
	ordered := make([]string, 0, 2)
	for _, want := range []string{CapChat, CapEmbedding} {
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
