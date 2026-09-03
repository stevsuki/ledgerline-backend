// Package router: assembly of every HTTP route.
package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Blank import: registers the swag-generated documentation.
	_ "github.com/stevensuki/ledgerline-backend/docs"
	"github.com/stevensuki/ledgerline-backend/internal/config"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/handler"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/middleware"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/validator"
)

// Dependencies: everything the router needs, injected from main.
type Dependencies struct {
	Config       *config.Config
	Logger       *slog.Logger
	TokenManager domain.TokenManager
	Auth         *handler.AuthHandler
	User         *handler.UserHandler
	Category     *handler.CategoryHandler
	Health       *handler.HealthHandler
	Role         *handler.RoleHandler
	AuditLog     *handler.AuditLogHandler
	Wallet       *handler.WalletHandler
}

// New assembles the Gin engine with global middleware and all routes.
func New(deps Dependencies) *gin.Engine {
	if deps.Config.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	validator.RegisterGinValidator()

	engine := gin.New()
	engine.RedirectTrailingSlash = true
	engine.HandleMethodNotAllowed = true

	rateLimiter := middleware.NewRateLimiter(deps.Config.HTTP.RateLimitRPS, deps.Config.HTTP.RateLimitBurst)
	engine.Use(
		middleware.RequestID(),
		middleware.Logger(deps.Logger),
		middleware.Recovery(deps.Logger),
		middleware.SecureHeaders(),
		middleware.CORS(deps.Config.HTTP.AllowedOrigins),
		middleware.Timeout(deps.Config.HTTP.RequestTimeout),
		rateLimiter.Middleware(),
	)

	engine.NoRoute(handler.NotFound)
	engine.NoMethod(handler.MethodNotAllowed)

	engine.GET("/health", deps.Health.Liveness)
	engine.GET("/health/ready", deps.Health.Readiness)

	if !deps.Config.App.IsProduction() {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	v1 := engine.Group("/api/v1")
	registerAuthRoutes(v1, deps)
	registerUserRoutes(v1, deps)
	registerCategoryRoutes(v1, deps)
	registerRoleRoutes(v1, deps)
	registerAuditLogRoutes(v1, deps)
	registerWalletRoutes(v1, deps)

	return engine
}

func registerAuthRoutes(rg *gin.RouterGroup, deps Dependencies) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", deps.Auth.Register)
		auth.POST("/login", deps.Auth.Login)
		auth.POST("/refresh", deps.Auth.Refresh)
		auth.GET("/me", middleware.Authenticate(deps.TokenManager), deps.Auth.Me)
		auth.POST("/forgot-password", deps.Auth.ForgotPassword)
		auth.POST("/verify-otp", deps.Auth.VerifyOTPResetPassword)
		auth.POST("/reset-password", deps.Auth.ResetPassword)
	}
}

func registerUserRoutes(rg *gin.RouterGroup, deps Dependencies) {
	users := rg.Group("/users", middleware.Authenticate(deps.TokenManager))
	{
		users.GET("", deps.User.List)
		users.GET("/:id", deps.User.GetByID)
		users.POST("", deps.User.Create)
		users.PATCH("/:id", deps.User.Update)
		users.DELETE("/:id", deps.User.Delete)

		// TODO: move create, update and delete behind middleware.RequireRoles(domain.RoleIDAdmin).
	}
}

func registerCategoryRoutes(rg *gin.RouterGroup, deps Dependencies) {
	categories := rg.Group("/categories", middleware.Authenticate(deps.TokenManager))
	{
		categories.POST("", deps.Category.Create)
		categories.GET("", deps.Category.List)
		categories.GET("/:id", deps.Category.GetByID)
		categories.PATCH("/:id", deps.Category.Update)
		categories.DELETE("/:id", deps.Category.Delete)
	}
}

func registerRoleRoutes(rg *gin.RouterGroup, deps Dependencies) {
	roles := rg.Group("/roles", middleware.Authenticate(deps.TokenManager))
	{
		roles.GET("", deps.Role.List)
		roles.GET("/:id", deps.Role.GetByID)
		roles.POST("", deps.Role.Create)
		roles.PATCH("/:id", deps.Role.Update)
		roles.DELETE("/:id", deps.Role.Delete)

		// TODO: move create, update and delete behind middleware.RequireRoles(domain.RoleIDAdmin).
	}
}

func registerAuditLogRoutes(rg *gin.RouterGroup, deps Dependencies) {
	auditLogs := rg.Group("/audit-logs", middleware.Authenticate(deps.TokenManager))
	{
		auditLogs.GET("", deps.AuditLog.List)
		auditLogs.GET("/overview", deps.AuditLog.Overview)
		auditLogs.GET("/options", deps.AuditLog.Options)
		auditLogs.GET("/export", deps.AuditLog.Export)
	}
}

func registerWalletRoutes(rg *gin.RouterGroup, deps Dependencies) {
	wallets := rg.Group("/wallets", middleware.Authenticate(deps.TokenManager))
	{
		wallets.GET("", deps.Wallet.List)
		wallets.GET("/:id", deps.Wallet.GetByID)
		wallets.POST("", deps.Wallet.Create)
		wallets.PATCH("/:id", deps.Wallet.Update)
		wallets.DELETE("/:id", deps.Wallet.Delete)
		wallets.GET("/overview", deps.Wallet.Overview)
	}
}
