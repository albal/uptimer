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

	"github.com/albal/uptimer/internal/config"
	"github.com/albal/uptimer/internal/database"
	"github.com/albal/uptimer/internal/monitor"
	"github.com/albal/uptimer/internal/repository"
	"github.com/albal/uptimer/internal/service"
	"github.com/albal/uptimer/internal/transport"
)

func main() {
	// Set up structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	slog.Info("starting Uptimer", "version", "1.0.0")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Connect to database
	ctx := context.Background()
	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepo(db.Pool)
	teamRepo := repository.NewTeamRepo(db.Pool)
	monitorRepo := repository.NewMonitorRepo(db.Pool)
	incidentRepo := repository.NewIncidentRepo(db.Pool)
	alertContactRepo := repository.NewAlertContactRepo(db.Pool)
	statusPageRepo := repository.NewStatusPageRepo(db.Pool)
	maintenanceWindowRepo := repository.NewMaintenanceWindowRepo(db.Pool)
	apiKeyRepo := repository.NewAPIKeyRepo(db.Pool)

	// Initialize services
	authService := service.NewAuthService(cfg, userRepo, teamRepo)
	monitorService := service.NewMonitorService(cfg, monitorRepo, alertContactRepo, teamRepo)
	incidentService := service.NewIncidentService(incidentRepo, monitorRepo, alertContactRepo)
	notifService := service.NewNotificationService()
	statusPageService := service.NewStatusPageService(statusPageRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)

	// Initialize monitoring engine
	engine := monitor.NewEngine(
		monitorRepo,
		incidentService,
		notifService,
		maintenanceWindowRepo,
		alertContactRepo,
		cfg.MonitorWorkers,
	)

	// Start monitoring engine
	engineCtx, engineCancel := context.WithCancel(ctx)
	engine.Start(engineCtx)

	// Set up HTTP router
	router := transport.NewRouter(
		cfg,
		authService,
		monitorService,
		incidentService,
		notifService,
		statusPageService,
		teamService,
		monitorRepo,
		incidentRepo,
		alertContactRepo,
		statusPageRepo,
		maintenanceWindowRepo,
		teamRepo,
		apiKeyRepo,
		userRepo,
	)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("HTTP server starting", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("received shutdown signal", "signal", sig)

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	// Stop monitoring engine
	engineCancel()
	engine.Stop()

	// Stop HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	}

	slog.Info("Uptimer stopped")
}
