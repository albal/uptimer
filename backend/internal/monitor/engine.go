package monitor

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
	"github.com/albal/uptimer/internal/service"
)

// Engine is the core monitoring engine that schedules and runs checks.
type Engine struct {
	monitorRepo      *repository.MonitorRepo
	incidentService  *service.IncidentService
	notifService     *service.NotificationService
	maintenanceRepo  *repository.MaintenanceWindowRepo
	alertContactRepo *repository.AlertContactRepo
	checkers         map[string]Checker
	workers          int
	jobs             chan *models.Monitor
	quit             chan struct{}
	wg               sync.WaitGroup

	// Track next check times
	mu        sync.RWMutex
	schedules map[uuid.UUID]time.Time
}

// NewEngine creates a new monitoring engine.
func NewEngine(
	monitorRepo *repository.MonitorRepo,
	incidentService *service.IncidentService,
	notifService *service.NotificationService,
	maintenanceRepo *repository.MaintenanceWindowRepo,
	alertContactRepo *repository.AlertContactRepo,
	workers int,
) *Engine {
	e := &Engine{
		monitorRepo:      monitorRepo,
		incidentService:  incidentService,
		notifService:     notifService,
		maintenanceRepo:  maintenanceRepo,
		alertContactRepo: alertContactRepo,
		workers:          workers,
		jobs:             make(chan *models.Monitor, 1000),
		quit:             make(chan struct{}),
		schedules:        make(map[uuid.UUID]time.Time),
	}

	// Register all checkers
	e.checkers = map[string]Checker{
		models.MonitorHTTP:      &HTTPChecker{},
		models.MonitorPing:      &PingChecker{},
		models.MonitorPort:      &PortChecker{},
		models.MonitorUDP:       &UDPChecker{},
		models.MonitorKeyword:   &KeywordChecker{},
		models.MonitorAPI:       &APIChecker{},
		models.MonitorSSL:       &SSLChecker{},
		models.MonitorDNS:       &DNSChecker{},
		models.MonitorDomain:    &DomainChecker{},
		models.MonitorHeartbeat: &HeartbeatChecker{},
	}

	return e
}

// Start begins the monitoring engine.
func (e *Engine) Start(ctx context.Context) {
	slog.Info("starting monitoring engine", "workers", e.workers)

	// Start worker pool
	for i := 0; i < e.workers; i++ {
		e.wg.Add(1)
		go e.worker(ctx, i)
	}

	// Start scheduler
	e.wg.Add(1)
	go e.scheduler(ctx)

	slog.Info("monitoring engine started")
}

// Stop gracefully stops the monitoring engine.
func (e *Engine) Stop() {
	slog.Info("stopping monitoring engine")
	close(e.quit)
	e.wg.Wait()
	slog.Info("monitoring engine stopped")
}

// scheduler periodically scans for monitors that need checking.
func (e *Engine) scheduler(ctx context.Context) {
	defer e.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.quit:
			return
		case <-ticker.C:
			e.scheduleChecks(ctx)
		}
	}
}

// scheduleChecks finds monitors due for checking and enqueues them.
func (e *Engine) scheduleChecks(ctx context.Context) {
	monitors, err := e.monitorRepo.FindAllActive(ctx)
	if err != nil {
		slog.Error("failed to fetch active monitors", "error", err)
		return
	}

	now := time.Now()
	for i := range monitors {
		m := &monitors[i]
		if m.Status == models.StatusPaused {
			continue
		}

		e.mu.RLock()
		nextCheck, exists := e.schedules[m.ID]
		e.mu.RUnlock()

		if !exists || now.After(nextCheck) {
			// Schedule next check
			e.mu.Lock()
			e.schedules[m.ID] = now.Add(time.Duration(m.IntervalSeconds) * time.Second)
			e.mu.Unlock()

			select {
			case e.jobs <- m:
			default:
				slog.Warn("job queue full, skipping monitor", "monitor_id", m.ID)
			}
		}
	}
}

// worker processes monitoring jobs from the queue.
func (e *Engine) worker(ctx context.Context, id int) {
	defer e.wg.Done()

	for {
		select {
		case <-e.quit:
			return
		case monitor := <-e.jobs:
			e.checkMonitor(ctx, monitor)
		}
	}
}

