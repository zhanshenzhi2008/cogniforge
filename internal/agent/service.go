package agent

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

type AgentService struct {
	db *gorm.DB
}

func NewAgentService() *AgentService {
	return &AgentService{db: database.DB}
}

// ListAgents 获取 Agent 列表
func (s *AgentService) ListAgents(userID string) ([]model.Agent, error) {
	var agents []model.Agent
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&agents).Error; err != nil {
		return nil, fmt.Errorf("查询 Agent 列表失败")
	}
	return agents, nil
}

// CreateAgent 创建 Agent
func (s *AgentService) CreateAgent(userID string, req *CreateAgentRequest) (*model.Agent, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("请选择模型")
	}

	tools := model.JSONBArray{}
	if req.Tools != nil {
		tools = req.Tools
	}

	agent := model.Agent{
		ID:           generateID(),
		UserID:       userID,
		Name:         req.Name,
		Description:  req.Description,
		Model:        req.Model,
		SystemPrompt: req.SystemPrompt,
		Tools:        tools,
		MemoryType:   "short_term",
		MemoryTurns:  10,
		InputFilter:  true,
		OutputFilter: true,
		Status:       "active",
		Metadata:     model.JSONBMap{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(&agent).Error; err != nil {
		return nil, fmt.Errorf("创建 Agent 失败")
	}

	return &agent, nil
}

// GetAgent 获取 Agent 详情
func (s *AgentService) GetAgent(userID, agentID string) (*model.Agent, error) {
	var agent model.Agent
	if err := s.db.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Agent 不存在")
		}
		return nil, fmt.Errorf("查询 Agent 失败")
	}
	return &agent, nil
}

// UpdateAgent 更新 Agent
func (s *AgentService) UpdateAgent(userID, agentID string, req *UpdateAgentRequest) (*model.Agent, error) {
	var agent model.Agent
	if err := s.db.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Agent 不存在")
		}
		return nil, fmt.Errorf("查询 Agent 失败")
	}

	if req.Name != "" {
		agent.Name = req.Name
	}
	if req.Description != "" {
		agent.Description = req.Description
	}
	if req.Model != "" {
		agent.Model = req.Model
	}
	if req.SystemPrompt != "" {
		agent.SystemPrompt = req.SystemPrompt
	}
	if req.Tools != nil {
		agent.Tools = req.Tools
	}
	if req.Status != "" {
		agent.Status = req.Status
	}
	agent.UpdatedAt = time.Now()

	if err := s.db.Save(&agent).Error; err != nil {
		return nil, fmt.Errorf("更新 Agent 失败")
	}

	return &agent, nil
}

// DeleteAgent 删除 Agent
func (s *AgentService) DeleteAgent(userID, agentID string) error {
	var agent model.Agent
	if err := s.db.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("Agent 不存在")
		}
		return fmt.Errorf("查询 Agent 失败")
	}

	if err := s.db.Delete(&agent).Error; err != nil {
		return fmt.Errorf("删除 Agent 失败")
	}
	return nil
}

// generateID 生成唯一ID
func generateID() string {
	return newID()
}
