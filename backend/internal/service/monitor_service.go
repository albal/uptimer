package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/config"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
)

// MonitorService handles monitor business logic.
type MonitorService struct {
	cfg              *config.Config
	monitorRepo      *repository.MonitorRepo
	alertContactRepo *repository.AlertContactRepo
	teamRepo         *repository.TeamRepo
}

// NewMonitorService creates a new MonitorService.
func NewMonitorService(cfg *config.Config, monitorRepo *repository.MonitorRepo, alertContactRepo *repository.AlertContactRepo, teamRepo *repository.TeamRepo) *MonitorService {
	return &MonitorService{
		cfg:              cfg,
		monitorRepo:      monitorRepo,
		alertContactRepo: alertContactRepo,
		teamRepo:         teamRepo,
	}
}

// Create creates a new monitor from a request.
func (s *MonitorService) Create(ctx context.Context, teamID uuid.UUID, userID uuid.UUID, req models.CreateMonitorRequest) (*models.Monitor, error) {
	// Validate team can add more monitors
	count, err := s.monitorRepo.CountByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, fmt.Errorf("team not found")
	}
	if count >= team.MaxMonitors {
		return nil, fmt.Errorf("monitor limit reached (%d/%d)", count, team.MaxMonitors)
	}

	// Set defaults
	interval := req.IntervalSeconds
	if interval == 0 {
		interval = s.cfg.DefaultInterval
	}
	if interval < s.cfg.MinInterval {
		interval = s.cfg.MinInterval
	}

	timeout := req.TimeoutSeconds
	if timeout == 0 {
		timeout = 30
	}

	// Validate monitor type
	switch req.Type {
	case models.MonitorHTTP, models.MonitorPing, models.MonitorPort, models.MonitorKeyword,
		models.MonitorAPI, models.MonitorUDP, models.MonitorSSL, models.MonitorDNS,
		models.MonitorDomain, models.MonitorHeartbeat:
		// valid
	default:
		return nil, fmt.Errorf("invalid monitor type: %s", req.Type)
	}

	monitor := &models.Monitor{
		TeamID:              teamID,
		Name:                req.Name,
		Type:                req.Type,
		URL:                 req.URL,
		IPAddress:           req.IPAddress,
		Port:                req.Port,
		IntervalSeconds:     interval,
		TimeoutSeconds:      timeout,
		HTTPMethod:          req.HTTPMethod,
		HTTPHeaders:         req.HTTPHeaders,
		HTTPBody:            req.HTTPBody,
		HTTPAuthType:        req.HTTPAuthType,
		HTTPUsername:         req.HTTPUsername,
		ExpectedStatusCodes: req.ExpectedStatusCodes,
		FollowRedirects:     req.FollowRedirects,
		Keyword:             req.Keyword,
		KeywordType:         req.KeywordType,
		APIAssertions:       req.APIAssertions,
		UDPData:             req.UDPData,
		UDPExpected:         req.UDPExpected,
		SSLExpiryReminder:   req.SSLExpiryReminder,
		DNSRecordType:       req.DNSRecordType,
		DNSExpectedValue:    req.DNSExpectedValue,
		DomainExpiryReminder: req.DomainExpiryReminder,
		MonitoringRegions:   req.MonitoringRegions,
		SlowThresholdMs:     req.SlowThresholdMs,
		HeartbeatGraceSec:   req.HeartbeatGraceSec,
		Status:              models.StatusPending,
		CreatedBy:           userID,
	}

	// Generate heartbeat token if needed
	if req.Type == models.MonitorHeartbeat {
		token, err := generateToken(32)
		if err != nil {
			return nil, err
		}
		monitor.HeartbeatToken = token
	}

	// Set default HTTP method
	if monitor.HTTPMethod == "" && (req.Type == models.MonitorHTTP || req.Type == models.MonitorKeyword || req.Type == models.MonitorAPI) {
		monitor.HTTPMethod = "GET"
	}

	// Set default expected status codes
	if len(monitor.ExpectedStatusCodes) == 0 {
		monitor.ExpectedStatusCodes = []int{200}
	}

	if err := s.monitorRepo.Create(ctx, monitor); err != nil {
		return nil, fmt.Errorf("creating monitor: %w", err)
	}

	// Link alert contacts
	for _, acID := range req.AlertContactIDs {
		if err := s.alertContactRepo.LinkToMonitor(ctx, monitor.ID, acID, 0); err != nil {
			slog.Warn("failed to link alert contact", "monitor_id", monitor.ID, "alert_contact_id", acID, "error", err)
		}
	}

	slog.Info("monitor created", "id", monitor.ID, "name", monitor.Name, "type", monitor.Type)
	return monitor, nil
}

