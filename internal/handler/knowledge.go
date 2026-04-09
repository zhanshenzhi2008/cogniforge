package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// 允许的文件类型
	AllowedFileTypePDF  = "pdf"
	AllowedFileTypeTXT  = "txt"
	AllowedFileTypeMD   = "md"
	AllowedFileTypeDOCX = "docx"
	AllowedFileTypeHTML = "html"

	// 文件大小限制 (50MB)
	MaxFileSize = 50 * 1024 * 1024

	// 存储目录
	UploadsDir = "uploads"
	DocumentsDir = "documents"
)

// 允许的文件扩展名映射到类型
var allowedExtensions = map[string]string{
	".pdf":  AllowedFileTypePDF,
	".txt":  AllowedFileTypeTXT,
	".md":   AllowedFileTypeMD,
	".docx": AllowedFileTypeDOCX,
	".html": AllowedFileTypeHTML,
}

// =============================================================================
// Helper functions
// =============================================================================

// getFileExt 获取文件扩展名（小写）
func getFileExt(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}

// getFileType 获取文件类型
func getFileType(filename string) (string, bool) {
	ext := getFileExt(filename)
	fileType, ok := allowedExtensions[ext]
	return fileType, ok
}

// generateDocID 生成文档ID
func generateDocID() string {
	return fmt.Sprintf("doc_%s_%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)
}

// ensureUploadDir 确保上传目录存在
func ensureUploadDir(userID, kbID string) (string, error) {
	dir := filepath.Join(UploadsDir, DocumentsDir, userID, kbID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %w", err)
	}
	return dir, nil
}

// saveUploadedFile 保存上传的文件
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

// =============================================================================
// Request/Response types

type CreateKBRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	VectorDB    string `json:"vector_db"`
	EmbeddingModel string `json:"embedding_model"`
}

type UpdateKBRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	VectorDB    string `json:"vector_db"`
	EmbeddingModel string `json:"embedding_model"`
	Status      string `json:"status"`
}

// ListKnowledgeBases 获取知识库列表
func ListKnowledgeBases(c *gin.Context) {
	userID := c.GetString("user_id")

	var kbs []model.KnowledgeBase
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&kbs).Error; err != nil {
		model.InternalError(c, "查询知识库列表失败")
		return
	}

	model.Success(c, kbs)
}

// CreateKnowledgeBase 创建知识库
func CreateKnowledgeBase(c *gin.Context) {
	userID := c.GetString("user_id")

	var req CreateKBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	if req.Name == "" {
		model.BadRequest(c, "知识库名称不能为空")
		return
	}

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

	// 设置默认值
	if kb.VectorDB == "" {
		kb.VectorDB = "chroma"
	}
	if kb.EmbeddingModel == "" {
		kb.EmbeddingModel = "text-embedding-ada-002"
	}

	if err := database.DB.Create(&kb).Error; err != nil {
		model.InternalError(c, "创建知识库失败")
		return
	}

	model.Created(c, kb)
}

// GetKnowledgeBase 获取知识库详情
func GetKnowledgeBase(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	var kb model.KnowledgeBase
	if err := database.DB.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "知识库不存在")
		} else {
			model.InternalError(c, "查询知识库失败")
		}
		return
	}

	model.Success(c, kb)
}

// UpdateKnowledgeBase 更新知识库
func UpdateKnowledgeBase(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	var kb model.KnowledgeBase
	if err := database.DB.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "知识库不存在")
		} else {
			model.InternalError(c, "查询知识库失败")
		}
		return
	}

	var req UpdateKBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// 更新字段
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

	if err := database.DB.Save(&kb).Error; err != nil {
		model.InternalError(c, "更新知识库失败")
		return
	}

	model.Success(c, kb)
}

// DeleteKnowledgeBase 删除知识库（软删除）
func DeleteKnowledgeBase(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	var kb model.KnowledgeBase
	if err := database.DB.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "知识库不存在")
		} else {
			model.InternalError(c, "查询知识库失败")
		}
		return
	}

	if err := database.DB.Delete(&kb).Error; err != nil {
		model.InternalError(c, "删除知识库失败")
		return
	}

	model.SuccessWithMessage(c, nil, "知识库已删除")
}

// =============================================================================
// Document APIs
// =============================================================================

// ListDocuments 获取知识库的文档列表
func ListDocuments(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	// 验证知识库存在且属于当前用户
	var kb model.KnowledgeBase
	if err := database.DB.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "知识库不存在")
		} else {
			model.InternalError(c, "查询知识库失败")
		}
		return
	}

	var docs []model.Document
	if err := database.DB.Where("knowledge_base_id = ?", kbID).Order("created_at DESC").Find(&docs).Error; err != nil {
		model.InternalError(c, "查询文档列表失败")
		return
	}

	model.Success(c, docs)
}

