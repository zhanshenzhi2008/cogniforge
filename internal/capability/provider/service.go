package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"cogniforge/internal/infra/crypto"
	"cogniforge/internal/model"
	"cogniforge/internal/infra/modelcache"
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
	p, err := s.loadSitePublicActiveFromDB()
	if err != nil {
		return
	}
	snap := s.buildSnapshot(p)
	snap.Rev = s.cache.CurrentRev()
	s.cache.Put(snap)
}

// loadSitePublicActiveFromDB 全站公开默认（供 Redis 快照 / 无用户上下文回退）
func (s *Service) loadSitePublicActiveFromDB() (*model.AIProvider, error) {
	if p, err := s.repo.GetPublicDefault(); err == nil && p.IsEnabled && model.HasCapability(p.Capabilities, model.CapChat) {
		return p, nil
	}
	// 兼容旧数据：is_default 但 scope 尚未回填
	if p, err := s.repo.GetDefault(); err == nil && p.IsEnabled && model.HasCapability(p.Capabilities, model.CapChat) {
		return p, nil
	}
	if p, err := s.repo.GetFirstEnabledVisible("", true, model.CapChat); err == nil {
		return p, nil
	}
	return nil, fmt.Errorf("no active provider configured")
}

func (s *Service) loadActiveFromDB() (*model.AIProvider, error) {
	return s.loadSitePublicActiveFromDB()
}

func (s *Service) loadEmbeddingFromDB() (*model.AIProvider, error) {
	if p, err := s.repo.GetPublicDefaultEmbedding(); err == nil && p.IsEnabled && model.HasCapability(p.Capabilities, model.CapEmbedding) {
		return p, nil
	}
	if p, err := s.repo.GetDefaultEmbedding(); err == nil && p.IsEnabled && model.HasCapability(p.Capabilities, model.CapEmbedding) {
		return p, nil
	}
	if p, err := s.repo.GetFirstEnabledVisible("", true, model.CapEmbedding); err == nil {
		return p, nil
	}
	return nil, fmt.Errorf("no embedding provider configured")
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
	// 快照模型列表仅含公开启用项（避免把别人私有模型名泄露进全局缓存）
	if list, err := s.repo.ListVisible("", false); err == nil {
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
			if !p.IsEnabled {
				continue
			}
			add(p.DefaultModel)
			for _, m := range model.CatalogModels(p.Provider) {
				add(m)
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
		Scope:        model.ProviderScopePublic,
	}
}

// RepairVendorLabels 修正误标成 OpenAI 的已知厂商，有改动时刷新缓存。
func (s *Service) RepairVendorLabels() {
	if s.repo == nil {
		return
	}
	n, err := s.repo.RepairOpenAIVendorLabels()
	if err != nil || n == 0 {
		return
	}
	s.refreshCache()
}

// ListFor 按查看者过滤列表
func (s *Service) ListFor(actor Actor) ([]model.AIProvider, error) {
	s.RepairVendorLabels()
	return s.repo.ListVisible(actor.UserID, actor.IsAdmin)
}

// List 兼容旧调用（无用户上下文 → 仅公开）
func (s *Service) List() ([]model.AIProvider, error) {
	return s.ListFor(Actor{})
}

// CachedModelsFor 某用户可见的模型下拉
func (s *Service) CachedModelsFor(actor Actor) []modelcache.ModelItem {
	list, err := s.repo.ListVisible(actor.UserID, actor.IsAdmin)
	if err != nil {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]modelcache.ModelItem, 0, 8)
	add := func(m string) {
		m = strings.TrimSpace(m)
		if m == "" {
			return
		}
		if _, ok := seen[m]; ok {
			return
		}
		seen[m] = struct{}{}
		out = append(out, modelcache.ModelItem{ID: m, Name: m})
	}
	for _, p := range list {
		if !p.IsEnabled || !model.HasCapability(p.Capabilities, model.CapChat) {
			continue
		}
		add(p.DefaultModel)
		for _, m := range model.CatalogModels(p.Provider) {
			add(m)
		}
	}
	return out
}

