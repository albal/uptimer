package models

import (
	"time"

	"github.com/google/uuid"
)

// IncidentStatus constants
const (
	IncidentOngoing      = "ongoing"
	IncidentResolved     = "resolved"
	IncidentAcknowledged = "acknowledged"
)

// Incident represents a downtime incident for a monitor.
type Incident struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	MonitorID       uuid.UUID  `json:"monitor_id" db:"monitor_id"`
	StartedAt       time.Time  `json:"started_at" db:"started_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" db:"duration_seconds"`
	Reason          string     `json:"reason,omitempty" db:"reason"`
	RootCause       string     `json:"root_cause,omitempty" db:"root_cause"`
	Status          string     `json:"status" db:"status"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`

	// Computed
	MonitorName string `json:"monitor_name,omitempty" db:"-"`
}
