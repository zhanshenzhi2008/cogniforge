package router

import (
	"github.com/gin-gonic/gin"

	"cogniforge/internal/agent"
	"cogniforge/internal/auth"
	"cogniforge/internal/chat"
	"cogniforge/internal/config"
	"cogniforge/internal/knowledge"
	"cogniforge/internal/middleware"
	"cogniforge/internal/monitor"
	"cogniforge/internal/rbac"
	"cogniforge/internal/user"
	"cogniforge/internal/workflow"
)

// SetupRoutes 配置所有路由
func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	// 健康检查
	r.GET("/health", healthHandler)
	r.GET("/ready", readyHandler)
	r.GET("/live", liveHandler)

	// 初始化各模块 Handler
	authHandler := auth.NewAuthHandler()
	userHandler := user.NewUserHandler()
	chatHandler := chat.NewChatHandler(cfg)
	workflowHandler := workflow.NewWorkflowHandler()
	pythonClient := knowledge.NewPythonServiceClient(cfg)
	knowledgeHandler := knowledge.NewKnowledgeHandler(pythonClient)
	agentHandler := agent.NewAgentHandler(cfg.AI.DefaultModel)
	monitorHandler := monitor.NewMonitorHandler()
	rbacHandler := rbac.NewRBACHandler()

	api := r.Group("/api/v1")
	{
		// 认证相关（公开接口）
		authHandler.RegisterRoutes(api)

		// API Key 路由（简化路径）
		authKeys := api.Group("/keys")
		authKeys.Use(middleware.AuthRequired())
		{
			authKeys.GET("", authHandler.ListApiKeys)
			authKeys.POST("", authHandler.CreateApiKey)
			authKeys.DELETE("/:id", authHandler.DeleteApiKey)
		}

		// 聊天/模型相关（公开接口）
		chatHandler.RegisterRoutes(api)

		// 需要认证的路由
		authenticated := api.Group("")
		authenticated.Use(middleware.AuthRequired())
		{
			// 用户管理
			userHandler.RegisterRoutes(authenticated)

			// 工作流
			workflowHandler.RegisterRoutes(authenticated)

			// 知识库
			knowledgeHandler.RegisterRoutes(authenticated)

			// 知识库路由（简化路径）
			kb := authenticated.Group("/knowledge")
			{
				kb.GET("", knowledgeHandler.ListKnowledgeBases)
				kb.POST("", knowledgeHandler.CreateKnowledgeBase)
				kb.GET("/:id", knowledgeHandler.GetKnowledgeBase)
				kb.PUT("/:id", knowledgeHandler.UpdateKnowledgeBase)
				kb.DELETE("/:id", knowledgeHandler.DeleteKnowledgeBase)
				kb.GET("/:id/documents", knowledgeHandler.ListDocuments)
				kb.DELETE("/:id/documents/:docId", knowledgeHandler.DeleteDocument)
				kb.POST("/:id/documents/upload", knowledgeHandler.UploadDocument)
				kb.POST("/:id/search", knowledgeHandler.SearchKnowledge)
			}

			// Agent
			agentHandler.RegisterRoutes(authenticated)

			// 监控
			monitorHandler.RegisterRoutes(authenticated)
		}

		// 管理员路由
		admin := api.Group("")
		admin.Use(middleware.AuthRequired())
		admin.Use(middleware.RequireAdmin())
		{
			// RBAC
			rbacHandler.RegisterRoutes(admin)

			// 管理员路由（简化路径）
			admin.GET("/admin/users", userHandler.GetUsers)
			admin.POST("/admin/users", userHandler.CreateUser)
			admin.GET("/admin/users/:id", userHandler.GetUser)
			admin.PUT("/admin/users/:id", userHandler.UpdateUser)
			admin.DELETE("/admin/users/:id", userHandler.DeleteUser)
			admin.PATCH("/admin/users/:id/status", userHandler.UpdateUserStatus)
			admin.POST("/admin/users/:id/roles", rbacHandler.AssignRole)
			admin.GET("/admin/users/:id/role", rbacHandler.GetUserRole)
		}
	}
}

// ============ 健康检查 ============

func healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "ok",
		"timestamp": "2024-01-01T00:00:00Z",
		"version":   "1.0.0",
	})
}

func readyHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ready"})
}

func liveHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "alive"})
}
