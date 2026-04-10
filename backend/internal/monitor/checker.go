package monitor

import (
	"context"
	"time"

	"github.com/albal/uptimer/internal/models"
)

// CheckResult holds the result of a single monitoring check.
type CheckResult struct {
	Status       string
	ResponseTime time.Duration
	StatusCode   int
	Error        error
	Region       string
}

// Checker defines the interface for all monitor type implementations.
type Checker interface {
	// Check performs a monitoring check and returns the result.
	Check(ctx context.Context, monitor *models.Monitor) CheckResult
	// Type returns the monitor type this checker handles.
	Type() string
}
