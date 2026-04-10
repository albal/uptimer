package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
)

// StatusPageService handles status page business logic.
type StatusPageService struct {
	statusPageRepo *repository.StatusPageRepo
}

// NewStatusPageService creates a new StatusPageService.
func NewStatusPageService(statusPageRepo *repository.StatusPageRepo) *StatusPageService {
	return &StatusPageService{statusPageRepo: statusPageRepo}
}

// Create creates a new status page.
func (s *StatusPageService) Create(ctx context.Context, sp *models.StatusPage) error {
	if sp.PrimaryColor == "" {
		sp.PrimaryColor = "#10B981"
	}
	if sp.Language == "" {
		sp.Language = "en"
	}
	return s.statusPageRepo.Create(ctx, sp)
}

// Update updates a status page.
func (s *StatusPageService) Update(ctx context.Context, sp *models.StatusPage, teamID uuid.UUID) error {
	existing, err := s.statusPageRepo.FindByID(ctx, sp.ID)
	if err != nil || existing == nil {
		return fmt.Errorf("status page not found")
	}
	if existing.TeamID != teamID {
		return fmt.Errorf("unauthorized")
	}
	return s.statusPageRepo.Update(ctx, sp)
}

// Delete deletes a status page.
func (s *StatusPageService) Delete(ctx context.Context, id uuid.UUID, teamID uuid.UUID) error {
	existing, err := s.statusPageRepo.FindByID(ctx, id)
	if err != nil || existing == nil {
		return fmt.Errorf("status page not found")
	}
	if existing.TeamID != teamID {
		return fmt.Errorf("unauthorized")
	}
	return s.statusPageRepo.Delete(ctx, id)
}
