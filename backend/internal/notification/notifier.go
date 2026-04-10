package notification

import (
	"context"
	"time"

	"github.com/albal/uptimer/internal/models"
)

// EventType represents the type of notification event.
type EventType string

const (
	EventDown     EventType = "down"
	EventUp       EventType = "up"
	EventDegraded EventType = "degraded"
	EventSSL      EventType = "ssl_expiring"
	EventDomain   EventType = "domain_expiring"
)

// Event holds the data for a notification.
type Event struct {
	Type        EventType
	MonitorName string
	MonitorURL  string
	MonitorType string
	Reason      string
	IncidentID  string
	StartedAt   time.Time
	ResponseMs  int
}

// Notifier defines the interface for sending notifications.
type Notifier interface {
	// Notify sends a notification to the given alert contact.
	Notify(ctx context.Context, contact models.AlertContact, event Event) error
}

// FormatMessage creates a human-readable message from an event.
func FormatMessage(event Event) string {
	switch event.Type {
	case EventDown:
		return "🔴 **" + event.MonitorName + "** is DOWN\n" +
			"URL: " + event.MonitorURL + "\n" +
			"Reason: " + event.Reason + "\n" +
			"Started: " + event.StartedAt.Format(time.RFC3339)
	case EventUp:
		return "🟢 **" + event.MonitorName + "** is UP\n" +
			"URL: " + event.MonitorURL + "\n" +
			"Monitor has recovered."
	case EventDegraded:
		return "🟡 **" + event.MonitorName + "** is DEGRADED\n" +
			"URL: " + event.MonitorURL + "\n" +
			"Response time: " + formatMs(event.ResponseMs)
	default:
		return "ℹ️ **" + event.MonitorName + "** — " + event.Reason
	}
}

// FormatSubject creates a subject line for email notifications.
func FormatSubject(event Event) string {
	switch event.Type {
	case EventDown:
		return "🔴 " + event.MonitorName + " is DOWN"
	case EventUp:
		return "🟢 " + event.MonitorName + " is UP"
	case EventDegraded:
		return "🟡 " + event.MonitorName + " is DEGRADED"
	default:
		return "Uptimer Alert: " + event.MonitorName
	}
}

func formatMs(ms int) string {
	if ms < 1000 {
		return string(rune(ms+'0')) + "ms"
	}
	return string(rune(ms/1000+'0')) + "." + string(rune((ms%1000)/100+'0')) + "s"
}