// checkMonitor performs a single check on a monitor.
func (e *Engine) checkMonitor(ctx context.Context, monitor *models.Monitor) {
	// Check if in maintenance window
	inMaintenance, err := e.maintenanceRepo.IsMonitorInMaintenance(ctx, monitor.ID)
	if err != nil {
		slog.Error("failed to check maintenance window", "error", err)
	}
	if inMaintenance {
		return
	}

	checker, ok := e.checkers[monitor.Type]
	if !ok {
		slog.Error("no checker for monitor type", "type", monitor.Type)
		return
	}

	// Create a timeout context for the check
	checkCtx, cancel := context.WithTimeout(ctx, time.Duration(monitor.TimeoutSeconds)*time.Second)
	defer cancel()

	result := checker.Check(checkCtx, monitor)
	responseMs := int(result.ResponseTime.Milliseconds())

	// Determine status
	status := result.Status
	if status == models.StatusUp && monitor.SlowThresholdMs != nil && responseMs > *monitor.SlowThresholdMs {
		status = models.StatusDegraded
	}

	// Store result
	monitorResult := &models.MonitorResult{
		MonitorID:    monitor.ID,
		Status:       status,
		ResponseTimeMs: &responseMs,
		Region:       result.Region,
		CheckedAt:    time.Now(),
	}
	if result.StatusCode > 0 {
		monitorResult.StatusCode = &result.StatusCode
	}
	if result.Error != nil {
		errMsg := result.Error.Error()
		monitorResult.ErrorMessage = errMsg
	}

	if err := e.monitorRepo.InsertResult(ctx, monitorResult); err != nil {
		slog.Error("failed to store check result", "monitor_id", monitor.ID, "error", err)
	}

	// Update monitor status
	if err := e.monitorRepo.UpdateStatus(ctx, monitor.ID, status, responseMs); err != nil {
		slog.Error("failed to update monitor status", "monitor_id", monitor.ID, "error", err)
	}

	// Update uptime percentage
	if err := e.monitorRepo.UpdateUptimePercentage(ctx, monitor.ID); err != nil {
		slog.Error("failed to update uptime percentage", "monitor_id", monitor.ID, "error", err)
	}

	// Handle incident management
	e.handleIncidents(ctx, monitor, status, result)
}

// handleIncidents manages creating and resolving incidents based on check results.
func (e *Engine) handleIncidents(ctx context.Context, monitor *models.Monitor, status string, result CheckResult) {
	contacts, err := e.alertContactRepo.FindByMonitorID(ctx, monitor.ID)
	if err != nil {
		slog.Error("failed to fetch alert contacts", "error", err)
		return
	}

	previousStatus := monitor.Status

	switch {
	case status == models.StatusDown && previousStatus != models.StatusDown:
		// Monitor went down — open incident and notify
		reason := "Monitor is down"
		if result.Error != nil {
			reason = result.Error.Error()
		}
		incident, err := e.incidentService.OpenIncident(ctx, monitor.ID, reason)
		if err != nil {
			slog.Error("failed to open incident", "error", err)
			return
		}
		e.notifService.NotifyDown(ctx, monitor, incident, contacts)

	case status == models.StatusUp && previousStatus == models.StatusDown:
		// Monitor recovered — resolve incident and notify
		if err := e.incidentService.ResolveIncident(ctx, monitor.ID); err != nil {
			slog.Error("failed to resolve incident", "error", err)
		}
		e.notifService.NotifyUp(ctx, monitor, contacts)

	case status == models.StatusDegraded && previousStatus != models.StatusDegraded:
		// Monitor is degraded (slow response)
		responseMs := int(result.ResponseTime.Milliseconds())
		e.notifService.NotifyDegraded(ctx, monitor, responseMs, contacts)
	}
}

// RemoveSchedule removes a monitor from the schedule (when deleted/paused).
func (e *Engine) RemoveSchedule(monitorID uuid.UUID) {
	e.mu.Lock()
	delete(e.schedules, monitorID)
	e.mu.Unlock()
}
