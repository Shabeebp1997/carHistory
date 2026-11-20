package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"carhistory/internal/config"
	"carhistory/internal/handlers"
	"carhistory/internal/middleware"
	"carhistory/internal/repository"
	jwtpkg "carhistory/pkg/jwt"
)

type Handlers struct {
	Auth           *handlers.AuthHandler
	User           *handlers.UserHandler
	Vehicle        *handlers.VehicleHandler
	Maintenance    *handlers.MaintenanceHandler
	ServiceRecord  *handlers.ServiceRecordHandler
	Insurance      *handlers.InsuranceHandler
	PUC            *handlers.PUCHandler
	Tyre           *handlers.TyreHandler
	Battery        *handlers.BatteryHandler
	Odometer       *handlers.OdometerHandler
	Reminder       *handlers.ReminderHandler
	Notification   *handlers.NotificationHandler
	Document       *handlers.DocumentHandler
	Expense        *handlers.ExpenseHandler
	Activity       *handlers.ActivityHandler
	Report         *handlers.ReportHandler
	Fuel           *handlers.FuelHandler
	Media          *handlers.MediaHandler
	Admin          *handlers.AdminHandler
}

func New(
	cfg *config.Config,
	jwtManager *jwtpkg.Manager,
	vehicleRepo *repository.VehicleRepository,
	userRepo *repository.UserRepository,
	h *Handlers,
) *gin.Engine {
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	origins := []string{cfg.App.BaseURL, "http://localhost:3000"}
	if len(cfg.App.CORSOrigins) > 0 {
		origins = append(origins, cfg.App.CORSOrigins...)
	}
	// In development, also allow the common Vite/Next.js dev ports.
	if cfg.Server.Env != "production" {
		origins = append(origins, "http://localhost:5173", "http://127.0.0.1:3000", "http://127.0.0.1:5173")
	}

	corsConfig := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}
	r.Use(cors.New(corsConfig))

	r.Static("/uploads", cfg.Storage.LocalPath)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	// ── Auth (public) ──────────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/logout", h.Auth.Logout)
		auth.POST("/refresh", h.Auth.RefreshToken)
		auth.POST("/forgot-password", h.Auth.ForgotPassword)
		auth.POST("/reset-password", h.Auth.ResetPassword)
	}

	// ── Protected routes ──────────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwtManager))

	// Users
	users := protected.Group("/users")
	{
		users.GET("/me", h.User.GetProfile)
		users.PUT("/me", h.User.UpdateProfile)
		users.PUT("/me/password", h.User.ChangePassword)
		users.POST("/me/avatar", h.User.UploadAvatar)
	}

	protected.GET("/vehicle-types", h.Vehicle.GetVehicleTypes)
	protected.GET("/maintenance-types", h.Maintenance.GetTypes)
	protected.GET("/dashboard", h.Report.UserDashboard)
	protected.GET("/notifications", h.Notification.List)

	invitations := protected.Group("/invitations")
	{
		invitations.GET("", h.Vehicle.GetPendingInvitations)
		invitations.POST("/:token/accept", h.Vehicle.AcceptInvitation)
		invitations.POST("/:token/reject", h.Vehicle.RejectInvitation)
	}

	// Vehicles
	vehicles := protected.Group("/vehicles")
	{
		vehicles.GET("", h.Vehicle.ListVehicles)
		vehicles.GET("/search", h.Vehicle.SearchVehicles)
		vehicles.POST("", h.Vehicle.CreateVehicle)
		vehicles.GET("/:vehicleId", h.Vehicle.GetVehicle)
		vehicles.PUT("/:vehicleId", middleware.VehicleAccessMiddleware(vehicleRepo), middleware.RequireRole("owner", "editor"), h.Vehicle.UpdateVehicle)
		vehicles.DELETE("/:vehicleId", middleware.VehicleAccessMiddleware(vehicleRepo), middleware.RequireRole("owner"), h.Vehicle.ArchiveVehicle)
	}

	// All vehicle sub-resources use vehicle access middleware
	vehicleSub := protected.Group("/vehicles/:vehicleId")
	vehicleSub.Use(middleware.VehicleAccessMiddleware(vehicleRepo))
	{
		// Members
		vehicleSub.GET("/members", h.Vehicle.GetMembers)
		vehicleSub.DELETE("/members/:userId", middleware.RequireRole("owner"), h.Vehicle.RemoveMember)
		vehicleSub.PUT("/members/:userId/role", middleware.RequireRole("owner"), h.Vehicle.UpdateMemberRole)

		// Invitations
		vehicleSub.GET("/invitations", middleware.RequireRole("owner"), h.Vehicle.GetInvitations)
		vehicleSub.POST("/invitations", middleware.RequireRole("owner"), h.Vehicle.InviteMember)

		// Maintenance
		vehicleSub.GET("/maintenance", h.Maintenance.List)
		vehicleSub.POST("/maintenance", middleware.RequireRole("owner", "editor"), h.Maintenance.Create)
		vehicleSub.GET("/maintenance/:id", h.Maintenance.Get)
		vehicleSub.PUT("/maintenance/:id", middleware.RequireRole("owner", "editor"), h.Maintenance.Update)
		vehicleSub.DELETE("/maintenance/:id", middleware.RequireRole("owner"), h.Maintenance.Delete)

		// Services
		vehicleSub.GET("/services", h.ServiceRecord.List)
		vehicleSub.POST("/services", middleware.RequireRole("owner", "editor"), h.ServiceRecord.Create)
		vehicleSub.GET("/services/:id", h.ServiceRecord.Get)
		vehicleSub.PUT("/services/:id", middleware.RequireRole("owner", "editor"), h.ServiceRecord.Update)
		vehicleSub.DELETE("/services/:id", middleware.RequireRole("owner"), h.ServiceRecord.Delete)

		// Insurance
		vehicleSub.GET("/insurance", h.Insurance.List)
		vehicleSub.POST("/insurance", middleware.RequireRole("owner", "editor"), h.Insurance.Create)
		vehicleSub.GET("/insurance/:id", h.Insurance.Get)
		vehicleSub.PUT("/insurance/:id", middleware.RequireRole("owner", "editor"), h.Insurance.Update)
		vehicleSub.DELETE("/insurance/:id", middleware.RequireRole("owner"), h.Insurance.Delete)

		// PUC
		vehicleSub.GET("/puc", h.PUC.List)
		vehicleSub.POST("/puc", middleware.RequireRole("owner", "editor"), h.PUC.Create)
		vehicleSub.GET("/puc/:id", h.PUC.Get)
		vehicleSub.PUT("/puc/:id", middleware.RequireRole("owner", "editor"), h.PUC.Update)
		vehicleSub.DELETE("/puc/:id", middleware.RequireRole("owner"), h.PUC.Delete)

		// Tyres
		vehicleSub.GET("/tyres", h.Tyre.List)
		vehicleSub.POST("/tyres", middleware.RequireRole("owner", "editor"), h.Tyre.Create)
		vehicleSub.GET("/tyres/:id", h.Tyre.Get)
		vehicleSub.PUT("/tyres/:id", middleware.RequireRole("owner", "editor"), h.Tyre.Update)
		vehicleSub.DELETE("/tyres/:id", middleware.RequireRole("owner"), h.Tyre.Delete)

		// Battery
		vehicleSub.GET("/batteries", h.Battery.List)
		vehicleSub.POST("/batteries", middleware.RequireRole("owner", "editor"), h.Battery.Create)
		vehicleSub.GET("/batteries/:id", h.Battery.Get)
		vehicleSub.PUT("/batteries/:id", middleware.RequireRole("owner", "editor"), h.Battery.Update)
		vehicleSub.DELETE("/batteries/:id", middleware.RequireRole("owner"), h.Battery.Delete)

		// Odometer
		vehicleSub.GET("/odometer", h.Odometer.List)
		vehicleSub.POST("/odometer", middleware.RequireRole("owner", "editor"), h.Odometer.Create)

		// Reminders
		vehicleSub.GET("/reminders", h.Reminder.List)
		vehicleSub.POST("/reminders", middleware.RequireRole("owner", "editor"), h.Reminder.Create)
		vehicleSub.GET("/reminders/:id", h.Reminder.Get)
		vehicleSub.PUT("/reminders/:id", middleware.RequireRole("owner", "editor"), h.Reminder.Update)
		vehicleSub.DELETE("/reminders/:id", middleware.RequireRole("owner"), h.Reminder.Delete)

		// Documents
		vehicleSub.GET("/documents", h.Document.List)
		vehicleSub.POST("/documents", middleware.RequireRole("owner", "editor"), h.Document.Upload)
		vehicleSub.GET("/documents/:id", h.Document.Get)
		vehicleSub.DELETE("/documents/:id", middleware.RequireRole("owner"), h.Document.Delete)

		// Expenses
		vehicleSub.GET("/expenses", h.Expense.List)
		vehicleSub.POST("/expenses", middleware.RequireRole("owner", "editor"), h.Expense.Create)
		vehicleSub.GET("/expenses/:id", h.Expense.Get)
		vehicleSub.PUT("/expenses/:id", middleware.RequireRole("owner", "editor"), h.Expense.Update)
		vehicleSub.DELETE("/expenses/:id", middleware.RequireRole("owner"), h.Expense.Delete)

		// Fuel / Charging logs
		vehicleSub.GET("/fuel", h.Fuel.List)
		vehicleSub.POST("/fuel", middleware.RequireRole("owner", "editor"), h.Fuel.Create)
		vehicleSub.GET("/fuel/summary", h.Fuel.Summary)
		vehicleSub.GET("/fuel/:id", h.Fuel.Get)
		vehicleSub.PUT("/fuel/:id", middleware.RequireRole("owner", "editor"), h.Fuel.Update)
		vehicleSub.DELETE("/fuel/:id", middleware.RequireRole("owner"), h.Fuel.Delete)

		// Activity
		vehicleSub.GET("/activity", h.Activity.List)

		// Reports
		vehicleSub.GET("/reports/summary", h.Report.VehicleSummary)
		vehicleSub.GET("/reports/expenses", h.Report.ExpenseReport)
	}

	// Media attachments (generic)
	media := protected.Group("/media")
	{
		media.POST("/upload", h.Media.Upload)
		media.GET("/:entityType/:entityId", h.Media.ListByEntity)
		media.DELETE("/:id", h.Media.Delete)
	}

	// Admin routes
	admin := protected.Group("/admin")
	admin.Use(middleware.AdminMiddleware(userRepo))
	{
		admin.GET("/analytics", h.Admin.Analytics)
		admin.GET("/admins", h.Admin.ListAdmins)
		admin.GET("/users", h.Admin.ListUsers)
		admin.GET("/users/:userId", h.Admin.GetUser)
		admin.PUT("/users/:userId/admin", h.Admin.SetAdminStatus)
		admin.PUT("/users/:userId/deactivate", h.Admin.DeactivateUser)
	}

	return r
}
