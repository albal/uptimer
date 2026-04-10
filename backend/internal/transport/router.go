package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/albal/uptimer/internal/config"
	"github.com/albal/uptimer/internal/repository"
	"github.com/albal/uptimer/internal/service"
)

// NewRouter creates and configures the HTTP router with all routes.
func NewRouter(
	cfg *config.Config,
	authService *service.AuthService,
	monitorService *service.MonitorService,
	incidentService *service.IncidentService,
	notifService *service.NotificationService,
	statusPageService *service.StatusPageService,
	teamService *service.TeamService,
	monitorRepo *repository.MonitorRepo,
	incidentRepo *repository.IncidentRepo,
	alertContactRepo *repository.AlertContactRepo,
	statusPageRepo *repository.StatusPageRepo,
	maintenanceWindowRepo *repository.MaintenanceWindowRepo,
	teamRepo *repository.TeamRepo,
	apiKeyRepo *repository.APIKeyRepo,
	userRepo *repository.UserRepo,
) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Compress(5))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL, "http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Create handlers
	authHandler := NewAuthHandler(authService, teamRepo, cfg)
	monitorHandler := NewMonitorHandler(monitorService, monitorRepo)
	incidentHandler := NewIncidentHandler(incidentRepo)
	alertContactHandler := NewAlertContactHandler(alertContactRepo)
	statusPageHandler := NewStatusPageHandler(statusPageService, statusPageRepo, monitorRepo)
	maintenanceHandler := NewMaintenanceHandler(maintenanceWindowRepo)
	teamHandler := NewTeamHandler(teamService, teamRepo)
	apiHandler := NewAPIHandler(monitorRepo, incidentRepo, alertContactRepo, statusPageRepo, maintenanceWindowRepo, authService, apiKeyRepo)

	// Auth routes (no auth required)
	r.Route("/api/auth", func(r chi.Router) {
		r.Get("/providers", authHandler.GetProviders)
		r.Get("/google", authHandler.GoogleLogin)
		r.Get("/google/callback", authHandler.GoogleCallback)
		r.Get("/microsoft", authHandler.MicrosoftLogin)
		r.Get("/microsoft/callback", authHandler.MicrosoftCallback)
		r.Get("/apple", authHandler.AppleLogin)
		r.Post("/apple/callback", authHandler.AppleCallback)
		r.Get("/me", authHandler.WithAuth(authHandler.GetMe))
		r.Post("/logout", authHandler.Logout)
	})

	// Heartbeat endpoint (no auth, uses token)
	r.Get("/api/heartbeat/{token}", monitorHandler.Heartbeat)

	// Protected API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)

		// Monitors
		r.Route("/monitors", func(r chi.Router) {
			r.Get("/", monitorHandler.List)
			r.Post("/", monitorHandler.Create)
			r.Get("/{id}", monitorHandler.Get)
			r.Put("/{id}", monitorHandler.Update)
			r.Delete("/{id}", monitorHandler.Delete)
			r.Post("/{id}/pause", monitorHandler.Pause)
			r.Post("/{id}/resume", monitorHandler.Resume)
			r.Get("/{id}/results", monitorHandler.GetResults)
		})

		// Incidents
		r.Route("/incidents", func(r chi.Router) {
			r.Get("/", incidentHandler.List)
			r.Get("/{id}", incidentHandler.Get)
		})

		// Alert Contacts
		r.Route("/alert-contacts", func(r chi.Router) {
			r.Get("/", alertContactHandler.List)
			r.Post("/", alertContactHandler.Create)
			r.Get("/{id}", alertContactHandler.Get)
			r.Put("/{id}", alertContactHandler.Update)
			r.Delete("/{id}", alertContactHandler.Delete)
		})

		// Status Pages
		r.Route("/status-pages", func(r chi.Router) {
			r.Get("/", statusPageHandler.List)
			r.Post("/", statusPageHandler.Create)
			r.Get("/{id}", statusPageHandler.Get)
			r.Put("/{id}", statusPageHandler.Update)
			r.Delete("/{id}", statusPageHandler.Delete)
			r.Put("/{id}/monitors", statusPageHandler.SetMonitors)
		})

		// Maintenance Windows
		r.Route("/maintenance-windows", func(r chi.Router) {
			r.Get("/", maintenanceHandler.List)
			r.Post("/", maintenanceHandler.Create)
			r.Get("/{id}", maintenanceHandler.Get)
			r.Delete("/{id}", maintenanceHandler.Delete)
		})

		// Team
		r.Route("/team", func(r chi.Router) {
			r.Get("/", teamHandler.Get)
			r.Get("/members", teamHandler.ListMembers)
			r.Post("/members", teamHandler.InviteMember)
			r.Delete("/members/{userId}", teamHandler.RemoveMember)
		})

		// API Keys
		r.Route("/api-keys", func(r chi.Router) {
			r.Get("/", apiHandler.ListKeys)
			r.Post("/", apiHandler.CreateKey)
			r.Delete("/{id}", apiHandler.DeleteKey)
		})
	})

	// Public Status Page (no auth)
	r.Get("/api/status/{slug}", statusPageHandler.GetPublicStatusPage)

	// Public API v1 (API key auth)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(apiHandler.APIKeyAuthMiddleware)
		r.Get("/monitors", apiHandler.ListMonitors)
		r.Get("/monitors/{id}", apiHandler.GetMonitor)
		r.Post("/monitors", apiHandler.CreateMonitor)
		r.Put("/monitors/{id}", apiHandler.UpdateMonitor)
		r.Delete("/monitors/{id}", apiHandler.DeleteMonitor)
		r.Get("/incidents", apiHandler.ListIncidents)
	})

	// Serve frontend static files in production
	staticDir := http.Dir("./static")
	fileServer := http.FileServer(staticDir)
	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		// If it's an API route or heartbeat, let those handlers handle it (they've already been matched or not)
		// But chi matches top-down, so this catch-all is only hit if nothing else matched.

		// Check if file exists
		f, err := staticDir.Open(path)
		if err != nil {
			// File doesn't exist, serve index.html
			http.ServeFile(w, req, "./static/index.html")
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, req)
	})

	return r
}
