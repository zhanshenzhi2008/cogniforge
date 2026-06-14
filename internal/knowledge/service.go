package knowledge

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
	"cogniforge/internal/trace"
)

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
	kb := model.KnowledgeBase{
		ID:             generateID(),
		UserID:         userID,
		Name:           req.Name,
		Description:    req.Description,
		VectorDB:       req.VectorDB,
		EmbeddingModel: req.EmbeddingModel,
		Status:         "active",
		Metadata:       model.JSONBMap{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if kb.VectorDB == "" {
		kb.VectorDB = "pgvector"
	}
	if kb.EmbeddingModel == "" {
		kb.EmbeddingModel = "text-embedding-ada-002"
	}

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

	if req.Name != "" {
		kb.Name = req.Name
	}
	if req.Description != "" {
		kb.Description = req.Description
	}
	if req.VectorDB != "" {
		kb.VectorDB = req.VectorDB
	}
	if req.EmbeddingModel != "" {
		kb.EmbeddingModel = req.EmbeddingModel
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

// UploadDocument 上传文档
func (s *KnowledgeService) UploadDocument(ctx context.Context, userID, kbID string, file *multipart.FileHeader) (*model.Document, error) {
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

	dir, err := ensureUploadDir(userID, kbID)
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
		Metadata:        model.JSONBMap{},
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

// SearchKnowledge 知识库检索
func (s *KnowledgeService) SearchKnowledge(c *gin.Context, userID, kbID string, req *SearchRequest) (*SearchResponse, error) {
	var kb model.KnowledgeBase
	if err := s.db.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在")
		}
		return nil, fmt.Errorf("查询知识库失败")
	}

	startTime := time.Now()

	if req.TopK <= 0 {
		req.TopK = 5
	}
	if req.TopK > 20 {
		req.TopK = 20
	}
	if req.MinScore <= 0 {
		req.MinScore = 0.3
	}

	// 创建带 trace_id 的 context
	ctx := context.WithValue(c.Request.Context(), trace.TraceIDKey, trace.GetTraceIDFromGin(c))

	// 调用 Python 服务进行向量检索
	pythonResp, err := s.pythonClient.Search(&SearchRequest{
		Context:        ctx,
		Query:          req.Query,
		CollectionName: kbID,
		TopK:           req.TopK,
		MinScore:       req.MinScore,
	})
	if err != nil {
		return nil, fmt.Errorf("检索失败: %w", err)
	}

	// 转换为标准响应格式
	var results []SearchResult
	for _, r := range pythonResp.Results {
		results = append(results, SearchResult{
			DocumentID:   r.ChunkID,
			DocumentName: "",
			ChunkID:      r.ChunkID,
			Content:      r.Content,
			Score:        r.Score,
		})
	}

	return &SearchResponse{
		Results:  results,
		Total:    len(results),
		Query:    req.Query,
		Duration: time.Since(startTime).Milliseconds(),
	}, nil
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

	// 调用 Python 服务进行文档处理和向量化
	result, err := s.pythonClient.ProcessDocument(&ProcessRequest{
		Context:        ctx,
		FilePath:       doc.FilePath,
		DocumentID:     docID,
		CollectionName: doc.KnowledgeBaseID,
		Metadata: map[string]interface{}{
			"file_type": doc.FileType,
			"file_name": doc.Name,
		},
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

func ensureUploadDir(userID, kbID string) (string, error) {
	dir := filepath.Join(UploadsDir, DocumentsDir, userID, kbID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %w", err)
	}
	return dir, nil
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

func readDocumentContent(filePath, fileType string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	switch fileType {
	case AllowedFileTypeTXT, AllowedFileTypeMD:
		return string(content), nil
	case AllowedFileTypeHTML:
		text := stripHTMLTags(string(content))
		return text, nil
	case AllowedFileTypePDF, AllowedFileTypeDOCX:
		return string(content), nil
	default:
		return string(content), nil
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
