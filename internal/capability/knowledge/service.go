package knowledge

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/capability/webfetch"
	"cogniforge/internal/infra/database"
	"cogniforge/internal/infra/trace"
	"cogniforge/internal/model"
	"cogniforge/internal/shared/secaudit"
)

// enforceKBSecurityLevel 金融扩展：KB_ENFORCE_SECURITY_LEVEL=1 时按人侧 clearance 过滤检索
func enforceKBSecurityLevel() bool {
	v := strings.TrimSpace(os.Getenv("KB_ENFORCE_SECURITY_LEVEL"))
	return v == "1" || strings.EqualFold(v, "true")
}

// authorizeSecurityLevelChange 抬到机密/绝密需 clearance≥目标，或 admin；降密/改公开·内部不拦
func (s *KnowledgeService) authorizeSecurityLevelChange(userID, oldLevel, newLevel string) error {
	newRank := model.SecurityLevelRank(newLevel)
	if newRank <= model.SecurityLevelRank(model.DocSecurityInternal) {
		return nil
	}
	oldRank := model.SecurityLevelRank(oldLevel)
	if newRank <= oldRank {
		return nil // 未抬高敏感档
	}
	clearance := s.resolveUserClearance(userID)
	if model.SecurityLevelRank(clearance) >= newRank {
		return nil
	}
	var user model.User
	if err := s.db.Select("role").Where("id = ?", userID).First(&user).Error; err == nil && user.Role == "admin" {
		return nil
	}
	return fmt.Errorf("无权将文档密级抬至 %s（需要不低于该级的 clearance，或管理员）", newLevel)
}

// resolveUserClearance 人员覆盖 → 角色上限 → 默认 secret（不限制）
func (s *KnowledgeService) resolveUserClearance(userID string) string {
	if userID == "" {
		return model.DocSecuritySecret
	}
	var user model.User
	if err := s.db.Select("id", "role", "max_security_level").Where("id = ?", userID).First(&user).Error; err != nil {
		return model.DocSecuritySecret
	}
	if lv := strings.TrimSpace(strings.ToLower(user.MaxSecurityLevel)); lv != "" {
		return lv
	}
	if user.Role == "" {
		return model.DocSecuritySecret
	}
	var role model.Role
	if err := s.db.Select("max_security_level").Where("code = ?", user.Role).First(&role).Error; err != nil {
		return model.DocSecuritySecret
	}
	if lv := strings.TrimSpace(strings.ToLower(role.MaxSecurityLevel)); lv != "" {
		return lv
	}
	return model.DocSecuritySecret
}

const (
	AllowedFileTypePDF  = "pdf"
	AllowedFileTypeTXT  = "txt"
	AllowedFileTypeMD   = "md"
	AllowedFileTypeDOCX = "docx"
	AllowedFileTypeHTML = "html"
	MaxFileSize         = 50 * 1024 * 1024
	UploadsDir          = "uploads"
	DocumentsDir        = "documents"
)

var allowedExtensions = map[string]string{
	".pdf":  AllowedFileTypePDF,
	".txt":  AllowedFileTypeTXT,
	".md":   AllowedFileTypeMD,
	".docx": AllowedFileTypeDOCX,
	".html": AllowedFileTypeHTML,
}

type KnowledgeService struct {
	db           *gorm.DB
	pythonClient *ServiceClient
}

func NewKnowledgeService(pythonClient *ServiceClient) *KnowledgeService {
	return &KnowledgeService{
		db:           database.DB,
		pythonClient: pythonClient,
	}
}

// ListFolders 知识库目录列表
func (s *KnowledgeService) ListFolders(userID string) ([]model.KnowledgeFolder, error) {
	var folders []model.KnowledgeFolder
	if err := s.db.Where("user_id = ?", userID).Order("sort_order ASC, created_at ASC").Find(&folders).Error; err != nil {
		return nil, fmt.Errorf("查询目录失败")
	}
	return folders, nil
}

// CreateFolder 创建知识库目录
func (s *KnowledgeService) CreateFolder(userID string, req *CreateFolderRequest) (*model.KnowledgeFolder, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("目录名称不能为空")
	}
	if len([]rune(name)) > 50 {
		return nil, fmt.Errorf("目录名称不能超过 50 字")
	}

	var n int64
	if err := s.db.Model(&model.KnowledgeFolder{}).
		Where("user_id = ? AND name = ?", userID, name).
		Count(&n).Error; err != nil {
		return nil, fmt.Errorf("查询目录失败")
	}
	if n > 0 {
		return nil, fmt.Errorf("目录已存在")
	}

	folder := model.KnowledgeFolder{
		ID:        generateID(),
		UserID:    userID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.db.Create(&folder).Error; err != nil {
		return nil, fmt.Errorf("创建目录失败")
	}
	return &folder, nil
}

func (s *KnowledgeService) ensureFolderOwned(userID, folderID string) error {
	if folderID == "" {
		return nil
	}
	var n int64
	if err := s.db.Model(&model.KnowledgeFolder{}).
		Where("id = ? AND user_id = ?", folderID, userID).
		Count(&n).Error; err != nil {
		return fmt.Errorf("查询目录失败")
	}
	if n == 0 {
		return fmt.Errorf("目录不存在")
	}
	return nil
}

// ListKnowledgeBases 获取知识库列表
func (s *KnowledgeService) ListKnowledgeBases(userID string) ([]model.KnowledgeBase, error) {
	var kbs []model.KnowledgeBase
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&kbs).Error; err != nil {
		return nil, fmt.Errorf("查询知识库列表失败")
	}
	return kbs, nil
}

// CreateKnowledgeBase 创建知识库
func (s *KnowledgeService) CreateKnowledgeBase(userID string, req *CreateKBRequest) (*model.KnowledgeBase, error) {
	if err := s.ensureFolderOwned(userID, req.FolderID); err != nil {
		return nil, err
	}

	kb := model.KnowledgeBase{
		ID:             generateID(),
		UserID:         userID,
		FolderID:       req.FolderID,
		Name:           req.Name,
		Description:    req.Description,
		VectorDB:       req.VectorDB,
		EmbeddingModel: req.EmbeddingModel,
		ChunkSize:      512,
		ChunkOverlap:   50,
		SplitMode:      "recursive",
		SplitSeparator: "\n\n",
		RagTopK:              5,
		RagMinScore:          0.3,
		RagTopKSemantic:      5,
		RagMinScoreSemantic:  0.3,
		RagTopKKeyword:       5,
		RagMinScoreKeyword:   0.3,
		RagTopKHybrid:        5,
		RagMinScoreHybrid:    0.3,
		HybridSemanticWeight: 0.7,
		SearchMode:           "semantic",
		Visibility:           "private",
		Status:               "active",
		Metadata:             model.JSONBMap{},
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if kb.VectorDB == "" {
		kb.VectorDB = "pgvector"
	}
	if kb.EmbeddingModel == "" {
		kb.EmbeddingModel = "text-embedding-ada-002"
	}
	storageDir, err := s.allocStorageDir(userID, kb.Name, "")
	if err != nil {
		return nil, err
	}
	kb.StorageDir = storageDir
	applyCreateParseDefaults(&kb, req)

	if err := s.db.Create(&kb).Error; err != nil {
		return nil, fmt.Errorf("创建知识库失败")
	}

	return &kb, nil
}

// GetKnowledgeBase 获取知识库详情
func (s *KnowledgeService) GetKnowledgeBase(userID, kbID string) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}
	return &kb, nil
}