// CachedModels 兼容：公开模型
func (s *Service) CachedModels() []modelcache.ModelItem {
	return s.CachedModelsFor(Actor{})
}

// RefreshCache 启动或迁移后重建快照
func (s *Service) RefreshCache() {
	s.refreshCache()
}

// GetFor 可见性检查后取详情
func (s *Service) GetFor(id string, actor Actor) (*model.AIProvider, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrProviderNotFound
	}
	if !model.ProviderVisibleTo(p, actor.UserID, actor.IsAdmin) {
		return nil, ErrProviderForbidden
	}
	return p, nil
}

// Get 兼容旧调用
func (s *Service) Get(id string) (*model.AIProvider, error) {
	return s.repo.GetByID(id)
}

// CreateFor 创建（非 admin 强制 private）
func (s *Service) CreateFor(req *CreateProviderRequest, actor Actor) (*model.AIProvider, error) {
	if actor.UserID == "" {
		return nil, ErrProviderForbidden
	}
	encryptedKey, err := crypto.Encrypt(req.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt api key: %w", err)
	}
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = model.ProviderScopePrivate
	}
	if scope != model.ProviderScopePrivate && scope != model.ProviderScopePublic {
		return nil, fmt.Errorf("%w: invalid scope", ErrProviderBadRequest)
	}
	if scope == model.ProviderScopePublic && !actor.IsAdmin {
		return nil, fmt.Errorf("%w: only admin can create public providers", ErrProviderForbidden)
	}
	p := &model.AIProvider{
		ID:             req.ID,
		Name:           req.Name,
		Provider:       req.Provider,
		BaseURL:        req.BaseURL,
		APIKey:         encryptedKey,
		DefaultModel:   req.DefaultModel,
		EmbeddingModel: req.EmbeddingModel,
		Capabilities:   model.JoinCapabilities(req.Capabilities),
		OwnerUserID:    actor.UserID,
		Scope:          scope,
		IsEnabled:      req.IsEnabled,
		Priority:       req.Priority,
		Status:         "active",
	}
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if req.ExtraHeaders != nil {
		data, _ := json.Marshal(req.ExtraHeaders)
		p.ExtraHeaders = model.JSONBMap{}
		_ = json.Unmarshal(data, &p.ExtraHeaders)
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	s.refreshCache()
	return p, nil
}

// Create 兼容（无 actor 时不可用）
func (s *Service) Create(req *CreateProviderRequest) (*model.AIProvider, error) {
	return s.CreateFor(req, Actor{})
}

