package skill

import (
	"fmt"
	"time"

	"cogniforge/internal/model"
	"gorm.io/gorm"
)

// Service SKILL 业务逻辑
type Service struct {
	db *gorm.DB
}

// NewService 创建 Skill Service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ListSkills 列出当前用户可用的 SKILL（内置 + 自己创建的）
func (s *Service) ListSkills(userID string) ([]model.Skill, error) {
	var skills []model.Skill
	for _, b := range model.BuiltInSkills {
		b := b
		skills = append(skills, b)
	}
	var custom []model.Skill
	if err := s.db.Where("user_id = ? AND is_built_in = ?", userID, false).Order("sort_order ASC, created_at DESC").Find(&custom).Error; err != nil {
		return nil, fmt.Errorf("查询 SKILL 列表失败")
	}
	skills = append(skills, custom...)
	return skills, nil
}

// GetSkill 获取 SKILL 详情
func (s *Service) GetSkill(userID, skillID string) (*model.Skill, error) {
	if b := model.GetBuiltInSkill(skillID); b != nil {
		return b, nil
	}
	var skill model.Skill
	if err := s.db.Where("id = ? AND user_id = ?", skillID, userID).First(&skill).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("SKILL 不存在")
		}
		return nil, fmt.Errorf("查询 SKILL 失败")
	}
	return &skill, nil
}

// CreateSkill 创建自定义 SKILL
func (s *Service) CreateSkill(userID string, req *CreateSkillRequest) (*model.Skill, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}
	skill := model.Skill{
		ID:           fmt.Sprintf("skill_%d", time.Now().UnixNano()),
		UserID:       userID,
		Name:         req.Name,
		Description:  req.Description,
		Icon:         req.Icon,
		Instructions: req.Instructions,
		Examples:     req.Examples,
		References:   req.References,
		Constraints:  req.Constraints,
		Model:        req.Model,
		McpServers:   req.McpServers,
		MemoryType:   req.MemoryType,
		MemoryTurns:  req.MemoryTurns,
		Category:     req.Category,
		Tags:         req.Tags,
		Version:      req.Version,
		Author:       req.Author,
		IsBuiltIn:    false,
		SortOrder:    req.SortOrder,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if skill.MemoryType == "" {
		skill.MemoryType = "short_term"
	}
	if skill.MemoryTurns == 0 {
		skill.MemoryTurns = 10
	}
	if err := s.db.Create(&skill).Error; err != nil {
		return nil, fmt.Errorf("创建 SKILL 失败")
	}
	return &skill, nil
}

// UpdateSkill 更新 SKILL
func (s *Service) UpdateSkill(userID, skillID string, req *UpdateSkillRequest) (*model.Skill, error) {
	if model.IsBuiltInSkill(skillID) {
		return nil, fmt.Errorf("内置 SKILL 不可编辑")
	}
	var skill model.Skill
	if err := s.db.Where("id = ? AND user_id = ?", skillID, userID).First(&skill).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("SKILL 不存在")
		}
		return nil, fmt.Errorf("查询 SKILL 失败")
	}
	if req.Name != nil && *req.Name != "" {
		skill.Name = *req.Name
	}
	if req.Description != nil {
		skill.Description = *req.Description
	}
	if req.Icon != nil {
		skill.Icon = *req.Icon
	}
	if req.Instructions != nil {
		skill.Instructions = *req.Instructions
	}
	if req.Examples != nil {
		skill.Examples = req.Examples
	}
	if req.References != nil {
		skill.References = req.References
	}
	if req.Constraints != nil {
		skill.Constraints = req.Constraints
	}
	if req.Model != nil {
		skill.Model = *req.Model
	}
	if req.McpServers != nil {
		skill.McpServers = req.McpServers
	}
	if req.MemoryType != nil {
		skill.MemoryType = *req.MemoryType
	}
	if req.MemoryTurns != nil {
		skill.MemoryTurns = *req.MemoryTurns
	}
	if req.Category != nil {
		skill.Category = *req.Category
	}
	if req.Tags != nil {
		skill.Tags = req.Tags
	}
	if req.SortOrder != nil {
		skill.SortOrder = *req.SortOrder
	}
	skill.UpdatedAt = time.Now()
	if err := s.db.Save(&skill).Error; err != nil {
		return nil, fmt.Errorf("更新 SKILL 失败")
	}
	return &skill, nil
}

// DeleteSkill 删除 SKILL（只能删自定义的）
func (s *Service) DeleteSkill(userID, skillID string) error {
	if model.IsBuiltInSkill(skillID) {
		return fmt.Errorf("内置 SKILL 不可删除")
	}
	var skill model.Skill
	if err := s.db.Where("id = ? AND user_id = ?", skillID, userID).First(&skill).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("SKILL 不存在")
		}
		return fmt.Errorf("查询 SKILL 失败")
	}
	if err := s.db.Delete(&skill).Error; err != nil {
		return fmt.Errorf("删除 SKILL 失败")
	}
	return nil
}

// ============ 请求结构 ============

type CreateSkillRequest struct {
	Name         string            `json:"name" binding:"required"`
	Description  string            `json:"description"`
	Icon         string            `json:"icon"`
	Instructions string            `json:"instructions"` // 核心指令
	Examples     model.JSONBArray `json:"examples"`
	References   model.JSONBArray `json:"references"`
	Constraints  model.JSONBArray `json:"constraints"`
	Model        string            `json:"model"`
	McpServers   model.JSONBArray `json:"mcp_servers"`
	MemoryType   string            `json:"memory_type"`
	MemoryTurns  int              `json:"memory_turns"`
	Category     string            `json:"category"`
	Tags         model.JSONBArray `json:"tags"`
	Version      string            `json:"version"`
	Author       string            `json:"author"`
	SortOrder    int              `json:"sort_order"`
}

type UpdateSkillRequest struct {
	Name         *string           `json:"name"`
	Description  *string           `json:"description"`
	Icon         *string           `json:"icon"`
	Instructions *string           `json:"instructions"`
	Examples     model.JSONBArray `json:"examples"`
	References   model.JSONBArray `json:"references"`
	Constraints  model.JSONBArray `json:"constraints"`
	Model        *string           `json:"model"`
	McpServers   model.JSONBArray `json:"mcp_servers"`
	MemoryType   *string           `json:"memory_type"`
	MemoryTurns  *int              `json:"memory_turns"`
	Category     *string           `json:"category"`
	Tags         model.JSONBArray `json:"tags"`
	SortOrder    *int              `json:"sort_order"`
}