// UpdateKnowledgeBase 更新知识库
func (s *KnowledgeService) UpdateKnowledgeBase(userID, kbID string, req *UpdateKBRequest) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}

	if req.Name != "" && req.Name != kb.Name {
		newDir, err := s.allocStorageDir(userID, req.Name, kb.ID)
		if err != nil {
			return nil, err
		}
		if err := s.moveUploadDir(userID, kb.StorageDir, newDir, kb.ID); err != nil {
			return nil, err
		}
		kb.Name = req.Name
		kb.StorageDir = newDir
	}
	if req.Description != "" {
		kb.Description = req.Description
	}
	if req.FolderID != nil {
		if err := s.ensureFolderOwned(userID, *req.FolderID); err != nil {
			return nil, err
		}
		kb.FolderID = *req.FolderID
	}
	if req.VectorDB != "" {
		kb.VectorDB = req.VectorDB
	}
	if req.EmbeddingModel != "" {
		kb.EmbeddingModel = req.EmbeddingModel
	}
	applyUpdateParseDefaults(&kb, req)
	if err := s.validateKBRerankProviders(userID, &kb); err != nil {
		return nil, err
	}
	if req.Status != "" {
		kb.Status = req.Status
	}
	kb.UpdatedAt = time.Now()

	if err := s.db.Save(&kb).Error; err != nil {
		return nil, fmt.Errorf("更新知识库失败")
	}

	return &kb, nil
}

// DeleteKnowledgeBase 删除知识库
func (s *KnowledgeService) DeleteKnowledgeBase(userID, kbID string) error {
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("知识库不存在")
		}
		return fmt.Errorf("查询知识库失败")
	}

	if err := s.db.Delete(&kb).Error; err != nil {
		return fmt.Errorf("删除知识库失败")
	}
	return nil
}

// ListDocuments 获取文档列表
func (s *KnowledgeService) ListDocuments(userID, kbID string) ([]model.Document, error) {
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}

	var docs []model.Document
	if err := s.db.Where("knowledge_base_id = ?", kbID).Order("created_at DESC").Find(&docs).Error; err != nil {
		return nil, fmt.Errorf("查询文档列表失败")
	}
	return docs, nil
}

const maxPreviewRunes = 120000

// GetDocument 文档详情；文本类附带正文预览
func (s *KnowledgeService) GetDocument(userID, kbID, docID string) (*DocumentPreview, error) {
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}

	var doc model.Document
	if err := s.db.Where("id = ? AND knowledge_base_id = ?", docID, kbID).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文档不存在")
		}
		return nil, fmt.Errorf("查询文档失败")
	}

	out := &DocumentPreview{Document: doc}
	switch doc.FileType {
	case AllowedFileTypeTXT, AllowedFileTypeMD, AllowedFileTypeHTML:
		// 预览/编辑用原文；HTML 不再剥标签（搜索仍走 strip）
		content, err := readDocumentRaw(doc.FilePath)
		if err != nil {
			out.Previewable = false
			out.PreviewNote = err.Error()
			return out, nil
		}
		runes := []rune(content)
		if len(runes) > maxPreviewRunes {
			content = string(runes[:maxPreviewRunes])
			out.Truncated = true
			out.Editable = false
			out.PreviewNote = "正文过长，仅可预览前段，暂不支持在线编辑"
		} else {
			out.Editable = true
		}
		out.Content = content
		out.Previewable = true
	default:
		out.Previewable = false
		out.PreviewNote = "该格式暂不支持在线预览，请下载后查看"
	}
	return out, nil
}

// UpdateDocumentContent 覆盖文本类文档文件内容，默认触发重新解析
func (s *KnowledgeService) UpdateDocumentContent(ctx context.Context, userID, kbID, docID string, req *UpdateDocumentContentRequest) (*DocumentPreview, error) {
	if req == nil {
		return nil, fmt.Errorf("请求无效")
	}
	if len(req.Content) > MaxFileSize {
		return nil, fmt.Errorf("正文超过大小限制（%dMB）", MaxFileSize/(1024*1024))
	}

	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}

	var doc model.Document
	if err := s.db.Where("id = ? AND knowledge_base_id = ?", docID, kbID).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文档不存在")
		}
		return nil, fmt.Errorf("查询文档失败")
	}

	switch doc.FileType {
	case AllowedFileTypeTXT, AllowedFileTypeMD, AllowedFileTypeHTML:
	default:
		return nil, fmt.Errorf("该格式不支持在线编辑")
	}

	if err := writeDocumentContent(doc.FilePath, req.Content); err != nil {
		return nil, err
	}

	now := time.Now()
	size := int64(len([]byte(req.Content)))
	if err := s.db.Model(&doc).Updates(map[string]any{
		"file_size":  size,
		"updated_at": now,
		"error":      "",
	}).Error; err != nil {
		return nil, fmt.Errorf("更新文档元数据失败")
	}
	doc.FileSize = size
	doc.UpdatedAt = now

	reparse := true
	if req.Reparse != nil {
		reparse = *req.Reparse
	}
	if reparse {
		if _, err := s.ReparseDocument(ctx, userID, kbID, docID); err != nil {
			slog.Warn("保存正文后重新解析失败", "doc_id", docID, "error", err.Error())
			return nil, fmt.Errorf("正文已保存，但重新解析失败: %w", err)
		}
	}

	return s.GetDocument(userID, kbID, docID)
}

// UpdateDocumentLifecycle 更新业务状态与有效期（与解析 status 无关）
func (s *KnowledgeService) UpdateDocumentLifecycle(userID, kbID, docID string, req *UpdateDocumentLifecycleRequest) (*model.Document, error) {
	if req == nil {
		return nil, fmt.Errorf("请求无效")
	}
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}
	var doc model.Document
	if err := s.db.Where("id = ? AND knowledge_base_id = ?", docID, kbID).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文档不存在")
		}
		return nil, fmt.Errorf("查询文档失败")
	}

	updates := map[string]any{}
	if req.LifecycleStatus != nil {
		st := strings.TrimSpace(*req.LifecycleStatus)
		switch st {
		case model.DocLifecycleDraft, model.DocLifecycleReview, model.DocLifecyclePublished, model.DocLifecycleAbolished:
			updates["lifecycle_status"] = st
		default:
			return nil, fmt.Errorf("无效生命周期状态")
		}
	}
	if req.SecurityLevel != nil {
		lv := strings.TrimSpace(strings.ToLower(*req.SecurityLevel))
		switch lv {
		case model.DocSecurityPublic, model.DocSecurityInternal, model.DocSecurityConfidential, model.DocSecuritySecret:
			oldLv := doc.SecurityLevel
			if err := s.authorizeSecurityLevelChange(userID, oldLv, lv); err != nil {
				s.writeSecurityAudit(userID, "document_security_level_change", "document", docID, kbID, oldLv, lv, "denied", err.Error())
				return nil, err
			}
			if oldLv != lv {
				updates["security_level"] = lv
			}
		default:
			return nil, fmt.Errorf("无效密级")
		}
	}
	if req.ClearEffective {
		updates["effective_at"] = nil
	} else if req.EffectiveAt != nil {
		updates["effective_at"] = *req.EffectiveAt
	}
	if req.ClearExpire {
		updates["expire_at"] = nil
	} else if req.ExpireAt != nil {
		updates["expire_at"] = *req.ExpireAt
	}
	if len(updates) == 0 {
		return &doc, nil
	}
	oldSec := doc.SecurityLevel
	if err := s.db.Model(&doc).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新文档生命周期失败")
	}
	if newSec, ok := updates["security_level"].(string); ok && newSec != oldSec {
		s.writeSecurityAudit(userID, "document_security_level_change", "document", docID, kbID, oldSec, newSec, "success", "")
	}
	_ = s.db.Where("id = ?", docID).First(&doc).Error
	return &doc, nil
}

