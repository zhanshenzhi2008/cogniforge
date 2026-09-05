package skill

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/response"
)

// Handler SKILL HTTP 处理
type Handler struct {
	svc *Service
}

// NewHandler 创建 Skill Handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		svc: NewService(db),
	}
}

// ListSkills 获取 SKILL 列表
func (h *Handler) ListSkills(c *gin.Context) {
	userID := c.GetString("user_id")
	skills, err := h.svc.ListSkills(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, skills)
}

// CreateSkill 创建自定义 SKILL
func (h *Handler) CreateSkill(c *gin.Context) {
	userID := c.GetString("user_id")
	var req CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	skill, err := h.svc.CreateSkill(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, skill)
}

// GetSkill 获取 SKILL 详情
func (h *Handler) GetSkill(c *gin.Context) {
	userID := c.GetString("user_id")
	skillID := c.Param("id")

	skill, err := h.svc.GetSkill(userID, skillID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, skill)
}

// UpdateSkill 更新 SKILL
func (h *Handler) UpdateSkill(c *gin.Context) {
	userID := c.GetString("user_id")
	skillID := c.Param("id")

	var req UpdateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	skill, err := h.svc.UpdateSkill(userID, skillID, &req)
	if err != nil {
		if err.Error() == "内置 SKILL 不可编辑" {
			response.Forbidden(c, err.Error())
		} else {
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Success(c, skill)
}

// DeleteSkill 删除 SKILL
func (h *Handler) DeleteSkill(c *gin.Context) {
	userID := c.GetString("user_id")
	skillID := c.Param("id")

	err := h.svc.DeleteSkill(userID, skillID)
	if err != nil {
		if err.Error() == "内置 SKILL 不可删除" {
			response.Forbidden(c, err.Error())
		} else {
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.SuccessWithMessage(c, nil, "SKILL 已删除")
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	skills := rg.Group("/skills")
	{
		skills.GET("", h.ListSkills)
		skills.POST("", h.CreateSkill)
		skills.GET("/:id", h.GetSkill)
		skills.PUT("/:id", h.UpdateSkill)
		skills.DELETE("/:id", h.DeleteSkill)
	}
}
