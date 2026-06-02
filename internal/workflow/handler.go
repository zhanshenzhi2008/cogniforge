package workflow

import (
	"github.com/gin-gonic/gin"

	"cogniforge/internal/response"
	"cogniforge/internal/trace"
)

type WorkflowHandler struct {
	service *WorkflowService
}

func NewWorkflowHandler() *WorkflowHandler {
	return &WorkflowHandler{
		service: NewWorkflowService(),
	}
}

// ListWorkflows 获取工作流列表
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflows, err := h.service.ListWorkflows(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, workflows)
}

// CreateWorkflow 创建工作流
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	workflow, err := h.service.CreateWorkflow(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, workflow)
}

// GetWorkflow 获取工作流详情
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")
	workflow, err := h.service.GetWorkflow(userID, workflowID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, workflow)
}

// UpdateWorkflow 更新工作流
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")
	var req UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	workflow, err := h.service.UpdateWorkflow(userID, workflowID, &req)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, workflow)
}

// DeleteWorkflow 删除工作流
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")
	err := h.service.DeleteWorkflow(userID, workflowID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, nil, "工作流已删除")
}

// ExecuteWorkflow 执行工作流
func (h *WorkflowHandler) ExecuteWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")
	var req ExecuteWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Input = make(map[string]any)
	}

	result, err := h.service.ExecuteWorkflowWithTraceID(userID, workflowID, &req, trace.GetTraceIDFromGin(c))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Accepted(c, result)
}

// ListWorkflowExecutions 获取执行记录列表
func (h *WorkflowHandler) ListWorkflowExecutions(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")
	executions, err := h.service.ListWorkflowExecutions(userID, workflowID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, executions)
}

// GetWorkflowExecution 获取执行记录详情
func (h *WorkflowHandler) GetWorkflowExecution(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	executionID := c.Param("executionId")
	execution, err := h.service.GetWorkflowExecution(userID, executionID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, execution)
}

// CancelWorkflowExecution 取消执行
func (h *WorkflowHandler) CancelWorkflowExecution(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	executionID := c.Param("executionId")
	err := h.service.CancelWorkflowExecution(userID, executionID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, nil, "执行已取消")
}

// DebugWorkflow 调试工作流
func (h *WorkflowHandler) DebugWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")
	var req DebugWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Input = make(map[string]any)
	}

	result, err := h.service.DebugWorkflowWithTraceID(userID, workflowID, &req, trace.GetTraceIDFromGin(c))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Accepted(c, result)
}

// RegisterRoutes 注册路由
func (h *WorkflowHandler) RegisterRoutes(rg *gin.RouterGroup) {
	workflows := rg.Group("/workflows")
	{
		workflows.GET("", h.ListWorkflows)
		workflows.POST("", h.CreateWorkflow)
		workflows.GET("/:id", h.GetWorkflow)
		workflows.PUT("/:id", h.UpdateWorkflow)
		workflows.DELETE("/:id", h.DeleteWorkflow)
		workflows.POST("/:id/execute", h.ExecuteWorkflow)
		workflows.POST("/:id/debug", h.DebugWorkflow)
		workflows.GET("/:id/executions", h.ListWorkflowExecutions)
		workflows.GET("/executions/:executionId", h.GetWorkflowExecution)
		workflows.POST("/executions/:executionId/cancel", h.CancelWorkflowExecution)
	}
}