func (s *KnowledgeService) writeSecurityAudit(actorID, action, resType, resID, kbID, oldV, newV, result, detail string) {
	secaudit.Write(s.db, actorID, action, resType, resID, kbID, oldV, newV, result, detail)
}

// ListSecurityAudits 管理员查看密级审计（新在前）
func (s *KnowledgeService) ListSecurityAudits(limit int) ([]model.SecurityAuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []model.SecurityAuditLog
	if err := s.db.Order("created_at DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询密级审计失败")
	}
	return rows, nil
}

// DeleteDocument 删除文档
func (s *KnowledgeService) DeleteDocument(userID, kbID, docID string) error {
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("知识库不存在")
		}
		return fmt.Errorf("查询知识库失败")
	}

	var doc model.Document
	if err := s.db.Where("id = ? AND knowledge_base_id = ?", docID, kbID).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("文档不存在")
		}
		return fmt.Errorf("查询文档失败")
	}

	// 删除向量
	if err := s.pythonClient.DeleteDocument(&DeleteDocumentRequest{Context: context.Background(), DocumentID: docID, CollectionName: kbID}); err != nil {
		slog.Warn("删除文档向量失败，继续删除记录", "doc_id", docID, "error", err.Error())
	}

	if err := s.db.Delete(&doc).Error; err != nil {
		return fmt.Errorf("删除文档失败")
	}

	s.db.Model(&kb).Update("doc_count", gorm.Expr("doc_count - 1"))
	return nil
}

// ReparseDocument 重新解析文档：先删除旧向量，重置状态，再重新处理
func (s *KnowledgeService) ReparseDocument(ctx context.Context, userID, kbID, docID string) (*model.Document, error) {
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}

	var doc model.Document
	if err := s.db.Where("id = ? AND knowledge_base_id = ?", docID, kbID).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文档不存在")
		}
		return nil, fmt.Errorf("查询文档失败")
	}

	// 1. 删除旧的向量数据
	if err := s.pythonClient.DeleteDocument(&DeleteDocumentRequest{Context: ctx, DocumentID: docID, CollectionName: kbID}); err != nil {
		slog.Warn("删除文档旧向量失败，继续重置状态", "doc_id", docID, "error", err.Error())
	}

	// 2. 重置文档状态
	now := time.Now()
	if err := s.db.Model(&doc).Updates(map[string]any{
		"status":       "pending",
		"chunk_count":  0,
		"vector_count": 0,
		"error":        "",
		"updated_at":   now,
	}).Error; err != nil {
		return nil, fmt.Errorf("重置文档状态失败")
	}

	// 3. 触发异步重新处理（从 context 提取 trace ID 传递下去）
	go func() {
		s.processDocumentAsync(docID, trace.GetTraceIDFromContext(ctx))
	}()

	doc.Status = "pending"
	doc.ChunkCount = 0
	doc.VectorCount = 0
	doc.Error = ""
	doc.UpdatedAt = now
	return &doc, nil
}

// UploadDocument 上传文档（opts 为空时用知识库默认分段）
func (s *KnowledgeService) UploadDocument(ctx context.Context, userID, kbID string, file *multipart.FileHeader, opts *ParseOptions) (*model.Document, error) {
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}

	if file.Size > MaxFileSize {
		return nil, fmt.Errorf("文件大小不能超过 %dMB", MaxFileSize/(1024*1024))
	}

	fileType, ok := getFileType(file.Filename)
	if !ok {
		return nil, fmt.Errorf("不支持的文件类型，支持：PDF、TXT、MD、DOCX、HTML")
	}

	dir, err := ensureUploadDir(s.userDirName(userID), kbStorageKey(&kb))
	if err != nil {
		return nil, fmt.Errorf("创建上传目录失败")
	}

	docID := generateDocID()
	ext := getFileExt(file.Filename)
	savedFilename := docID + ext
	destPath := filepath.Join(dir, savedFilename)

	if err := saveUploadedFile(file, destPath); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	resolved := resolveParseOptions(&kb, opts)
	meta := model.JSONBMap{
		"chunk_size":      resolved.ChunkSize,
		"chunk_overlap":   resolved.ChunkOverlap,
		"split_mode":      resolved.SplitMode,
		"split_separator": resolved.SplitSeparator,
	}

	doc := model.Document{
		ID:              docID,
		KnowledgeBaseID: kbID,
		UserID:          userID,
		Name:            file.Filename,
		FileName:        file.Filename,
		FileSize:        file.Size,
		FileType:        fileType,
		FilePath:        destPath,
		Status:          "pending",
		ChunkCount:      0,
		VectorCount:     0,
		Metadata:        meta,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.db.Create(&doc).Error; err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("创建文档记录失败")
	}

	s.db.Model(&kb).Update("doc_count", gorm.Expr("doc_count + 1"))

	go func() {
		s.processDocumentAsync(docID, trace.GetTraceIDFromContext(ctx))
	}()

	return &doc, nil
}

// CreateDocumentFromText 把自定义文本写成 md/txt 后入库并异步解析
func (s *KnowledgeService) CreateDocumentFromText(ctx context.Context, userID, kbID string, req *CreateTextDocumentRequest) (*model.Document, error) {
	name := strings.TrimSpace(req.Name)
	content := strings.TrimSpace(req.Content)
	if name == "" {
		return nil, fmt.Errorf("请填写名称")
	}
	if content == "" {
		return nil, fmt.Errorf("请填写正文")
	}
	if len(content) > MaxFileSize {
		return nil, fmt.Errorf("正文过长，单文件最多约 %dMB", MaxFileSize/(1024*1024))
	}

	fileType := strings.ToLower(strings.TrimSpace(req.FileType))
	if fileType != AllowedFileTypeTXT {
		fileType = AllowedFileTypeMD
	}
	ext := "." + fileType
	if !strings.HasSuffix(strings.ToLower(name), ext) {
		name = name + ext
	}

	opts := &ParseOptions{
		ChunkSize:      req.ChunkSize,
		ChunkOverlap:   req.ChunkOverlap,
		SplitMode:      req.SplitMode,
		SplitSeparator: req.SplitSeparator,
	}
	return s.createDocumentFromBytes(ctx, userID, kbID, name, fileType, []byte(content), opts, map[string]interface{}{
		"source": "text",
	})
}

// CreateDocumentFromURL 抓取网页后入库（Jina / Firecrawl / 简易 GET）
func (s *KnowledgeService) CreateDocumentFromURL(ctx context.Context, userID, kbID string, req *CreateURLDocumentRequest) (*model.Document, error) {
	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		return nil, fmt.Errorf("请填写网页地址")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("网页地址无效，请使用 http:// 或 https:// 开头")
	}

	fetched, err := webfetch.Fetch(ctx, s.db, userID, parsed.String(), req.Engine)
	if err != nil {
		return nil, err
	}

	fileType := AllowedFileTypeHTML
	ext := ".html"
	if fetched.FileType == "md" || fetched.FileType == AllowedFileTypeMD {
		fileType = AllowedFileTypeMD
		ext = ".md"
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = parsed.Host
		if base := filepath.Base(parsed.Path); base != "" && base != "/" && base != "." {
			name = base
		}
	}
	lowerName := strings.ToLower(name)
	if !strings.HasSuffix(lowerName, ext) {
		for _, old := range []string{".html", ".htm", ".md"} {
			if strings.HasSuffix(lowerName, old) {
				name = name[:len(name)-len(old)]
				break
			}
		}
		if strings.TrimSpace(name) == "" {
			name = parsed.Host
		}
		name = name + ext
	}

	opts := &ParseOptions{
		ChunkSize:      req.ChunkSize,
		ChunkOverlap:   req.ChunkOverlap,
		SplitMode:      req.SplitMode,
		SplitSeparator: req.SplitSeparator,
	}
	return s.createDocumentFromBytes(ctx, userID, kbID, name, fileType, fetched.Content, opts, map[string]interface{}{
		"source": "web",
		"url":    parsed.String(),
		"engine": fetched.Engine,
	})
}

