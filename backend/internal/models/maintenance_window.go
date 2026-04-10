package models

import (
	"time"

	"github.com/google/uuid"
)

// MaintenanceWindow represents a planned maintenance period.
type MaintenanceWindow struct {
	ID             uuid.UUID `json:"id" db:"id"`
	TeamID         uuid.UUID `json:"team_id" db:"team_id"`
	Name           string    `json:"name" db:"name"`
	StartTime      time.Time `json:"start_time" db:"start_time"`
	EndTime        time.Time `json:"end_time" db:"end_time"`
	Recurring      bool      `json:"recurring" db:"recurring"`
	RecurrenceRule string    `json:"recurrence_rule,omitempty" db:"recurrence_rule"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`

	// Associated monitor IDs
	MonitorIDs []uuid.UUID `json:"monitor_ids,omitempty" db:"-"`
}

// APIKey represents a REST API access key.
type APIKey struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	TeamID    uuid.UUID  `json:"team_id" db:"team_id"`
	Name      string     `json:"name" db:"name"`
	KeyHash   string     `json:"-" db:"key_hash"`
	Prefix    string     `json:"prefix" db:"prefix"`
	Scopes    []string   `json:"scopes" db:"scopes"`
	LastUsed  *time.Time `json:"last_used,omitempty" db:"last_used"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// NotificationLog records sent notifications for audit.
type NotificationLog struct {
	ID             uuid.UUID `json:"id" db:"id"`
	IncidentID     uuid.UUID `json:"incident_id" db:"incident_id"`
	AlertContactID uuid.UUID `json:"alert_contact_id" db:"alert_contact_id"`
	Type           string    `json:"type" db:"type"`
	Status         string    `json:"status" db:"status"`
	ErrorMessage   string    `json:"error_message,omitempty" db:"error_message"`
	SentAt         time.Time `json:"sent_at" db:"sent_at"`
}
