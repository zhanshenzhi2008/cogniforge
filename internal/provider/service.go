package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"cogniforge/internal/crypto"
	"cogniforge/internal/model"
)

// Service 业务逻辑层
type Service struct {
	repo *Repository
}

// NewService 创建 Service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List 获取所有供应商
func (s *Service) List() ([]model.AIProvider, error) {
	return s.repo.ListAll()
}

// Get 获取单个供应商
func (s *Service) Get(id string) (*model.AIProvider, error) {
	return s.repo.GetByID(id)
}

// Create 创建供应商
func (s *Service) Create(req *CreateProviderRequest) (*model.AIProvider, error) {
	// 加密存储 API Key
	encryptedKey, err := crypto.Encrypt(req.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt api key: %w", err)
	}
	p := &model.AIProvider{
		ID:           req.ID,
		Name:         req.Name,
		Provider:     req.Provider,
		BaseURL:      req.BaseURL,
		APIKey:       encryptedKey,
		DefaultModel: req.DefaultModel,
		IsEnabled:    req.IsEnabled,
		Priority:     req.Priority,
		Status:       "active",
	}
	// 自动生成 ID（如果未指定）
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if req.ExtraHeaders != nil {
		data, _ := json.Marshal(req.ExtraHeaders)
		p.ExtraHeaders = model.JSONBMap{}
		json.Unmarshal(data, &p.ExtraHeaders)
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

// Update 更新供应商
func (s *Service) Update(id string, req *UpdateProviderRequest) (*model.AIProvider, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("provider not found")
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.BaseURL != nil {
		p.BaseURL = *req.BaseURL
	}
	if req.APIKey != nil {
		encryptedKey, err := crypto.Encrypt(*req.APIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt api key: %w", err)
		}
		p.APIKey = encryptedKey
	}
	if req.DefaultModel != nil {
		p.DefaultModel = *req.DefaultModel
	}
	// 自动生成 ID（如果未指定）
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if req.ExtraHeaders != nil {
		data, _ := json.Marshal(req.ExtraHeaders)
		p.ExtraHeaders = model.JSONBMap{}
		json.Unmarshal(data, &p.ExtraHeaders)
	}
	if req.IsEnabled != nil {
		p.IsEnabled = *req.IsEnabled
	}
	if req.Priority != nil {
		p.Priority = *req.Priority
	}
	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete 删除供应商
func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}

// SetDefault 设为默认
func (s *Service) SetDefault(id string) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("provider not found")
	}
	return s.repo.SetDefault(id)
}

// GetActive 获取当前生效的供应商配置
// 优先级：有默认配置 > 第一个启用的配置
func (s *Service) GetActive() (*model.AIProvider, error) {
	if p, err := s.repo.GetDefault(); err == nil && p.IsEnabled {
		return p, nil
	}
	if p, err := s.repo.GetFirstEnabled(); err == nil {
		return p, nil
	}
	return nil, fmt.Errorf("no active provider configured")
}

// TestConnection 测试连接
func (s *Service) TestConnection(id string) (*TestResult, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("provider not found")
	}

	// 解密 API Key
	apiKey, err := crypto.Decrypt(p.APIKey)
	if err != nil {
		return &TestResult{Success: false, Message: "failed to decrypt api key"}, nil
	}

	// 构建测试请求
	testURL := p.BaseURL
	if testURL == "" {
		testURL = "https://api.openai.com/v1/models"
	}
	if testURL[len(testURL)-1] == '/' {
		testURL = testURL + "models"
	} else {
		testURL = testURL + "/models"
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return &TestResult{Success: false, Message: err.Error()}, nil
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range p.ExtraHeaders {
		req.Header.Set(k, fmt.Sprintf("%v", v))
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)

	result := &TestResult{LatencyMs: latency.Milliseconds()}
	if err != nil {
		result.Success = false
		result.Message = err.Error()
		s.repo.UpdateStatus(id, "error", err.Error())
		return result, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 400 {
		result.Success = true
		result.Message = fmt.Sprintf("连接成功 (HTTP %d)", resp.StatusCode)
		s.repo.UpdateStatus(id, "active", "")
	} else {
		result.Success = false
		result.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		// 读取错误体
		var body bytes.Buffer
		body.ReadFrom(resp.Body)
		if body.Len() > 0 {
			result.Message = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, body.String())
		}
		s.repo.UpdateStatus(id, "error", result.Message)
	}
	return result, nil
}

// GetActiveForChat 获取 chat 用的配置（供 chat service 调用）
func (s *Service) GetActiveForChat() (baseURL, apiKey string, headers map[string]string, err error) {
	p, err := s.GetActive()
	if err != nil {
		return "", "", nil, err
	}
	decryptedKey, err := crypto.Decrypt(p.APIKey)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to decrypt api key: %w", err)
	}
	h := make(map[string]string)
	for k, v := range p.ExtraHeaders {
		h[k] = fmt.Sprintf("%v", v)
	}
	return p.BaseURL, decryptedKey, h, nil
}
