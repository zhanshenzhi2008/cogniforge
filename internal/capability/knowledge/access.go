package knowledge

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"cogniforge/internal/model"
)

const (
	VisibilityPrivate = "private"
	VisibilityOrg     = "org"
)

// normalizeVisibility 仅允许 private|org；空或非法回落 private。
func normalizeVisibility(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case VisibilityOrg:
		return VisibilityOrg
	default:
		return VisibilityPrivate
	}
}

// loadKBForWrite 仅所有者可写。
func (s *KnowledgeService) loadKBForWrite(userID, kbID string) (*model.KnowledgeBase, error) {
	return s.loadKBForWriteCtx(context.Background(), userID, kbID)
}

func (s *KnowledgeService) loadKBForWriteCtx(ctx context.Context, userID, kbID string) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}
	return &kb, nil
}

// loadKBForRead 所有者或 visibility=org 可读。
func (s *KnowledgeService) loadKBForRead(userID, kbID string) (*model.KnowledgeBase, error) {
	return s.loadKBForReadCtx(context.Background(), userID, kbID)
}

func (s *KnowledgeService) loadKBForReadCtx(ctx context.Context, userID, kbID string) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	err := s.db.WithContext(ctx).
		Where("id = ? AND (user_id = ? OR visibility = ?)", kbID, userID, VisibilityOrg).
		First(&kb).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}
	return &kb, nil
}
