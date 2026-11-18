package provider

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"cogniforge/internal/model"

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

// ListVisible 按查看者过滤：admin 全量；匿名仅 public；登录用户 public + 自己的 private
func (r *Repository) ListVisible(userID string, isAdmin bool) ([]model.AIProvider, error) {
	if isAdmin {
		return r.ListAll()
	}
	var providers []model.AIProvider
	q := r.db.Where("deleted_at IS NULL")
	if userID == "" {
		q = q.Where("scope = ?", model.ProviderScopePublic)
	} else {
		q = q.Where("scope = ? OR owner_user_id = ?", model.ProviderScopePublic, userID)
	}
	err := q.Order("priority ASC, created_at ASC").Find(&providers).Error
	return providers, err
}

// GetPublicDefault 全站公开对话默认
func (r *Repository) GetPublicDefault() (*model.AIProvider, error) {
	var p model.AIProvider
	err := r.db.Where("is_default = ? AND scope = ? AND deleted_at IS NULL", true, model.ProviderScopePublic).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPublicDefaultEmbedding 全站公开向量默认
func (r *Repository) GetPublicDefaultEmbedding() (*model.AIProvider, error) {
	var p model.AIProvider
	err := r.db.Where("is_default_embedding = ? AND scope = ? AND deleted_at IS NULL", true, model.ProviderScopePublic).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetFirstEnabledVisible 可见范围内第一个启用且具备能力的供应商
func (r *Repository) GetFirstEnabledVisible(userID string, isAdmin bool, cap string) (*model.AIProvider, error) {
	list, err := r.ListVisible(userID, isAdmin)
	if err != nil {
		return nil, err
	}
	for i := range list {
		p := &list[i]
		if !p.IsEnabled {
			continue
		}
		if model.HasCapability(p.Capabilities, cap) {
			return p, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// GetPreferredProviderIDs 读用户偏好
func (r *Repository) GetPreferredProviderIDs(userID string) (chatID, embedID string) {
	if userID == "" {
		return "", ""
	}
	var s model.UserSettings
	if err := r.db.Where("user_id = ?", userID).First(&s).Error; err != nil {
		return "", ""
	}
	return s.PreferredChatProviderID, s.PreferredEmbeddingProviderID
}

// SetPreferredProvider 写入用户个人默认（不改全站 is_default）
func (r *Repository) SetPreferredProvider(userID, purpose, providerID string) error {
	if userID == "" {
		return gorm.ErrInvalidData
	}
	var s model.UserSettings
	err := r.db.Where("user_id = ?", userID).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		s = model.UserSettings{
			ID:        uuid.New().String(),
			UserID:    userID,
			Theme:     "light",
			Language:  "zh-CN",
			Timezone:  "Asia/Shanghai",
			Metadata:  model.JSONBMap{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if purpose == model.CapEmbedding {
			s.PreferredEmbeddingProviderID = providerID
		} else {
			s.PreferredChatProviderID = providerID
		}
		return r.db.Create(&s).Error
	}
	if err != nil {
		return err
	}
	updates := map[string]any{"updated_at": time.Now()}
	if purpose == model.CapEmbedding {
		updates["preferred_embedding_provider_id"] = providerID
	} else {
		updates["preferred_chat_provider_id"] = providerID
	}
	return r.db.Model(&s).Updates(updates).Error
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
	return r.GetFirstEnabledFor(model.CapChat)
}

func (r *Repository) GetDefaultEmbedding() (*model.AIProvider, error) {
	var p model.AIProvider
	err := r.db.Where("is_default_embedding = ? AND deleted_at IS NULL", true).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetFirstEnabledFor(cap string) (*model.AIProvider, error) {
	var list []model.AIProvider
	err := r.db.Where("is_enabled = ? AND deleted_at IS NULL", true).Order("priority ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	for i := range list {
		if model.HasCapability(list[i].Capabilities, cap) {
			return &list[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
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

func (r *Repository) SetDefaultEmbedding(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AIProvider{}).Where("is_default_embedding = ?", true).
			Updates(map[string]any{"is_default_embedding": false}).Error; err != nil {
			return err
		}
		return tx.Model(&model.AIProvider{}).Where("id = ?", id).
			Updates(map[string]any{"is_default_embedding": true}).Error
	})
}

// RepairOpenAIVendorLabels 把「厂商仍写成 openai、但地址明显是别的厂商」的行改成对应厂商键。
// 只动 provider=openai 的行，已经选过智谱/自定义的不动。返回改了多少行。
func (r *Repository) RepairOpenAIVendorLabels() (int64, error) {
	var changed int64
	for _, h := range model.VendorURLHints() {
		if h.Provider == "" || h.Provider == string(model.ProviderOpenAI) {
			continue
		}
		res := r.db.Model(&model.AIProvider{}).
			Where("provider = ? AND LOWER(base_url) LIKE ?", string(model.ProviderOpenAI), "%"+strings.ToLower(h.Host)+"%").
			Update("provider", h.Provider)
		if res.Error != nil {
			return changed, res.Error
		}
		changed += res.RowsAffected
	}
	return changed, nil
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
