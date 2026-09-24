package mcp

import (
	"fmt"
	"time"

	"cogniforge/internal/model"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

// Service MCP 业务逻辑
type Service struct {
	db *gorm.DB
}

// NewService 创建 MCP Service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ListServers 列出当前用户的 MCP Server（内置 + 用户自定义）
func (s *Service) ListServers(userID string) ([]model.McpServer, error) {
	var servers []model.McpServer
	// 查用户自定义的
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("查询 MCP Server 列表失败")
	}
	return servers, nil
}

// CreateServer 创建 MCP Server
func (s *Service) CreateServer(userID string, req *CreateServerRequest) (*model.McpServer, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}
	if req.Type == "" {
		req.Type = model.McpServerTypeHTTP
	}
	if req.Type != model.McpServerTypeBuiltIn && req.URL == "" {
		return nil, fmt.Errorf("URL 不能为空")
	}

	srv := model.McpServer{
		ID:         uuid.New().String(),
		UserID:     userID,
		Name:       req.Name,
		Type:       req.Type,
		URL:        req.URL,
		AuthHeader: req.AuthHeader,
		Enabled:    true,
		Metadata:   req.Metadata,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.db.Create(&srv).Error; err != nil {
		return nil, fmt.Errorf("创建 MCP Server 失败")
	}
	return &srv, nil
}

// GetServer 获取 MCP Server 详情
func (s *Service) GetServer(userID, serverID string) (*model.McpServer, error) {
	// 内置的直接返回
	if model.IsBuiltInServer(serverID) {
		b := model.GetBuiltInServer(serverID)
		if b == nil {
			return nil, fmt.Errorf("MCP Server 不存在")
		}
		return b, nil
	}

	var srv model.McpServer
	if err := s.db.Where("id = ? AND user_id = ?", serverID, userID).First(&srv).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("MCP Server 不存在")
		}
		return nil, fmt.Errorf("查询 MCP Server 失败")
	}
	return &srv, nil
}

// UpdateServer 更新 MCP Server
func (s *Service) UpdateServer(userID, serverID string, req *UpdateServerRequest) (*model.McpServer, error) {
	if model.IsBuiltInServer(serverID) {
		return nil, fmt.Errorf("内置 MCP Server 不可编辑")
	}

	var srv model.McpServer
	if err := s.db.Where("id = ? AND user_id = ?", serverID, userID).First(&srv).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("MCP Server 不存在")
		}
		return nil, fmt.Errorf("查询 MCP Server 失败")
	}

	if req.Name != nil && *req.Name != "" {
		srv.Name = *req.Name
	}
	if req.URL != nil && *req.URL != "" {
		srv.URL = *req.URL
	}
	if req.AuthHeader != nil && *req.AuthHeader != "" {
		srv.AuthHeader = *req.AuthHeader
	}
	if req.Enabled != nil {
		srv.Enabled = *req.Enabled
	}
	if req.Metadata != nil {
		srv.Metadata = *req.Metadata
	}
	srv.UpdatedAt = time.Now()

	if err := s.db.Save(&srv).Error; err != nil {
		return nil, fmt.Errorf("更新 MCP Server 失败")
	}
	return &srv, nil
}

// DeleteServer 删除 MCP Server
func (s *Service) DeleteServer(userID, serverID string) error {
	if model.IsBuiltInServer(serverID) {
		return fmt.Errorf("内置 MCP Server 不可删除")
	}

	var srv model.McpServer
	if err := s.db.Where("id = ? AND user_id = ?", serverID, userID).First(&srv).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("MCP Server 不存在")
		}
		return fmt.Errorf("查询 MCP Server 失败")
	}

	if err := s.db.Delete(&srv).Error; err != nil {
		return fmt.Errorf("删除 MCP Server 失败")
	}
	return nil
}

// GetServerTools 获取某 Server 支持的工具列表
func (s *Service) GetServerTools(serverID string, registry *Registry) []ToolDefinition {
	return registry.GetToolsForServer(serverID)
}

// ============ 请求结构 ============

type CreateServerRequest struct {
	Name        string            `json:"name" binding:"required"`
	Type        string            `json:"type"` // "built-in" | "http"
	URL         string            `json:"url"`
	AuthHeader  string            `json:"auth_header"`
	Metadata    model.JSONBMap    `json:"metadata"`
}

type UpdateServerRequest struct {
	Name       *string          `json:"name"`
	URL        *string          `json:"url"`
	AuthHeader *string          `json:"auth_header"`
	Enabled    *bool            `json:"enabled"`
	Metadata   *model.JSONBMap  `json:"metadata"`
}
