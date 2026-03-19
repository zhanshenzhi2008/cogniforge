package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/orjrs/gateway/pkg/orjrs/gw/handler"
	"github.com/orjrs/gateway/pkg/orjrs/gw/middleware"
)

func main() {
	// Set Gin mode
	gin.SetMode(os.Getenv("GIN_MODE"))

	r := gin.Default()

	// Middleware
	r.Use(middleware.Cors())
	r.Use(middleware.Logger())

	// Health check
	r.GET("/health", handler.Health)
	r.GET("/ready", handler.Ready)
	r.GET("/live", handler.Live)

	// API routes
	api := r.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.POST("/logout", middleware.AuthRequired(), handler.Logout)
			auth.GET("/me", middleware.AuthRequired(), handler.GetCurrentUser)
		}

		// User routes
		user := api.Group("/users")
		user.Use(middleware.AuthRequired())
		{
			user.GET("/", handler.ListUsers)
			user.GET("/:id", handler.GetUser)
			user.PUT("/:id", handler.UpdateUser)
			user.DELETE("/:id", handler.DeleteUser)
		}

		// API Key routes
		apikey := api.Group("/keys")
		apikey.Use(middleware.AuthRequired())
		{
			apikey.POST("/", handler.CreateApiKey)
			apikey.GET("/", handler.ListApiKeys)
			apikey.DELETE("/:id", handler.DeleteApiKey)
		}

		// Model routes
		model := api.Group("/models")
		{
			model.GET("/", handler.ListModels)
			model.GET("/:id", handler.GetModel)
			model.POST("/chat", handler.Chat)
			model.POST("/chat/stream", handler.ChatStream)
		}

		// Agent routes
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

		// Workflow routes
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

		// Knowledge routes
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

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