// DeleteDocument 删除文档
func DeleteDocument(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")
	docID := c.Param("docId")

	// 验证知识库存在且属于当前用户
	var kb model.KnowledgeBase
	if err := database.DB.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "知识库不存在")
		} else {
			model.InternalError(c, "查询知识库失败")
		}
		return
	}

	// 验证文档存在且属于该知识库
	var doc model.Document
	if err := database.DB.Where("id = ? AND knowledge_base_id = ?", docID, kbID).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "文档不存在")
		} else {
			model.InternalError(c, "查询文档失败")
		}
		return
	}

	if err := database.DB.Delete(&doc).Error; err != nil {
		model.InternalError(c, "删除文档失败")
		return
	}

	// 更新知识库的文档计数
	database.DB.Model(&kb).Update("doc_count", gorm.Expr("doc_count - 1"))

	model.SuccessWithMessage(c, nil, "文档已删除")
}

// UploadDocument 上传文档
func UploadDocument(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	// 验证知识库存在且属于当前用户
	var kb model.KnowledgeBase
	if err := database.DB.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "知识库不存在")
		} else {
			model.InternalError(c, "查询知识库失败")
		}
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		model.BadRequest(c, "请选择要上传的文件")
		return
	}

	// 验证文件大小
	if file.Size > MaxFileSize {
		model.BadRequest(c, fmt.Sprintf("文件大小不能超过 %dMB", MaxFileSize/(1024*1024)))
		return
	}

	// 验证文件类型
	fileType, ok := getFileType(file.Filename)
	if !ok {
		model.BadRequest(c, "不支持的文件类型，支持：PDF、TXT、MD、DOCX、HTML")
		return
	}

	// 确保上传目录存在
	dir, err := ensureUploadDir(userID, kbID)
	if err != nil {
		model.InternalError(c, "创建上传目录失败")
		return
	}

	// 生成文档ID和文件名
	docID := generateDocID()
	ext := getFileExt(file.Filename)
	savedFilename := docID + ext
	destPath := filepath.Join(dir, savedFilename)

	// 保存文件
	if err := saveUploadedFile(file, destPath); err != nil {
		model.InternalError(c, "保存文件失败: "+err.Error())
		return
	}

	// 创建文档记录
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

	if err := database.DB.Create(&doc).Error; err != nil {
		// 删除已保存的文件
		os.Remove(destPath)
		model.InternalError(c, "创建文档记录失败")
		return
	}

	// 更新知识库的文档计数
	database.DB.Model(&kb).Update("doc_count", gorm.Expr("doc_count + 1"))

	// TODO: 触发异步分块处理任务
	// 这里暂时将状态设为 processing，后续可接入消息队列处理
	go func() {
		processDocumentAsync(docID)
	}()

	model.Created(c, doc)
}

// processDocumentAsync 异步处理文档（分块、向量化）
func processDocumentAsync(docID string) {
	// 更新状态为 processing
	database.DB.Model(&model.Document{}).Where("id = ?", docID).Updates(map[string]any{
		"status":     "processing",
		"updated_at": time.Now(),
	})

	// 读取文件内容（仅处理文本类文件）
	var doc model.Document
	if err := database.DB.Where("id = ?", docID).First(&doc).Error; err != nil {
		return
	}

	content, err := readDocumentContent(doc.FilePath, doc.FileType)
	if err != nil {
		database.DB.Model(&doc).Updates(map[string]any{
			"status": "failed",
			"error":  err.Error(),
		})
		return
	}

	// 简单分块处理（按段落或固定长度）
	chunks := chunkText(content, 512, 50)

	// 更新分块数量
	database.DB.Model(&doc).Updates(map[string]any{
		"chunk_count":  len(chunks),
		"vector_count": len(chunks), // 假设每个 chunk 都能成功向量化
		"status":       "completed",
		"updated_at":   time.Now(),
	})
}

// readDocumentContent 读取文档内容
func readDocumentContent(filePath, fileType string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	switch fileType {
	case AllowedFileTypeTXT, AllowedFileTypeMD:
		return string(content), nil
	case AllowedFileTypeHTML:
		// 简单的 HTML 标签去除
		text := stripHTMLTags(string(content))
		return text, nil
	case AllowedFileTypePDF, AllowedFileTypeDOCX:
		// PDF 和 DOCX 需要额外的解析库，这里暂时返回原始内容
		// 后续可接入 pdfgo 或者调用 Python 服务处理
		return string(content), nil
	default:
		return string(content), nil
	}
}

