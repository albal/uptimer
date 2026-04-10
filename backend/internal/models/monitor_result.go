package models

import (
	"time"

	"github.com/google/uuid"
)

// MonitorResult stores the result of a single monitoring check.
type MonitorResult struct {
	ID             uuid.UUID `json:"id" db:"id"`
	MonitorID      uuid.UUID `json:"monitor_id" db:"monitor_id"`
	Status         string    `json:"status" db:"status"`
	ResponseTimeMs *int      `json:"response_time_ms,omitempty" db:"response_time_ms"`
	StatusCode     *int      `json:"status_code,omitempty" db:"status_code"`
	ErrorMessage   string    `json:"error_message,omitempty" db:"error_message"`
	Region         string    `json:"region,omitempty" db:"region"`
	CheckedAt      time.Time `json:"checked_at" db:"checked_at"`
}

// MonitorResultsPage holds a paginated list of monitor results.
type MonitorResultsPage struct {
	Results    []MonitorResult `json:"results"`
	TotalCount int             `json:"total_count"`
	HasMore    bool            `json:"has_more"`
	NextCursor string          `json:"next_cursor,omitempty"`
}
