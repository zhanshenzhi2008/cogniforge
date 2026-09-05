package router

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/agent"
	"cogniforge/internal/auth"
	"cogniforge/internal/chat"
	"cogniforge/internal/config"
	"cogniforge/internal/httpclient"
	"cogniforge/internal/skillimport"
	"cogniforge/internal/knowledge"
	"cogniforge/internal/mail"
	"cogniforge/internal/mcp"
	"cogniforge/internal/middleware"
	"cogniforge/internal/modelcache"
	"cogniforge/internal/memory"
	"cogniforge/internal/monitor"
	"cogniforge/internal/provider"
	"cogniforge/internal/quota"
	"cogniforge/internal/rbac"
	"cogniforge/internal/skill"
	"cogniforge/internal/token"
	"cogniforge/internal/user"
	"cogniforge/internal/workflow"
)

// SetupRoutes 配置所有路由
func SetupRoutes(r *gin.Engine, cfg *config.Config, db *gorm.DB) {
	// 健康检查
	r.GET("/health", healthHandler)
	r.GET("/ready", readyHandler)
	r.GET("/live", liveHandler)

	// 初始化各模块 Handler
	rdb := modelcache.DialRedis(cfg)
	providerRepo := provider.NewRepository(db)
	mc := modelcache.NewFromRedis(rdb)
	providerSvc := provider.NewService(providerRepo, mc)
	providerSvc.RefreshCache()
	providerHandler := provider.NewHandler(providerSvc)

	authSvc := auth.NewAuthServiceWithDeps(db, rdb, buildMailer(cfg), cfg.Mail.PublicURL)
	authHandler := auth.NewAuthHandlerWithService(authSvc)
	userHandler := user.NewUserHandler()

	var quotaStore quota.Store
	if rdb != nil {
		quotaStore = quota.NewRedisStore(rdb)
	}
	quotaSvc := quota.New(db, quotaStore)
	quotaSvc.EnsureDefaultPolicy()
	quotaHandler := quota.NewHandler(quotaSvc)

	chatHandler := chat.NewChatHandler(providerSvc, db, quotaSvc)
	memoryHandler := memory.NewMemoryHandler(db)
	workflowHandler := workflow.NewWorkflowHandler()
	pythonClient := knowledge.NewServiceClient(httpclient.NewClient(cfg.RAG.PythonServiceURL))
	knowledgeHandler := knowledge.NewKnowledgeHandler(pythonClient)
	agentHandler := agent.NewAgentHandler(providerSvc, chatHandler.Service(), quotaSvc)
	monitorHandler := monitor.NewMonitorHandler()
	mcpHandler := mcp.NewHandler(db)
	skillHandler := skill.NewHandler(db)
	importHandler := skillimport.NewHandler(db)
	rbacHandler := rbac.NewRBACHandler()
	tokenSvc := token.NewService(rdb)
	tokenHandler := token.NewHandler(tokenSvc)

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

		// 模型列表 / embeddings 公开；对话 completions 可选登录（Python 回调不带 JWT）
		chatHandler.RegisterPublicRoutes(api)
		api.POST("/chat/completions", middleware.AuthOptional(), chatHandler.Chat)

		// 需要认证的路由
		authenticated := api.Group("")
		authenticated.Use(middleware.AuthRequired())
		{
			authenticated.POST("/chat/stream", chatHandler.ChatStream)
			quotaHandler.RegisterUserRoutes(authenticated)

			// LLM 临时凭证（阶段十四：Chat 记忆，Python 直调用）
			tokenHandler.RegisterRoutes(authenticated)

			// 聊天历史
			chatHandler.RegisterConversationRoutes(authenticated)

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
				kb.POST("/:id/documents/reparse", knowledgeHandler.ReparseDocument)
				kb.POST("/:id/search", knowledgeHandler.SearchKnowledge)
			}

			// AI 供应商配置
			providerHandler.RegisterRoutes(authenticated)

			// Agent
			agentHandler.RegisterRoutes(authenticated)

			// MCP Server 管理（阶段十五 15.1）
			mcpHandler.RegisterRoutes(authenticated)

			// SKILL 管理（阶段十五 15.3）
			skillHandler.RegisterRoutes(authenticated)

			// SKILL 导入（Markdown / ZIP）
			importHandler.RegisterRoutes(authenticated)

			// 长期记忆 CRUD（阶段十四 14.5）
			memoryHandler.RegisterRoutes(authenticated)

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
			admin.POST("/admin/users/:id/reset-password", userHandler.AdminResetPassword)
			admin.POST("/admin/users/:id/roles", rbacHandler.AssignRole)
			admin.GET("/admin/users/:id/role", rbacHandler.GetUserRole)
			quotaHandler.RegisterAdminRoutes(admin)
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

func buildMailer(cfg *config.Config) mail.Sender {
	if cfg == nil {
		return mail.Nop{}
	}
	provider := strings.ToLower(strings.TrimSpace(cfg.Mail.Provider))
	switch provider {
	case "", "smtp", "qq", "163":
		s := mail.NewSMTP(
			cfg.Mail.SMTPHost,
			cfg.Mail.SMTPPort,
			cfg.Mail.SMTPUser,
			cfg.Mail.SMTPPassword,
			cfg.Mail.From,
		)
		if s.Enabled() {
			return s
		}
	case "resend":
		r := mail.NewResend(cfg.Mail.APIKey, cfg.Mail.From)
		if r.Enabled() {
			return r
		}
	}
	// 兜底：哪种配齐用哪种
	s := mail.NewSMTP(cfg.Mail.SMTPHost, cfg.Mail.SMTPPort, cfg.Mail.SMTPUser, cfg.Mail.SMTPPassword, cfg.Mail.From)
	if s.Enabled() {
		return s
	}
	r := mail.NewResend(cfg.Mail.APIKey, cfg.Mail.From)
	if r.Enabled() {
		return r
	}
	return mail.Nop{}
}