// stripHTMLTags 去除 HTML 标签
func stripHTMLTags(html string) string {
	// 简单实现，实际应使用 html/template 或专门库
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

// chunkText 将文本分块
func chunkText(content string, chunkSize, overlap int) []string {
	if len(content) == 0 {
		return nil
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
		chunks = append(chunks, chunk)

		start = end - overlap
		if start >= len(runes) {
			break
		}
		if start < 0 {
			start = 0
		}
	}

	return chunks
}

// =============================================================================
// Search APIs
// =============================================================================

// SearchRequest 检索请求
type SearchRequest struct {
	Query    string  `json:"query" binding:"required"`
	TopK     int     `json:"top_k"`
	MinScore float64 `json:"min_score"`
}

// SearchResult 检索结果
type SearchResult struct {
	DocumentID   string  `json:"document_id"`
	DocumentName string  `json:"document_name"`
	ChunkID      string  `json:"chunk_id"`
	Content      string  `json:"content"`
	Score        float64 `json:"score"`
}

// SearchResponse 检索响应
type SearchResponse struct {
	Results  []SearchResult `json:"results"`
	Total    int            `json:"total"`
	Query    string         `json:"query"`
	Duration int64          `json:"duration_ms"`
}

// SearchKnowledge 知识库语义检索
func SearchKnowledge(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	startTime := time.Now()

	// 验证知识库存在且属于当前用户
	var kb model.KnowledgeBase
	if err := database.DB.Where("id = ? AND user_id = ?", kbID, userID).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "知识库不存在")
		} else {
			model.InternalError(c, "查询知识库失败")
		}
		return
	}

	// 解析请求
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, "请提供检索关键词")
		return
	}

	// 设置默认值
	if req.TopK <= 0 {
		req.TopK = 5
	}
	if req.TopK > 20 {
		req.TopK = 20
	}
	if req.MinScore <= 0 {
		req.MinScore = 0.3
	}

	// 获取该知识库下的所有文档
	var docs []model.Document
	if err := database.DB.Where("knowledge_base_id = ? AND status = ?", kbID, "completed").Find(&docs).Error; err != nil {
		model.InternalError(c, "查询文档失败")
		return
	}

	if len(docs) == 0 {
		model.Success(c, SearchResponse{
			Results:  []SearchResult{},
			Total:    0,
			Query:    req.Query,
			Duration: time.Since(startTime).Milliseconds(),
		})
		return
	}

	// 简单的文本检索（基于关键词匹配和 TF-IDF 相似度）
	// 后续可接入向量数据库实现真正的语义检索
	results := performTextSearch(req.Query, docs, req.TopK, req.MinScore)

	duration := time.Since(startTime).Milliseconds()
	model.Success(c, SearchResponse{
		Results:  results,
		Total:    len(results),
		Query:    req.Query,
		Duration: duration,
	})
}

// performTextSearch 简单的文本检索
func performTextSearch(query string, docs []model.Document, topK int, minScore float64) []SearchResult {
	queryLower := strings.ToLower(query)
	queryWords := tokenize(queryLower)

	if len(queryWords) == 0 {
		return nil
	}

	var results []SearchResult

	for _, doc := range docs {
		// 读取文档内容
		content, err := readDocumentContent(doc.FilePath, doc.FileType)
		if err != nil {
			continue
		}

		// 将文档分块
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

	// 按分数排序
	sortSearchResults(results)

	// 取前 topK 个
	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

// tokenize 简单的分词
func tokenize(text string) []string {
	// 简单按空格和标点分词
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
	// 过滤掉停用词和太短的词
	var filtered []string
	for _, w := range words {
		if len(w) >= 2 {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

// calculateTextSimilarity 计算文本相似度（基于词频）
func calculateTextSimilarity(query, text string, queryWords []string) float64 {
	if len(text) == 0 || len(queryWords) == 0 {
		return 0
	}

	// 计算查询词在文本中出现的次数
	matchCount := 0
	for _, word := range queryWords {
		if strings.Contains(text, word) {
			matchCount++
		}
	}

	// 计算匹配率
	coverage := float64(matchCount) / float64(len(queryWords))

	// 计算查询词在文本中的密度
	textWords := tokenize(text)
	keywordDensity := float64(matchCount) / float64(len(textWords)+1)

	// 综合评分
	score := coverage*0.6 + keywordDensity*0.4

	// 如果查询完全包含在文本中，增加分数
	if strings.Contains(text, query) {
		score = score*1.2 + 0.1
		if score > 1 {
			score = 1
		}
	}

	return score
}

// sortSearchResults 按分数排序（降序）
func sortSearchResults(results []SearchResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// truncateText 截断文本
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
