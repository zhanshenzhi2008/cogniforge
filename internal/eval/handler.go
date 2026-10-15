package eval

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/chat"
	"cogniforge/internal/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(db *gorm.DB, chatSvc *chat.ChatService) *Handler {
	return &Handler{svc: NewService(db, chatSvc)}
}

func (h *Handler) ListSuites(c *gin.Context) {
	list, err := h.svc.ListSuites(c.GetString("user_id"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, list)
}

func (h *Handler) CreateSuite(c *gin.Context) {
	var req CreateSuiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	suite, err := h.svc.CreateSuite(c.GetString("user_id"), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, suite)
}

func (h *Handler) GetSuite(c *gin.Context) {
	suite, err := h.svc.GetSuite(c.GetString("user_id"), c.Param("id"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, suite)
}

func (h *Handler) UpdateSuite(c *gin.Context) {
	var req UpdateSuiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	suite, err := h.svc.UpdateSuite(c.GetString("user_id"), c.Param("id"), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, suite)
}

func (h *Handler) DeleteSuite(c *gin.Context) {
	if err := h.svc.DeleteSuite(c.GetString("user_id"), c.Param("id")); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "评测集已删除")
}

func (h *Handler) ListCases(c *gin.Context) {
	list, err := h.svc.ListCases(c.GetString("user_id"), c.Param("id"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, list)
}

func (h *Handler) CreateCase(c *gin.Context) {
	var req CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.CreateCase(c.GetString("user_id"), c.Param("id"), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, item)
}

func (h *Handler) UpdateCase(c *gin.Context) {
	var req UpdateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.UpdateCase(c.GetString("user_id"), c.Param("id"), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *Handler) DeleteCase(c *gin.Context) {
	if err := h.svc.DeleteCase(c.GetString("user_id"), c.Param("id")); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "用例已删除")
}

func (h *Handler) CreateRun(c *gin.Context) {
	var req CreateRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	run, results, err := h.svc.CreateAndExecuteRun(c.GetString("user_id"), c.Param("id"), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, RunDetailResponse{Run: run, Results: results})
}

func (h *Handler) ListRuns(c *gin.Context) {
	list, err := h.svc.ListRuns(c.GetString("user_id"), c.Query("suite_id"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, list)
}

func (h *Handler) GetRun(c *gin.Context) {
	run, results, err := h.svc.GetRun(c.GetString("user_id"), c.Param("id"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, RunDetailResponse{Run: run, Results: results})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	evals := rg.Group("/evals")
	{
		evals.GET("/suites", h.ListSuites)
		evals.POST("/suites", h.CreateSuite)
		evals.GET("/suites/:id", h.GetSuite)
		evals.PUT("/suites/:id", h.UpdateSuite)
		evals.DELETE("/suites/:id", h.DeleteSuite)
		evals.GET("/suites/:id/cases", h.ListCases)
		evals.POST("/suites/:id/cases", h.CreateCase)
		evals.POST("/suites/:id/runs", h.CreateRun)

		evals.PUT("/cases/:id", h.UpdateCase)
		evals.DELETE("/cases/:id", h.DeleteCase)

		evals.GET("/runs", h.ListRuns)
		evals.GET("/runs/:id", h.GetRun)
	}
}