func (s *KnowledgeService) createDocumentFromBytes(
	ctx context.Context,
	userID, kbID, displayName, fileType string,
	data []byte,
	opts *ParseOptions,
	extraMeta map[string]interface{},
) (*model.Document, error) {
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("内容为空")
	}
	if int64(len(data)) > MaxFileSize {
		return nil, fmt.Errorf("文件大小不能超过 %dMB", MaxFileSize/(1024*1024))
	}

	dir, err := ensureUploadDir(s.userDirName(userID), kbStorageKey(&kb))
	if err != nil {
		return nil, fmt.Errorf("创建上传目录失败")
	}

	docID := generateDocID()
	ext := "." + fileType
	destPath := filepath.Join(dir, docID+ext)
	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	resolved := resolveParseOptions(&kb, opts)
	meta := model.JSONBMap{
		"chunk_size":      resolved.ChunkSize,
		"chunk_overlap":   resolved.ChunkOverlap,
		"split_mode":      resolved.SplitMode,
		"split_separator": resolved.SplitSeparator,
	}
	for k, v := range extraMeta {
		meta[k] = v
	}

	doc := model.Document{
		ID:              docID,
		KnowledgeBaseID: kbID,
		UserID:          userID,
		Name:            displayName,
		FileName:        displayName,
		FileSize:        int64(len(data)),
		FileType:        fileType,
		FilePath:        destPath,
		Status:          "pending",
		ChunkCount:      0,
		VectorCount:     0,
		Metadata:        meta,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := s.db.Create(&doc).Error; err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("创建文档记录失败")
	}
	s.db.Model(&kb).Update("doc_count", gorm.Expr("doc_count + 1"))

	go func() {
		s.processDocumentAsync(docID, trace.GetTraceIDFromContext(ctx))
	}()
	return &doc, nil
}

// PreviewChunks 临时保存文件并调用 AI 预览分块（不入库、不向量化）
func (s *KnowledgeService) PreviewChunks(ctx context.Context, userID, kbID string, file *multipart.FileHeader, opts *ParseOptions) (*ChunkPreviewResponse, error) {
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}

	if file.Size > MaxFileSize {
		return nil, fmt.Errorf("文件大小不能超过 %dMB", MaxFileSize/(1024*1024))
	}
	if _, ok := getFileType(file.Filename); !ok {
		return nil, fmt.Errorf("不支持的文件类型，支持：PDF、TXT、MD、DOCX、HTML")
	}

	dir, err := ensureUploadDir(s.userDirName(userID), kbStorageKey(&kb))
	if err != nil {
		return nil, fmt.Errorf("创建上传目录失败")
	}
	previewDir := filepath.Join(dir, "_preview")
	if err := os.MkdirAll(previewDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建预览目录失败")
	}

	tmpName := "preview_" + strconv.FormatInt(time.Now().UnixNano(), 10) + getFileExt(file.Filename)
	destPath := filepath.Join(previewDir, tmpName)
	if err := saveUploadedFile(file, destPath); err != nil {
		return nil, fmt.Errorf("保存预览文件失败: %w", err)
	}
	defer os.Remove(destPath)

	resolved := resolveParseOptions(&kb, opts)
	aiPath := toSharedUploadPath(destPath)
	result, err := s.pythonClient.PreviewChunks(&PreviewChunksRequest{
		Context:      ctx,
		FilePath:     aiPath,
		ChunkSize:    resolved.ChunkSize,
		ChunkOverlap: resolved.ChunkOverlap,
		SplitMode:    resolved.SplitMode,
		Separator:    resolved.SplitSeparator,
		MaxChunks:    20,
	})
	if err != nil {
		return nil, fmt.Errorf("预览分块失败（请确认 PDF 可提取文字且 AI 服务可读上传目录）: %w", err)
	}
	return result, nil
}

// SearchKnowledge 知识库检索（HTTP 入口，带 Gin trace）
func (s *KnowledgeService) SearchKnowledge(c *gin.Context, userID, kbID string, req *SearchRequest) (*SearchResponse, error) {
	ctx := context.WithValue(c.Request.Context(), trace.TraceIDKey, trace.GetTraceIDFromGin(c))
	return s.SearchByUser(ctx, userID, kbID, req)
}

// SearchByUser 供 Chat 等内部调用：按用户校验知识库后做向量检索
func (s *KnowledgeService) SearchByUser(ctx context.Context, userID, kbID string, req *SearchRequest) (*SearchResponse, error) {
	var kb model.KnowledgeBase
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}

	startTime := time.Now()

	mode := normalizeSearchMode(req.SearchMode)
	if req.SearchMode == "" {
		mode = normalizeSearchMode(kb.SearchMode)
	}
	topK, minScore := resolveModeRetrievalParams(&kb, mode, req.TopK, req.MinScore)

	hy := kb.HybridSemanticWeight
	if req.HybridSemanticWeight != nil {
		hy = *req.HybridSemanticWeight
	}
	hy = clampHybridSemanticWeight(hy)
	hyPtr := hy

	rerank, rerankProv := resolveSearchRerank(&kb, mode, req)

	enforceSec := enforceKBSecurityLevel()
	clearance := ""
	if enforceSec {
		clearance = s.resolveUserClearance(userID)
	}

	// 多取一些再按生命周期/密级过滤，尽量凑满 TopK
	fetchK := topK
	if !req.IncludeDrafts || !req.IgnoreValidity || enforceSec {
		fetchK = topK * 3
		if fetchK < topK {
			fetchK = topK
		}
		if fetchK > 50 {
			fetchK = 50
		}
	}

	pythonResp, err := s.pythonClient.Search(&SearchRequest{
		Context:              ctx,
		Query:                req.Query,
		CollectionName:       kbID,
		TopK:                 fetchK,
		MinScore:             minScore,
		SearchMode:           mode,
		Rerank:               rerank,
		RerankProviderID:     rerankProv,
		HybridSemanticWeight: &hyPtr,
	})
	if err != nil {
		return nil, fmt.Errorf("检索失败: %w", err)
	}

	allowed := s.searchableDocIDs(ctx, kbID, req.IncludeDrafts, req.IgnoreValidity, clearance, enforceSec)

	var results []SearchResult
	for _, r := range pythonResp.Results {
		docID := ""
		docName := ""
		if r.Metadata != nil {
			if id, ok := r.Metadata["document_id"].(string); ok {
				docID = id
			}
			if n, ok := r.Metadata["document_name"].(string); ok {
				docName = n
			} else if n, ok := r.Metadata["file_name"].(string); ok {
				docName = n
			}
		}
		if docID == "" {
			// 旧向量无 document_id：尝试从 chunk_id 前缀推断
			docID = inferDocIDFromChunk(r.ChunkID)
		}
		if allowed != nil {
			if docID == "" || !allowed[docID] {
				continue
			}
		}
		results = append(results, SearchResult{
			DocumentID:   docID,
			DocumentName: docName,
			ChunkID:      r.ChunkID,
			Content:      r.Content,
			Score:        r.Score,
		})
		if len(results) >= topK {
			break
		}
	}

	return &SearchResponse{
		Results:  results,
		Total:    len(results),
		Query:    req.Query,
		Duration: time.Since(startTime).Milliseconds(),
	}, nil
}

