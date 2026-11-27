package secaudit

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"cogniforge/internal/model"
)

// Write 写入密级相关审计（失败只打日志）
func Write(db *gorm.DB, actorID, action, resType, resID, kbID, oldV, newV, result, detail string) {
	if db == nil {
		return
	}
	row := model.SecurityAuditLog{
		ID:              uuid.New().String(),
		ActorUserID:     actorID,
		Action:          action,
		ResourceType:    resType,
		ResourceID:      resID,
		KnowledgeBaseID: kbID,
		OldValue:        oldV,
		NewValue:        newV,
		Result:          result,
		Detail:          detail,
		CreatedAt:       time.Now(),
	}
	if err := db.Create(&row).Error; err != nil {
		slog.Warn("security audit write failed", "error", err, "action", action, "resource_id", resID)
	}
}
