package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	ServerPort    int    `envconfig:"SERVER_PORT" default:"8080"`
	ServerHost    string `envconfig:"SERVER_HOST" default:"0.0.0.0"`
	FrontendURL   string `envconfig:"FRONTEND_URL" default:"http://localhost:5173"`
	BaseURL       string `envconfig:"BASE_URL" default:"http://localhost:8080"`

	// Database
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// JWT
	JWTSecret          string `envconfig:"JWT_SECRET" required:"true"`
	JWTExpiryHours     int    `envconfig:"JWT_EXPIRY_HOURS" default:"24"`
	JWTRefreshDays     int    `envconfig:"JWT_REFRESH_DAYS" default:"30"`

	// OAuth - Google
	GoogleClientID     string `envconfig:"OAUTH_GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `envconfig:"OAUTH_GOOGLE_CLIENT_SECRET"`

	// OAuth - Microsoft
	MicrosoftClientID     string `envconfig:"OAUTH_MICROSOFT_CLIENT_ID"`
	MicrosoftClientSecret string `envconfig:"OAUTH_MICROSOFT_CLIENT_SECRET"`

	// OAuth - Apple
	AppleClientID   string `envconfig:"OAUTH_APPLE_CLIENT_ID"`
	AppleTeamID     string `envconfig:"OAUTH_APPLE_TEAM_ID"`
	AppleKeyID      string `envconfig:"OAUTH_APPLE_KEY_ID"`
	ApplePrivateKey string `envconfig:"OAUTH_APPLE_PRIVATE_KEY"`

	// SMTP (for email notifications)
	SMTPHost     string `envconfig:"SMTP_HOST"`
	SMTPPort     int    `envconfig:"SMTP_PORT" default:"587"`
	SMTPUsername string `envconfig:"SMTP_USERNAME"`
	SMTPPassword string `envconfig:"SMTP_PASSWORD"`
	SMTPFrom     string `envconfig:"SMTP_FROM" default:"noreply@uptimer.local"`

	// Monitoring Engine
	MonitorWorkers    int `envconfig:"MONITOR_WORKERS" default:"100"`
	DefaultInterval   int `envconfig:"DEFAULT_INTERVAL" default:"300"`
	MinInterval       int `envconfig:"MIN_INTERVAL" default:"30"`

	// Team Settings
	DefaultMaxSeats   int `envconfig:"DEFAULT_MAX_SEATS" default:"5"`
	DefaultMaxMonitors int `envconfig:"DEFAULT_MAX_MONITORS" default:"1000"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return &cfg, nil
}
