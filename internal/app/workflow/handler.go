package workflow

import (
	"strconv"

	"github.com/gin-gonic/gin"

	orchwf "cogniforge/internal/orchestration/workflow"
	"cogniforge/internal/infra/trace"
	"cogniforge/internal/shared/response"
)

type WorkflowHandler struct {
	service *orchwf.WorkflowService
}

func NewWorkflowHandler() *WorkflowHandler {
	return NewWorkflowHandlerWithService(orchwf.NewWorkflowService())
}

// NewWorkflowHandlerWithService 注入带 AI 依赖的 WorkflowService
func NewWorkflowHandlerWithService(svc *orchwf.WorkflowService) *WorkflowHandler {
	if svc == nil {
		svc = orchwf.NewWorkflowService()
	}
	return &WorkflowHandler{service: svc}
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

	var req orchwf.CreateWorkflowRequest
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
	var req orchwf.UpdateWorkflowRequest
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

// CloneWorkflow 复制工作流
func (h *WorkflowHandler) CloneWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")
	workflow, err := h.service.CloneWorkflow(userID, workflowID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Created(c, workflow)
}

// ListWorkflowVersions 历史版本列表
func (h *WorkflowHandler) ListWorkflowVersions(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}
	list, err := h.service.ListWorkflowVersions(userID, c.Param("id"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, list)
}

// RollbackWorkflowVersion 回滚到指定历史版本
func (h *WorkflowHandler) RollbackWorkflowVersion(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}
	ver, err := strconv.Atoi(c.Param("version"))
	if err != nil || ver < 1 {
		response.BadRequest(c, "无效版本号")
		return
	}
	workflow, err := h.service.RollbackWorkflowVersion(userID, c.Param("id"), ver)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, workflow)
}

// ExecuteWorkflow 执行工作流
func (h *WorkflowHandler) ExecuteWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")
	var req orchwf.ExecuteWorkflowRequest
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
	detail, err := h.service.GetWorkflowExecutionDetail(userID, executionID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, detail)
}

// ListExecutionNodeRuns 节点运行列表
func (h *WorkflowHandler) ListExecutionNodeRuns(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}
	executionID := c.Param("executionId")
	runs, err := h.service.ListNodeRuns(userID, executionID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, runs)
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

// ListExecutionCheckpoints 检查点列表
func (h *WorkflowHandler) ListExecutionCheckpoints(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}
	executionID := c.Param("executionId")
	list, err := h.service.ListCheckpoints(userID, executionID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, list)
}

// ResumeWorkflowExecution 从 checkpoint 继续
func (h *WorkflowHandler) ResumeWorkflowExecution(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}
	executionID := c.Param("executionId")
	var req orchwf.ResumeExecutionRequest
	_ = c.ShouldBindJSON(&req)
	result, err := h.service.ResumeExecution(userID, executionID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Accepted(c, result)
}

// ApproveWorkflowExecution 审批通过（W3）
func (h *WorkflowHandler) ApproveWorkflowExecution(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}
	executionID := c.Param("executionId")
	var req orchwf.ApproveExecutionRequest
	_ = c.ShouldBindJSON(&req)
	result, err := h.service.ApproveExecution(userID, executionID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Accepted(c, result)
}

// RejectWorkflowExecution 审批拒绝（W3）
func (h *WorkflowHandler) RejectWorkflowExecution(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}
	executionID := c.Param("executionId")
	var req orchwf.RejectExecutionRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.service.RejectExecution(userID, executionID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "已拒绝")
}

// CallbackWorkflowExecution 外部回调 approve/reject（W3）
func (h *WorkflowHandler) CallbackWorkflowExecution(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}
	executionID := c.Param("executionId")
	var req orchwf.CallbackExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "action 必填（approve|reject）")
		return
	}
	result, err := h.service.CallbackExecution(userID, executionID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if result.Status == "failed" {
		response.Success(c, result)
		return
	}
	response.Accepted(c, result)
}

// DebugWorkflow 调试工作流
func (h *WorkflowHandler) DebugWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")
	var req orchwf.DebugWorkflowRequest
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
		workflows.POST("/:id/clone", h.CloneWorkflow)
		workflows.GET("/:id/versions", h.ListWorkflowVersions)
		workflows.POST("/:id/versions/:version/rollback", h.RollbackWorkflowVersion)
		workflows.POST("/:id/execute", h.ExecuteWorkflow)
		workflows.POST("/:id/debug", h.DebugWorkflow)
		workflows.GET("/:id/executions", h.ListWorkflowExecutions)
		workflows.GET("/executions/:executionId", h.GetWorkflowExecution)
		workflows.GET("/executions/:executionId/nodes", h.ListExecutionNodeRuns)
		workflows.GET("/executions/:executionId/checkpoints", h.ListExecutionCheckpoints)
		workflows.POST("/executions/:executionId/resume", h.ResumeWorkflowExecution)
		workflows.POST("/executions/:executionId/approve", h.ApproveWorkflowExecution)
		workflows.POST("/executions/:executionId/reject", h.RejectWorkflowExecution)
		workflows.POST("/executions/:executionId/callback", h.CallbackWorkflowExecution)
		workflows.POST("/executions/:executionId/cancel", h.CancelWorkflowExecution)
	}
}
