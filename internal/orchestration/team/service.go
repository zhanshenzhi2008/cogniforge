package team

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"cogniforge/internal/infra/database"
	"cogniforge/internal/model"
)

// Service 小队 CRUD + 成员校验（执行内核后续再接 Chat / 工作流）
type Service struct {
	db *gorm.DB
}

func NewService() *Service {
	return &Service{db: database.DB}
}

func NewServiceWithDB(db *gorm.DB) *Service {
	return &Service{db: db}
}

func normalizeMode(mode string) (string, error) {
	m := strings.TrimSpace(strings.ToLower(mode))
	if m == "" {
		return model.TeamModeSupervisor, nil
	}
	switch m {
	case model.TeamModePipeline, model.TeamModeSupervisor, model.TeamModeHandoff:
		return m, nil
	default:
		return "", fmt.Errorf("mode 必须是 pipeline / supervisor / handoff")
	}
}

func (s *Service) assertAgentsOwned(userID string, ids []string) error {
	seen := map[string]struct{}{}
	var uniq []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	if len(uniq) == 0 {
		return nil
	}
	var count int64
	if err := s.db.Model(&model.Agent{}).
		Where("user_id = ? AND id IN ?", userID, uniq).
		Count(&count).Error; err != nil {
		return fmt.Errorf("校验 Agent 失败")
	}
	if int(count) != len(uniq) {
		return fmt.Errorf("部分 Agent 不存在或不属于当前用户")
	}
	return nil
}

func mergeConfig(cfg model.JSONBMap) model.JSONBMap {
	out := model.TeamConfigDefaults()
	for k, v := range cfg {
		out[k] = v
	}
	return out
}

// List 当前用户小队
func (s *Service) List(userID string) ([]model.AgentTeam, error) {
	var rows []model.AgentTeam
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询小队列表失败")
	}
	return rows, nil
}

// Get 详情
func (s *Service) Get(userID, id string) (*model.AgentTeam, error) {
	var row model.AgentTeam
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("小队不存在")
		}
		return nil, fmt.Errorf("查询小队失败")
	}
	return &row, nil
}

// Create 创建
func (s *Service) Create(userID string, req *CreateTeamRequest) (*model.AgentTeam, error) {
	if req == nil || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("名称不能为空")
	}
	mode, err := normalizeMode(req.Mode)
	if err != nil {
		return nil, err
	}
	members := model.JSONBArray{}
	if req.MemberAgentIDs != nil {
		members = req.MemberAgentIDs
	}
	checkIDs := append([]string{}, members...)
	if sid := strings.TrimSpace(req.SupervisorAgentID); sid != "" {
		checkIDs = append(checkIDs, sid)
	}
	if err := s.assertAgentsOwned(userID, checkIDs); err != nil {
		return nil, err
	}
	if mode == model.TeamModeSupervisor && strings.TrimSpace(req.SupervisorAgentID) == "" {
		return nil, fmt.Errorf("supervisor 模式必须指定 supervisor_agent_id")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	row := model.AgentTeam{
		ID:                newID(),
		UserID:            userID,
		Name:              strings.TrimSpace(req.Name),
		Description:       req.Description,
		Mode:              mode,
		SupervisorAgentID: strings.TrimSpace(req.SupervisorAgentID),
		MemberAgentIDs:    members,
		WorkflowID:        strings.TrimSpace(req.WorkflowID),
		Config:            mergeConfig(req.Config),
		Status:            status,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := s.db.Create(&row).Error; err != nil {
		return nil, fmt.Errorf("创建小队失败")
	}
	return &row, nil
}

// Update 更新
func (s *Service) Update(userID, id string, req *UpdateTeamRequest) (*model.AgentTeam, error) {
	row, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return row, nil
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("名称不能为空")
		}
		row.Name = name
	}
	if req.Description != nil {
		row.Description = *req.Description
	}
	if req.Mode != nil {
		mode, err := normalizeMode(*req.Mode)
		if err != nil {
			return nil, err
		}
		row.Mode = mode
	}
	if req.SupervisorAgentID != nil {
		row.SupervisorAgentID = strings.TrimSpace(*req.SupervisorAgentID)
	}
	if req.MemberAgentIDs != nil {
		row.MemberAgentIDs = *req.MemberAgentIDs
	}
	if req.WorkflowID != nil {
		row.WorkflowID = strings.TrimSpace(*req.WorkflowID)
	}
	if req.Config != nil {
		row.Config = mergeConfig(*req.Config)
	}
	if req.Status != nil {
		st := strings.TrimSpace(*req.Status)
		if st != "" {
			row.Status = st
		}
	}
	checkIDs := append([]string{}, row.MemberAgentIDs...)
	if row.SupervisorAgentID != "" {
		checkIDs = append(checkIDs, row.SupervisorAgentID)
	}
	if err := s.assertAgentsOwned(userID, checkIDs); err != nil {
		return nil, err
	}
	if row.Mode == model.TeamModeSupervisor && row.SupervisorAgentID == "" {
		return nil, fmt.Errorf("supervisor 模式必须指定 supervisor_agent_id")
	}
	row.UpdatedAt = time.Now()
	if err := s.db.Save(row).Error; err != nil {
		return nil, fmt.Errorf("更新小队失败")
	}
	return row, nil
}