// UpdateFor 更新（鉴权）
func (s *Service) UpdateFor(id string, req *UpdateProviderRequest, actor Actor) (*model.AIProvider, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrProviderNotFound
	}
	if !model.ProviderWritableBy(p, actor.UserID, actor.IsAdmin) {
		return nil, ErrProviderForbidden
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Provider != nil {
		if v := strings.TrimSpace(*req.Provider); v != "" {
			p.Provider = v
		}
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
	if req.EmbeddingModel != nil {
		p.EmbeddingModel = *req.EmbeddingModel
	}
	if req.Capabilities != nil {
		p.Capabilities = model.JoinCapabilities(req.Capabilities)
	}
	if req.ExtraHeaders != nil {
		data, _ := json.Marshal(req.ExtraHeaders)
		p.ExtraHeaders = model.JSONBMap{}
		_ = json.Unmarshal(data, &p.ExtraHeaders)
	}
	if req.IsEnabled != nil {
		p.IsEnabled = *req.IsEnabled
	}
	if req.Priority != nil {
		p.Priority = *req.Priority
	}
	if req.Scope != nil {
		scope := strings.TrimSpace(*req.Scope)
		if scope != model.ProviderScopePrivate && scope != model.ProviderScopePublic {
			return nil, fmt.Errorf("%w: invalid scope", ErrProviderBadRequest)
		}
		if scope == model.ProviderScopePublic && !actor.IsAdmin {
			return nil, fmt.Errorf("%w: only admin can set public", ErrProviderForbidden)
		}
		// 非 admin 不能把别人的或已公开的改来改去——写权限已拦；此处仅允许 admin 改 scope
		if !actor.IsAdmin && p.Scope == model.ProviderScopePublic {
			return nil, ErrProviderForbidden
		}
		p.Scope = scope
		// 降为 private 时清掉全站默认标记
		if scope == model.ProviderScopePrivate {
			p.IsDefault = false
			p.IsDefaultEmbedding = false
		}
	}
	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	s.refreshCache()
	return p, nil
}

func (s *Service) Update(id string, req *UpdateProviderRequest) (*model.AIProvider, error) {
	return s.UpdateFor(id, req, Actor{IsAdmin: true})
}

// DeleteFor 删除（鉴权）
func (s *Service) DeleteFor(id string, actor Actor) error {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return ErrProviderNotFound
	}
	if !model.ProviderWritableBy(p, actor.UserID, actor.IsAdmin) {
		return ErrProviderForbidden
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.refreshCache()
	return nil
}

func (s *Service) Delete(id string) error {
	return s.DeleteFor(id, Actor{IsAdmin: true})
}

// SetDefaultFor 设默认：admin 可设全站（须 public）；普通用户只写个人偏好
func (s *Service) SetDefaultFor(id string, purpose string, actor Actor, personal bool) error {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return ErrProviderNotFound
	}
	if !model.ProviderVisibleTo(p, actor.UserID, actor.IsAdmin) {
		return ErrProviderForbidden
	}
	purpose = strings.TrimSpace(strings.ToLower(purpose))
	if purpose == "" {
		purpose = model.CapChat
	}
	if purpose != model.CapChat && purpose != model.CapEmbedding {
		return fmt.Errorf("%w: unknown purpose %s", ErrProviderBadRequest, purpose)
	}
	if !model.HasCapability(p.Capabilities, purpose) {
		return fmt.Errorf("%w: provider has no %s capability", ErrProviderBadRequest, purpose)
	}

	// 普通用户或显式 personal → 个人偏好
	if personal || !actor.IsAdmin {
		if actor.UserID == "" {
			return ErrProviderForbidden
		}
		return s.repo.SetPreferredProvider(actor.UserID, purpose, id)
	}

	// 全站默认：仅 public
	if p.Scope != model.ProviderScopePublic && p.Scope != "" {
		return fmt.Errorf("%w: site default must be public", ErrProviderBadRequest)
	}
	if !p.IsEnabled {
		return fmt.Errorf("%w: provider must be enabled", ErrProviderBadRequest)
	}
	if purpose == model.CapEmbedding {
		if err := s.repo.SetDefaultEmbedding(id); err != nil {
			return err
		}
	} else {
		if err := s.repo.SetDefault(id); err != nil {
			return err
		}
	}
	s.refreshCache()
	return nil
}

func (s *Service) SetDefault(id string, purpose string) error {
	return s.SetDefaultFor(id, purpose, Actor{IsAdmin: true}, false)
}

// ResolveForChat 按用户解析对话供应商：偏好 → 全站 public 默认 → 可见启用第一条
func (s *Service) ResolveForChat(userID string) (*model.AIProvider, error) {
	actor := ResolveActor(userID)
	prefChat, _ := s.repo.GetPreferredProviderIDs(userID)
	if prefChat != "" {
		if p, err := s.repo.GetByID(prefChat); err == nil &&
			p.IsEnabled &&
			model.HasCapability(p.Capabilities, model.CapChat) &&
			model.ProviderVisibleTo(p, actor.UserID, actor.IsAdmin) {
			return p, nil
		}
	}
	if p, err := s.repo.GetPublicDefault(); err == nil && p.IsEnabled && model.HasCapability(p.Capabilities, model.CapChat) {
		return p, nil
	}
	if p, err := s.repo.GetDefault(); err == nil && p.IsEnabled && model.HasCapability(p.Capabilities, model.CapChat) &&
		model.ProviderVisibleTo(p, actor.UserID, actor.IsAdmin) {
		return p, nil
	}
	if p, err := s.repo.GetFirstEnabledVisible(actor.UserID, actor.IsAdmin, model.CapChat); err == nil {
		return p, nil
	}
	return nil, fmt.Errorf("no active provider configured")
}

