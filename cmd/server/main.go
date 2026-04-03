package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/config"
	"cogniforge/internal/database"
	"cogniforge/internal/handler"
	"cogniforge/internal/logger"
	"cogniforge/internal/middleware"
	"cogniforge/internal/model"
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
		&model.ApiKey{},
		&model.Agent{},
		&model.Workflow{},
		&model.WorkflowNode{},
		&model.WorkflowEdge{},
		&model.WorkflowExecution{},
	); err != nil {
		slog.Error("failed to migrate database", "error", err)
		return
	}

	handler.InitDefaultAdmin()

	r := gin.Default()
	r.Use(middleware.Cors())
	r.Use(middleware.Logger())

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
		}

		user := api.Group("/users")
		user.Use(middleware.AuthRequired())
		{
			user.GET("/", handler.ListUsers)
			user.GET("/:id", handler.GetUser)
			user.PUT("/:id", handler.UpdateUser)
			user.DELETE("/:id", handler.DeleteUser)
		}

		apikey := api.Group("/keys")
		apikey.Use(middleware.AuthRequired())
		{
			apikey.POST("/", handler.CreateApiKey)
			apikey.GET("/", handler.ListApiKeys)
			apikey.DELETE("/:id", handler.DeleteApiKey)
		}

		model := api.Group("/models")
		{
			model.GET("/", handler.ListModels)
			model.GET("/:id", handler.GetModel)
			model.POST("/chat", handler.Chat)
		}

		agent := api.Group("/agents")
		agent.Use(middleware.AuthRequired())
		{
			agent.GET("/", handler.ListAgents)
			agent.POST("/", handler.CreateAgent)
			agent.GET("/:id", handler.GetAgent)
			agent.PUT("/:id", handler.UpdateAgent)
			agent.DELETE("/:id", handler.DeleteAgent)
			agent.POST("/:id/chat", handler.AgentChat)
		}

		workflow := api.Group("/workflows")
		workflow.Use(middleware.AuthRequired())
		{
			workflow.GET("/", handler.ListWorkflows)
			workflow.POST("/", handler.CreateWorkflow)
			workflow.GET("/:id", handler.GetWorkflow)
			workflow.PUT("/:id", handler.UpdateWorkflow)
			workflow.DELETE("/:id", handler.DeleteWorkflow)
			workflow.POST("/:id/execute", handler.ExecuteWorkflow)
		}

		knowledge := api.Group("/knowledge")
		knowledge.Use(middleware.AuthRequired())
		{
			knowledge.GET("/", handler.ListKnowledgeBases)
			knowledge.POST("/", handler.CreateKnowledgeBase)
			knowledge.GET("/:id", handler.GetKnowledgeBase)
			knowledge.PUT("/:id", handler.UpdateKnowledgeBase)
			knowledge.DELETE("/:id", handler.DeleteKnowledgeBase)
			knowledge.POST("/:id/documents", handler.UploadDocument)
			knowledge.GET("/:id/documents", handler.ListDocuments)
			knowledge.DELETE("/:id/documents/:docId", handler.DeleteDocument)
			knowledge.POST("/:id/search", handler.SearchKnowledge)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, stop := signal.NotifyContext(ctx, os.Signal(syscall.SIGINT, syscall.SIGTERM))
	defer stop()

	slog.Info("server starting", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("failed to start server", "error", err)
		return
	}
}
