package prompt

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"cogniforge/internal/model"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) List(userID string) ([]model.Prompt, error) {
	var list []model.Prompt
	err := s.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&list).Error
	return list, err
}

func (s *Service) Create(userID string, req *CreatePromptRequest) (*model.Prompt, *model.PromptVersion, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, nil, errors.New("name 不能为空")
	}
	p := &model.Prompt{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        name,
		Description: req.Description,
	}
	var ver *model.PromptVersion
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		content := strings.TrimSpace(req.Content)
		if content == "" {
			content = "You are a helpful assistant."
		}
		ver = &model.PromptVersion{
			ID:        uuid.New().String(),
			PromptID:  p.ID,
			UserID:    userID,
			Version:   1,
			Content:   content,
			Changelog: req.Changelog,
		}
		if ver.Changelog == "" {
			ver.Changelog = "初始版本"
		}
		if err := tx.Create(ver).Error; err != nil {
			return err
		}
		p.CurrentVersionID = ver.ID
		return tx.Save(p).Error
	})
	if err != nil {
		return nil, nil, err
	}
	return p, ver, nil
}

func (s *Service) Get(userID, id string) (*model.Prompt, *model.PromptVersion, error) {
	var p model.Prompt
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&p).Error; err != nil {
		return nil, nil, errors.New("提示词不存在")
	}
	var ver *model.PromptVersion
	if p.CurrentVersionID != "" {
		var v model.PromptVersion
		if err := s.db.Where("id = ? AND user_id = ?", p.CurrentVersionID, userID).First(&v).Error; err == nil {
			ver = &v
		}
	}
	return &p, ver, nil
}

func (s *Service) Update(userID, id string, req *UpdatePromptRequest) (*model.Prompt, error) {
	p, _, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		n := strings.TrimSpace(*req.Name)
		if n == "" {
			return nil, errors.New("name 不能为空")
		}
		p.Name = n
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if err := s.db.Save(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(userID, id string) error {
	p, _, err := s.Get(userID, id)
	if err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("prompt_id = ?", p.ID).Delete(&model.PromptVersion{}).Error; err != nil {
			return err
		}
		return tx.Delete(p).Error
	})
}

func (s *Service) ListVersions(userID, promptID string) ([]model.PromptVersion, error) {
	if _, _, err := s.Get(userID, promptID); err != nil {
		return nil, err
	}
	var list []model.PromptVersion
	err := s.db.Where("prompt_id = ? AND user_id = ?", promptID, userID).
		Order("version DESC").Find(&list).Error
	return list, err
}

func (s *Service) CreateVersion(userID, promptID string, req *CreateVersionRequest) (*model.PromptVersion, error) {
	p, _, err := s.Get(userID, promptID)
	if err != nil {
		return nil, err
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errors.New("content 不能为空")
	}
	var maxVer int
	_ = s.db.Model(&model.PromptVersion{}).
		Where("prompt_id = ?", promptID).
		Select("COALESCE(MAX(version), 0)").Scan(&maxVer).Error

	ver := &model.PromptVersion{
		ID:        uuid.New().String(),
		PromptID:  promptID,
		UserID:    userID,
		Version:   maxVer + 1,
		Content:   content,
		Changelog: req.Changelog,
	}
	if ver.Changelog == "" {
		ver.Changelog = "新版本"
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ver).Error; err != nil {
			return err
		}
		p.CurrentVersionID = ver.ID
		return tx.Save(p).Error
	})
	if err != nil {
		return nil, err
	}
	return ver, nil
}

func (s *Service) GetVersion(userID, versionID string) (*model.PromptVersion, error) {
	var ver model.PromptVersion
	if err := s.db.Where("id = ? AND user_id = ?", versionID, userID).First(&ver).Error; err != nil {
		return nil, errors.New("提示词版本不存在")
	}
	return &ver, nil
}
