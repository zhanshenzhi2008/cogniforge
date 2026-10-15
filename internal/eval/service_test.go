package eval

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"cogniforge/internal/model"
)

func setupEvalDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.EvalSuite{},
		&model.EvalCase{},
		&model.EvalRun{},
		&model.EvalResult{},
		&model.Agent{},
		&model.Prompt{},
		&model.PromptVersion{},
	))
	return db
}

func TestSuiteCRUD_AndUserIsolation(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)

	a, err := svc.CreateSuite("u1", &CreateSuiteRequest{Name: "A", Description: "da"})
	require.NoError(t, err)
	_, err = svc.CreateSuite("u2", &CreateSuiteRequest{Name: "B"})
	require.NoError(t, err)

	list1, err := svc.ListSuites("u1")
	require.NoError(t, err)
	require.Len(t, list1, 1)
	assert.Equal(t, "A", list1[0].Name)

	_, err = svc.GetSuite("u2", a.ID)
	require.Error(t, err)

	name := "A2"
	updated, err := svc.UpdateSuite("u1", a.ID, &UpdateSuiteRequest{Name: &name})
	require.NoError(t, err)
	assert.Equal(t, "A2", updated.Name)

	require.NoError(t, svc.DeleteSuite("u1", a.ID))
	list1, err = svc.ListSuites("u1")
	require.NoError(t, err)
	assert.Empty(t, list1)
}

func TestCaseCRUD_AndCascadeDelete(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)

	suite, err := svc.CreateSuite("u1", &CreateSuiteRequest{Name: "S"})
	require.NoError(t, err)

	c, err := svc.CreateCase("u1", suite.ID, &CreateCaseRequest{
		Input:     "你好",
		Expected:  "你好,hello",
		JudgeType: model.EvalJudgeContains,
	})
	require.NoError(t, err)
	assert.Equal(t, model.EvalJudgeContains, c.JudgeType)

	cases, err := svc.ListCases("u1", suite.ID)
	require.NoError(t, err)
	require.Len(t, cases, 1)

	exp := "ok"
	_, err = svc.UpdateCase("u1", c.ID, &UpdateCaseRequest{Expected: &exp})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteSuite("u1", suite.ID))
	var n int64
	require.NoError(t, db.Model(&model.EvalCase{}).Count(&n).Error)
	assert.Equal(t, int64(0), n)
}

func TestCreateCase_RejectsBadJudge(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)
	suite, err := svc.CreateSuite("u1", &CreateSuiteRequest{Name: "S"})
	require.NoError(t, err)
	_, err = svc.CreateCase("u1", suite.ID, &CreateCaseRequest{
		Input:     "x",
		JudgeType: "magic",
	})
	require.Error(t, err)
}

func TestCreateAndExecuteRun_WithoutChatService_MarksErrors(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)
	suite, err := svc.CreateSuite("u1", &CreateSuiteRequest{Name: "S"})
	require.NoError(t, err)
	_, err = svc.CreateCase("u1", suite.ID, &CreateCaseRequest{
		Input:    "首都是哪里",
		Expected: "北京",
	})
	require.NoError(t, err)

	run, results, err := svc.CreateAndExecuteRun("u1", suite.ID, &CreateRunRequest{
		TargetType: model.EvalTargetModel,
		TargetID:   "deepseek-chat",
	})
	require.NoError(t, err)
	require.NotNil(t, run)
	assert.Equal(t, model.EvalRunCompleted, run.Status)
	require.Len(t, results, 1)
	assert.Equal(t, model.EvalResultError, results[0].Status)
	assert.Equal(t, 1, run.FailCount)
	assert.Equal(t, 0, run.PassCount)

	got, gotResults, err := svc.GetRun("u1", run.ID)
	require.NoError(t, err)
	assert.Equal(t, run.ID, got.ID)
	require.Len(t, gotResults, 1)
}

func TestCreateAndExecuteRun_EmptySuite(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)
	suite, err := svc.CreateSuite("u1", &CreateSuiteRequest{Name: "S"})
	require.NoError(t, err)
	_, _, err = svc.CreateAndExecuteRun("u1", suite.ID, &CreateRunRequest{
		TargetType: model.EvalTargetModel,
		TargetID:   "m",
	})
	require.Error(t, err)
}

