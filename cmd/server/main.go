package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/auth"
	"cogniforge/internal/config"
	"cogniforge/internal/crypto"
	"cogniforge/internal/database"
	"cogniforge/internal/logger"
	"cogniforge/internal/middleware"
	"cogniforge/internal/model"
	"cogniforge/internal/router"
	"cogniforge/internal/util"
)

func main() {
	logger.Init()

	gin.SetMode(os.Getenv("GIN_MODE"))

	cfg := config.Load()

	crypto.Init(cfg.Encryption.Key)

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
		&model.AIProvider{},
	); err != nil {
		slog.Error("failed to migrate database", "error", err)
		return
	}

	// 初始化默认 AI 供应商
	if err := model.InitDefaultProviders(db); err != nil {
		slog.Warn("failed to init default AI providers", "error", err)
	}

	// 初始化默认管理员
	authHandler := auth.NewAuthHandler()
	authHandler.InitDefaultAdmin()

	// 初始化默认角色和权限
	// 注意：rbac.NewRBACHandler().InitDefaultRoles() 已由 auth.InitDefaultAdmin() 内部调用
	// 如需单独调用，可取消下面注释
	// rbacHandler := rbac.NewRBACHandler()
	// rbacHandler.InitDefaultRoles()

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

	r.RemoveExtraSlash = true
	r.MaxMultipartMemory = 50 << 20

	// 设置路由
	router.SetupRoutes(r, cfg, db)

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
