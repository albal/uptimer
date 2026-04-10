package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/models"
)

type mockRepo struct {
	results []*models.CheckResult
}

func (r *mockRepo) CreateResult(ctx context.Context, res *models.CheckResult) error {
	r.results = append(r.results, res)
	return nil
}

func (r *mockRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.MonitorStatus, lastResponseMs int) error {
	return nil
}

func TestEngine_Schedule(t *testing.T) {
	// Minimal setup to test if engine starts and can process a task
	repo := &mockRepo{}
	engine := NewEngine(nil, nil, repo, nil) // passing nil for notifiers/incidents for simplicity
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	monitor := &models.Monitor{
		ID:              uuid.New(),
		Type:            models.MonitorHTTP,
		URL:             "http://example.com",
		IntervalSeconds: 1,
		Status:          models.StatusPending,
	}

	// This is a unit test, we won't run the full Wait loop, 
	// but we can test the check logic
	res := engine.checkMonitor(ctx, monitor)
	if res == nil {
		t.Fatal("expected check result, got nil")
	}
	
	if res.MonitorID != monitor.ID {
		t.Errorf("expected monitor ID %s, got %s", monitor.ID, res.MonitorID)
	}
}
