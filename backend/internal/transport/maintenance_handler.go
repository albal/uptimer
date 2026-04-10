package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
)

type MaintenanceHandler struct {
	repo *repository.MaintenanceWindowRepo
}

func NewMaintenanceHandler(repo *repository.MaintenanceWindowRepo) *MaintenanceHandler {
	return &MaintenanceHandler{repo: repo}
}

func (h *MaintenanceHandler) List(w http.ResponseWriter, r *http.Request) {
	teamID := GetTeamID(r.Context())
	windows, err := h.repo.FindByTeamID(r.Context(), teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list maintenance windows")
		return
	}
	if windows == nil {
		windows = []models.MaintenanceWindow{}
	}
	writeJSON(w, http.StatusOK, windows)
}

func (h *MaintenanceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var mw models.MaintenanceWindow
	if err := readJSON(r, &mw); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mw.TeamID = GetTeamID(r.Context())
	if err := h.repo.Create(r.Context(), &mw); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create maintenance window")
		return
	}

	writeJSON(w, http.StatusCreated, mw)
}

func (h *MaintenanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	mw, err := h.repo.FindByID(r.Context(), id)
	if err != nil || mw == nil {
		writeError(w, http.StatusNotFound, "maintenance window not found")
		return
	}
	if mw.TeamID != GetTeamID(r.Context()) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, mw)
}

func (h *MaintenanceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	teamID := GetTeamID(r.Context())
	mw, err := h.repo.FindByID(r.Context(), id)
	if err != nil || mw == nil || mw.TeamID != teamID {
		writeError(w, http.StatusNotFound, "maintenance window not found")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete maintenance window")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "maintenance window deleted"})
}
