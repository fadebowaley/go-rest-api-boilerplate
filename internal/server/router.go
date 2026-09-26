package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/fadebowaley/applico/internal/applicant"
	"github.com/fadebowaley/applico/internal/audit"
	"github.com/fadebowaley/applico/internal/auth"
	"github.com/fadebowaley/applico/internal/config"
	"github.com/fadebowaley/applico/internal/disbursement"
	"github.com/fadebowaley/applico/internal/errors"
	"github.com/fadebowaley/applico/internal/document"
	"github.com/fadebowaley/applico/internal/form"
	"github.com/fadebowaley/applico/internal/grant"
	"github.com/fadebowaley/applico/internal/project"
	"github.com/fadebowaley/applico/internal/health"
	"github.com/fadebowaley/applico/internal/middleware"
	"github.com/fadebowaley/applico/internal/permission"
	"github.com/fadebowaley/applico/internal/program"
	"github.com/fadebowaley/applico/internal/stats"
	"github.com/fadebowaley/applico/internal/tenant"
	"github.com/fadebowaley/applico/internal/user"
	"github.com/fadebowaley/applico/internal/workflow"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(
	userHandler *user.Handler,
	authService auth.Service,
	cfg *config.Config,
	db *gorm.DB,
	permissionHandler *permission.Handler,
	auditHandler *audit.Handler,
	grantHandler *grant.Handler,
	documentHandler *document.Handler,
	tenantHandler *tenant.Handler,
	tenantSvc tenant.Service,
	programHandler *program.Handler,
	applicantHandler *applicant.Handler,
	workflowHandler *workflow.Handler,
	disbursementHandler *disbursement.Handler,
	projectHandler *project.Handler,
	statsHandler *stats.Handler,
	formHandler *form.Handler,
) *gin.Engine {
	router := gin.New()

	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	skipPaths := config.GetSkipPaths(cfg.App.Environment)
	loggerConfig := middleware.NewLoggerConfig(
		cfg.Logging.GetLogLevel(),
		skipPaths,
	)
	router.Use(middleware.Logger(loggerConfig))
	router.Use(errors.ErrorHandler())
	router.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	router.Use(cors.New(corsConfig))

	var checkers []health.Checker
	if cfg.Health.DatabaseCheckEnabled {
		dbChecker := health.NewDatabaseChecker(db)
		checkers = append(checkers, dbChecker)
	}
	healthService := health.NewService(checkers, cfg.App.Version, cfg.App.Environment)
	healthHandler := health.NewHandler(healthService)

	router.GET("/health", healthHandler.Health)
	router.GET("/health/live", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	rlCfg := cfg.Ratelimit
	if rlCfg.Enabled {
		router.Use(
			middleware.NewRateLimitMiddleware(
				rlCfg.Window,
				rlCfg.Requests,
				func(c *gin.Context) string {
					ip := c.ClientIP()
					if ip == "" {
						ip = c.GetHeader("X-Forwarded-For")
						if ip == "" {
							ip = c.GetHeader("X-Real-IP")
						}
						if ip == "" {
							ip = "unknown"
						}
					}
					return ip
				},
				nil,
			),
		)
	}

	v1 := router.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", userHandler.Register)
			authGroup.POST("/login", userHandler.Login)
			authGroup.POST("/refresh", userHandler.RefreshToken)
			authGroup.POST("/logout", auth.AuthMiddleware(authService), userHandler.Logout)
			authGroup.GET("/me", auth.AuthMiddleware(authService), userHandler.GetMe)
			authGroup.PUT("/me", auth.AuthMiddleware(authService), userHandler.UpdateMe)
			authGroup.GET("/me/permissions", auth.AuthMiddleware(authService), permissionHandler.GetMyPermissions)
			authGroup.GET("/me/tenants", auth.AuthMiddleware(authService), tenantHandler.GetMyTenants)
			authGroup.POST("/change-password", auth.AuthMiddleware(authService), userHandler.ChangePassword)
			authGroup.POST("/forgot-password", userHandler.ForgotPassword)
			authGroup.POST("/reset-password", userHandler.ResetPassword)
		}

		usersGroup := v1.Group("/users")
		usersGroup.Use(auth.AuthMiddleware(authService))
		{
			usersGroup.GET("/:id", userHandler.GetUser)
			usersGroup.PUT("/:id", userHandler.UpdateUser)
			usersGroup.DELETE("/:id", userHandler.DeleteUser)
		}

		grantsGroup := v1.Group("/grants")
		grantsGroup.Use(auth.AuthMiddleware(authService), middleware.RequireTenant(tenantSvc))
		{
			grantsGroup.POST("", grantHandler.CreateGrant)
			grantsGroup.GET("", grantHandler.ListMyGrants)
			grantsGroup.GET("/:id", grantHandler.GetGrant)
			grantsGroup.PUT("/:id", grantHandler.UpdateGrant)
			grantsGroup.DELETE("/:id", grantHandler.DeleteGrant)
			grantsGroup.POST("/:id/submit", grantHandler.SubmitGrant)

			grantsGroup.POST("/:id/documents", documentHandler.Upload)
			grantsGroup.GET("/:id/documents", documentHandler.List)
			grantsGroup.GET("/:id/documents/:docId", documentHandler.GetByID)
			grantsGroup.PUT("/:id/documents/:docId", documentHandler.Replace)
			grantsGroup.DELETE("/:id/documents/:docId", documentHandler.Delete)
			grantsGroup.GET("/:id/documents/:docId/download", documentHandler.Download)
			grantsGroup.POST("/:id/documents/:docId/verify", middleware.RequireAnyPermission("document:verify", "admin:document:verify"), documentHandler.Verify)
			grantsGroup.POST("/:id/documents/:docId/reject", middleware.RequireAnyPermission("document:reject", "admin:document:reject"), documentHandler.Reject)

			grantsGroup.GET("/:id/workflow", workflowHandler.GetWorkflow)
			grantsGroup.POST("/:id/workflow/approve", middleware.RequirePermission("grant:approve"), workflowHandler.Approve)
			grantsGroup.POST("/:id/workflow/reject", middleware.RequirePermission("grant:reject"), workflowHandler.Reject)
			grantsGroup.POST("/:id/workflow/return", middleware.RequirePermission("grant:return"), workflowHandler.Return)
			grantsGroup.GET("/:id/workflow/history", workflowHandler.GetHistory)

			grantsGroup.GET("/:id/disbursements", disbursementHandler.ListByGrant)
			grantsGroup.POST("/:id/disbursements", middleware.RequirePermission("disbursement:create"), disbursementHandler.Create)

			grantsGroup.GET("/:id/updates", projectHandler.ListByGrant)
			grantsGroup.POST("/:id/updates", projectHandler.Create)
		}

		tenantUsers := v1.Group("/tenants")
		tenantUsers.Use(auth.AuthMiddleware(authService))
		{
			tenantUsers.GET("/:id", tenantHandler.GetMyTenant)
			tenantUsers.POST("/:id/invite", tenantHandler.InviteUser)
			tenantUsers.GET("/:id/users", tenantHandler.GetUsers)
			tenantUsers.GET("/:id/roles", tenantHandler.ListRoles)
			tenantUsers.POST("/:id/roles", tenantHandler.CreateRole)
			tenantUsers.GET("/:id/roles/:roleId", tenantHandler.GetRole)
			tenantUsers.PUT("/:id/roles/:roleId", tenantHandler.UpdateRole)
			tenantUsers.DELETE("/:id/roles/:roleId", tenantHandler.DeleteRole)
		}

		tenantScoped := v1.Group("")
		tenantScoped.Use(auth.AuthMiddleware(authService), middleware.RequireTenant(tenantSvc))
		{
			tenantScoped.GET("/programs", programHandler.List)
			tenantScoped.POST("/programs", programHandler.Create)
			tenantScoped.GET("/programs/:id", programHandler.GetByID)
			tenantScoped.PUT("/programs/:id", programHandler.Update)
			tenantScoped.DELETE("/programs/:id", programHandler.Delete)

			tenantScoped.GET("/applicants", applicantHandler.List)
			tenantScoped.POST("/applicants", applicantHandler.Create)
			tenantScoped.GET("/applicants/:id", applicantHandler.GetByID)
			tenantScoped.PUT("/applicants/:id", applicantHandler.Update)
			tenantScoped.DELETE("/applicants/:id", applicantHandler.Delete)

			tenantScoped.GET("/workflows", workflowHandler.ListTemplates)
			tenantScoped.POST("/workflows", workflowHandler.CreateTemplate)
			tenantScoped.GET("/workflows/:id", workflowHandler.GetTemplate)
			tenantScoped.PUT("/workflows/:id", workflowHandler.UpdateTemplate)
			tenantScoped.DELETE("/workflows/:id", workflowHandler.DeleteTemplate)
			tenantScoped.GET("/workflows/:id/steps", workflowHandler.ListSteps)
			tenantScoped.POST("/workflows/:id/steps", workflowHandler.CreateStep)
			tenantScoped.GET("/steps/:stepId", workflowHandler.GetStep)
			tenantScoped.PUT("/steps/:stepId", workflowHandler.UpdateStep)
			tenantScoped.DELETE("/steps/:stepId", workflowHandler.DeleteStep)

			tenantScoped.GET("/workflow/pending", workflowHandler.ListPendingApprovals)

			tenantScoped.GET("/disbursements/:id", disbursementHandler.GetByID)
			tenantScoped.PUT("/disbursements/:id", middleware.RequirePermission("disbursement:update"), disbursementHandler.Update)
			tenantScoped.DELETE("/disbursements/:id", middleware.RequirePermission("disbursement:delete"), disbursementHandler.Delete)

			tenantScoped.GET("/project-updates/:id", projectHandler.GetByID)
			tenantScoped.PUT("/project-updates/:id", projectHandler.Update)
			tenantScoped.DELETE("/project-updates/:id", projectHandler.Delete)

			tenantScoped.GET("/forms/templates", formHandler.List)
			tenantScoped.POST("/forms/templates", formHandler.Create)
			tenantScoped.GET("/forms/templates/:id", formHandler.GetByID)
			tenantScoped.PUT("/forms/templates/:id", formHandler.Update)
			tenantScoped.DELETE("/forms/templates/:id", formHandler.Delete)
			tenantScoped.POST("/forms/templates/:id/publish", formHandler.Publish)
			tenantScoped.GET("/forms/templates/:id/versions", formHandler.ListVersions)

			tenantScoped.GET("/dashboard/summary", statsHandler.Summary)
			tenantScoped.GET("/dashboard/finance-summary", statsHandler.FinanceSummary)
			tenantScoped.GET("/dashboard/audit-summary", statsHandler.AuditSummary)
			tenantScoped.GET("/dashboard/program-stats/:id", statsHandler.ProgramStats)
			tenantScoped.GET("/dashboard/workflow-stats", statsHandler.WorkflowStats)
		}

		adminGroup := v1.Group("/admin")
		adminGroup.Use(auth.AuthMiddleware(authService), middleware.RequireAnyPermission("admin:tenant:list", "admin:user:list", "admin:permission:list"))
		{
			adminGroup.POST("/tenants", tenantHandler.Create)
			adminGroup.GET("/tenants", tenantHandler.List)
			adminGroup.GET("/tenants/:id", tenantHandler.GetByID)
			adminGroup.PUT("/tenants/:id", tenantHandler.Update)
			adminGroup.DELETE("/tenants/:id", tenantHandler.Delete)
			adminGroup.POST("/tenants/:id/users", tenantHandler.AddUser)
			adminGroup.GET("/tenants/:id/users", tenantHandler.GetUsers)
			adminGroup.DELETE("/tenants/:id/users/:userId", tenantHandler.RemoveUser)

			adminGroup.GET("/users", userHandler.ListUsers)
			adminGroup.GET("/users/:id", userHandler.GetUser)
			adminGroup.PUT("/users/:id", userHandler.UpdateUser)
			adminGroup.DELETE("/users/:id", userHandler.DeleteUser)
			adminGroup.POST("/users/:id/roles", userHandler.AssignUserRole)
			adminGroup.GET("/users/:id/roles", userHandler.GetUserRoles)
			adminGroup.DELETE("/users/:id/roles/:roleName", userHandler.RemoveUserRole)

			permissionGroup := adminGroup.Group("/permissions")
			{
				permissionGroup.POST("", permissionHandler.CreatePermission)
				permissionGroup.GET("", permissionHandler.ListPermissions)
				permissionGroup.DELETE("/:id", permissionHandler.DeletePermission)
			}

			rolePermGroup := adminGroup.Group("/roles/:roleId/permissions")
			{
				rolePermGroup.GET("", permissionHandler.GetPermissionsForRole)
				rolePermGroup.PUT("", permissionHandler.AssignPermissionsToRole)
			}

			auditLogGroup := adminGroup.Group("/audit-logs")
			{
				auditLogGroup.GET("", auditHandler.ListAuditLogs)
				auditLogGroup.POST("", auditHandler.LogActionFromRequest)
			}
		}

		adminTenantScoped := v1.Group("/admin")
		adminTenantScoped.Use(auth.AuthMiddleware(authService), middleware.RequireAnyPermission("admin:grant:list", "admin:document:verify"), middleware.RequireTenant(tenantSvc))
		{
			adminTenantScoped.GET("/grants", grantHandler.ListAllGrants)
			adminTenantScoped.GET("/grants/:id", grantHandler.GetGrantAsAdmin)

			adminTenantScoped.POST("/documents/:docId/verify", documentHandler.Verify)
			adminTenantScoped.POST("/documents/:docId/reject", documentHandler.Reject)
		}
	}

	return router
}
