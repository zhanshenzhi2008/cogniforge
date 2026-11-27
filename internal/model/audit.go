package model

import "time"

// SecurityAuditLog 密级相关操作审计（金融扩展）
type SecurityAuditLog struct {
	ID              string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	ActorUserID     string    `gorm:"column:actor_user_id;type:varchar(64);index:idx_secaudit_actor_ct,priority:1" json:"actor_user_id"`
	Action          string    `gorm:"type:varchar(64);not null;index:idx_secaudit_action" json:"action"`
	ResourceType    string    `gorm:"type:varchar(32);not null" json:"resource_type"`
	ResourceID      string    `gorm:"type:varchar(64);index:idx_secaudit_res" json:"resource_id"`
	KnowledgeBaseID string    `gorm:"column:knowledge_base_id;type:varchar(64);index:idx_secaudit_kb" json:"knowledge_base_id,omitempty"`
	OldValue        string    `gorm:"type:varchar(32)" json:"old_value,omitempty"`
	NewValue        string    `gorm:"type:varchar(32)" json:"new_value,omitempty"`
	Result          string    `gorm:"type:varchar(16);not null" json:"result"` // success | denied
	Detail          string    `gorm:"type:text" json:"detail,omitempty"`
	CreatedAt       time.Time `gorm:"index:idx_secaudit_actor_ct,priority:2,sort:desc;index:idx_secaudit_ct,sort:desc" json:"created_at"`
}

func (SecurityAuditLog) TableName() string {
	return "security_audit_logs"
}
