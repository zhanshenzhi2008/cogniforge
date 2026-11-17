package eval

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"cogniforge/internal/model"
)

func TestJudge_ContainsAny(t *testing.T) {
	c := &model.EvalCase{
		Expected:  "北京,天气",
		JudgeType: model.EvalJudgeContains,
	}
	st, sc, d := Judge(c, "今天北京多云")
	assert.Equal(t, model.EvalResultPass, st)
	assert.Equal(t, 1.0, sc)
	assert.Contains(t, d["hit"], "北京")
}

func TestJudge_ContainsAllFail(t *testing.T) {
	c := &model.EvalCase{
		Expected:    "北京,晴天",
		JudgeType:   model.EvalJudgeContains,
		JudgeConfig: model.JSONBMap{"match": "all"},
	}
	st, sc, _ := Judge(c, "北京多云")
	assert.Equal(t, model.EvalResultFail, st)
	assert.Equal(t, 0.0, sc)
}

func TestJudge_Exact(t *testing.T) {
	c := &model.EvalCase{Expected: "ok", JudgeType: model.EvalJudgeExact}
	st, _, _ := Judge(c, " ok ")
	assert.Equal(t, model.EvalResultPass, st)
	st, _, _ = Judge(c, "no")
	assert.Equal(t, model.EvalResultFail, st)
}

func TestJudge_LLMSkip(t *testing.T) {
	c := &model.EvalCase{JudgeType: model.EvalJudgeLLM}
	st, _, _ := Judge(c, "anything")
	assert.Equal(t, model.EvalResultSkip, st)
}
