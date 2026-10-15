package eval

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"cogniforge/internal/chat"
	"cogniforge/internal/mcp"
	"cogniforge/internal/model"
)

type Service struct {
	db      *gorm.DB
	chatSvc *chat.ChatService
	mcpReg  *mcp.Registry
}

func NewService(db *gorm.DB, chatSvc *chat.ChatService) *Service {
	return &Service{db: db, chatSvc: chatSvc}
}

// SetMCPRegistry 注入 MCP Registry，使 target=agent 时可跑工具循环
func (s *Service) SetMCPRegistry(reg *mcp.Registry) {
	s.mcpReg = reg
}

func (s *Service) ListSuites(userID string) ([]model.EvalSuite, error) {
	var list []model.EvalSuite
	err := s.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&list).Error
	return list, err
}

func (s *Service) CreateSuite(userID string, req *CreateSuiteRequest) (*model.EvalSuite, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name 不能为空")
	}
	suite := &model.EvalSuite{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        name,
		Description: req.Description,
	}
	if err := s.db.Create(suite).Error; err != nil {
		return nil, err
	}
	return suite, nil
}

func (s *Service) GetSuite(userID, id string) (*model.EvalSuite, error) {
	var suite model.EvalSuite
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&suite).Error; err != nil {
		return nil, errors.New("评测集不存在")
	}
	return &suite, nil
}

func (s *Service) UpdateSuite(userID, id string, req *UpdateSuiteRequest) (*model.EvalSuite, error) {
	suite, err := s.GetSuite(userID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		n := strings.TrimSpace(*req.Name)
		if n == "" {
			return nil, errors.New("name 不能为空")
		}
		suite.Name = n
	}
	if req.Description != nil {
		suite.Description = *req.Description
	}
	if err := s.db.Save(suite).Error; err != nil {
		return nil, err
	}
	return suite, nil
}

