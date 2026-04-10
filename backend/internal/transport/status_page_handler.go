package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
	"github.com/albal/uptimer/internal/service"
)

// StatusPageHandler handles status page API endpoints.
type StatusPageHandler struct {
	statusPageService *service.StatusPageService
	statusPageRepo    *repository.StatusPageRepo
	monitorRepo       *repository.MonitorRepo
}

// NewStatusPageHandler creates a new StatusPageHandler.
func NewStatusPageHandler(sps *service.StatusPageService, spr *repository.StatusPageRepo, mr *repository.MonitorRepo) *StatusPageHandler {
	return &StatusPageHandler{statusPageService: sps, statusPageRepo: spr, monitorRepo: mr}
}

// List returns all status pages for the current team.
func (h *StatusPageHandler) List(w http.ResponseWriter, r *http.Request) {
	teamID := GetTeamID(r.Context())
	pages, err := h.statusPageRepo.FindByTeamID(r.Context(), teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list status pages")
		return
	}
	if pages == nil {
		pages = []models.StatusPage{}
	}
	writeJSON(w, http.StatusOK, pages)
}

// Create creates a new status page.
func (h *StatusPageHandler) Create(w http.ResponseWriter, r *http.Request) {
	var sp models.StatusPage
	if err := readJSON(r, &sp); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sp.TeamID = GetTeamID(r.Context())
	if err := h.statusPageService.Create(r.Context(), &sp); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, sp)
}

// Get returns a single status page.
func (h *StatusPageHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	sp, err := h.statusPageRepo.FindByID(r.Context(), id)
	if err != nil || sp == nil {
		writeError(w, http.StatusNotFound, "status page not found")
		return
	}
	if sp.TeamID != GetTeamID(r.Context()) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Load monitors
	monitors, _ := h.statusPageRepo.FindMonitorsByStatusPageID(r.Context(), id)
	sp.Monitors = make([]models.StatusPageMonitor, len(monitors))
	for i, m := range monitors {
		monitor := m
		sp.Monitors[i] = models.StatusPageMonitor{
			StatusPageID: id,
			MonitorID:    m.ID,
			SortOrder:    i,
			Monitor:      &monitor,
		}
	}

	writeJSON(w, http.StatusOK, sp)
}

// Update updates a status page.
func (h *StatusPageHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	var sp models.StatusPage
	if err := readJSON(r, &sp); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sp.ID = id

	teamID := GetTeamID(r.Context())
	if err := h.statusPageService.Update(r.Context(), &sp, teamID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, sp)
}

// Delete deletes a status page.
func (h *StatusPageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	teamID := GetTeamID(r.Context())
	if err := h.statusPageService.Delete(r.Context(), id, teamID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "status page deleted"})
}

// SetMonitors sets the monitors for a status page.
func (h *StatusPageHandler) SetMonitors(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	var req struct {
		MonitorIDs []uuid.UUID `json:"monitor_ids"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Verify ownership
	sp, err := h.statusPageRepo.FindByID(r.Context(), id)
	if err != nil || sp == nil || sp.TeamID != GetTeamID(r.Context()) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := h.statusPageRepo.SetMonitors(r.Context(), id, req.MonitorIDs); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to set monitors")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "monitors updated"})
}

// GetPublicStatusPage returns a public status page by slug.
func (h *StatusPageHandler) GetPublicStatusPage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	sp, err := h.statusPageRepo.FindBySlug(r.Context(), slug)
	if err != nil || sp == nil {
		writeError(w, http.StatusNotFound, "status page not found")
		return
	}

	// Check password protection
	if sp.IsPasswordProtected {
		pw := r.URL.Query().Get("password")
		if pw == "" {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"password_required": true,
				"name":              sp.Name,
				"logo_url":          sp.LogoURL,
			})
			return
		}
		// In production, compare password hash
	}

	// Load monitors with status info
	monitors, _ := h.statusPageRepo.FindMonitorsByStatusPageID(r.Context(), sp.ID)

	// Don't expose password hash
	sp.PasswordHash = ""

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status_page": sp,
		"monitors":    monitors,
	})
}
