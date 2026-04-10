package models

import (
	"time"

	"github.com/google/uuid"
)

// StatusPage represents a public status page.
type StatusPage struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	TeamID              uuid.UUID `json:"team_id" db:"team_id"`
	Name                string    `json:"name" db:"name"`
	Slug                string    `json:"slug" db:"slug"`
	CustomDomain        string    `json:"custom_domain,omitempty" db:"custom_domain"`
	LogoURL             string    `json:"logo_url,omitempty" db:"logo_url"`
	PrimaryColor        string    `json:"primary_color" db:"primary_color"`
	IsPasswordProtected bool      `json:"is_password_protected" db:"is_password_protected"`
	PasswordHash        string    `json:"-" db:"password_hash"`
	HideFromSearch      bool      `json:"hide_from_search" db:"hide_from_search"`
	Announcement        string    `json:"announcement,omitempty" db:"announcement"`
	Language            string    `json:"language" db:"language"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`

	// Associated monitors
	Monitors []StatusPageMonitor `json:"monitors,omitempty" db:"-"`
}

// StatusPageMonitor links a monitor to a status page with ordering.
type StatusPageMonitor struct {
	StatusPageID uuid.UUID `json:"status_page_id" db:"status_page_id"`
	MonitorID    uuid.UUID `json:"monitor_id" db:"monitor_id"`
	SortOrder    int       `json:"sort_order" db:"sort_order"`
	Monitor      *Monitor  `json:"monitor,omitempty" db:"-"`
}
