package models

import (
	"time"

	"github.com/google/uuid"
)

// MonitorType constants
const (
	MonitorHTTP      = "http"
	MonitorPing      = "ping"
	MonitorPort      = "port"
	MonitorKeyword   = "keyword"
	MonitorAPI       = "api"
	MonitorUDP       = "udp"
	MonitorSSL       = "ssl"
	MonitorDNS       = "dns"
	MonitorDomain    = "domain"
	MonitorHeartbeat = "heartbeat"
)

// MonitorStatus constants
const (
	StatusUp       = "up"
	StatusDown     = "down"
	StatusDegraded = "degraded"
	StatusPaused   = "paused"
	StatusPending  = "pending"
)

// Monitor represents a monitoring target.
type Monitor struct {
	ID              uuid.UUID `json:"id" db:"id"`
	TeamID          uuid.UUID `json:"team_id" db:"team_id"`
	Name            string    `json:"name" db:"name"`
	Type            string    `json:"type" db:"type"`
	URL             string    `json:"url,omitempty" db:"url"`
	IPAddress       string    `json:"ip_address,omitempty" db:"ip_address"`
	Port            *int      `json:"port,omitempty" db:"port"`
	IntervalSeconds int       `json:"interval_seconds" db:"interval_seconds"`
	TimeoutSeconds  int       `json:"timeout_seconds" db:"timeout_seconds"`

	// HTTP options
	HTTPMethod         string            `json:"http_method,omitempty" db:"http_method"`
	HTTPHeaders        map[string]string `json:"http_headers,omitempty" db:"http_headers"`
	HTTPBody           string            `json:"http_body,omitempty" db:"http_body"`
	HTTPAuthType       string            `json:"http_auth_type,omitempty" db:"http_auth_type"`
	HTTPUsername        string            `json:"http_username,omitempty" db:"http_username"`
	HTTPPasswordEnc    string            `json:"-" db:"http_password_enc"`
	ExpectedStatusCodes []int            `json:"expected_status_codes,omitempty" db:"expected_status_codes"`
	FollowRedirects    bool              `json:"follow_redirects" db:"follow_redirects"`

	// Keyword options
	Keyword     string `json:"keyword,omitempty" db:"keyword"`
	KeywordType string `json:"keyword_type,omitempty" db:"keyword_type"`

	// API monitoring options
	APIAssertions []APIAssertion `json:"api_assertions,omitempty" db:"api_assertions"`

	// UDP options
	UDPData     string `json:"udp_data,omitempty" db:"udp_data"`
	UDPExpected string `json:"udp_expected,omitempty" db:"udp_expected"`

	// SSL options
	SSLExpiryReminder int `json:"ssl_expiry_reminder,omitempty" db:"ssl_expiry_reminder"`

	// DNS options
	DNSRecordType  string `json:"dns_record_type,omitempty" db:"dns_record_type"`
	DNSExpectedValue string `json:"dns_expected_value,omitempty" db:"dns_expected_value"`

	// Domain expiry options
	DomainExpiryReminder int `json:"domain_expiry_reminder,omitempty" db:"domain_expiry_reminder"`

	// Location-specific monitoring
	MonitoringRegions []string `json:"monitoring_regions,omitempty" db:"monitoring_regions"`

	// Slow response alert
	SlowThresholdMs *int `json:"slow_threshold_ms,omitempty" db:"slow_threshold_ms"`

	// Heartbeat
	HeartbeatToken    string     `json:"heartbeat_token,omitempty" db:"heartbeat_token"`
	HeartbeatGraceSec int        `json:"heartbeat_grace_sec,omitempty" db:"heartbeat_grace_sec"`
	HeartbeatLastPing *time.Time `json:"heartbeat_last_ping,omitempty" db:"heartbeat_last_ping"`

	// Status tracking
	Status           string     `json:"status" db:"status"`
	LastCheckedAt    *time.Time `json:"last_checked_at,omitempty" db:"last_checked_at"`
	LastResponseMs   *int       `json:"last_response_ms,omitempty" db:"last_response_ms"`
	UptimePercentage float64    `json:"uptime_percentage" db:"uptime_percentage"`
	TotalChecks      int64      `json:"total_checks" db:"total_checks"`
	TotalDowntimeSec int64      `json:"total_downtime_sec" db:"total_downtime_sec"`

	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Computed fields (not stored)
	AlertContacts []AlertContact `json:"alert_contacts,omitempty" db:"-"`
}

// APIAssertion defines an assertion for API monitoring.
type APIAssertion struct {
	Source     string `json:"source"`      // "status_code", "response_time", "header", "body"
	Property   string `json:"property"`    // JSONPath or header name
	Comparison string `json:"comparison"`  // "equals", "contains", "greater_than", "less_than"
	Value      string `json:"value"`       // expected value
}

// CreateMonitorRequest is the API request to create a monitor.
type CreateMonitorRequest struct {
	Name              string            `json:"name" validate:"required"`
	Type              string            `json:"type" validate:"required"`
	URL               string            `json:"url,omitempty"`
	IPAddress         string            `json:"ip_address,omitempty"`
	Port              *int              `json:"port,omitempty"`
	IntervalSeconds   int               `json:"interval_seconds"`
	TimeoutSeconds    int               `json:"timeout_seconds"`
	HTTPMethod        string            `json:"http_method,omitempty"`
	HTTPHeaders       map[string]string `json:"http_headers,omitempty"`
	HTTPBody          string            `json:"http_body,omitempty"`
	HTTPAuthType      string            `json:"http_auth_type,omitempty"`
	HTTPUsername       string            `json:"http_username,omitempty"`
	HTTPPassword       string           `json:"http_password,omitempty"`
	ExpectedStatusCodes []int           `json:"expected_status_codes,omitempty"`
	FollowRedirects   bool              `json:"follow_redirects"`
	Keyword           string            `json:"keyword,omitempty"`
	KeywordType       string            `json:"keyword_type,omitempty"`
	APIAssertions     []APIAssertion    `json:"api_assertions,omitempty"`
	UDPData           string            `json:"udp_data,omitempty"`
	UDPExpected       string            `json:"udp_expected,omitempty"`
	SSLExpiryReminder int               `json:"ssl_expiry_reminder,omitempty"`
	DNSRecordType     string            `json:"dns_record_type,omitempty"`
	DNSExpectedValue  string            `json:"dns_expected_value,omitempty"`
	DomainExpiryReminder int            `json:"domain_expiry_reminder,omitempty"`
	MonitoringRegions []string          `json:"monitoring_regions,omitempty"`
	SlowThresholdMs   *int              `json:"slow_threshold_ms,omitempty"`
	HeartbeatGraceSec int               `json:"heartbeat_grace_sec,omitempty"`
	AlertContactIDs   []uuid.UUID       `json:"alert_contact_ids,omitempty"`
}