func TestCreateAndExecuteRun_WithPassViaFakeOutput(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)
	suite, err := svc.CreateSuite("u1", &CreateSuiteRequest{Name: "S"})
	require.NoError(t, err)
	_, err = svc.CreateCase("u1", suite.ID, &CreateCaseRequest{
		Input:     "ping",
		Expected:  "",
		JudgeType: model.EvalJudgeContains,
	})
	require.NoError(t, err)

	run, results, err := svc.CreateAndExecuteRun("u1", suite.ID, &CreateRunRequest{
		TargetType: model.EvalTargetModel,
		TargetID:   "m",
	})
	require.NoError(t, err)
	assert.Equal(t, model.EvalRunCompleted, run.Status)
	require.Len(t, results, 1)
}

func TestCreateAndExecuteRun_AgentMissing(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)
	suite, err := svc.CreateSuite("u1", &CreateSuiteRequest{Name: "S"})
	require.NoError(t, err)
	_, err = svc.CreateCase("u1", suite.ID, &CreateCaseRequest{Input: "hi", Expected: "hi"})
	require.NoError(t, err)

	_, _, err = svc.CreateAndExecuteRun("u1", suite.ID, &CreateRunRequest{
		TargetType: model.EvalTargetAgent,
		TargetID:   "no-such-agent",
	})
	require.Error(t, err)
}

func TestCreateAndExecuteRun_SystemPromptOverride(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)
	suite, err := svc.CreateSuite("u1", &CreateSuiteRequest{Name: "S"})
	require.NoError(t, err)
	_, err = svc.CreateCase("u1", suite.ID, &CreateCaseRequest{Input: "hi", Expected: "x"})
	require.NoError(t, err)

	run, _, err := svc.CreateAndExecuteRun("u1", suite.ID, &CreateRunRequest{
		TargetType:   model.EvalTargetModel,
		TargetID:     "deepseek-chat",
		SystemPrompt: "你是严格裁判",
	})
	require.NoError(t, err)
	assert.Equal(t, "你是严格裁判", run.SystemPrompt)
	assert.Equal(t, "deepseek-chat", run.Model)
}

func TestCreateAndExecuteRun_PromptVersion(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)
	ver := &model.PromptVersion{
		ID:       "pv1",
		PromptID: "p1",
		UserID:   "u1",
		Version:  1,
		Content:  "提示词正文 A",
	}
	require.NoError(t, db.Create(&model.Prompt{
		ID: "p1", UserID: "u1", Name: "N", CurrentVersionID: "pv1",
	}).Error)
	require.NoError(t, db.Create(ver).Error)

	suite, err := svc.CreateSuite("u1", &CreateSuiteRequest{Name: "S"})
	require.NoError(t, err)
	_, err = svc.CreateCase("u1", suite.ID, &CreateCaseRequest{Input: "q", Expected: "a"})
	require.NoError(t, err)

	_, _, err = svc.CreateAndExecuteRun("u1", suite.ID, &CreateRunRequest{
		TargetType: model.EvalTargetPrompt,
		TargetID:   "pv1",
	})
	require.Error(t, err) // missing model

	run, _, err := svc.CreateAndExecuteRun("u1", suite.ID, &CreateRunRequest{
		TargetType: model.EvalTargetPrompt,
		TargetID:   "pv1",
		Model:      "deepseek-chat",
	})
	require.NoError(t, err)
	assert.Equal(t, model.EvalTargetPrompt, run.TargetType)
	assert.Equal(t, "提示词正文 A", run.SystemPrompt)
	assert.Equal(t, "deepseek-chat", run.Model)
}

func TestCreateAndExecuteRun_AgentExists_UsesAgentModel(t *testing.T) {
	db := setupEvalDB(t)
	svc := NewService(db, nil)
	require.NoError(t, db.Create(&model.Agent{
		ID:           "ag1",
		UserID:       "u1",
		Name:         "Helper",
		Model:        "agent-model",
		SystemPrompt: "be short",
	}).Error)

	suite, err := svc.CreateSuite("u1", &CreateSuiteRequest{Name: "S"})
	require.NoError(t, err)
	_, err = svc.CreateCase("u1", suite.ID, &CreateCaseRequest{Input: "hi", Expected: "x"})
	require.NoError(t, err)

	run, results, err := svc.CreateAndExecuteRun("u1", suite.ID, &CreateRunRequest{
		TargetType: model.EvalTargetAgent,
		TargetID:   "ag1",
	})
	require.NoError(t, err)
	assert.Equal(t, model.EvalTargetAgent, run.TargetType)
	assert.Equal(t, "ag1", run.TargetID)
	assert.Equal(t, model.EvalRunCompleted, run.Status)
	require.Len(t, results, 1)
	assert.Equal(t, model.EvalResultError, results[0].Status) // no ChatService
}
