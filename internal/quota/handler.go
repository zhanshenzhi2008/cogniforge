package quota

import (
	"github.com/gin-gonic/gin"

	"cogniforge/internal/model"
	"cogniforge/internal/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	snap, err := h.svc.Me(c.Request.Context(), userID)
	if err != nil {
		if err == ErrUnavailable {
			WriteError(c, err)
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, snap)
}

func (h *Handler) Usage(c *gin.Context) {
	userID := c.GetString("user_id")
	role, err := h.svc.loadUser(userID)
	if err != nil {
		response.Unauthorized(c, "用户不存在")
		return
	}
	q := UsageQuery{
		Range:   c.DefaultQuery("range", "7d"),
		Scope:   c.DefaultQuery("scope", "self"),
		UserID:  userID,
		IsAdmin: role == "admin",
	}
	if q.IsAdmin {
		if uid := c.Query("user_id"); uid != "" {
			q.UserID = uid
			q.Scope = "user"
		}
	}
	report, err := h.svc.Usage(q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, report)
}

func (h *Handler) GetPolicy(c *gin.Context) {
	p, err := h.svc.GetDefaultPolicy()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, p)
}

type policyBody struct {
	DailyRequests  *int   `json:"daily_requests"`
	DailyTokens    *int64 `json:"daily_tokens"`
	MonthlyTokens  *int64 `json:"monthly_tokens"`
	RPM            *int   `json:"rpm"`
	AdminUnlimited *bool  `json:"admin_unlimited"`
	Enabled        *bool  `json:"enabled"`
}

func applyPolicyBody(dst *model.QuotaPolicy, body policyBody) {
	if body.DailyRequests != nil {
		dst.DailyRequests = *body.DailyRequests
	}
	if body.DailyTokens != nil {
		dst.DailyTokens = *body.DailyTokens
	}
	if body.MonthlyTokens != nil {
		dst.MonthlyTokens = *body.MonthlyTokens
	}
	if body.RPM != nil {
		dst.RPM = *body.RPM
	}
	if body.AdminUnlimited != nil {
		dst.AdminUnlimited = *body.AdminUnlimited
	}
	if body.Enabled != nil {
		dst.Enabled = *body.Enabled
	}
}

func (h *Handler) PutPolicy(c *gin.Context) {
	var body policyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	cur, err := h.svc.GetDefaultPolicy()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	applyPolicyBody(cur, body)
	p, err := h.svc.UpdateDefaultPolicy(*cur)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, p)
}

func (h *Handler) PutUserPolicy(c *gin.Context) {
	uid := c.Param("id")
	if uid == "" {
		response.BadRequest(c, "用户 ID 不能为空")
		return
	}
	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if isClearOverride(raw) {
		if _, err := h.svc.UpsertUserPolicy(uid, nil); err != nil {
			response.InternalError(c, err.Error())
			return
		}
		response.Success(c, gin.H{"cleared": true})
		return
	}
	cur := model.QuotaPolicy{
		DailyRequests:  DefaultDailyRequests,
		DailyTokens:    DefaultDailyTokens,
		MonthlyTokens:  DefaultMonthlyTokens,
		RPM:            DefaultRPM,
		AdminUnlimited: false,
		Enabled:        true,
	}
	if def, err := h.svc.GetDefaultPolicy(); err == nil {
		cur.DailyRequests = def.DailyRequests
		cur.DailyTokens = def.DailyTokens
		cur.MonthlyTokens = def.MonthlyTokens
		cur.RPM = def.RPM
		cur.Enabled = def.Enabled
	}
	applyPolicyBody(&cur, policyFromMap(raw))
	p, err := h.svc.UpsertUserPolicy(uid, &cur)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, p)
}

func isClearOverride(raw map[string]any) bool {
	if v, ok := raw["clear"]; ok {
		b, _ := v.(bool)
		return b
	}
	return false
}

func policyFromMap(raw map[string]any) policyBody {
	var b policyBody
	if v, ok := raw["daily_requests"].(float64); ok {
		n := int(v)
		b.DailyRequests = &n
	}
	if v, ok := raw["daily_tokens"].(float64); ok {
		n := int64(v)
		b.DailyTokens = &n
	}
	if v, ok := raw["monthly_tokens"].(float64); ok {
		n := int64(v)
		b.MonthlyTokens = &n
	}
	if v, ok := raw["rpm"].(float64); ok {
		n := int(v)
		b.RPM = &n
	}
	if v, ok := raw["admin_unlimited"].(bool); ok {
		b.AdminUnlimited = &v
	}
	if v, ok := raw["enabled"].(bool); ok {
		b.Enabled = &v
	}
	return b
}

func (h *Handler) RegisterUserRoutes(rg *gin.RouterGroup) {
	rg.GET("/quota/me", h.Me)
	rg.GET("/quota/usage", h.Usage)
}

func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/admin/quota/policy", h.GetPolicy)
	rg.PUT("/admin/quota/policy", h.PutPolicy)
	rg.PUT("/admin/quota/users/:id", h.PutUserPolicy)
}
