package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"cogniforge/internal/model"
)

// agentConfigSnap 工作室可回滚字段（写入 agent_versions.config）
type agentConfigSnap struct {
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	Model               string   `json:"model"`
	SystemPrompt        string   `json:"system_prompt"`
	PromptID            string   `json:"prompt_id"`
	OpeningMessage      string   `json:"opening_message"`
	SuggestedPrompts    []string `json:"suggested_prompts"`
	RAGKnowledgeBaseIDs []string `json:"rag_kb_ids"`
	AllowAttachments    bool     `json:"allow_attachments"`
	Tools               []string `json:"tools"`
	SkillID             string   `json:"skill_id"`
	Status              string   `json:"status"`
}

func snapFromAgent(a *model.Agent) agentConfigSnap {
	return agentConfigSnap{
		Name:                a.Name,
		Description:         a.Description,
		Model:               a.Model,
		SystemPrompt:        a.SystemPrompt,
		PromptID:            a.PromptID,
		OpeningMessage:      a.OpeningMessage,
		SuggestedPrompts:    []string(a.SuggestedPrompts),
		RAGKnowledgeBaseIDs: []string(a.RAGKnowledgeBaseIDs),
		AllowAttachments:    a.AllowAttachments,
		Tools:               []string(a.Tools),
		SkillID:             a.SkillID,
		Status:              a.Status,
	}
}

func encodeSnap(s agentConfigSnap) (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func decodeSnap(raw string) (agentConfigSnap, error) {
	var s agentConfigSnap
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return s, err
	}
	return s, nil
}

func applySnap(a *model.Agent, s agentConfigSnap) {
	a.Name = s.Name
	a.Description = s.Description
	a.Model = s.Model
	a.SystemPrompt = s.SystemPrompt
	a.PromptID = s.PromptID
	a.OpeningMessage = s.OpeningMessage
	a.SuggestedPrompts = model.JSONBArray(s.SuggestedPrompts)
	a.RAGKnowledgeBaseIDs = model.JSONBArray(s.RAGKnowledgeBaseIDs)
	a.AllowAttachments = s.AllowAttachments
	a.Tools = model.JSONBArray(s.Tools)
	a.SkillID = s.SkillID
	if s.Status != "" {
		a.Status = s.Status
	}
}

func (s *AgentService) snapshotVersion(agentID string, version int, cfg agentConfigSnap, note string) {
	if s.db == nil || agentID == "" {
		return
	}
	ver := version
	if ver < 1 {
		ver = 1
	}
	var exists int64
	_ = s.db.Model(&model.AgentVersion{}).
		Where("agent_id = ? AND version = ?", agentID, ver).
		Count(&exists).Error
	if exists > 0 {
		return
	}
	raw, err := encodeSnap(cfg)
	if err != nil {
		return
	}
	_ = s.db.Create(&model.AgentVersion{
		ID:        generateID(),
		AgentID:   agentID,
		Version:   ver,
		Config:    raw,
		Note:      note,
		CreatedAt: time.Now(),
	}).Error
}

func configChanged(before, after agentConfigSnap) bool {
	b1, _ := encodeSnap(before)
	b2, _ := encodeSnap(after)
	return b1 != b2
}

// ListAgentVersions 配置历史（最多 50 条）
func (s *AgentService) ListAgentVersions(userID, agentID string) ([]model.AgentVersion, error) {
	if _, err := s.GetAgent(userID, agentID); err != nil {
		return nil, err
	}
	var list []model.AgentVersion
	if err := s.db.Where("agent_id = ?", agentID).
		Order("version DESC").
		Limit(50).
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("查询版本失败")
	}
	return list, nil
}

// RollbackAgentVersion 回滚到历史版本：先快照当前，再恢复目标并 bump version
func (s *AgentService) RollbackAgentVersion(userID, agentID string, targetVersion int) (*model.Agent, error) {
	agent, err := s.GetAgent(userID, agentID)
	if err != nil {
		return nil, err
	}
	if targetVersion < 1 {
		return nil, fmt.Errorf("无效版本号")
	}
	var snap model.AgentVersion
	if err := s.db.Where("agent_id = ? AND version = ?", agentID, targetVersion).First(&snap).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("版本不存在")
		}
		return nil, fmt.Errorf("查询版本失败")
	}
	cfg, err := decodeSnap(snap.Config)
	if err != nil {
		return nil, fmt.Errorf("版本数据损坏")
	}
	s.snapshotVersion(agent.ID, agent.Version, snapFromAgent(agent), "before_rollback")
	applySnap(agent, cfg)
	if agent.Version < 1 {
		agent.Version = 1
	}
	agent.Version++
	agent.UpdatedAt = time.Now()
	if err := s.db.Save(agent).Error; err != nil {
		return nil, fmt.Errorf("回滚失败")
	}
	return agent, nil
}

func normalizeTools(in []string) []string {
	out := make([]string, 0, len(in))
	for _, t := range in {
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
