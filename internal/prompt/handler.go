package prompt

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{svc: NewService(db)}
}

// Service 暴露给 eval / agent 复用解析逻辑
func (h *Handler) Service() *Service { return h.svc }

func (h *Handler) List(c *gin.Context) {
	list, err := h.svc.List(c.GetString("user_id"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, list)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, ver, err := h.svc.Create(c.GetString("user_id"), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, PromptDetail{Prompt: p, CurrentVersion: ver})
}

func (h *Handler) Get(c *gin.Context) {
	p, ver, err := h.svc.Get(c.GetString("user_id"), c.Param("id"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, PromptDetail{Prompt: p, CurrentVersion: ver})
}

func (h *Handler) Update(c *gin.Context) {
	var req UpdatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.Update(c.GetString("user_id"), c.Param("id"), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, p)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.GetString("user_id"), c.Param("id")); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "提示词已删除")
}

func (h *Handler) ListVersions(c *gin.Context) {
	list, err := h.svc.ListVersions(c.GetString("user_id"), c.Param("id"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, list)
}

func (h *Handler) CreateVersion(c *gin.Context) {
	var req CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	ver, err := h.svc.CreateVersion(c.GetString("user_id"), c.Param("id"), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, ver)
}

func (h *Handler) GetVersion(c *gin.Context) {
	ver, err := h.svc.GetVersion(c.GetString("user_id"), c.Param("id"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, ver)
}

func (h *Handler) DeleteVersion(c *gin.Context) {
	if err := h.svc.DeleteVersion(c.GetString("user_id"), c.Param("id")); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "版本已删除")
}

func (h *Handler) PublishVersion(c *gin.Context) {
	p, ver, err := h.svc.PublishVersion(c.GetString("user_id"), c.Param("id"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, PromptDetail{Prompt: p, CurrentVersion: ver})
}

func (h *Handler) SetCurrentVersion(c *gin.Context) {
	var req SetCurrentVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.SetCurrentVersion(c.GetString("user_id"), c.Param("id"), req.VersionID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, p)
}

func (h *Handler) Render(c *gin.Context) {
	var req RenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.svc.Render(c.GetString("user_id"), c.Param("id"), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, out)
}

func (h *Handler) Import(c *gin.Context) {
	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, ver, err := h.svc.ImportFromAgent(c.GetString("user_id"), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, PromptDetail{Prompt: p, CurrentVersion: ver})
}

func (h *Handler) ListUsages(c *gin.Context) {
	list, err := h.svc.ListUsages(c.GetString("user_id"), c.Param("id"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, list)
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	p := rg.Group("/prompts")
	{
		p.GET("", h.List)
		p.POST("", h.Create)
		p.POST("/import", h.Import)
		// 静态段放在 /:id 之前，避免 "versions" / "import" 被当成 id
		p.GET("/versions/:id", h.GetVersion)
		p.DELETE("/versions/:id", h.DeleteVersion)
		p.POST("/versions/:id/publish", h.PublishVersion)
		p.GET("/:id/versions", h.ListVersions)
		p.POST("/:id/versions", h.CreateVersion)
		p.PUT("/:id/current-version", h.SetCurrentVersion)
		p.POST("/:id/render", h.Render)
		p.GET("/:id/usages", h.ListUsages)
		p.GET("/:id", h.Get)
		p.PUT("/:id", h.Update)
		p.DELETE("/:id", h.Delete)
	}
}
