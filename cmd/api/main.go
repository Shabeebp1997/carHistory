package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"carhistory/internal/config"
	"carhistory/internal/database"
	"carhistory/internal/handlers"
	"carhistory/internal/repository"
	"carhistory/internal/router"
	"carhistory/internal/services"
	"carhistory/internal/worker"
	"carhistory/migrations"
	jwtpkg "carhistory/pkg/jwt"
)

func main() {
	// ── Config ──────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ── Database ─────────────────────────────────────────────
	db, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// ── Auto-migrate (creates tables if they do not exist) ───
	if err := database.AutoMigrate(db, migrations.Schema); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	// ── Seed admin user from env (idempotent) ─────────────────
	if err := database.SeedAdmin(db, cfg.App.AdminEmail, cfg.App.AdminPassword); err != nil {
		log.Printf("Warning: admin seed failed: %v", err)
	}

	// ── JWT Manager ──────────────────────────────────────────
	jwtManager := jwtpkg.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiryMinutes,
		cfg.JWT.RefreshTokenExpiryDays,
	)

	// ── Repositories ─────────────────────────────────────────
	userRepo         := repository.NewUserRepository(db)
	vehicleRepo      := repository.NewVehicleRepository(db)
	maintenanceRepo  := repository.NewMaintenanceRepository(db)
	serviceRepo      := repository.NewServiceRepository(db)
	insuranceRepo    := repository.NewInsuranceRepository(db)
	pucRepo          := repository.NewPUCRepository(db)
	tyreRepo         := repository.NewTyreRepository(db)
	batteryRepo      := repository.NewBatteryRepository(db)
	odometerRepo     := repository.NewOdometerRepository(db)
	reminderRepo     := repository.NewReminderRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	documentRepo     := repository.NewDocumentRepository(db)
	expenseRepo      := repository.NewExpenseRepository(db)
	activityRepo     := repository.NewActivityRepository(db)
	fuelRepo         := repository.NewFuelRepository(db)
	mediaRepo        := repository.NewMediaRepository(db)
	adminRepo        := repository.NewAdminRepository(db)

	// ── Services ─────────────────────────────────────────────
	emailSvc   := services.NewEmailService(cfg.SMTP)
	smsSvc     := services.NewSMSService(cfg.SMS)
	authSvc    := services.NewAuthService(userRepo, jwtManager, cfg, emailSvc)
	userSvc    := services.NewUserService(userRepo)
	vehicleSvc := services.NewVehicleService(vehicleRepo, activityRepo, userRepo, emailSvc, cfg)

	// ── Handlers ─────────────────────────────────────────────
	h := &router.Handlers{
		Auth:          handlers.NewAuthHandler(authSvc, userRepo, jwtManager),
		User:          handlers.NewUserHandler(userSvc, userRepo, cfg),
		Vehicle:       handlers.NewVehicleHandler(vehicleSvc, vehicleRepo),
		Maintenance:   handlers.NewMaintenanceHandler(maintenanceRepo, activityRepo),
		ServiceRecord: handlers.NewServiceRecordHandler(serviceRepo, activityRepo),
		Insurance:     handlers.NewInsuranceHandler(insuranceRepo, activityRepo),
		PUC:           handlers.NewPUCHandler(pucRepo, activityRepo),
		Tyre:          handlers.NewTyreHandler(tyreRepo, activityRepo),
		Battery:       handlers.NewBatteryHandler(batteryRepo, activityRepo),
		Odometer:      handlers.NewOdometerHandler(odometerRepo, vehicleRepo, activityRepo),
		Reminder:      handlers.NewReminderHandler(reminderRepo, activityRepo),
		Notification:  handlers.NewNotificationHandler(notificationRepo),
		Document:      handlers.NewDocumentHandler(documentRepo, activityRepo, cfg),
		Expense:       handlers.NewExpenseHandler(expenseRepo, activityRepo),
		Activity:      handlers.NewActivityHandler(activityRepo),
		Report: handlers.NewReportHandler(
			vehicleRepo,
			maintenanceRepo,
			expenseRepo,
			insuranceRepo,
			pucRepo,
			serviceRepo,
			fuelRepo,
			reminderRepo,
			activityRepo,
			odometerRepo,
		),
		Fuel:  handlers.NewFuelHandler(fuelRepo),
		Media: handlers.NewMediaHandler(mediaRepo, cfg),
		Admin: handlers.NewAdminHandler(adminRepo, userRepo),
	}

	// ── Router ───────────────────────────────────────────────
	r := router.New(cfg, jwtManager, vehicleRepo, userRepo, h)

	// ── Background Worker ─────────────────────────────────────
	reminderWorker := worker.NewReminderWorker(
		reminderRepo,
		notificationRepo,
		vehicleRepo,
		userRepo,
		emailSvc,
		smsSvc,
		cfg.App.ReminderCronSchedule,
	)
	reminderWorker.Start()
	defer reminderWorker.Stop()

	// ── HTTP Server ───────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s (env: %s)", cfg.Server.Port, cfg.Server.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// ── Graceful Shutdown ─────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
