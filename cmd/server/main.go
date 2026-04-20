package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/config"
	"cogniforge/internal/database"
	"cogniforge/internal/handler"
	"cogniforge/internal/logger"
	"cogniforge/internal/middleware"
	"cogniforge/internal/model"
	"cogniforge/internal/util"
)

func main() {
	logger.Init()

	gin.SetMode(os.Getenv("GIN_MODE"))

	cfg := config.Load()
	handler.SetChatConfig(cfg)

	db, err := database.Connect(cfg)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		return
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserSettings{},
		&model.UserSession{},
		&model.ApiKey{},
		&model.Agent{},
		&model.Workflow{},
		&model.WorkflowNode{},
		&model.WorkflowEdge{},
		&model.WorkflowExecution{},
		&model.KnowledgeBase{},
		&model.Document{},
		&model.RequestLog{},
		&model.Permission{},
		&model.Role{},
		&model.RolePermission{},
	); err != nil {
		slog.Error("failed to migrate database", "error", err)
		return
	}

	handler.InitDefaultAdmin()
	handler.InitDefaultRoles()

	// 初始化 GeoIP 数据库
	if err := util.InitGeoIP(util.DefaultGeoIPPath()); err != nil {
		slog.Warn("failed to initialize GeoIP database", "error", err, "path", util.DefaultGeoIPPath())
	} else {
		slog.Info("GeoIP database initialized", "path", util.DefaultGeoIPPath())
	}

	// 创建 gin engine
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Cors())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Logger())

	// 移除 URL 中多余的斜杠（如 /api//v1 -> /api/v1）
	r.RemoveExtraSlash = true

	// 设置文件上传大小限制为 50MB
	r.MaxMultipartMemory = 50 << 20

	r.GET("/health", handler.Health)
	r.GET("/ready", handler.Ready)
	r.GET("/live", handler.Live)

	api := r.Group("/api/v1")
	{
		api.POST("/chat/stream", handler.ChatStream)

		auth := api.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.POST("/logout", middleware.AuthRequired(), handler.Logout)
			auth.GET("/me", middleware.AuthRequired(), handler.GetCurrentUser)
			auth.PUT("/me", middleware.AuthRequired(), handler.UpdateCurrentUser) // 新增
		}

		user := api.Group("/users")
		user.Use(middleware.AuthRequired())
		{
			user.GET("", handler.ListUsers)
			user.GET("/:id", handler.GetUser)
			user.PUT("/:id", handler.UpdateUser)
			user.DELETE("/:id", handler.DeleteUser)
		}

		apikey := api.Group("/keys")
		apikey.Use(middleware.AuthRequired())
		{
			apikey.POST("", handler.CreateApiKey)
			apikey.GET("", handler.ListApiKeys)
			apikey.DELETE("/:id", handler.DeleteApiKey)
		}

		model := api.Group("/models")
		{
			model.GET("", handler.ListModels)
			model.GET("/:id", handler.GetModel)
			model.POST("/chat", handler.Chat)
		}

		agent := api.Group("/agents")
		agent.Use(middleware.AuthRequired())
		{
			agent.GET("", handler.ListAgents)
			agent.POST("", handler.CreateAgent)
			agent.GET("/:id", handler.GetAgent)
			agent.PUT("/:id", handler.UpdateAgent)
			agent.DELETE("/:id", handler.DeleteAgent)
			agent.POST("/:id/chat", handler.AgentChat)
		}

		workflow := api.Group("/workflows")
		workflow.Use(middleware.AuthRequired())
		{
			workflow.GET("", handler.ListWorkflows)
			workflow.POST("", handler.CreateWorkflow)
			workflow.GET("/:id", handler.GetWorkflow)
			workflow.PUT("/:id", handler.UpdateWorkflow)
			workflow.DELETE("/:id", handler.DeleteWorkflow)
			workflow.POST("/:id/execute", handler.ExecuteWorkflow)
		}

		knowledge := api.Group("/knowledge")
		knowledge.Use(middleware.AuthRequired())
		{
			// 同时支持 /api/v1/knowledge 和 /api/v1/knowledge/
			knowledge.GET("", handler.ListKnowledgeBases)
			knowledge.GET("/", handler.ListKnowledgeBases)
			knowledge.POST("", handler.CreateKnowledgeBase)
			knowledge.POST("/", handler.CreateKnowledgeBase)
			knowledge.GET("/:id", handler.GetKnowledgeBase)
			knowledge.PUT("/:id", handler.UpdateKnowledgeBase)
			knowledge.DELETE("/:id", handler.DeleteKnowledgeBase)
			knowledge.POST("/:id/documents", handler.UploadDocument)
			knowledge.GET("/:id/documents", handler.ListDocuments)
			knowledge.DELETE("/:id/documents/:docId", handler.DeleteDocument)
			knowledge.POST("/:id/search", handler.SearchKnowledge)
		}

		// 监控中心 API
		monitor := api.Group("/monitor")
		monitor.Use(middleware.AuthRequired())
		{
			monitor.GET("/logs", handler.ListRequestLogs)
			monitor.GET("/logs/:id", handler.GetRequestLog)
			monitor.GET("/stats", handler.GetUsageStats)
			monitor.GET("/stats/realtime", handler.GetRealtimeStats)
		}

		// 个人设置 API
		settings := api.Group("/settings")
		settings.Use(middleware.AuthRequired())
		{
			settings.GET("", handler.GetSettings)
			settings.PUT("", handler.UpdateSettings)
			settings.POST("/password", handler.ChangePassword)
			settings.POST("/avatar", handler.UploadAvatar)
			settings.GET("/sessions", handler.GetSessions)
			settings.DELETE("/sessions/:id", handler.RevokeSession)
		}

		// 用户管理 API（管理员）
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired())
		admin.Use(middleware.RequireAdmin())
		{
			admin.GET("/users", handler.GetUsers)
			admin.POST("/users", handler.CreateUser)
			admin.GET("/users/:id", handler.GetUser)
			admin.PUT("/users/:id", handler.UpdateUser)
			admin.DELETE("/users/:id", handler.DeleteUser)
			admin.PATCH("/users/:id/status", handler.UpdateUserStatus)

			// 角色管理
			admin.GET("/roles", handler.ListRoles)
			admin.POST("/roles", handler.CreateRole)
			admin.GET("/roles/:id", handler.GetRole)
			admin.PUT("/roles/:id", handler.UpdateRole)
			admin.DELETE("/roles/:id", handler.DeleteRole)

			// 权限管理
			admin.GET("/permissions", handler.ListPermissions)
			admin.POST("/permissions", handler.CreatePermission)
			admin.DELETE("/permissions/:id", handler.DeletePermission)

			// 用户角色分配
			admin.POST("/users/:id/roles", handler.AssignRole)
			admin.GET("/users/:id/role", handler.GetUserRole)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("server starting", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("failed to start server", "error", err)
		return
	}
}
