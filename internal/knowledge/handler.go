package knowledge

import (
	"github.com/gin-gonic/gin"

	"cogniforge/internal/response"
)

type KnowledgeHandler struct {
	service *KnowledgeService
}

func NewKnowledgeHandler(pythonClient *PythonServiceClient) *KnowledgeHandler {
	return &KnowledgeHandler{
		service: NewKnowledgeService(pythonClient),
	}
}

// ListKnowledgeBases 获取知识库列表
func (h *KnowledgeHandler) ListKnowledgeBases(c *gin.Context) {
	userID := c.GetString("user_id")
	kbs, err := h.service.ListKnowledgeBases(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, kbs)
}

// CreateKnowledgeBase 创建知识库
func (h *KnowledgeHandler) CreateKnowledgeBase(c *gin.Context) {
	userID := c.GetString("user_id")
	var req CreateKBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	if req.Name == "" {
		response.BadRequest(c, "知识库名称不能为空")
		return
	}

	kb, err := h.service.CreateKnowledgeBase(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, kb)
}

// GetKnowledgeBase 获取知识库详情
func (h *KnowledgeHandler) GetKnowledgeBase(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	kb, err := h.service.GetKnowledgeBase(userID, kbID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, kb)
}

// UpdateKnowledgeBase 更新知识库
func (h *KnowledgeHandler) UpdateKnowledgeBase(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	var req UpdateKBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	kb, err := h.service.UpdateKnowledgeBase(userID, kbID, &req)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, kb)
}

// DeleteKnowledgeBase 删除知识库
func (h *KnowledgeHandler) DeleteKnowledgeBase(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	err := h.service.DeleteKnowledgeBase(userID, kbID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "知识库已删除")
}

// ListDocuments 获取文档列表
func (h *KnowledgeHandler) ListDocuments(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	docs, err := h.service.ListDocuments(userID, kbID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, docs)
}

// DeleteDocument 删除文档
func (h *KnowledgeHandler) DeleteDocument(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")
	docID := c.Param("docId")

	err := h.service.DeleteDocument(userID, kbID, docID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "文档已删除")
}

// UploadDocument 上传文档
func (h *KnowledgeHandler) UploadDocument(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	doc, err := h.service.UploadDocument(userID, kbID, file)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, doc)
}

// SearchKnowledge 知识库检索
func (h *KnowledgeHandler) SearchKnowledge(c *gin.Context) {
	userID := c.GetString("user_id")
	kbID := c.Param("id")

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供检索关键词")
		return
	}

	result, err := h.service.SearchKnowledge(c, userID, kbID, &req)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, result)
}

// RegisterRoutes 注册路由
func (h *KnowledgeHandler) RegisterRoutes(rg *gin.RouterGroup) {
	kbs := rg.Group("/knowledge-bases")
	{
		kbs.GET("", h.ListKnowledgeBases)
		kbs.POST("", h.CreateKnowledgeBase)
		kbs.GET("/:id", h.GetKnowledgeBase)
		kbs.PUT("/:id", h.UpdateKnowledgeBase)
		kbs.DELETE("/:id", h.DeleteKnowledgeBase)
		kbs.GET("/:id/documents", h.ListDocuments)
		kbs.DELETE("/:id/documents/:docId", h.DeleteDocument)
		kbs.POST("/:id/documents/upload", h.UploadDocument)
		kbs.POST("/:id/search", h.SearchKnowledge)
	}
}
