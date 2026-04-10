package transport

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
)

// IncidentHandler handles incident API endpoints.
type IncidentHandler struct {
	incidentRepo *repository.IncidentRepo
}

// NewIncidentHandler creates a new IncidentHandler.
func NewIncidentHandler(incidentRepo *repository.IncidentRepo) *IncidentHandler {
	return &IncidentHandler{incidentRepo: incidentRepo}
}

// List returns all incidents for the current team.
func (h *IncidentHandler) List(w http.ResponseWriter, r *http.Request) {
	teamID := GetTeamID(r.Context())

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	incidents, total, err := h.incidentRepo.FindByTeamID(r.Context(), teamID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list incidents")
		return
	}
	if incidents == nil {
		incidents = []models.Incident{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"incidents":   incidents,
		"total_count": total,
		"has_more":    offset+limit < total,
	})
}

// Get returns a single incident.
func (h *IncidentHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid incident ID")
		return
	}

	writeError(w, http.StatusNotImplemented, "use list endpoint with filters")
}
