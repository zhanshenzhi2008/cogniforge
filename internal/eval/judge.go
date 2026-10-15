package eval

import (
	"strings"

	"cogniforge/internal/model"
)

// Judge 对实际输出打分。返回 status / score / detail。
func Judge(c *model.EvalCase, actual string) (status string, score float64, detail map[string]interface{}) {
	detail = map[string]interface{}{}
	expected := strings.TrimSpace(c.Expected)
	actualTrim := strings.TrimSpace(actual)

	switch c.JudgeType {
	case model.EvalJudgeExact:
		ok := actualTrim == expected
		detail["mode"] = "exact"
		if ok {
			return model.EvalResultPass, 1, detail
		}
		return model.EvalResultFail, 0, detail

	case model.EvalJudgeLLM:
		detail["mode"] = "llm_judge"
		detail["reason"] = "一期未实现，已跳过"
		return model.EvalResultSkip, 0, detail

	default: // contains
		detail["mode"] = "contains"
		if expected == "" {
			detail["reason"] = "expected 为空，视为通过"
			return model.EvalResultPass, 1, detail
		}
		parts := splitKeywords(expected)
		matchAll := false
		if c.JudgeConfig != nil {
			if m, ok := c.JudgeConfig["match"].(string); ok && strings.EqualFold(m, "all") {
				matchAll = true
			}
		}
		detail["keywords"] = parts
		detail["match"] = map[bool]string{true: "all", false: "any"}[matchAll]

		lowerActual := strings.ToLower(actualTrim)
		var hit, miss []string
		for _, p := range parts {
			if strings.Contains(lowerActual, strings.ToLower(p)) {
				hit = append(hit, p)
			} else {
				miss = append(miss, p)
			}
		}
		detail["hit"] = hit
		detail["miss"] = miss
		ok := false
		if matchAll {
			ok = len(miss) == 0 && len(parts) > 0
		} else {
			ok = len(hit) > 0
		}
		if ok {
			return model.EvalResultPass, 1, detail
		}
		return model.EvalResultFail, 0, detail
	}
}

func splitKeywords(expected string) []string {
	raw := strings.FieldsFunc(expected, func(r rune) bool {
		return r == ',' || r == '，' || r == ';' || r == '；' || r == '\n'
	})
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 && strings.TrimSpace(expected) != "" {
		return []string{strings.TrimSpace(expected)}
	}
	return out
}