// searchableDocIDs 返回当前可被召回的文档 ID 集合；
// includeDrafts+ignoreValidity 且不强制密级时返回 nil（不过滤）
func (s *KnowledgeService) searchableDocIDs(
	ctx context.Context,
	kbID string,
	includeDrafts, ignoreValidity bool,
	maxClearance string,
	enforceSecurity bool,
) map[string]bool {
	if includeDrafts && ignoreValidity && !enforceSecurity {
		return nil
	}
	var docs []model.Document
	q := s.db.WithContext(ctx).Where("knowledge_base_id = ?", kbID)
	if !includeDrafts {
		q = q.Where("lifecycle_status = ? OR lifecycle_status = '' OR lifecycle_status IS NULL", model.DocLifecyclePublished)
	} else {
		q = q.Where("lifecycle_status <> ?", model.DocLifecycleAbolished)
	}
	if err := q.Find(&docs).Error; err != nil {
		return map[string]bool{}
	}
	now := time.Now()
	out := make(map[string]bool, len(docs))
	for i := range docs {
		d := &docs[i]
		ok := false
		if includeDrafts {
			if d.LifecycleStatus == model.DocLifecycleAbolished {
				continue
			}
			if !ignoreValidity {
				if d.EffectiveAt != nil && now.Before(*d.EffectiveAt) {
					continue
				}
				if d.ExpireAt != nil && !now.Before(*d.ExpireAt) {
					continue
				}
			}
			ok = true
		} else if d.IsSearchableNow(now) || (ignoreValidity && (d.LifecycleStatus == "" || d.LifecycleStatus == model.DocLifecyclePublished)) {
			ok = true
		}
		if !ok {
			continue
		}
		if enforceSecurity && !d.AccessibleByClearance(maxClearance) {
			continue
		}
		out[d.ID] = true
	}
	return out
}

func inferDocIDFromChunk(chunkID string) string {
	if chunkID == "" {
		return ""
	}
	const sep = "_chunk_"
	if i := strings.LastIndex(chunkID, sep); i > 0 {
		return chunkID[:i]
	}
	return ""
}

