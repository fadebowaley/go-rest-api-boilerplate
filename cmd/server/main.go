package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"

	_ "github.com/fadebowaley/applico/api/docs"
	"github.com/fadebowaley/applico/internal/applicant"
	"github.com/fadebowaley/applico/internal/audit"
	"github.com/fadebowaley/applico/internal/auth"
	"github.com/fadebowaley/applico/internal/config"
	"github.com/fadebowaley/applico/internal/db"
	"github.com/fadebowaley/applico/internal/disbursement"
	"github.com/fadebowaley/applico/internal/form"
	"github.com/fadebowaley/applico/internal/project"
	"github.com/fadebowaley/applico/internal/stats"
	"github.com/fadebowaley/applico/internal/document"
	"github.com/fadebowaley/applico/internal/grant"
	"github.com/fadebowaley/applico/internal/migrate"
	"github.com/fadebowaley/applico/internal/permission"
	"github.com/fadebowaley/applico/internal/program"
	"github.com/fadebowaley/applico/internal/server"
	"github.com/fadebowaley/applico/internal/tenant"
	"github.com/fadebowaley/applico/internal/user"
	"github.com/fadebowaley/applico/internal/workflow"
)

// @title Applico API
// @version 1.0
// @description Multi-tenant grant management platform for faith-based and community organizations. Supports tenants, users, roles, grant programs, dynamic forms, applicant management, grant applications, document management, configurable workflows, disbursements, project updates, dashboards, and audit logging.
// @termsOfService http://swagger.io/terms/

// @contact.name Applico Support
// @contact.url https://applico.dev/support
// @contact.email support@applico.dev

// @license.name Proprietary
// @license.url https://applico.dev/license

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

type workflowStarterAdapter struct {
	svc workflow.Service
}

type userLookupAdapter struct {
	svc user.Service
}

func (a *userLookupAdapter) FindByEmail(ctx context.Context, email string) (uint, string, error) {
	u, err := a.svc.GetUserByEmail(ctx, email)
	if err != nil {
		return 0, "", err
	}
	return u.ID, u.Email, nil
}

func (a *workflowStarterAdapter) StartWorkflow(ctx context.Context, applicationID, workflowTemplateID uint) error {
	_, err := a.svc.StartWorkflow(ctx, applicationID, workflowTemplateID)
	return err
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	logger := slog.Default()
	logger.Info("Starting Go REST API Boilerplate...")

	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		return err
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("Configuration validation failed", "error", err)
		return err
	}

	cfg.LogSafeConfig(logger)

	database, err := db.NewPostgresDBFromDatabaseConfig(cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return err
	}

	if os.Getenv("SKIP_MIGRATION_CHECK") == "" {
		if err := checkMigrationStatus(database, &cfg.Migrations); err != nil {
			logger.Warn("Migration check", "status", "⚠️", "error", err)
		} else {
			logger.Info("Migration check", "status", "✓")
		}
	}

	authService := auth.NewServiceWithRepo(&cfg.JWT, database)
	userRepo := user.NewRepository(database)
	userService := user.NewService(userRepo)

	tenantRepo := tenant.NewRepository(database)
	tenantService := tenant.NewServiceWithUsers(tenantRepo, &userLookupAdapter{svc: userService})
	tenantHandler := tenant.NewHandler(tenantService)

	userHandler := user.NewHandler(userService, authService, tenantService)

	permissionRepo := permission.NewRepository(database)
	permissionService := permission.NewService(permissionRepo)
	permissionHandler := permission.NewHandler(permissionService)

	auditRepo := audit.NewRepository(database)
	auditService := audit.NewService(auditRepo)
	auditHandler := audit.NewHandler(auditService)

	workflowRepo := workflow.NewRepository(database)
	workflowService := workflow.NewService(workflowRepo)

	grantRepo := grant.NewRepository(database)
	grantService := grant.NewServiceWithWorkflow(grantRepo, &workflowStarterAdapter{svc: workflowService})
	grantHandler := grant.NewHandler(grantService, auditService)

	workflowService.SetGrantUpdater(grantService)
	workflowHandler := workflow.NewHandler(workflowService, auditService)

	documentRepo := document.NewRepository(database)
	documentStorage := document.NewLocalStorage("uploads")
	documentService := document.NewService(documentRepo, grantService, documentStorage)
	documentHandler := document.NewHandler(documentService, auditService)

	programRepo := program.NewRepository(database)
	programService := program.NewService(programRepo)
	programHandler := program.NewHandler(programService)

	applicantRepo := applicant.NewRepository(database)
	applicantService := applicant.NewService(applicantRepo)
	applicantHandler := applicant.NewHandler(applicantService)

	disbursementRepo := disbursement.NewRepository(database)
	disbursementService := disbursement.NewService(disbursementRepo, grantService)
	disbursementHandler := disbursement.NewHandler(disbursementService, auditService)

	projectRepo := project.NewRepository(database)
	projectService := project.NewService(projectRepo, grantService)
	projectHandler := project.NewHandler(projectService, auditService)

	statsRepo := stats.NewRepository(database)
	statsService := stats.NewService(statsRepo)
	statsHandler := stats.NewHandler(statsService)

	formRepo := form.NewRepository(database)
	formService := form.NewService(formRepo)
	formHandler := form.NewHandler(formService, auditService)

	router := server.SetupRouter(userHandler, authService, cfg, database, permissionHandler, auditHandler, grantHandler, documentHandler, tenantHandler, tenantService, programHandler, applicantHandler, workflowHandler, disbursementHandler, projectHandler, statsHandler, formHandler)

	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	maxHeaderBytes := cfg.Server.MaxHeaderBytes
	if maxHeaderBytes == 0 {
		maxHeaderBytes = 1 << 20
	}

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}

	go func() {
		logger.Info("Server starting", "address", srv.Addr)
		logger.Info("Swagger UI available", "url", fmt.Sprintf("http://localhost:%s/swagger/index.html", port))
		logger.Info("Health check available", "url", fmt.Sprintf("http://localhost:%s/health", port))
		logger.Info("Liveness probe available", "url", fmt.Sprintf("http://localhost:%s/health/live", port))
		logger.Info("Readiness probe available", "url", fmt.Sprintf("http://localhost:%s/health/ready", port))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("Received shutdown signal", "signal", sig)
	logger.Info("Shutting down server gracefully...")

	sqlDB, err := database.DB()
	if err == nil {
		logger.Info("Closing database connections...")
		if err := sqlDB.Close(); err != nil {
			logger.Error("Error closing database", "error", err)
		}
	}

	shutdownTimeout := time.Duration(cfg.Server.ShutdownTimeout) * time.Second
	if shutdownTimeout == 0 {
		shutdownTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	logger.Info("Server exited gracefully")
	return nil
}

func checkMigrationStatus(database *gorm.DB, cfg *config.MigrationsConfig) error {
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	migrator, err := migrate.New(sqlDB, migrate.Config{
		MigrationsDir: cfg.Directory,
		Timeout:       time.Duration(cfg.Timeout) * time.Second,
		LockTimeout:   time.Duration(cfg.LockTimeout) * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	version, dirty, err := migrator.Version()
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database in dirty state at version %d", version)
	}

	slog.Info("Database schema", "version", version)
	return nil
}
