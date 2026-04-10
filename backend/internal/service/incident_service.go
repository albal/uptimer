package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
)

// IncidentService handles incident business logic.
type IncidentService struct {
	incidentRepo     *repository.IncidentRepo
	monitorRepo      *repository.MonitorRepo
	alertContactRepo *repository.AlertContactRepo
}

// NewIncidentService creates a new IncidentService.
func NewIncidentService(incidentRepo *repository.IncidentRepo, monitorRepo *repository.MonitorRepo, alertContactRepo *repository.AlertContactRepo) *IncidentService {
	return &IncidentService{
		incidentRepo:     incidentRepo,
		monitorRepo:      monitorRepo,
		alertContactRepo: alertContactRepo,
	}
}

// OpenIncident creates a new incident for a monitor going down.
func (s *IncidentService) OpenIncident(ctx context.Context, monitorID uuid.UUID, reason string) (*models.Incident, error) {
	// Check if there's already an ongoing incident
	existing, err := s.incidentRepo.FindOngoingByMonitorID(ctx, monitorID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil // Already have an open incident
	}

	incident := &models.Incident{
		MonitorID: monitorID,
		Reason:    reason,
		Status:    models.IncidentOngoing,
	}

	if err := s.incidentRepo.Create(ctx, incident); err != nil {
		return nil, err
	}

	slog.Info("incident opened", "incident_id", incident.ID, "monitor_id", monitorID, "reason", reason)
	return incident, nil
}

// ResolveIncident marks an incident as resolved.
func (s *IncidentService) ResolveIncident(ctx context.Context, monitorID uuid.UUID) error {
	incident, err := s.incidentRepo.FindOngoingByMonitorID(ctx, monitorID)
	if err != nil {
		return err
	}
	if incident == nil {
		return nil // No ongoing incident to resolve
	}

	if err := s.incidentRepo.Resolve(ctx, incident.ID); err != nil {
		return err
	}

	slog.Info("incident resolved", "incident_id", incident.ID, "monitor_id", monitorID)
	return nil
}

// GetAlertContactsForMonitor returns the alert contacts that should be notified for a monitor's incidents.
func (s *IncidentService) GetAlertContactsForMonitor(ctx context.Context, monitorID uuid.UUID) ([]models.AlertContact, error) {
	return s.alertContactRepo.FindByMonitorID(ctx, monitorID)
}
