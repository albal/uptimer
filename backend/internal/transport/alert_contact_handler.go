package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
)

type AlertContactHandler struct {
	repo *repository.AlertContactRepo
}

func NewAlertContactHandler(repo *repository.AlertContactRepo) *AlertContactHandler {
	return &AlertContactHandler{repo: repo}
}

func (h *AlertContactHandler) List(w http.ResponseWriter, r *http.Request) {
	teamID := GetTeamID(r.Context())
	contacts, err := h.repo.FindByTeamID(r.Context(), teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alert contacts")
		return
	}
	if contacts == nil {
		contacts = []models.AlertContact{}
	}
	writeJSON(w, http.StatusOK, contacts)
}

func (h *AlertContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	var ac models.AlertContact
	if err := readJSON(r, &ac); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ac.TeamID = GetTeamID(r.Context())
	if err := h.repo.Create(r.Context(), &ac); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create alert contact")
		return
	}

	writeJSON(w, http.StatusCreated, ac)
}

func (h *AlertContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	ac, err := h.repo.FindByID(r.Context(), id)
	if err != nil || ac == nil {
		writeError(w, http.StatusNotFound, "alert contact not found")
		return
	}
	if ac.TeamID != GetTeamID(r.Context()) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, ac)
}

func (h *AlertContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	var ac models.AlertContact
	if err := readJSON(r, &ac); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ac.ID = id
	ac.TeamID = GetTeamID(r.Context())

	if err := h.repo.Update(r.Context(), &ac); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update alert contact")
		return
	}

	writeJSON(w, http.StatusOK, ac)
}

func (h *AlertContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	teamID := GetTeamID(r.Context())
	ac, err := h.repo.FindByID(r.Context(), id)
	if err != nil || ac == nil || ac.TeamID != teamID {
		writeError(w, http.StatusNotFound, "alert contact not found")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete alert contact")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "alert contact deleted"})
}
