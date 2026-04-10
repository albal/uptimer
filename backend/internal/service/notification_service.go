package service

import (
	"context"
	"log/slog"

	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/notification"
)

// NotificationService handles sending notifications via configured integrations.
type NotificationService struct {
	notifiers map[string]notification.Notifier
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService() *NotificationService {
	return &NotificationService{
		notifiers: map[string]notification.Notifier{
			models.AlertEmail:      &notification.EmailNotifier{},
			models.AlertSlack:      &notification.SlackNotifier{},
			models.AlertTeams:      &notification.TeamsNotifier{},
			models.AlertDiscord:    &notification.DiscordNotifier{},
			models.AlertTelegram:   &notification.TelegramNotifier{},
			models.AlertPagerDuty:  &notification.PagerDutyNotifier{},
			models.AlertWebhook:    &notification.WebhookNotifier{},
			models.AlertGoogleChat: &notification.GoogleChatNotifier{},
			models.AlertPushbullet: &notification.PushbulletNotifier{},
			models.AlertPushover:   &notification.PushoverNotifier{},
			models.AlertMattermost: &notification.MattermostNotifier{},
			models.AlertZapier:     &notification.ZapierNotifier{},
		},
	}
}

// NotifyDown sends notifications for a monitor going down.
func (s *NotificationService) NotifyDown(ctx context.Context, monitor *models.Monitor, incident *models.Incident, contacts []models.AlertContact) {
	event := notification.Event{
		Type:        notification.EventDown,
		MonitorName: monitor.Name,
		MonitorURL:  monitor.URL,
		MonitorType: monitor.Type,
		Reason:      incident.Reason,
		IncidentID:  incident.ID.String(),
		StartedAt:   incident.StartedAt,
	}

	for _, contact := range contacts {
		notifier, ok := s.notifiers[contact.Type]
		if !ok {
			slog.Warn("unknown notifier type", "type", contact.Type)
			continue
		}
		go func(c models.AlertContact) {
			if err := notifier.Notify(ctx, c, event); err != nil {
				slog.Error("notification failed", "type", c.Type, "contact", c.Name, "error", err)
			} else {
				slog.Info("notification sent", "type", c.Type, "contact", c.Name, "monitor", monitor.Name)
			}
		}(contact)
	}
}

// NotifyUp sends notifications for a monitor recovering.
func (s *NotificationService) NotifyUp(ctx context.Context, monitor *models.Monitor, contacts []models.AlertContact) {
	event := notification.Event{
		Type:        notification.EventUp,
		MonitorName: monitor.Name,
		MonitorURL:  monitor.URL,
		MonitorType: monitor.Type,
	}

	for _, contact := range contacts {
		notifier, ok := s.notifiers[contact.Type]
		if !ok {
			continue
		}
		go func(c models.AlertContact) {
			if err := notifier.Notify(ctx, c, event); err != nil {
				slog.Error("recovery notification failed", "type", c.Type, "error", err)
			}
		}(contact)
	}
}

// NotifyDegraded sends notifications for a monitor being degraded (slow response).
func (s *NotificationService) NotifyDegraded(ctx context.Context, monitor *models.Monitor, responseMs int, contacts []models.AlertContact) {
	event := notification.Event{
		Type:        notification.EventDegraded,
		MonitorName: monitor.Name,
		MonitorURL:  monitor.URL,
		MonitorType: monitor.Type,
		ResponseMs:  responseMs,
	}

	for _, contact := range contacts {
		notifier, ok := s.notifiers[contact.Type]
		if !ok {
			continue
		}
		go func(c models.AlertContact) {
			if err := notifier.Notify(ctx, c, event); err != nil {
				slog.Error("degraded notification failed", "type", c.Type, "error", err)
			}
		}(contact)
	}
}
