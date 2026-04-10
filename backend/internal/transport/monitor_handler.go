package transport

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
	"github.com/albal/uptimer/internal/service"
)

// MonitorHandler handles monitor API endpoints.
type MonitorHandler struct {
	monitorService *service.MonitorService
	monitorRepo    *repository.MonitorRepo
}

// NewMonitorHandler creates a new MonitorHandler.
func NewMonitorHandler(monitorService *service.MonitorService, monitorRepo *repository.MonitorRepo) *MonitorHandler {
	return &MonitorHandler{monitorService: monitorService, monitorRepo: monitorRepo}
}

// List returns all monitors for the current team.
func (h *MonitorHandler) List(w http.ResponseWriter, r *http.Request) {
	teamID := GetTeamID(r.Context())
	monitors, err := h.monitorRepo.FindByTeamID(r.Context(), teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list monitors")
		return
	}
	if monitors == nil {
		monitors = []models.Monitor{}
	}
	writeJSON(w, http.StatusOK, monitors)
}

// Get returns a single monitor by ID.
func (h *MonitorHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	teamID := GetTeamID(r.Context())
	if monitor.TeamID != teamID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, monitor)
}

// Create creates a new monitor.
func (h *MonitorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMonitorRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Type == "" {
		writeError(w, http.StatusBadRequest, "name and type are required")
		return
	}

	teamID := GetTeamID(r.Context())
	userID := GetUserID(r.Context())

	monitor, err := h.monitorService.Create(r.Context(), teamID, userID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, monitor)
}

// Update updates a monitor.
func (h *MonitorHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid monitor ID")
		return
	}

	var req models.CreateMonitorRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	teamID := GetTeamID(r.Context())
	monitor, err := h.monitorService.Update(r.Context(), id, teamID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, monitor)
}

// Delete deletes a monitor.
func (h *MonitorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid monitor ID")
		return
	}

	teamID := GetTeamID(r.Context())
	if err := h.monitorService.Delete(r.Context(), id, teamID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "monitor deleted"})
}

// Pause pauses a monitor.
func (h *MonitorHandler) Pause(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid monitor ID")
		return
	}

	teamID := GetTeamID(r.Context())
	if err := h.monitorService.Pause(r.Context(), id, teamID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "monitor paused"})
}

// Resume resumes a paused monitor.
func (h *MonitorHandler) Resume(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid monitor ID")
		return
	}

	teamID := GetTeamID(r.Context())
	if err := h.monitorService.Resume(r.Context(), id, teamID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "monitor resumed"})
}

// GetResults returns monitoring check results for a monitor.
func (h *MonitorHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid monitor ID")
		return
	}

	// Verify ownership
	monitor, err := h.monitorRepo.FindByID(r.Context(), id)
	if err != nil || monitor == nil {
		writeError(w, http.StatusNotFound, "monitor not found")
		return
	}
	if monitor.TeamID != GetTeamID(r.Context()) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	limit := 100
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 500 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	results, total, err := h.monitorRepo.FindResults(r.Context(), id, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch results")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results":     results,
		"total_count": total,
		"has_more":    offset+limit < total,
	})
}

// Heartbeat handles incoming heartbeat pings.
func (h *MonitorHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "missing token")
		return
	}

	monitor, err := h.monitorRepo.FindByHeartbeatToken(r.Context(), token)
	if err != nil || monitor == nil {
		writeError(w, http.StatusNotFound, "heartbeat monitor not found")
		return
	}

	if err := h.monitorRepo.UpdateHeartbeatPing(r.Context(), monitor.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update heartbeat")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
