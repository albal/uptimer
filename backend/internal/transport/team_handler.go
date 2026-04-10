package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/repository"
	"github.com/albal/uptimer/internal/service"
)

type TeamHandler struct {
	teamService *service.TeamService
	teamRepo    *repository.TeamRepo
}

func NewTeamHandler(teamService *service.TeamService, teamRepo *repository.TeamRepo) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
		teamRepo:    teamRepo,
	}
}

func (h *TeamHandler) Get(w http.ResponseWriter, r *http.Request) {
	teamID := GetTeamID(r.Context())
	team, err := h.teamRepo.FindByID(r.Context(), teamID)
	if err != nil || team == nil {
		writeError(w, http.StatusNotFound, "team not found")
		return
	}
	writeJSON(w, http.StatusOK, team)
}

func (h *TeamHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	teamID := GetTeamID(r.Context())
	members, err := h.teamRepo.FindMembers(r.Context(), teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list members")
		return
	}
	writeJSON(w, http.StatusOK, members)
}

func (h *TeamHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := readJSON(r, &req); err != nil || req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	teamID := GetTeamID(r.Context())
	if err := h.teamService.InviteMember(r.Context(), teamID, req.Email, req.Role); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "member invited"})
}

func (h *TeamHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	teamID := GetTeamID(r.Context())
	callerID := GetUserID(r.Context())
	if err := h.teamService.RemoveMember(r.Context(), teamID, userID, callerID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}