// Delete 软删
func (s *Service) Delete(userID, id string) error {
	res := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.AgentTeam{})
	if res.Error != nil {
		return fmt.Errorf("删除小队失败")
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("小队不存在")
	}
	return nil
}

// ExpandResult 一键展开结果
type ExpandResult struct {
	Team     *model.AgentTeam `json:"team"`
	Workflow *model.Workflow  `json:"workflow"`
}

// ExpandOptions 一键展开选项
type ExpandOptions struct {
	Layout string // 空=按 mode；parallel=工人扇出汇合；linear=强制线性；team=强制 team 节点
}

// ExpandToWorkflow 从小队生成工作流图并回写 workflow_id。
// 默认：pipeline → 线性链；supervisor/handoff → team 节点。layout=parallel → 扇出图。
func (s *Service) ExpandToWorkflow(userID, teamID string, opts *ExpandOptions) (*ExpandResult, error) {
	t, err := s.Get(userID, teamID)
	if err != nil {
		return nil, err
	}

	layout := ""
	if opts != nil {
		layout = strings.TrimSpace(strings.ToLower(opts.Layout))
	}

	idsForNames := pipelineAgentOrder(t)
	if layout == "parallel" {
		idsForNames = fanoutWorkerIDs(t)
	}
	names, err := s.agentNameMap(userID, idsForNames)
	if err != nil {
		return nil, err
	}

	var def map[string]any
	switch layout {
	case "parallel":
		def, err = BuildParallelFanoutDefinition(t, names)
	case "linear":
		def, err = BuildLinearPipelineDefinition(t, names)
	case "team":
		def, err = BuildTeamNodeDefinition(t)
	default:
		if t.Mode == model.TeamModePipeline {
			def, err = BuildLinearPipelineDefinition(t, names)
		} else {
			def, err = BuildTeamNodeDefinition(t)
		}
	}
	if err != nil {
		return nil, err
	}

	defJSON, err := json.Marshal(def)
	if err != nil {
		return nil, fmt.Errorf("序列化工作流失败")
	}

	wfName := strings.TrimSpace(t.Name) + " · 流水线"
	desc := "由小队「" + t.Name + "」一键生成"
	switch layout {
	case "parallel":
		wfName = strings.TrimSpace(t.Name) + " · 并行扇出"
		desc = "由小队「" + t.Name + "」一键生成（工人并行，最多 8）"
	case "team":
		wfName = strings.TrimSpace(t.Name) + " · 小队图"
	default:
		if t.Mode != model.TeamModePipeline && layout != "linear" {
			wfName = strings.TrimSpace(t.Name) + " · 小队图"
		}
	}

	var wf model.Workflow
	if t.WorkflowID != "" {
		err = s.db.Where("id = ? AND user_id = ?", t.WorkflowID, userID).First(&wf).Error
		if err == nil {
			if err := s.db.Model(&wf).Updates(map[string]any{
				"name":        wfName,
				"description": desc,
				"definition":  string(defJSON),
				"version":     wf.Version + 1,
			}).Error; err != nil {
				return nil, fmt.Errorf("更新工作流失败")
			}
			_ = s.db.Where("id = ?", wf.ID).First(&wf)
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("查询工作流失败")
		}
	}

	if wf.ID == "" {
		wf = model.Workflow{
			ID:          uuid.New().String(),
			UserID:      userID,
			Name:        wfName,
			Description: desc,
			Definition:  string(defJSON),
			Status:      "draft",
			Version:     1,
		}
		if err := s.db.Create(&wf).Error; err != nil {
			return nil, fmt.Errorf("创建工作流失败")
		}
		t.WorkflowID = wf.ID
		t.UpdatedAt = time.Now()
		if err := s.db.Model(t).Updates(map[string]any{
			"workflow_id": t.WorkflowID,
			"updated_at":  t.UpdatedAt,
		}).Error; err != nil {
			return nil, fmt.Errorf("回写小队 workflow_id 失败")
		}
	}

	return &ExpandResult{Team: t, Workflow: &wf}, nil
}

func (s *Service) agentNameMap(userID string, ids []string) (map[string]string, error) {
	out := map[string]string{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []model.Agent
	if err := s.db.Where("user_id = ? AND id IN ?", userID, ids).
		Select("id", "name").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("加载 Agent 名称失败")
	}
	for _, a := range rows {
		out[a.ID] = a.Name
	}
	return out, nil
}

// InitRunState 根据小队配置生成 RunState（供 chat/run 与 checkpoint 使用）
func InitRunState(t *model.AgentTeam, goal string) (*RunState, error) {
	if t == nil {
		return nil, fmt.Errorf("team is nil")
	}
	active := strings.TrimSpace(t.SupervisorAgentID)
	if active == "" && len(t.MemberAgentIDs) > 0 {
		active = t.MemberAgentIDs[0]
	}
	if active == "" {
		return nil, fmt.Errorf("小队没有可用的入口 Agent")
	}
	maxDepth := 8
	maxSteps := 40
	if t.Config != nil {
		if v, ok := asInt(t.Config["max_depth"]); ok {
			maxDepth = v
		}
		if v, ok := asInt(t.Config["max_steps"]); ok {
			maxSteps = v
		}
	}
	return NewRunState(t.ID, active, goal, maxDepth, maxSteps), nil
}

func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}