// ResolveForEmbedding 向量供应商解析
func (s *Service) ResolveForEmbedding(userID string) (*model.AIProvider, error) {
	actor := ResolveActor(userID)
	_, prefEmbed := s.repo.GetPreferredProviderIDs(userID)
	if prefEmbed != "" {
		if p, err := s.repo.GetByID(prefEmbed); err == nil &&
			p.IsEnabled &&
			model.HasCapability(p.Capabilities, model.CapEmbedding) &&
			model.ProviderVisibleTo(p, actor.UserID, actor.IsAdmin) {
			return p, nil
		}
	}
	if p, err := s.repo.GetPublicDefaultEmbedding(); err == nil && p.IsEnabled && model.HasCapability(p.Capabilities, model.CapEmbedding) {
		return p, nil
	}
	if p, err := s.repo.GetDefaultEmbedding(); err == nil && p.IsEnabled && model.HasCapability(p.Capabilities, model.CapEmbedding) &&
		model.ProviderVisibleTo(p, actor.UserID, actor.IsAdmin) {
		return p, nil
	}
	if p, err := s.repo.GetFirstEnabledVisible(actor.UserID, actor.IsAdmin, model.CapEmbedding); err == nil {
		return p, nil
	}
	return nil, fmt.Errorf("no embedding provider configured")
}

// GetActive 全站公开默认（无用户上下文）
func (s *Service) GetActive() (*model.AIProvider, error) {
	if s.cache != nil {
		if snap, ok := s.cache.Get(); ok && snap.ID != "" {
			return snapshotToProvider(snap), nil
		}
	}
	p, err := s.loadSitePublicActiveFromDB()
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

// GetActiveForUser 按用户取「当前生效」供应商（含偏好）
func (s *Service) GetActiveForUser(userID string) (*model.AIProvider, error) {
	return s.ResolveForChat(userID)
}

// TestConnectionFor 测连（须可写或至少可见；改状态需可写）
func (s *Service) TestConnectionFor(id string, actor Actor) (*TestResult, error) {
	p, err := s.GetFor(id, actor)
	if err != nil {
		return nil, err
	}
	return s.testConnection(p, model.ProviderWritableBy(p, actor.UserID, actor.IsAdmin))
}

func (s *Service) TestConnection(id string) (*TestResult, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrProviderNotFound
	}
	return s.testConnection(p, true)
}

