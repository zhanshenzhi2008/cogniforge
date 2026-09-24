package skillimport

import (
	"io"

	"cogniforge/internal/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler 导入 HTTP 处理
type Handler struct {
	svc *Service
}

// NewHandler 创建 Import Handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		svc: NewService(db),
	}
}

// ImportSkillFromMarkdown 导入 SKILL（Markdown 文本）
// POST /api/v1/import/skills/markdown
func (h *Handler) ImportSkillFromMarkdown(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "未登录")
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result := h.svc.ImportSkillFromMarkdown(userID, []byte(req.Content))
	if len(result.Errors) > 0 {
		response.BadRequest(c, result.Errors[0])
		return
	}

	response.Created(c, gin.H{
		"skill":    result.Skill,
		"warnings": result.Warnings,
	})
}

// ImportSkillFromFile 导入 SKILL（文件上传：.md 或 .zip）
// POST /api/v1/import/skills/file
func (h *Handler) ImportSkillFromFile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "未登录")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件: "+err.Error())
		return
	}
	defer file.Close()

	cfg := LoadSecurityConfig()
	vResult := ValidateMultipartFile(header, cfg)
	if !vResult.Valid {
		response.BadRequest(c, vResult.Errors[0])
		return
	}

	ext := getFileExt(header.Filename)

	switch ext {
	case ".md":
		content, err := io.ReadAll(file)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}
		result := h.svc.ImportSkillFromMarkdown(userID, content)
		if len(result.Errors) > 0 {
			response.BadRequest(c, result.Errors[0])
			return
		}
		response.Created(c, gin.H{
			"skill":    result.Skill,
			"warnings": result.Warnings,
		})
	case ".zip":
		result := h.svc.ImportSkillFromZIP(userID, file, header.Size)
		// ZIP 批量结果整体返回，不因部分失败而 400
		response.Created(c, gin.H{
			"total":   result.Total,
			"success": result.Success,
			"failed":  result.Failed,
			"results": result.Results,
		})
	default:
		response.BadRequest(c, "不支持的文件类型: "+ext)
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	imp := rg.Group("/import")
	{
		skills := imp.Group("/skills")
		{
			skills.POST("/markdown", h.ImportSkillFromMarkdown)
			skills.POST("/file", h.ImportSkillFromFile)
		}
	}
}

// =============================================================================
// 工具函数
// =============================================================================

func getFileExt(filename string) string {
	if len(filename) < 4 {
		return ""
	}
	if filename[len(filename)-4] == '.' {
		return filename[len(filename)-4:]
	}
	if len(filename) >= 5 && filename[len(filename)-5] == '.' {
		return filename[len(filename)-5:]
	}
	return ""
}

