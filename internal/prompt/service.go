package prompt

import (
	"errors"
	"strings"
	"time"

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
	now := time.Now()
	p := &model.Prompt{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        name,
		Description: req.Description,
		Variables:   normalizeVarDefs(req.Variables),
		Tags:        model.JSONBArray(req.Tags),
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
			ID:          uuid.New().String(),
			PromptID:    p.ID,
			UserID:      userID,
			Version:     1,
			Status:      model.PromptVersionPublished,
			Content:     content,
			Changelog:   req.Changelog,
			PublishedAt: &now,
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
	if req.Variables != nil {
		p.Variables = normalizeVarDefs(*req.Variables)
	}
	if req.Tags != nil {
		p.Tags = model.JSONBArray(*req.Tags)
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
		// 解除 Agent 绑定
		if err := tx.Model(&model.Agent{}).
			Where("user_id = ? AND prompt_id = ?", userID, p.ID).
			Update("prompt_id", "").Error; err != nil {
			return err
		}
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

	status := model.PromptVersionDraft
	var publishedAt *time.Time
	if req.Publish {
		status = model.PromptVersionPublished
		now := time.Now()
		publishedAt = &now
	}

	ver := &model.PromptVersion{
		ID:          uuid.New().String(),
		PromptID:    promptID,
		UserID:      userID,
		Version:     maxVer + 1,
		Status:      status,
		Content:     content,
		Changelog:   req.Changelog,
		PublishedAt: publishedAt,
	}
	if ver.Changelog == "" {
		if req.Publish {
			ver.Changelog = "发布新版本"
		} else {
			ver.Changelog = "草稿"
		}
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ver).Error; err != nil {
			return err
		}
		if req.Publish {
			p.CurrentVersionID = ver.ID
			return tx.Save(p).Error
		}
		return nil
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

// PublishVersion 将草稿发布为正式版，并设为当前。
func (s *Service) PublishVersion(userID, versionID string) (*model.Prompt, *model.PromptVersion, error) {
	ver, err := s.GetVersion(userID, versionID)
	if err != nil {
		return nil, nil, err
	}
	p, _, err := s.Get(userID, ver.PromptID)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	err = s.db.Transaction(func(tx *gorm.DB) error {
		ver.Status = model.PromptVersionPublished
		ver.PublishedAt = &now
		if err := tx.Save(ver).Error; err != nil {
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

// DeleteVersion 删除单个版本。至少保留一版；若删的是当前版，自动切到剩余最高已发布版（否则任意最高版）。
func (s *Service) DeleteVersion(userID, versionID string) error {
	ver, err := s.GetVersion(userID, versionID)
	if err != nil {
		return err
	}
	p, _, err := s.Get(userID, ver.PromptID)
	if err != nil {
		return err
	}
	var count int64
	if err := s.db.Model(&model.PromptVersion{}).
		Where("prompt_id = ? AND user_id = ?", ver.PromptID, userID).
		Count(&count).Error; err != nil {
		return err
	}
	if count <= 1 {
		return errors.New("至少保留一个版本；若要清空请删除整条提示词")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(ver).Error; err != nil {
			return err
		}
		if p.CurrentVersionID != ver.ID {
			return nil
		}
		var next model.PromptVersion
		q := tx.Where("prompt_id = ? AND user_id = ?", ver.PromptID, userID)
		if err := q.Where("status = ?", model.PromptVersionPublished).
			Order("version DESC").First(&next).Error; err != nil {
			if err := tx.Where("prompt_id = ? AND user_id = ?", ver.PromptID, userID).
				Order("version DESC").First(&next).Error; err != nil {
				return err
			}
		}
		p.CurrentVersionID = next.ID
		return tx.Save(p).Error
	})
}

// SetCurrentVersion 将已发布版本设为当前（草稿不可上线）。
func (s *Service) SetCurrentVersion(userID, promptID, versionID string) (*model.Prompt, error) {
	p, _, err := s.Get(userID, promptID)
	if err != nil {
		return nil, err
	}
	ver, err := s.GetVersion(userID, versionID)
	if err != nil {
		return nil, err
	}
	if ver.PromptID != promptID {
		return nil, errors.New("版本不属于该提示词")
	}
	if ver.Status != model.PromptVersionPublished {
		return nil, errors.New("只能将已发布版本设为当前；请先发布")
	}
	p.CurrentVersionID = ver.ID
	if err := s.db.Save(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Render(userID, promptID string, req *RenderRequest) (*RenderResponse, error) {
	p, cur, err := s.Get(userID, promptID)
	if err != nil {
		return nil, err
	}
	var ver *model.PromptVersion
	if strings.TrimSpace(req.VersionID) != "" {
		ver, err = s.GetVersion(userID, req.VersionID)
		if err != nil {
			return nil, err
		}
		if ver.PromptID != promptID {
			return nil, errors.New("版本不属于该提示词")
		}
	} else {
		ver = cur
	}
	if ver == nil {
		return nil, errors.New("没有可用版本")
	}
	rendered, err := RenderContent(ver.Content, req.Variables, p.Variables, true)
	if err != nil {
		return nil, err
	}
	return &RenderResponse{
		Content:   rendered,
		VarNames:  ExtractVarNames(ver.Content),
		VersionID: ver.ID,
		Version:   ver.Version,
	}, nil
}

// ResolvePublishedContent 供 Agent / Eval 取「当前已发布」正文并渲染。
func (s *Service) ResolvePublishedContent(userID, promptID string, variables map[string]string) (string, error) {
	p, ver, err := s.Get(userID, promptID)
	if err != nil {
		return "", err
	}
	if ver == nil {
		return "", errors.New("提示词没有当前版本")
	}
	if ver.Status != model.PromptVersionPublished {
		return "", errors.New("当前版本未发布")
	}
	return RenderContent(ver.Content, variables, p.Variables, true)
}

func (s *Service) ImportFromAgent(userID string, req *ImportRequest) (*model.Prompt, *model.PromptVersion, error) {
	agentID := strings.TrimSpace(req.AgentID)
	if agentID == "" {
		return nil, nil, errors.New("agent_id 不能为空")
	}
	var ag model.Agent
	if err := s.db.Where("id = ? AND user_id = ?", agentID, userID).First(&ag).Error; err != nil {
		return nil, nil, errors.New("Agent 不存在")
	}
	content := strings.TrimSpace(ag.SystemPrompt)
	if content == "" {
		return nil, nil, errors.New("该 Agent 没有 system_prompt 可导入")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = ag.Name + " 提示词"
	}
	return s.Create(userID, &CreatePromptRequest{
		Name:        name,
		Description: "从 Agent「" + ag.Name + "」导入",
		Content:     content,
		Changelog:   "从 Agent 导入",
		Tags:        []string{"imported", "agent"},
	})
}

func (s *Service) ListUsages(userID, promptID string) ([]PromptUsage, error) {
	if _, _, err := s.Get(userID, promptID); err != nil {
		return nil, err
	}
	var agents []model.Agent
	if err := s.db.Select("id, name").
		Where("user_id = ? AND prompt_id = ?", userID, promptID).
		Find(&agents).Error; err != nil {
		return nil, err
	}
	out := make([]PromptUsage, 0, len(agents))
	for _, a := range agents {
		out = append(out, PromptUsage{AgentID: a.ID, AgentName: a.Name})
	}
	return out, nil
}

func normalizeVarDefs(defs model.PromptVarDefs) model.PromptVarDefs {
	if len(defs) == 0 {
		return model.PromptVarDefs{}
	}
	out := make(model.PromptVarDefs, 0, len(defs))
	seen := map[string]struct{}{}
	for _, d := range defs {
		name := strings.TrimSpace(d.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		d.Name = name
		out = append(out, d)
	}
	return out
}