// ResolveKBIDsForDocuments 根据文档 ID 解析所属知识库（用户隔离）
func (s *KnowledgeService) ResolveKBIDsForDocuments(ctx context.Context, userID string, docIDs []string) []string {
	if len(docIDs) == 0 {
		return nil
	}
	var docs []model.Document
	if err := s.db.WithContext(ctx).
		Select("knowledge_base_id").
		Where("user_id = ? AND id IN ?", userID, docIDs).
		Find(&docs).Error; err != nil {
		slog.Warn("resolve document KB failed", "error", err)
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, d := range docs {
		if d.KnowledgeBaseID == "" {
			continue
		}
		if _, ok := seen[d.KnowledgeBaseID]; ok {
			continue
		}
		seen[d.KnowledgeBaseID] = struct{}{}
		out = append(out, d.KnowledgeBaseID)
	}
	return out
}

// ============ 辅助函数 ============

func (s *KnowledgeService) processDocumentAsync(docID, parentTraceID string) {
	s.db.Model(&model.Document{}).Where("id = ?", docID).Updates(map[string]any{
		"status":     "processing",
		"updated_at": time.Now(),
	})

	var doc model.Document
	if err := s.db.Where("id = ?", docID).First(&doc).Error; err != nil {
		slog.Error("查询文档失败", "doc_id", docID, "error", err.Error())
		return
	}

	// 为后台任务生成 context，优先使用上游传入的 trace ID
	traceID := parentTraceID
	if traceID == "" {
		traceID = trace.GenerateTraceID()
	}
	ctx := context.WithValue(context.Background(), trace.TraceIDKey, traceID)

	slog.Info("开始处理文档", "doc_id", docID, "file", doc.Name, "trace_id", traceID)

	var kb model.KnowledgeBase
	_ = s.db.Where("id = ?", doc.KnowledgeBaseID).First(&kb).Error
	opts := parseOptionsFromMetadata(doc.Metadata, &kb)

	// 调用 Python 服务进行文档处理和向量化
	result, err := s.pythonClient.ProcessDocument(&ProcessRequest{
		Context:        ctx,
		FilePath:       toSharedUploadPath(doc.FilePath),
		DocumentID:     docID,
		CollectionName: doc.KnowledgeBaseID,
		Metadata: map[string]interface{}{
			"file_type": doc.FileType,
			"file_name": doc.Name,
		},
		ChunkSize:    opts.ChunkSize,
		ChunkOverlap: opts.ChunkOverlap,
		SplitMode:    opts.SplitMode,
		Separator:    opts.SplitSeparator,
	})

	if err != nil || !result.Success {
		errMsg := ""
		if err != nil {
			errMsg = fmt.Sprintf("调用 Python 服务失败: %v", err)
		} else if result.Error != "" {
			errMsg = result.Error
		}
		slog.Error("文档处理失败", "doc_id", docID, "error", errMsg, "trace_id", traceID)
		s.db.Model(&doc).Updates(map[string]any{
			"status": "failed",
			"error":  errMsg,
		})
		return
	}

	s.db.Model(&doc).Updates(map[string]any{
		"chunk_count":  result.ChunksCreated,
		"vector_count": result.ChunksCreated,
		"status":       "completed",
		"updated_at":   time.Now(),
	})
	slog.Info("文档处理完成", "doc_id", docID, "chunks", result.ChunksCreated, "trace_id", traceID)
}

func getFileExt(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}

func getFileType(filename string) (string, bool) {
	ext := getFileExt(filename)
	fileType, ok := allowedExtensions[ext]
	return fileType, ok
}

func generateDocID() string {
	return fmt.Sprintf("doc_%s_%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)
}

func applyCreateParseDefaults(kb *model.KnowledgeBase, req *CreateKBRequest) {
	if req.ChunkSize != nil && *req.ChunkSize > 0 {
		kb.ChunkSize = clampChunkSize(*req.ChunkSize)
	}
	if req.ChunkOverlap != nil && *req.ChunkOverlap >= 0 {
		kb.ChunkOverlap = clampOverlap(*req.ChunkOverlap, kb.ChunkSize)
	}
	if req.SplitMode != "" {
		kb.SplitMode = normalizeSplitMode(req.SplitMode)
	}
	if req.SplitSeparator != "" {
		kb.SplitSeparator = req.SplitSeparator
	}
	applyRetrievalConfig(kb, retrievalConfigPatch{
		RagTopK:              req.RagTopK,
		RagMinScore:          req.RagMinScore,
		RagTopKSemantic:      req.RagTopKSemantic,
		RagMinScoreSemantic:  req.RagMinScoreSemantic,
		RagTopKKeyword:       req.RagTopKKeyword,
		RagMinScoreKeyword:   req.RagMinScoreKeyword,
		RagTopKHybrid:        req.RagTopKHybrid,
		RagMinScoreHybrid:    req.RagMinScoreHybrid,
		HybridSemanticWeight: req.HybridSemanticWeight,
		SearchMode:           req.SearchMode,
	})
	if req.RerankSemantic != nil {
		kb.RerankSemantic = *req.RerankSemantic
	}
	if req.RerankKeyword != nil {
		kb.RerankKeyword = *req.RerankKeyword
	}
	if req.RerankHybrid != nil {
		kb.RerankHybrid = *req.RerankHybrid
	}
	if req.RerankProviderSemantic != nil {
		kb.RerankProviderSemantic = strings.TrimSpace(*req.RerankProviderSemantic)
	}
	if req.RerankProviderKeyword != nil {
		kb.RerankProviderKeyword = strings.TrimSpace(*req.RerankProviderKeyword)
	}
	if req.RerankProviderHybrid != nil {
		kb.RerankProviderHybrid = strings.TrimSpace(*req.RerankProviderHybrid)
	}
	if req.Visibility != "" {
		// 阶段三：暂仅 private；org 等后续 ACL
		kb.Visibility = "private"
	}
}

func applyUpdateParseDefaults(kb *model.KnowledgeBase, req *UpdateKBRequest) {
	if req.ChunkSize != nil && *req.ChunkSize > 0 {
		kb.ChunkSize = clampChunkSize(*req.ChunkSize)
	}
	if req.ChunkOverlap != nil && *req.ChunkOverlap >= 0 {
		kb.ChunkOverlap = clampOverlap(*req.ChunkOverlap, kb.ChunkSize)
	}
	if req.SplitMode != "" {
		kb.SplitMode = normalizeSplitMode(req.SplitMode)
	}
	if req.SplitSeparator != nil {
		kb.SplitSeparator = *req.SplitSeparator
	}
	applyRetrievalConfig(kb, retrievalConfigPatch{
		RagTopK:              req.RagTopK,
		RagMinScore:          req.RagMinScore,
		RagTopKSemantic:      req.RagTopKSemantic,
		RagMinScoreSemantic:  req.RagMinScoreSemantic,
		RagTopKKeyword:       req.RagTopKKeyword,
		RagMinScoreKeyword:   req.RagMinScoreKeyword,
		RagTopKHybrid:        req.RagTopKHybrid,
		RagMinScoreHybrid:    req.RagMinScoreHybrid,
		HybridSemanticWeight: req.HybridSemanticWeight,
		SearchMode:           req.SearchMode,
	})
	if req.RerankSemantic != nil {
		kb.RerankSemantic = *req.RerankSemantic
	}
	if req.RerankKeyword != nil {
		kb.RerankKeyword = *req.RerankKeyword
	}
	if req.RerankHybrid != nil {
		kb.RerankHybrid = *req.RerankHybrid
	}
	if req.RerankProviderSemantic != nil {
		kb.RerankProviderSemantic = strings.TrimSpace(*req.RerankProviderSemantic)
	}
	if req.RerankProviderKeyword != nil {
		kb.RerankProviderKeyword = strings.TrimSpace(*req.RerankProviderKeyword)
	}
	if req.RerankProviderHybrid != nil {
		kb.RerankProviderHybrid = strings.TrimSpace(*req.RerankProviderHybrid)
	}
	if req.Visibility != "" {
		kb.Visibility = "private"
	}
}

func (s *KnowledgeService) validateKBRerankProviders(userID string, kb *model.KnowledgeBase) error {
	checks := []struct {
		on  bool
		id  string
		tag string
	}{
		{kb.RerankSemantic, kb.RerankProviderSemantic, "语义"},
		{kb.RerankKeyword, kb.RerankProviderKeyword, "关键字"},
		{kb.RerankHybrid, kb.RerankProviderHybrid, "混合"},
	}
	for _, c := range checks {
		if !c.on {
			continue
		}
		if strings.TrimSpace(c.id) == "" {
			return fmt.Errorf("%s检索已开启 Rerank，请选择 Rerank 模型", c.tag)
		}
		if err := s.ensureRerankProvider(userID, c.id); err != nil {
			return err
		}
	}
	// 未开启时若仍写了 id，也校验合法性（避免脏数据）
	for _, id := range []string{kb.RerankProviderSemantic, kb.RerankProviderKeyword, kb.RerankProviderHybrid} {
		if strings.TrimSpace(id) == "" {
			continue
		}
		if err := s.ensureRerankProvider(userID, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *KnowledgeService) ensureRerankProvider(userID, providerID string) error {
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return nil
	}
	var p model.AIProvider
	if err := s.db.Where("id = ?", providerID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("Rerank 模型不存在")
		}
		return fmt.Errorf("查询 Rerank 模型失败")
	}
	if !p.IsEnabled {
		return fmt.Errorf("Rerank 模型未启用，请先在「模型」页启用")
	}
	if !model.HasCapability(p.Capabilities, model.CapRerank) {
		return fmt.Errorf("所选配置不是 Rerank 类型模型")
	}
	if !model.ProviderVisibleTo(&p, userID, false) {
		return fmt.Errorf("无权使用该 Rerank 模型")
	}
	return nil
}

type retrievalConfigPatch struct {
	RagTopK              *int
	RagMinScore          *float64
	RagTopKSemantic      *int
	RagMinScoreSemantic  *float64
	RagTopKKeyword       *int
	RagMinScoreKeyword   *float64
	RagTopKHybrid        *int
	RagMinScoreHybrid    *float64
	HybridSemanticWeight *float64
	SearchMode           string
}

func applyRetrievalConfig(kb *model.KnowledgeBase, patch retrievalConfigPatch) {
	if patch.SearchMode != "" {
		kb.SearchMode = normalizeSearchMode(patch.SearchMode)
	}

	legacyOnly := patch.RagTopK != nil || patch.RagMinScore != nil
	modeSpecific := patch.RagTopKSemantic != nil || patch.RagMinScoreSemantic != nil ||
		patch.RagTopKKeyword != nil || patch.RagMinScoreKeyword != nil ||
		patch.RagTopKHybrid != nil || patch.RagMinScoreHybrid != nil

	if patch.RagTopKSemantic != nil && *patch.RagTopKSemantic > 0 {
		kb.RagTopKSemantic = clampTopK(*patch.RagTopKSemantic)
	}
	if patch.RagMinScoreSemantic != nil && *patch.RagMinScoreSemantic >= 0 {
		kb.RagMinScoreSemantic = clampMinScore(*patch.RagMinScoreSemantic)
	}
	if patch.RagTopKKeyword != nil && *patch.RagTopKKeyword > 0 {
		kb.RagTopKKeyword = clampTopK(*patch.RagTopKKeyword)
	}
	if patch.RagMinScoreKeyword != nil && *patch.RagMinScoreKeyword >= 0 {
		kb.RagMinScoreKeyword = clampMinScore(*patch.RagMinScoreKeyword)
	}
	if patch.RagTopKHybrid != nil && *patch.RagTopKHybrid > 0 {
		kb.RagTopKHybrid = clampTopK(*patch.RagTopKHybrid)
	}
	if patch.RagMinScoreHybrid != nil && *patch.RagMinScoreHybrid >= 0 {
		kb.RagMinScoreHybrid = clampMinScore(*patch.RagMinScoreHybrid)
	}
	if patch.HybridSemanticWeight != nil {
		kb.HybridSemanticWeight = clampHybridSemanticWeight(*patch.HybridSemanticWeight)
	}

	// 旧客户端只传 rag_top_k / rag_min_score：同步到三种模式
	if legacyOnly && !modeSpecific {
		if patch.RagTopK != nil && *patch.RagTopK > 0 {
			v := clampTopK(*patch.RagTopK)
			kb.RagTopK = v
			kb.RagTopKSemantic = v
			kb.RagTopKKeyword = v
			kb.RagTopKHybrid = v
		}
		if patch.RagMinScore != nil && *patch.RagMinScore >= 0 {
			v := clampMinScore(*patch.RagMinScore)
			kb.RagMinScore = v
			kb.RagMinScoreSemantic = v
			kb.RagMinScoreKeyword = v
			kb.RagMinScoreHybrid = v
		}
	}

	syncLegacyRetrievalFromMode(kb)
}

func syncLegacyRetrievalFromMode(kb *model.KnowledgeBase) {
	topK, minScore := modeRetrievalDefaults(kb, kb.SearchMode)
	kb.RagTopK = topK
	kb.RagMinScore = minScore
	if kb.HybridSemanticWeight < 0 || kb.HybridSemanticWeight > 1 {
		kb.HybridSemanticWeight = 0.7
	}
}

func modeRetrievalDefaults(kb *model.KnowledgeBase, mode string) (int, float64) {
	topK := 0
	minScore := 0.0
	switch normalizeSearchMode(mode) {
	case "keyword":
		topK, minScore = kb.RagTopKKeyword, kb.RagMinScoreKeyword
	case "hybrid":
		topK, minScore = kb.RagTopKHybrid, kb.RagMinScoreHybrid
	default:
		topK, minScore = kb.RagTopKSemantic, kb.RagMinScoreSemantic
	}
	if topK <= 0 {
		topK = kb.RagTopK
	}
	if topK <= 0 {
		topK = 5
	}
	if minScore <= 0 {
		minScore = kb.RagMinScore
	}
	if minScore <= 0 {
		minScore = 0.3
	}
	return clampTopK(topK), clampMinScore(minScore)
}

func resolveModeRetrievalParams(kb *model.KnowledgeBase, mode string, reqTopK int, reqMinScore float64) (int, float64) {
	topK, minScore := modeRetrievalDefaults(kb, mode)
	if reqTopK > 0 {
		topK = clampTopK(reqTopK)
	}
	if reqMinScore > 0 {
		minScore = clampMinScore(reqMinScore)
	}
	return topK, minScore
}

func clampTopK(v int) int {
	if v < 1 {
		return 1
	}
	if v > 20 {
		return 20
	}
	return v
}

func clampMinScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func clampHybridSemanticWeight(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func normalizeSplitMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "fixed", "fixed_size":
		return "fixed"
	case "separator", "by_separator":
		return "separator"
	default:
		return "recursive"
	}
}

func normalizeSearchMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "keyword", "fulltext", "bm25":
		return "keyword"
	case "hybrid", "mix":
		return "hybrid"
	default:
		return "semantic"
	}
}