func (s *Service) DeleteSuite(userID, id string) error {
	suite, err := s.GetSuite(userID, id)
	if err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var runIDs []string
		if err := tx.Model(&model.EvalRun{}).Where("suite_id = ?", suite.ID).Pluck("id", &runIDs).Error; err != nil {
			return err
		}
		if len(runIDs) > 0 {
			if err := tx.Where("run_id IN ?", runIDs).Delete(&model.EvalResult{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("suite_id = ?", suite.ID).Delete(&model.EvalRun{}).Error; err != nil {
			return err
		}
		if err := tx.Where("suite_id = ?", suite.ID).Delete(&model.EvalCase{}).Error; err != nil {
			return err
		}
		return tx.Delete(suite).Error
	})
}

func (s *Service) ListCases(userID, suiteID string) ([]model.EvalCase, error) {
	if _, err := s.GetSuite(userID, suiteID); err != nil {
		return nil, err
	}
	var list []model.EvalCase
	err := s.db.Where("suite_id = ? AND user_id = ?", suiteID, userID).
		Order("sort_order ASC, created_at ASC").Find(&list).Error
	return list, err
}

func (s *Service) CreateCase(userID, suiteID string, req *CreateCaseRequest) (*model.EvalCase, error) {
	if _, err := s.GetSuite(userID, suiteID); err != nil {
		return nil, err
	}
	jt := req.JudgeType
	if jt == "" {
		jt = model.EvalJudgeContains
	}
	if jt != model.EvalJudgeContains && jt != model.EvalJudgeExact && jt != model.EvalJudgeLLM {
		return nil, fmt.Errorf("不支持的 judge_type: %s", jt)
	}
	weight := 1.0
	if req.Weight != nil && *req.Weight > 0 {
		weight = *req.Weight
	}
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	c := &model.EvalCase{
		ID:          uuid.New().String(),
		SuiteID:     suiteID,
		UserID:      userID,
		Name:        req.Name,
		Input:       strings.TrimSpace(req.Input),
		Expected:    req.Expected,
		JudgeType:   jt,
		JudgeConfig: model.JSONBMap(req.JudgeConfig),
		Weight:      weight,
		SortOrder:   sortOrder,
	}
	if c.Input == "" {
		return nil, errors.New("input 不能为空")
	}
	if err := s.db.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) UpdateCase(userID, caseID string, req *UpdateCaseRequest) (*model.EvalCase, error) {
	var c model.EvalCase
	if err := s.db.Where("id = ? AND user_id = ?", caseID, userID).First(&c).Error; err != nil {
		return nil, errors.New("用例不存在")
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Input != nil {
		in := strings.TrimSpace(*req.Input)
		if in == "" {
			return nil, errors.New("input 不能为空")
		}
		c.Input = in
	}
	if req.Expected != nil {
		c.Expected = *req.Expected
	}
	if req.JudgeType != nil {
		jt := *req.JudgeType
		if jt != model.EvalJudgeContains && jt != model.EvalJudgeExact && jt != model.EvalJudgeLLM {
			return nil, fmt.Errorf("不支持的 judge_type: %s", jt)
		}
		c.JudgeType = jt
	}
	if req.JudgeConfig != nil {
		c.JudgeConfig = model.JSONBMap(req.JudgeConfig)
	}
	if req.Weight != nil && *req.Weight > 0 {
		c.Weight = *req.Weight
	}
	if req.SortOrder != nil {
		c.SortOrder = *req.SortOrder
	}
	if err := s.db.Save(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) DeleteCase(userID, caseID string) error {
	res := s.db.Where("id = ? AND user_id = ?", caseID, userID).Delete(&model.EvalCase{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("用例不存在")
	}
	return nil
}

func (s *Service) ListRuns(userID, suiteID string) ([]model.EvalRun, error) {
	q := s.db.Where("user_id = ?", userID)
	if suiteID != "" {
		q = q.Where("suite_id = ?", suiteID)
	}
	var list []model.EvalRun
	err := q.Order("created_at DESC").Limit(100).Find(&list).Error
	return list, err
}

func (s *Service) GetRun(userID, runID string) (*model.EvalRun, []model.EvalResult, error) {
	var run model.EvalRun
	if err := s.db.Where("id = ? AND user_id = ?", runID, userID).First(&run).Error; err != nil {
		return nil, nil, errors.New("运行记录不存在")
	}
	var results []model.EvalResult
	if err := s.db.Where("run_id = ?", run.ID).Order("created_at ASC").Find(&results).Error; err != nil {
		return nil, nil, err
	}
	return &run, results, nil
}

// CreateAndExecuteRun 同步跑完整卷（一期）
func (s *Service) CreateAndExecuteRun(userID, suiteID string, req *CreateRunRequest) (*model.EvalRun, []model.EvalResult, error) {
	if _, err := s.GetSuite(userID, suiteID); err != nil {
		return nil, nil, err
	}
	tt := req.TargetType
	if tt != model.EvalTargetModel && tt != model.EvalTargetAgent && tt != model.EvalTargetPrompt {
		return nil, nil, fmt.Errorf("不支持的 target_type: %s", tt)
	}
	if strings.TrimSpace(req.TargetID) == "" {
		return nil, nil, errors.New("target_id 不能为空")
	}
	if tt == model.EvalTargetPrompt && strings.TrimSpace(req.Model) == "" {
		return nil, nil, errors.New("target_type=prompt 时必须传 model")
	}
	cases, err := s.ListCases(userID, suiteID)
	if err != nil {
		return nil, nil, err
	}
	if len(cases) == 0 {
		return nil, nil, errors.New("评测集没有用例")
	}
	if len(cases) > 50 {
		return nil, nil, errors.New("一期单次最多 50 条用例")
	}

	systemPrompt := "You are a helpful assistant for evaluation. Answer concisely."
	modelName := strings.TrimSpace(req.TargetID)
	var toolServerIDs []string
	if tt == model.EvalTargetAgent {
		var ag struct {
			Model        string
			SystemPrompt string
			Tools        model.JSONBArray
		}
		if err := s.db.Table("agents").
			Select("model, system_prompt, tools").
			Where("id = ? AND user_id = ?", strings.TrimSpace(req.TargetID), userID).
			Take(&ag).Error; err != nil {
			return nil, nil, errors.New("Agent 不存在")
		}
		if ag.Model != "" {
			modelName = ag.Model
		}
		if ag.SystemPrompt != "" {
			systemPrompt = ag.SystemPrompt
		}
		toolServerIDs = []string(ag.Tools)
	}
	if tt == model.EvalTargetPrompt {
		var ver model.PromptVersion
		if err := s.db.Where("id = ? AND user_id = ?", strings.TrimSpace(req.TargetID), userID).
			First(&ver).Error; err != nil {
			return nil, nil, errors.New("提示词版本不存在")
		}
		systemPrompt = ver.Content
		modelName = strings.TrimSpace(req.Model)
	}
	if strings.TrimSpace(req.Model) != "" && tt != model.EvalTargetPrompt {
		modelName = strings.TrimSpace(req.Model)
	}
	if override := strings.TrimSpace(req.SystemPrompt); override != "" {
		systemPrompt = override
	}

	now := time.Now()
	run := &model.EvalRun{
		ID:           uuid.New().String(),
		SuiteID:      suiteID,
		UserID:       userID,
		TargetType:   tt,
		TargetID:     strings.TrimSpace(req.TargetID),
		Model:        modelName,
		SystemPrompt: systemPrompt,
		Status:       model.EvalRunRunning,
		StartedAt:    &now,
	}
	if err := s.db.Create(run).Error; err != nil {
		return nil, nil, err
	}

	results := make([]model.EvalResult, 0, len(cases))
	var passN, failN, skipN int
	var weightedPass, weightedTotal float64

	for _, c := range cases {
		start := time.Now()
		actual, callErr := s.callCase(modelName, systemPrompt, c.Input, toolServerIDs)
		latency := time.Since(start).Milliseconds()

		var status string
		var score float64
		var detail map[string]interface{}
		if callErr != nil {
			status = model.EvalResultError
			score = 0
			detail = map[string]interface{}{"error": callErr.Error()}
			actual = ""
			failN++
		} else {
			status, score, detail = Judge(&c, actual)
			switch status {
			case model.EvalResultPass:
				passN++
			case model.EvalResultSkip:
				skipN++
			default:
				failN++
			}
		}
		w := c.Weight
		if w <= 0 {
			w = 1
		}
		if status != model.EvalResultSkip {
			weightedTotal += w
			weightedPass += w * score
		}
		r := model.EvalResult{
			ID:           uuid.New().String(),
			RunID:        run.ID,
			CaseID:       c.ID,
			Status:       status,
			Score:        score,
			ActualOutput: actual,
			LatencyMS:    latency,
			Detail:       model.JSONBMap(detail),
		}
		if err := s.db.Create(&r).Error; err != nil {
			s.failRun(run, err.Error())
			return run, results, err
		}
		results = append(results, r)
	}

	fin := time.Now()
	run.PassCount = passN
	run.FailCount = failN
	run.SkipCount = skipN
	if weightedTotal > 0 {
		run.Score = weightedPass / weightedTotal
	}
	run.Status = model.EvalRunCompleted
	run.FinishedAt = &fin
	_ = s.db.Save(run).Error
	return run, results, nil
}

func (s *Service) callCase(modelName, systemPrompt, userInput string, toolServerIDs []string) (string, error) {
	if s.chatSvc == nil {
		return "", errors.New("ChatService 未配置")
	}
	req := &chat.ChatRequest{
		Model: modelName,
		Messages: []chat.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userInput},
		},
	}
	chat.InjectMCPTools(req, s.mcpReg, toolServerIDs)
	var (
		resp *chat.ChatResponse
		err  error
	)
	if len(req.Tools) > 0 {
		resp, err = s.chatSvc.ChatWithTools(req, chat.NewMCPToolExecutor(s.mcpReg))
	} else {
		resp, err = s.chatSvc.Chat(req)
	}
	if err != nil {
		return "", err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return "", errors.New("模型无返回")
	}
	switch v := resp.Choices[0].Message.Content.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprint(v), nil
	}
}

func (s *Service) failRun(run *model.EvalRun, msg string) {
	fin := time.Now()
	run.Status = model.EvalRunFailed
	run.ErrorMessage = msg
	run.FinishedAt = &fin
	_ = s.db.Save(run).Error
}
