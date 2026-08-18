package chat

import (
	"encoding/json"
	"strings"
)

type sseUsageScan struct {
	rest  string
	text  strings.Builder
	usage ChatUsage
}

func (s *sseUsageScan) feed(p []byte) {
	s.rest += string(p)
	for {
		i := strings.IndexByte(s.rest, '\n')
		if i < 0 {
			return
		}
		line := strings.TrimRight(s.rest[:i], "\r")
		s.rest = s.rest[i+1:]
		s.handleLine(line)
	}
}

func (s *sseUsageScan) handleLine(line string) {
	if !strings.HasPrefix(line, "data:") {
		return
	}
	data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	if data == "" || data == "[DONE]" {
		return
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return
	}
	if u, ok := raw["usage"].(map[string]any); ok {
		s.usage.PromptTokens = jsonInt(u["prompt_tokens"])
		s.usage.CompletionTokens = jsonInt(u["completion_tokens"])
		s.usage.TotalTokens = jsonInt(u["total_tokens"])
		if s.usage.TotalTokens == 0 {
			s.usage.TotalTokens = s.usage.PromptTokens + s.usage.CompletionTokens
		}
	}
	choices, _ := raw["choices"].([]any)
	for _, ch := range choices {
		cm, _ := ch.(map[string]any)
		delta, _ := cm["delta"].(map[string]any)
		if c, ok := delta["content"].(string); ok {
			s.text.WriteString(c)
		}
	}
}

func jsonInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

// EstimateUsage 上游没给 usage 时按字数估算。
func EstimateUsage(messages []ChatMessage, completion string) ChatUsage {
	prompt := 0
	for _, m := range messages {
		prompt += tokenEstimate(m.Content)
	}
	comp := tokenEstimate(completion)
	return ChatUsage{
		PromptTokens:     prompt,
		CompletionTokens: comp,
		TotalTokens:      prompt + comp,
		Estimated:        true,
	}
}

func tokenEstimate(s string) int {
	if s == "" {
		return 0
	}
	ascii := 0
	n := 0
	for _, r := range s {
		n++
		if r < 128 {
			ascii++
		}
	}
	cjk := n - ascii
	est := ascii/4 + cjk
	if est < 1 {
		return 1
	}
	return est
}