// resolveSearchRerank：请求显式传了 rerank 则用请求；否则按当前检索方式取库默认
func resolveSearchRerank(kb *model.KnowledgeBase, mode string, req *SearchRequest) (*bool, string) {
	if kb == nil {
		return req.Rerank, strings.TrimSpace(req.RerankProviderID)
	}
	var on bool
	var prov string
	switch mode {
	case "keyword":
		on, prov = kb.RerankKeyword, kb.RerankProviderKeyword
	case "hybrid":
		on, prov = kb.RerankHybrid, kb.RerankProviderHybrid
	default:
		on, prov = kb.RerankSemantic, kb.RerankProviderSemantic
	}
	rerank := req.Rerank
	if rerank == nil {
		v := on
		rerank = &v
	}
	provID := strings.TrimSpace(req.RerankProviderID)
	if provID == "" {
		provID = strings.TrimSpace(prov)
	}
	return rerank, provID
}

func clampChunkSize(size int) int {
	if size < 64 {
		return 64
	}
	if size > 4000 {
		return 4000
	}
	return size
}

func clampOverlap(overlap, chunkSize int) int {
	if overlap < 0 {
		return 0
	}
	if overlap >= chunkSize {
		return chunkSize - 1
	}
	return overlap
}

func resolveParseOptions(kb *model.KnowledgeBase, opts *ParseOptions) ParseOptions {
	out := ParseOptions{
		ChunkSize:      512,
		ChunkOverlap:   50,
		SplitMode:      "recursive",
		SplitSeparator: "\n\n",
	}
	if kb != nil {
		if kb.ChunkSize > 0 {
			out.ChunkSize = kb.ChunkSize
		}
		if kb.ChunkOverlap >= 0 {
			out.ChunkOverlap = kb.ChunkOverlap
		}
		if kb.SplitMode != "" {
			out.SplitMode = kb.SplitMode
		}
		if kb.SplitSeparator != "" {
			out.SplitSeparator = kb.SplitSeparator
		}
	}
	if opts != nil {
		if opts.ChunkSize > 0 {
			out.ChunkSize = opts.ChunkSize
		}
		if opts.ChunkOverlap >= 0 && opts.ChunkSize > 0 {
			out.ChunkOverlap = opts.ChunkOverlap
		} else if opts.ChunkOverlap > 0 {
			out.ChunkOverlap = opts.ChunkOverlap
		}
		if opts.SplitMode != "" {
			out.SplitMode = opts.SplitMode
		}
		if opts.SplitSeparator != "" {
			out.SplitSeparator = opts.SplitSeparator
		}
	}
	out.ChunkSize = clampChunkSize(out.ChunkSize)
	out.ChunkOverlap = clampOverlap(out.ChunkOverlap, out.ChunkSize)
	out.SplitMode = normalizeSplitMode(out.SplitMode)
	if out.SplitSeparator == "" {
		out.SplitSeparator = "\n\n"
	}
	return out
}

func parseOptionsFromMetadata(meta model.JSONBMap, kb *model.KnowledgeBase) ParseOptions {
	opts := &ParseOptions{}
	has := false
	if meta != nil {
		if v, ok := metaInt(meta, "chunk_size"); ok {
			opts.ChunkSize = v
			has = true
		}
		if v, ok := metaInt(meta, "chunk_overlap"); ok {
			opts.ChunkOverlap = v
			has = true
		}
		if v, ok := metaString(meta, "split_mode"); ok {
			opts.SplitMode = v
			has = true
		}
		if v, ok := metaString(meta, "split_separator"); ok {
			opts.SplitSeparator = v
			has = true
		}
	}
	if !has {
		opts = nil
	}
	return resolveParseOptions(kb, opts)
}

func metaInt(meta model.JSONBMap, key string) (int, bool) {
	raw, ok := meta[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case int64:
		return int(v), true
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}

func metaString(meta model.JSONBMap, key string) (string, bool) {
	raw, ok := meta[key]
	if !ok || raw == nil {
		return "", false
	}
	s, ok := raw.(string)
	return s, ok && s != ""
}

func kbStorageKey(kb *model.KnowledgeBase) string {
	if kb != nil && strings.TrimSpace(kb.StorageDir) != "" {
		return kb.StorageDir
	}
	if kb != nil && kb.ID != "" {
		return kb.ID
	}
	return "kb"
}

func sanitizeStorageDir(name string) string {
	name = strings.TrimSpace(name)
	var b strings.Builder
	for _, r := range name {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '\n', '\r', '\t':
			b.WriteRune('-')
		default:
			if r < 32 {
				b.WriteRune('-')
			} else {
				b.WriteRune(r)
			}
		}
	}
	s := strings.Trim(b.String(), " .-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	if s == "" || s == "." || s == ".." {
		s = "kb"
	}
	rs := []rune(s)
	if len(rs) > 80 {
		s = string(rs[:80])
		s = strings.TrimRight(s, " .-")
	}
	if s == "" {
		return "kb"
	}
	return s
}

