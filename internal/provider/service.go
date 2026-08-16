package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"cogniforge/internal/crypto"
	"cogniforge/internal/model"
	"cogniforge/internal/modelcache"
)

// Service 业务逻辑层
type Service struct {
	repo  *Repository
	cache *modelcache.Cache
}

// NewService 创建 Service。cache 可为 nil（仅查库）。
func NewService(repo *Repository, cache *modelcache.Cache) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) refreshCache() {
	if s.cache == nil {
		return
	}
	s.cache.Invalidate()
	p, err := s.loadActiveFromDB()
	if err != nil {
		return
	}
	snap := s.buildSnapshot(p)
	snap.Rev = s.cache.CurrentRev()
	s.cache.Put(snap)
}

func (s *Service) loadActiveFromDB() (*model.AIProvider, error) {
	if p, err := s.repo.GetDefault(); err == nil && p.IsEnabled {
		return p, nil
	}
	if p, err := s.repo.GetFirstEnabled(); err == nil {
		return p, nil
	}
	return nil, fmt.Errorf("no active provider configured")
}

func (s *Service) buildSnapshot(active *model.AIProvider) *modelcache.Snapshot {
	headers := map[string]string{}
	for k, v := range active.ExtraHeaders {
		headers[k] = fmt.Sprintf("%v", v)
	}
	snap := &modelcache.Snapshot{
		ID:           active.ID,
		Name:         active.Name,
		Provider:     active.Provider,
		BaseURL:      active.BaseURL,
		DefaultModel: active.DefaultModel,
		ExtraHeaders: headers,
		EncryptedKey: active.APIKey,
	}
	if list, err := s.repo.ListAll(); err == nil {
		seen := map[string]struct{}{}
		add := func(m string) {
			m = strings.TrimSpace(m)
			if m == "" {
				return
			}
			if _, ok := seen[m]; ok {
				return
			}
			seen[m] = struct{}{}
			snap.Models = append(snap.Models, modelcache.ModelItem{ID: m, Name: m})
		}
		add(active.DefaultModel)
		for _, p := range list {
			if p.IsEnabled {
				add(p.DefaultModel)
			}
		}
	}
	return snap
}

func snapshotToProvider(snap *modelcache.Snapshot) *model.AIProvider {
	headers := model.JSONBMap{}
	for k, v := range snap.ExtraHeaders {
		headers[k] = v
	}
	return &model.AIProvider{
		ID:           snap.ID,
		Name:         snap.Name,
		Provider:     snap.Provider,
		BaseURL:      snap.BaseURL,
		APIKey:       snap.EncryptedKey,
		DefaultModel: snap.DefaultModel,
		ExtraHeaders: headers,
		IsEnabled:    true,
	}
}

// List 获取所有供应商
func (s *Service) List() ([]model.AIProvider, error) {
	return s.repo.ListAll()
}

// CachedModels 下拉用的模型名（命中缓存则不查库）
func (s *Service) CachedModels() []modelcache.ModelItem {
	if s.cache != nil {
		if snap, ok := s.cache.Get(); ok && len(snap.Models) > 0 {
			return snap.Models
		}
	}
	p, err := s.GetActive()
	if err != nil {
		return nil
	}
	if s.cache != nil {
		if snap, ok := s.cache.Get(); ok {
			return snap.Models
		}
	}
	if p.DefaultModel == "" {
		return nil
	}
	return []modelcache.ModelItem{{ID: p.DefaultModel, Name: p.DefaultModel}}
}

// Get 获取单个供应商
func (s *Service) Get(id string) (*model.AIProvider, error) {
	return s.repo.GetByID(id)
}

// Create 创建供应商
func (s *Service) Create(req *CreateProviderRequest) (*model.AIProvider, error) {
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
	s.refreshCache()
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
	s.refreshCache()
	return p, nil
}

// Delete 删除供应商
func (s *Service) Delete(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.refreshCache()
	return nil
}

// SetDefault 设为默认
func (s *Service) SetDefault(id string) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("provider not found")
	}
	if err := s.repo.SetDefault(id); err != nil {
		return err
	}
	s.refreshCache()
	return nil
}

// GetActive 获取当前生效的供应商配置
func (s *Service) GetActive() (*model.AIProvider, error) {
	if s.cache != nil {
		if snap, ok := s.cache.Get(); ok && snap.ID != "" {
			return snapshotToProvider(snap), nil
		}
	}
	p, err := s.loadActiveFromDB()
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		snap := s.buildSnapshot(p)
		snap.Rev = s.cache.CurrentRev()
		s.cache.Put(snap)
	}
	return p, nil
}

// TestConnection 测试连接
func (s *Service) TestConnection(id string) (*TestResult, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("provider not found")
	}

	apiKey, err := crypto.Decrypt(p.APIKey)
	if err != nil {
		return &TestResult{Success: false, Message: "failed to decrypt api key"}, nil
	}

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
	if s.cache != nil {
		if snap, plain, hdrs, ok := s.cache.GetHot(); ok && snap != nil {
			if plain != "" {
				return snap.BaseURL, plain, hdrs, nil
			}
			if snap.EncryptedKey != "" {
				decrypted, decErr := crypto.Decrypt(snap.EncryptedKey)
				if decErr == nil {
					h := snap.ExtraHeaders
					if h == nil {
						h = map[string]string{}
					}
					s.cache.PutHot(snap, decrypted, h)
					return snap.BaseURL, decrypted, h, nil
				}
			}
		}
	}

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
	if s.cache != nil {
		if snap, ok := s.cache.Get(); ok {
			s.cache.PutHot(snap, decryptedKey, h)
		}
	}
	return p.BaseURL, decryptedKey, h, nil
}