func (s *Service) testConnection(p *model.AIProvider, updateStatus bool) (*TestResult, error) {
	apiKey, err := crypto.Decrypt(p.APIKey)
	if err != nil {
		return &TestResult{Success: false, Message: "failed to decrypt api key"}, nil
	}

	client := &http.Client{Timeout: 20 * time.Second}
	var req *http.Request

	switch strings.ToLower(strings.TrimSpace(p.Provider)) {
	case string(model.ProviderJina):
		base := strings.TrimRight(strings.TrimSpace(p.BaseURL), "/")
		if base == "" {
			base = "https://r.jina.ai"
		}
		req, err = http.NewRequest("GET", base+"/https://example.com", nil)
		if err != nil {
			return &TestResult{Success: false, Message: err.Error()}, nil
		}
		req.Header.Set("Accept", "text/markdown")
		if apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	case string(model.ProviderFirecrawl):
		base := strings.TrimRight(strings.TrimSpace(p.BaseURL), "/")
		if base == "" {
			base = "https://api.firecrawl.dev"
		}
		req, err = http.NewRequest("GET", base+"/v1/team/credit-usage", nil)
		if err != nil {
			return &TestResult{Success: false, Message: err.Error()}, nil
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
	default:
		testURL := p.BaseURL
		if testURL == "" {
			testURL = "https://api.openai.com/v1/models"
		}
		if testURL[len(testURL)-1] == '/' {
			testURL = testURL + "models"
		} else {
			testURL = testURL + "/models"
		}
		req, err = http.NewRequest("GET", testURL, nil)
		if err != nil {
			return &TestResult{Success: false, Message: err.Error()}, nil
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
	}

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
		if updateStatus {
			_ = s.repo.UpdateStatus(p.ID, "error", err.Error())
		}
		return result, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 400 {
		result.Success = true
		result.Message = fmt.Sprintf("连接成功 (HTTP %d)", resp.StatusCode)
		if updateStatus {
			_ = s.repo.UpdateStatus(p.ID, "active", "")
		}
	} else {
		result.Success = false
		result.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		var body bytes.Buffer
		_, _ = body.ReadFrom(resp.Body)
		if body.Len() > 0 {
			result.Message = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, body.String())
		}
		if updateStatus {
			_ = s.repo.UpdateStatus(p.ID, "error", result.Message)
		}
	}
	return result, nil
}

func (s *Service) decryptProvider(p *model.AIProvider) (baseURL, apiKey string, headers map[string]string, err error) {
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

// GetActiveForChat 按用户解析；userID 空则回退全站公开默认（并可用热缓存）
func (s *Service) GetActiveForChat(userID ...string) (baseURL, apiKey string, headers map[string]string, err error) {
	uid := ""
	if len(userID) > 0 {
		uid = userID[0]
	}

	if uid != "" {
		p, rerr := s.ResolveForChat(uid)
		if rerr != nil {
			return "", "", nil, rerr
		}
		return s.decryptProvider(p)
	}

	// 无用户：沿用全站快照热路径
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
	baseURL, apiKey, headers, err = s.decryptProvider(p)
	if err != nil {
		return "", "", nil, err
	}
	if s.cache != nil {
		if snap, ok := s.cache.Get(); ok {
			s.cache.PutHot(snap, apiKey, headers)
		}
	}
	return baseURL, apiKey, headers, nil
}

func (s *Service) GetActiveForEmbedding(userID ...string) (baseURL, apiKey string, headers map[string]string, modelName string, err error) {
	uid := ""
	if len(userID) > 0 {
		uid = userID[0]
	}
	var p *model.AIProvider
	if uid != "" {
		p, err = s.ResolveForEmbedding(uid)
	} else {
		p, err = s.loadEmbeddingFromDB()
	}
	if err != nil {
		return "", "", nil, "", err
	}
	modelName = model.EmbeddingModelName(p)
	if modelName == "" {
		return "", "", nil, "", fmt.Errorf("embedding provider has no embedding model")
	}
	baseURL, apiKey, headers, err = s.decryptProvider(p)
	if err != nil {
		return "", "", nil, "", err
	}
	return baseURL, apiKey, headers, modelName, nil
}

// PreferredIDs 供 handler 组装响应
func (s *Service) PreferredIDs(userID string) (chatID, embedID string) {
	return s.repo.GetPreferredProviderIDs(userID)
}

// MapServiceError 把业务错误映射成是否禁止/找不到
func MapServiceError(err error) (forbidden, notFound, badRequest bool) {
	if err == nil {
		return false, false, false
	}
	if errors.Is(err, ErrProviderForbidden) {
		return true, false, false
	}
	if errors.Is(err, ErrProviderNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
		return false, true, false
	}
	if errors.Is(err, ErrProviderBadRequest) {
		return false, false, true
	}
	return false, false, false
}
