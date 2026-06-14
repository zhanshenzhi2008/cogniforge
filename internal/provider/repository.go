package provider

import (
	"cogniforge/internal/model"
	"time"

	"gorm.io/gorm"
)

// Repository 数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// List 获取所有启用的供应商（按 priority 升序）
func (r *Repository) List() ([]model.AIProvider, error) {
	var providers []model.AIProvider
	err := r.db.Where("deleted_at IS NULL").Order("priority ASC, created_at ASC").Find(&providers).Error
	return providers, err
}

// ListAll 获取所有供应商（含禁用的）
func (r *Repository) ListAll() ([]model.AIProvider, error) {
	var providers []model.AIProvider
	err := r.db.Where("deleted_at IS NULL").Order("priority ASC, created_at ASC").Find(&providers).Error
	return providers, err
}

// GetByID 根据 ID 获取
func (r *Repository) GetByID(id string) (*model.AIProvider, error) {
	var p model.AIProvider
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetDefault 获取默认供应商
func (r *Repository) GetDefault() (*model.AIProvider, error) {
	var p model.AIProvider
	err := r.db.Where("is_default = ? AND deleted_at IS NULL", true).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetFirstEnabled 获取第一个启用的供应商
func (r *Repository) GetFirstEnabled() (*model.AIProvider, error) {
	var p model.AIProvider
	err := r.db.Where("is_enabled = ? AND deleted_at IS NULL", true).Order("priority ASC").First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Create 创建供应商
func (r *Repository) Create(p *model.AIProvider) error {
	return r.db.Create(p).Error
}

// Update 更新供应商
func (r *Repository) Update(p *model.AIProvider) error {
	return r.db.Save(p).Error
}

// Delete 删除供应商（软删除）
func (r *Repository) Delete(id string) error {
	return r.db.Delete(&model.AIProvider{}, "id = ?", id).Error
}

// SetDefault 将指定供应商设为默认
func (r *Repository) SetDefault(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先取消所有默认
		if err := tx.Model(&model.AIProvider{}).Where("is_default = ?", true).
			Updates(map[string]any{"is_default": false}).Error; err != nil {
			return err
		}
		// 再设置新的默认
		return tx.Model(&model.AIProvider{}).Where("id = ?", id).
			Updates(map[string]any{"is_default": true}).Error
	})
}

// UpdateStatus 更新供应商状态
func (r *Repository) UpdateStatus(id, status, lastError string) error {
	updates := map[string]any{
		"status": status,
	}
	now := time.Now()
	updates["last_test_at"] = &now
	if lastError != "" {
		updates["last_error"] = lastError
	}
	return r.db.Model(&model.AIProvider{}).Where("id = ?", id).Updates(updates).Error
}