// Update updates a monitor.
func (s *MonitorService) Update(ctx context.Context, id uuid.UUID, teamID uuid.UUID, req models.CreateMonitorRequest) (*models.Monitor, error) {
	monitor, err := s.monitorRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if monitor == nil {
		return nil, fmt.Errorf("monitor not found")
	}
	if monitor.TeamID != teamID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Update fields
	monitor.Name = req.Name
	monitor.Type = req.Type
	monitor.URL = req.URL
	monitor.IPAddress = req.IPAddress
	monitor.Port = req.Port
	monitor.IntervalSeconds = req.IntervalSeconds
	monitor.TimeoutSeconds = req.TimeoutSeconds
	monitor.HTTPMethod = req.HTTPMethod
	monitor.HTTPHeaders = req.HTTPHeaders
	monitor.HTTPBody = req.HTTPBody
	monitor.HTTPAuthType = req.HTTPAuthType
	monitor.HTTPUsername = req.HTTPUsername
	monitor.ExpectedStatusCodes = req.ExpectedStatusCodes
	monitor.FollowRedirects = req.FollowRedirects
	monitor.Keyword = req.Keyword
	monitor.KeywordType = req.KeywordType
	monitor.APIAssertions = req.APIAssertions
	monitor.UDPData = req.UDPData
	monitor.UDPExpected = req.UDPExpected
	monitor.SSLExpiryReminder = req.SSLExpiryReminder
	monitor.DNSRecordType = req.DNSRecordType
	monitor.DNSExpectedValue = req.DNSExpectedValue
	monitor.DomainExpiryReminder = req.DomainExpiryReminder
	monitor.MonitoringRegions = req.MonitoringRegions
	monitor.SlowThresholdMs = req.SlowThresholdMs
	monitor.HeartbeatGraceSec = req.HeartbeatGraceSec

	if err := s.monitorRepo.Update(ctx, monitor); err != nil {
		return nil, err
	}

	return monitor, nil
}

// Delete deletes a monitor.
func (s *MonitorService) Delete(ctx context.Context, id uuid.UUID, teamID uuid.UUID) error {
	monitor, err := s.monitorRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if monitor == nil {
		return fmt.Errorf("monitor not found")
	}
	if monitor.TeamID != teamID {
		return fmt.Errorf("unauthorized")
	}

	return s.monitorRepo.Delete(ctx, id)
}

// Pause pauses a monitor.
func (s *MonitorService) Pause(ctx context.Context, id uuid.UUID, teamID uuid.UUID) error {
	monitor, err := s.monitorRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if monitor == nil || monitor.TeamID != teamID {
		return fmt.Errorf("monitor not found")
	}
	return s.monitorRepo.UpdateStatus(ctx, id, models.StatusPaused, 0)
}

// Resume resumes a paused monitor.
func (s *MonitorService) Resume(ctx context.Context, id uuid.UUID, teamID uuid.UUID) error {
	monitor, err := s.monitorRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if monitor == nil || monitor.TeamID != teamID {
		return fmt.Errorf("monitor not found")
	}
	return s.monitorRepo.UpdateStatus(ctx, id, models.StatusPending, 0)
}

func generateToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
