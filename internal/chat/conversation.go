package chat

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"cogniforge/internal/model"
)

const (
	maxConversationTitleRunes = 40
	maxConversationList       = 100
)

var errConversationNotFound = errors.New("对话不存在")

type ConversationService struct {
	db *gorm.DB
}

func NewConversationService(db *gorm.DB) *ConversationService {
	return &ConversationService{db: db}
}

type ConversationSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	AgentID   string    `json:"agent_id"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateConversationRequest struct {
	Title    string                      `json:"title"`
	AgentID  string                      `json:"agent_id"`
	Model    string                      `json:"model"`
	Messages []model.ConversationMessage `json:"messages"`
}

type UpdateConversationRequest struct {
	Title    *string                      `json:"title"`
	AgentID  *string                      `json:"agent_id"`
	Model    *string                      `json:"model"`
	Messages *[]model.ConversationMessage `json:"messages"`
}

func titleFromMessages(msgs []model.ConversationMessage) string {
	for _, m := range msgs {
		if m.Role != "user" {
			continue
		}
		text := strings.TrimSpace(m.Content)
		if text == "" {
			continue
		}
		if utf8.RuneCountInString(text) <= maxConversationTitleRunes {
			return text
		}
		runes := []rune(text)
		return string(runes[:maxConversationTitleRunes]) + "…"
	}
	return ""
}

func toSummary(row model.ChatConversation) ConversationSummary {
	return ConversationSummary{
		ID:        row.ID,
		Title:     row.Title,
		AgentID:   row.AgentID,
		Model:     row.Model,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func (s *ConversationService) List(userID string) ([]ConversationSummary, error) {
	var rows []model.ChatConversation
	err := s.db.Where("user_id = ?", userID).
		Select("id", "user_id", "agent_id", "title", "model", "created_at", "updated_at").
		Order("updated_at DESC").
		Limit(maxConversationList).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]ConversationSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, toSummary(row))
	}
	return out, nil
}

func (s *ConversationService) Create(userID string, req *CreateConversationRequest) (*model.ChatConversation, error) {
	msgs := req.Messages
	if msgs == nil {
		msgs = []model.ConversationMessage{}
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = titleFromMessages(msgs)
	}
	if title == "" {
		title = "New chat"
	}
	now := time.Now()
	row := &model.ChatConversation{
		ID:        newID(),
		UserID:    userID,
		AgentID:   strings.TrimSpace(req.AgentID),
		Title:     title,
		Model:     strings.TrimSpace(req.Model),
		Messages:  msgs,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.db.Create(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (s *ConversationService) Get(userID, id string) (*model.ChatConversation, error) {
	var row model.ChatConversation
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errConversationNotFound
		}
		return nil, err
	}
	if row.Messages == nil {
		row.Messages = []model.ConversationMessage{}
	}
	return &row, nil
}

func (s *ConversationService) Update(userID, id string, req *UpdateConversationRequest) (*model.ChatConversation, error) {
	row, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title != "" {
			row.Title = title
		}
	}
	if req.AgentID != nil {
		row.AgentID = strings.TrimSpace(*req.AgentID)
	}
	if req.Model != nil {
		row.Model = strings.TrimSpace(*req.Model)
	}
	if req.Messages != nil {
		row.Messages = *req.Messages
		if row.Messages == nil {
			row.Messages = []model.ConversationMessage{}
		}
		if req.Title == nil && (row.Title == "" || row.Title == "New chat") {
			if generated := titleFromMessages(row.Messages); generated != "" {
				row.Title = generated
			}
		}
	}
	row.UpdatedAt = time.Now()
	if err := s.db.Save(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (s *ConversationService) Delete(userID, id string) error {
	res := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.ChatConversation{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errConversationNotFound
	}
	return nil
}
