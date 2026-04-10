package transport

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
	"github.com/albal/uptimer/internal/service"
)

// APIHandler handles the public API v1 and API key management.
type APIHandler struct {
	monitorRepo       *repository.MonitorRepo
	incidentRepo      *repository.IncidentRepo
	alertContactRepo  *repository.AlertContactRepo
	statusPageRepo    *repository.StatusPageRepo
	maintenanceRepo   *repository.MaintenanceWindowRepo
	authService       *service.AuthService
	apiKeyRepo        *repository.APIKeyRepo
}

// NewAPIHandler creates a new APIHandler.
func NewAPIHandler(
	mr *repository.MonitorRepo,
	ir *repository.IncidentRepo,
	acr *repository.AlertContactRepo,
	spr *repository.StatusPageRepo,
	mwr *repository.MaintenanceWindowRepo,
	as *service.AuthService,
	akr *repository.APIKeyRepo,
) *APIHandler {
	return &APIHandler{
		monitorRepo:      mr,
		incidentRepo:     ir,
		alertContactRepo: acr,
		statusPageRepo:   spr,
		maintenanceRepo:  mwr,
		authService:      as,
		apiKeyRepo:       akr,
	}
}

// APIKeyAuthMiddleware authenticates requests using API keys.
func (h *APIHandler) APIKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if apiKey == "" {
			writeError(w, http.StatusUnauthorized, "API key required")
			return
		}

		teamID, err := h.authService.ValidateAPIKey(r.Context(), h.apiKeyRepo, apiKey)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid API key")
			return
		}

		ctx := context.WithValue(r.Context(), ContextTeamID, teamID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ListKeys lists API keys for the current team.
func (h *APIHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	teamID := GetTeamID(r.Context())
	keys, err := h.apiKeyRepo.FindByTeamID(r.Context(), teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list API keys")
		return
	}
	if keys == nil {
		keys = []models.APIKey{}
	}
	writeJSON(w, http.StatusOK, keys)
}

// CreateKey creates a new API key.
func (h *APIHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}
	if err := readJSON(r, &req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	if len(req.Scopes) == 0 {
		req.Scopes = []string{"read"}
	}

	// Generate a random API key
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate key")
		return
	}
	rawKey := "utm_" + hex.EncodeToString(rawBytes)
	prefix := rawKey[:8]

	// Hash the key for storage
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	key := &models.APIKey{
		TeamID:  GetTeamID(r.Context()),
		Name:    req.Name,
		KeyHash: keyHash,
		Prefix:  prefix,
		Scopes:  req.Scopes,
	}

	if err := h.apiKeyRepo.Create(r.Context(), key); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create API key")
		return
	}

	// Return the raw key only this once
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":     key.ID,
		"name":   key.Name,
		"key":    rawKey,
		"prefix": prefix,
		"scopes": key.Scopes,
	})
}

// DeleteKey deletes an API key.
func (h *APIHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	teamID := GetTeamID(r.Context())
	if err := h.apiKeyRepo.Delete(r.Context(), id, teamID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete API key")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "API key deleted"})
}

// Public API v1 endpoints

// ListMonitors lists all monitors via API.
func (h *APIHandler) ListMonitors(w http.ResponseWriter, r *http.Request) {
	teamID := GetTeamID(r.Context())
	monitors, err := h.monitorRepo.FindByTeamID(r.Context(), teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list monitors")
		return
	}
	if monitors == nil {
		monitors = []models.Monitor{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"monitors": monitors,
		"total":    len(monitors),
	})
}

// GetMonitor gets a single monitor via API.
func (h *APIHandler) GetMonitor(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid monitor ID")
		return
	}

	monitor, err := h.monitorRepo.FindByID(r.Context(), id)
	if err != nil || monitor == nil {
		writeError(w, http.StatusNotFound, "monitor not found")
		return
	}
	if monitor.TeamID != GetTeamID(r.Context()) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"monitor": monitor})
}

// CreateMonitor creates a monitor via API.
func (h *APIHandler) CreateMonitor(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "use the dashboard API")
}

// UpdateMonitor updates a monitor via API.
func (h *APIHandler) UpdateMonitor(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "use the dashboard API")
}

// DeleteMonitor deletes a monitor via API.
func (h *APIHandler) DeleteMonitor(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "use the dashboard API")
}

// ListIncidents lists incidents via API.
func (h *APIHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	teamID := GetTeamID(r.Context())
	incidents, total, err := h.incidentRepo.FindByTeamID(r.Context(), teamID, 50, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list incidents")
		return
	}
	if incidents == nil {
		incidents = []models.Incident{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"incidents": incidents,
		"total":     total,
	})
}
