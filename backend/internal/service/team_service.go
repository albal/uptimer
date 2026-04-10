package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
)

// TeamService handles team business logic.
type TeamService struct {
	teamRepo *repository.TeamRepo
	userRepo *repository.UserRepo
}

// NewTeamService creates a new TeamService.
func NewTeamService(teamRepo *repository.TeamRepo, userRepo *repository.UserRepo) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

// GetTeam returns team details by ID.
func (s *TeamService) GetTeam(ctx context.Context, teamID uuid.UUID) (*models.Team, error) {
	return s.teamRepo.FindByID(ctx, teamID)
}

// ListMembers returns all members of the team.
func (s *TeamService) ListMembers(ctx context.Context, teamID uuid.UUID) ([]models.TeamMember, error) {
	return s.teamRepo.FindMembers(ctx, teamID)
}

// InviteMember invites a user to a team by email.
func (s *TeamService) InviteMember(ctx context.Context, teamID uuid.UUID, email string, role string) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil || team == nil {
		return fmt.Errorf("team not found")
	}

	// Check seat limit
	count, err := s.teamRepo.CountMembers(ctx, teamID)
	if err != nil {
		return err
	}
	if count >= team.MaxSeats {
		return fmt.Errorf("team seat limit reached (%d/%d)", count, team.MaxSeats)
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found — they must sign up first")
	}

	if role == "" {
		role = models.RoleMember
	}

	return s.teamRepo.AddMember(ctx, teamID, user.ID, role)
}

// RemoveMember removes a user from a team.
func (s *TeamService) RemoveMember(ctx context.Context, teamID, userID, callerID uuid.UUID) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil || team == nil {
		return fmt.Errorf("team not found")
	}

	// Basic authorization check
	if team.OwnerID != callerID {
		// In a real app, we'd check if the caller has Admin role
		// For now, only owner can remove members in this simplified logic
		return fmt.Errorf("unauthorized")
	}

	if team.OwnerID == userID {
		return fmt.Errorf("cannot remove team owner")
	}
	
	return s.teamRepo.RemoveMember(ctx, teamID, userID)
}
