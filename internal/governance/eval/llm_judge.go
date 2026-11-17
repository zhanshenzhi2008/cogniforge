package eval

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cogniforge/internal/chat"
	"cogniforge/internal/model"
)

var (
	jsonScoreRe = regexp.MustCompile(`(?s)\{[^{}]*"score"\s*:\s*([0-9]*\.?[0-9]+)[^{}]*\}`)
	bareScoreRe = regexp.MustCompile(`(?m)^\s*([01](?:\.\d+)?|0?\.\d+)\s*$`)
)

// judgeWithLLM 用裁判模型打 0–1 分；失败则降级 skip（不拖垮整卷）
func (s *Service) judgeWithLLM(c *model.EvalCase, userInput, actual, fallbackModel string) (status string, score float64, detail map[string]interface{}) {
	detail = map[string]interface{}{"mode": "llm_judge"}
	if s.chatSvc == nil {
		detail["reason"] = "ChatService 未配置，已跳过"
		detail["degraded"] = true
		return model.EvalResultSkip, 0, detail
	}

	judgeModel := fallbackModel
	threshold := 0.6
	if c.JudgeConfig != nil {
		if m, ok := c.JudgeConfig["model"].(string); ok && strings.TrimSpace(m) != "" {
			judgeModel = strings.TrimSpace(m)
		}
		if t, ok := asFloat(c.JudgeConfig["threshold"]); ok {
			threshold = t
		}
	}
	if strings.TrimSpace(judgeModel) == "" {
		detail["reason"] = "未指定裁判模型，已跳过"
		detail["degraded"] = true
		return model.EvalResultSkip, 0, detail
	}
	detail["judge_model"] = judgeModel
	detail["threshold"] = threshold

	rubric := strings.TrimSpace(c.Expected)
	if rubric == "" {
		rubric = "回答应正确、完整、简洁。"
	}
	prompt := fmt.Sprintf(`你是严格的评测裁判。根据「用户问题」「期望/评分标准」「模型回答」打分。
只输出一行 JSON，不要其它文字：{"score":0到1的小数,"reason":"简短中文理由"}

用户问题：
%s

期望/评分标准：
%s

模型回答：
%s
`, userInput, rubric, actual)

	resp, err := s.chatSvc.Chat(&chat.ChatRequest{
		Model: judgeModel,
		Messages: []chat.ChatMessage{
			{Role: "system", Content: "You output only compact JSON for evaluation scores."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		detail["reason"] = "裁判调用失败: " + err.Error()
		detail["degraded"] = true
		return model.EvalResultSkip, 0, detail
	}
	raw := ""
	if resp != nil && len(resp.Choices) > 0 {
		switch v := resp.Choices[0].Message.Content.(type) {
		case string:
			raw = v
		default:
			raw = fmt.Sprint(v)
		}
	}
	detail["judge_raw"] = truncate(raw, 500)

	parsed, reason, ok := parseLLMJudgeScore(raw)
	if !ok {
		detail["reason"] = "无法解析裁判分数，已跳过"
		detail["degraded"] = true
		return model.EvalResultSkip, 0, detail
	}
	detail["llm_score"] = parsed
	if reason != "" {
		detail["llm_reason"] = reason
	}
	if parsed >= threshold {
		return model.EvalResultPass, parsed, detail
	}
	return model.EvalResultFail, parsed, detail
}

func parseLLMJudgeScore(raw string) (score float64, reason string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, "", false
	}
	// 尝试整段 JSON
	var obj struct {
		Score  float64 `json:"score"`
		Reason string  `json:"reason"`
	}
	if err := json.Unmarshal([]byte(raw), &obj); err == nil {
		if obj.Score >= 0 && obj.Score <= 1 {
			return obj.Score, obj.Reason, true
		}
	}
	if m := jsonScoreRe.FindStringSubmatch(raw); len(m) == 2 {
		if f, err := strconv.ParseFloat(m[1], 64); err == nil && f >= 0 && f <= 1 {
			reason = ""
			var wrap struct {
				Reason string `json:"reason"`
			}
			_ = json.Unmarshal([]byte(m[0]), &wrap)
			return f, wrap.Reason, true
		}
	}
	if m := bareScoreRe.FindStringSubmatch(raw); len(m) == 2 {
		if f, err := strconv.ParseFloat(m[1], 64); err == nil && f >= 0 && f <= 1 {
			return f, "", true
		}
	}
	return 0, "", false
}

func asFloat(v interface{}) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
