package eval

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/model"
)

func TestParseLLMJudgeScore_JSON(t *testing.T) {
	sc, reason, ok := parseLLMJudgeScore(`{"score":0.85,"reason":"基本正确"}`)
	require.True(t, ok)
	assert.InDelta(t, 0.85, sc, 1e-9)
	assert.Equal(t, "基本正确", reason)
}

func TestParseLLMJudgeScore_EmbeddedJSON(t *testing.T) {
	sc, _, ok := parseLLMJudgeScore("裁判结果：\n{\"score\": 0.4, \"reason\": \"偏题\"}\n")
	require.True(t, ok)
	assert.InDelta(t, 0.4, sc, 1e-9)
}

func TestParseLLMJudgeScore_Bare(t *testing.T) {
	sc, _, ok := parseLLMJudgeScore("0.9")
	require.True(t, ok)
	assert.InDelta(t, 0.9, sc, 1e-9)
}

func TestParseLLMJudgeScore_Invalid(t *testing.T) {
	_, _, ok := parseLLMJudgeScore("很好")
	assert.False(t, ok)
}

func TestJudgeWithLLM_NoChat_DegradesToSkip(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)
	c := &model.EvalCase{
		Input:       "q",
		Expected:    "应简洁",
		JudgeType:   model.EvalJudgeLLM,
		JudgeConfig: model.JSONBMap{"threshold": 0.5},
	}
	st, sc, d := svc.judgeWithLLM(c, c.Input, "a long answer", "deepseek-chat")
	assert.Equal(t, model.EvalResultSkip, st)
	assert.Equal(t, 0.0, sc)
	assert.Equal(t, true, d["degraded"])
}

func TestCreateAndExecuteRun_LLMJudge_DegradesWithoutChat(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)
	suite, err := svc.CreateSuite("u1", &CreateSuiteRequest{Name: "S"})
	require.NoError(t, err)
	_, err = svc.CreateCase("u1", suite.ID, &CreateCaseRequest{
		Input:     "写一首短诗",
		Expected:  "有意象",
		JudgeType: model.EvalJudgeLLM,
	})
	require.NoError(t, err)

	run, results, err := svc.CreateAndExecuteRun("u1", suite.ID, &CreateRunRequest{
		TargetType: model.EvalTargetModel,
		TargetID:   "m",
	})
	require.NoError(t, err)
	assert.Equal(t, model.EvalRunCompleted, run.Status)
	require.Len(t, results, 1)
	// 无 ChatService：答题 error，不会走到 llm_judge；造一条有输出的路径用 mock 更合适
	assert.Equal(t, model.EvalResultError, results[0].Status)
}
