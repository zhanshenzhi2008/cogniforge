package memory

import (
	"context"
	"time"

	"gorm.io/gorm"

	"cogniforge/internal/model"
)

type MemoryService struct {
	db *gorm.DB
}

func NewMemoryService(db *gorm.DB) *MemoryService {
	return &MemoryService{db: db}
}

// List 返回用户记忆列表，支持 kind / agent_id 过滤。
func (s *MemoryService) List(ctx context.Context, userID string, kind, agentID string, limit int) ([]model.ChatMemory, error) {
	q := s.db.WithContext(ctx).Where("user_id = ?", userID)
	if kind != "" {
		q = q.Where("kind = ?", kind)
	}
	if agentID != "" {
		q = q.Where("agent_id = ? OR agent_id IS NULL OR agent_id = ''", agentID)
	}
	if limit <= 0 {
		limit = 50
	}
	var rows []model.ChatMemory
	err := q.Order("importance DESC, created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

// Get 返回一条记忆。
func (s *MemoryService) Get(ctx context.Context, userID, id string) (*model.ChatMemory, error) {
	var row model.ChatMemory
	err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Delete 软删除一条记忆。
func (s *MemoryService) Delete(ctx context.Context, userID, id string) error {
	return s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.ChatMemory{}).Error
}

// ClearAll 软删除用户全部记忆。
func (s *MemoryService) ClearAll(ctx context.Context, userID string) error {
	return s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.ChatMemory{}).Error
}

// Upsert 插入或更新一条记忆（id 已有则更新 content/importance）。
func (s *MemoryService) Upsert(ctx context.Context, m *model.ChatMemory) error {
	return s.db.WithContext(ctx).
		Where("id = ?", m.ID).
		Assign(model.ChatMemory{
			Content:    m.Content,
			Kind:       m.Kind,
			Importance: m.Importance,
			UpdatedAt:  time.Now(),
		}).
		FirstOrCreate(m).Error
}

// TouchAccessed 更新 last_accessed_at。
func (s *MemoryService) TouchAccessed(ctx context.Context, userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).
		Model(&model.ChatMemory{}).
		Where("user_id = ? AND id IN ?", userID, ids).
		Update("last_accessed_at", time.Now()).Error
}

// Count 统计用户记忆数量。
func (s *MemoryService) Count(ctx context.Context, userID, kind, agentID string) (int64, error) {
	var total int64
	q := s.db.WithContext(ctx).Model(&model.ChatMemory{}).Where("user_id = ?", userID)
	if kind != "" {
		q = q.Where("kind = ?", kind)
	}
	if agentID != "" {
		q = q.Where("agent_id = ?", agentID)
	}
	err := q.Count(&total).Error
	return total, err
}
