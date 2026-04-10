package models

import (
	"time"

	"github.com/google/uuid"
)

// AlertContactType constants
const (
	AlertEmail      = "email"
	AlertSMS        = "sms"
	AlertSlack      = "slack"
	AlertTeams      = "teams"
	AlertDiscord    = "discord"
	AlertTelegram   = "telegram"
	AlertPagerDuty  = "pagerduty"
	AlertWebhook    = "webhook"
	AlertGoogleChat = "googlechat"
	AlertPushbullet = "pushbullet"
	AlertPushover   = "pushover"
	AlertMattermost = "mattermost"
	AlertZapier     = "zapier"
)

// AlertContact represents a notification channel.
type AlertContact struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	TeamID    uuid.UUID              `json:"team_id" db:"team_id"`
	Type      string                 `json:"type" db:"type"`
	Name      string                 `json:"name" db:"name"`
	Value     string                 `json:"value" db:"value"`
	Config    map[string]interface{} `json:"config,omitempty" db:"config"`
	IsActive  bool                   `json:"is_active" db:"is_active"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// MonitorAlertContact links a monitor to an alert contact with threshold config.
type MonitorAlertContact struct {
	MonitorID       uuid.UUID `json:"monitor_id" db:"monitor_id"`
	AlertContactID  uuid.UUID `json:"alert_contact_id" db:"alert_contact_id"`
	ThresholdSeconds int      `json:"threshold_seconds" db:"threshold_seconds"`
}