func (s *KnowledgeService) allocStorageDir(userID, name, excludeID string) (string, error) {
	base := sanitizeStorageDir(name)
	candidate := base
	for i := 2; i < 1000; i++ {
		q := s.db.Model(&model.KnowledgeBase{}).Where("user_id = ? AND storage_dir = ?", userID, candidate)
		if excludeID != "" {
			q = q.Where("id <> ?", excludeID)
		}
		var n int64
		if err := q.Count(&n).Error; err != nil {
			return "", fmt.Errorf("检查存储目录失败")
		}
		if n == 0 {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return "", fmt.Errorf("存储目录重名过多")
}

func (s *KnowledgeService) moveUploadDir(userID, oldDir, newDir, kbID string) error {
	if oldDir == "" || oldDir == newDir {
		return nil
	}
	oldPath := filepath.Join(UploadsDir, DocumentsDir, s.userDirName(userID), oldDir)
	newPath := filepath.Join(UploadsDir, DocumentsDir, s.userDirName(userID), newDir)
	info, err := os.Stat(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取存储目录失败")
	}
	if !info.IsDir() {
		return nil
	}
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("存储目录已存在")
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return fmt.Errorf("创建存储目录失败")
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("重命名存储目录失败")
	}
	var docs []model.Document
	if err := s.db.Where("knowledge_base_id = ?", kbID).Find(&docs).Error; err != nil {
		return fmt.Errorf("查询文档失败")
	}
	oldPrefix := filepath.Clean(oldPath)
	newPrefix := filepath.Clean(newPath)
	for _, d := range docs {
		p := filepath.Clean(d.FilePath)
		if !strings.HasPrefix(p, oldPrefix) {
			continue
		}
		np := newPrefix + strings.TrimPrefix(p, oldPrefix)
		if err := s.db.Model(&model.Document{}).Where("id = ?", d.ID).Update("file_path", np).Error; err != nil {
			return fmt.Errorf("更新文档路径失败")
		}
	}
	return nil
}

func (s *KnowledgeService) userDirName(userID string) string {
	var u model.User
	if err := s.db.Select("name").Where("id = ?", userID).First(&u).Error; err != nil {
		return sanitizeStorageDir(userID)
	}
	if strings.TrimSpace(u.Name) == "" {
		return sanitizeStorageDir(userID)
	}
	return sanitizeStorageDir(u.Name)
}

// RelocateUserUploadRoot 用户改名时，把 uploads/documents/{旧用户名} 改成新用户名，并改文档路径。
func RelocateUserUploadRoot(db *gorm.DB, userID, oldName, newName string) error {
	oldSeg := sanitizeStorageDir(oldName)
	newSeg := sanitizeStorageDir(newName)
	if oldSeg == newSeg {
		return nil
	}
	oldPath := filepath.Join(UploadsDir, DocumentsDir, oldSeg)
	newPath := filepath.Join(UploadsDir, DocumentsDir, newSeg)
	info, err := os.Stat(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取用户目录失败")
	}
	if !info.IsDir() {
		return nil
	}
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("用户目录已存在")
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("重命名用户目录失败")
	}
	var docs []model.Document
	if err := db.Where("user_id = ?", userID).Find(&docs).Error; err != nil {
		return fmt.Errorf("查询文档失败")
	}
	oldPrefix := filepath.Clean(oldPath)
	newPrefix := filepath.Clean(newPath)
	for _, d := range docs {
		p := filepath.Clean(d.FilePath)
		if !strings.HasPrefix(p, oldPrefix) {
			continue
		}
		np := newPrefix + strings.TrimPrefix(p, oldPrefix)
		if err := db.Model(&model.Document{}).Where("id = ?", d.ID).Update("file_path", np).Error; err != nil {
			return fmt.Errorf("更新文档路径失败")
		}
	}
	return nil
}

func ensureUploadDir(userDir, dirName string) (string, error) {
	dir := filepath.Join(UploadsDir, DocumentsDir, userDir, dirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %w", err)
	}
	return dir, nil
}

// toSharedUploadPath 把本地上传路径转成 AI 容器可读形式（uploads/... 或 /app/uploads/...）。
func toSharedUploadPath(localPath string) string {
	s := filepath.ToSlash(strings.TrimSpace(localPath))
	if s == "" {
		return s
	}
	for _, prefix := range []string{
		"/app/uploads/",
		"/opt/data/cogniforge/uploads/",
		"/opt/project/cogniforge/data/uploads/",
	} {
		if strings.HasPrefix(s, prefix) {
			return "uploads/" + strings.TrimPrefix(s, prefix)
		}
	}
	if strings.HasPrefix(s, "uploads/") {
		return s
	}
	if idx := strings.Index(s, "/uploads/"); idx >= 0 {
		return "uploads/" + s[idx+len("/uploads/"):]
	}
	return s
}

func saveUploadedFile(file *multipart.FileHeader, destPath string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}
	return nil
}

func readDocumentRaw(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	return string(content), nil
}

func writeDocumentContent(filePath, content string) error {
	if filePath == "" {
		return fmt.Errorf("文件路径为空")
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

func readDocumentContent(filePath, fileType string) (string, error) {
	content, err := readDocumentRaw(filePath)
	if err != nil {
		return "", err
	}

	switch fileType {
	case AllowedFileTypeTXT, AllowedFileTypeMD:
		return content, nil
	case AllowedFileTypeHTML:
		return stripHTMLTags(content), nil
	case AllowedFileTypePDF, AllowedFileTypeDOCX:
		return content, nil
	default:
		return content, nil
	}
}

func stripHTMLTags(html string) string {
	var result strings.Builder
	inTag := false
	for _, r := range html {
		if r == '<' {
			inTag = true
			result.WriteRune(' ')
		} else if r == '>' {
			inTag = false
			result.WriteRune(' ')
		} else if !inTag {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

func chunkText(content string, chunkSize, overlap int) []string {
	if len(content) == 0 {
		return nil
	}

	if chunkSize <= 0 {
		return []string{content}
	}

	var chunks []string
	runes := []rune(content)
	start := 0

	for start < len(runes) {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunk := string(runes[start:end])
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		// Move to next chunk position
		nextStart := start + chunkSize
		if overlap >= chunkSize {
			overlap = chunkSize / 2 // Prevent infinite loop
		}
		start = nextStart - overlap

		if start >= len(runes) {
			break
		}
		if start < 0 {
			start = 0
		}
	}

	return chunks
}

func performTextSearch(query string, docs []model.Document, topK int, minScore float64) []SearchResult {
	queryLower := strings.ToLower(query)
	queryWords := tokenize(queryLower)

	if len(queryWords) == 0 {
		return nil
	}

	var results []SearchResult

	for _, doc := range docs {
		content, err := readDocumentContent(doc.FilePath, doc.FileType)
		if err != nil {
			continue
		}

		chunks := chunkText(content, 512, 50)

		for i, chunk := range chunks {
			chunkLower := strings.ToLower(chunk)
			score := calculateTextSimilarity(queryLower, chunkLower, queryWords)

			if score >= minScore {
				results = append(results, SearchResult{
					DocumentID:   doc.ID,
					DocumentName: doc.Name,
					ChunkID:      fmt.Sprintf("%s_chunk_%d", doc.ID, i),
					Content:      truncateText(chunk, 500),
					Score:        score,
				})
			}
		}
	}

	sortSearchResults(results)

	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

func tokenize(text string) []string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "，", " ")
	text = strings.ReplaceAll(text, "。", " ")
	text = strings.ReplaceAll(text, "！", " ")
	text = strings.ReplaceAll(text, "？", " ")
	text = strings.ReplaceAll(text, "、", " ")
	text = strings.ReplaceAll(text, ";", " ")
	text = strings.ReplaceAll(text, ":", " ")
	text = strings.ReplaceAll(text, "\"", " ")
	text = strings.ReplaceAll(text, "'", " ")

	words := strings.Fields(text)
	var filtered []string
	for _, w := range words {
		if len(w) >= 2 {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

func calculateTextSimilarity(query, text string, queryWords []string) float64 {
	if len(text) == 0 || len(queryWords) == 0 {
		return 0
	}

	matchCount := 0
	for _, word := range queryWords {
		if strings.Contains(text, word) {
			matchCount++
		}
	}

	coverage := float64(matchCount) / float64(len(queryWords))
	textWords := tokenize(text)
	keywordDensity := float64(matchCount) / float64(len(textWords)+1)

	score := coverage*0.6 + keywordDensity*0.4

	if strings.Contains(text, query) {
		score = score*1.2 + 0.1
		if score > 1 {
			score = 1
		}
	}

	return score
}

func sortSearchResults(results []SearchResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// generateID 生成唯一ID
func generateID() string {
	return newID()
}
