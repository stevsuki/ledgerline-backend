package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/stevensuki/ledgerline-backend/internal/config"
	"github.com/stevensuki/ledgerline-backend/internal/database"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/handler"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/router"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres"
	"github.com/stevensuki/ledgerline-backend/internal/server"
	"github.com/stevensuki/ledgerline-backend/internal/service"
	"github.com/stevensuki/ledgerline-backend/pkg/hash"
	"github.com/stevensuki/ledgerline-backend/pkg/jwt"
	"github.com/stevensuki/ledgerline-backend/pkg/logger"
	"github.com/stevensuki/ledgerline-backend/pkg/mailer"
	"github.com/stevensuki/ledgerline-backend/pkg/otp"
)

// @title						LedgerLine API
// @version					1.0
// @description				LedgerLine backend API.
// @termsOfService				https://example.com/terms
// @contact.name				Tim LedgerLine
// @contact.email				dev@example.com
// @license.name				MIT
// @host						localhost:8080
// @BasePath					/api/v1
// @schemes					http https
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Fill with: Bearer {token}
func main() {
	if err := run(); err != nil {
		slog.Error("application stopped because of an error", slog.Any("error", err))
		os.Exit(1)
	}
}

// run separates startup logic from main so defer still runs on error.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format).With(
		slog.String("service", cfg.App.Name),
		slog.String("env", cfg.App.Env),
	)
	slog.SetDefault(log)

	// Context cancelled when SIGINT/SIGTERM arrives.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.New(ctx, cfg.Database, log, !cfg.App.IsProduction())
	if err != nil {
		return err
	}
	defer func() { _ = database.Close(db) }()
	log.Info("database connected")

	// Dependency wiring: repository -> service -> handler.
	userRepo := postgres.NewUserRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)
	passwordResetTokenRepo := postgres.NewPasswordResetTokenRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	menuRepo := postgres.NewMenuRepository(db)
	txManager := postgres.NewTxManager(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	walletRepo := postgres.NewWalletRepository(db)

	hasher := hash.NewBcrypt(0)
	tokenManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL, cfg.JWT.ResetTokenTTL)
	otpGenerator := otp.NewGenerator(cfg.OTP.Length)

	// SMTP disabled -> emails go to the log so local dev needs no mail server.
	var mail domain.Mailer = mailer.NewLog(log)
	if cfg.SMTP.Enabled {
		smtpMailer, err := mailer.NewSMTP(mailer.Config{
			Host:        cfg.SMTP.Host,
			Port:        cfg.SMTP.Port,
			Username:    cfg.SMTP.Username,
			Password:    cfg.SMTP.Password,
			FromName:    cfg.SMTP.FromName,
			FromAddress: cfg.SMTP.FromAddress,
			AppName:     cfg.App.Name,
			Timeout:     cfg.SMTP.Timeout,
			TLS:         cfg.SMTP.TLS,
		})
		if err != nil {
			return err
		}
		mail = smtpMailer
	}

	userService := service.NewUserService(userRepo, hasher, auditLogRepo)
	authService := service.NewAuthService(userRepo, hasher, tokenManager, mail, otpGenerator, passwordResetTokenRepo, menuRepo, txManager, cfg.OTP.TTL, auditLogRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	roleService := service.NewRoleService(roleRepo, auditLogRepo)
	auditLogService := service.NewAuditLogService(auditLogRepo)
	walletService := service.NewWalletService(walletRepo, auditLogRepo)

	engine := router.New(router.Dependencies{
		Config:       cfg,
		Logger:       log,
		TokenManager: tokenManager,
		Auth:         handler.NewAuthHandler(authService),
		User:         handler.NewUserHandler(userService),
		Category:     handler.NewCategoryHandler(categoryService),
		Health:       handler.NewHealthHandler(db, cfg.App.Version),
		Role:         handler.NewRoleHandler(roleService),
		AuditLog:     handler.NewAuditLogHandler(auditLogService),
		Wallet:       handler.NewWalletHandler(walletService),
	})

	return server.New(cfg.HTTP, engine, log).Run(ctx)
}
