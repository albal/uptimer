package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/albal/uptimer/internal/models"
)

// SlackNotifier sends notifications to Slack via incoming webhooks.
type SlackNotifier struct{}

func (n *SlackNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	webhookURL := contact.Value

	color := "#36a64f" // green
	if event.Type == EventDown {
		color = "#ff0000"
	} else if event.Type == EventDegraded {
		color = "#ffaa00"
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  FormatSubject(event),
				"text":   FormatMessage(event),
				"footer": "Uptimer Monitoring",
			},
		},
	}

	return postJSON(ctx, webhookURL, payload)
}

// TeamsNotifier sends notifications to Microsoft Teams via incoming webhooks.
type TeamsNotifier struct{}

func (n *TeamsNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	webhookURL := contact.Value

	color := "00FF00"
	if event.Type == EventDown {
		color = "FF0000"
	} else if event.Type == EventDegraded {
		color = "FFAA00"
	}

	payload := map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": color,
		"summary":    FormatSubject(event),
		"sections": []map[string]interface{}{
			{
				"activityTitle": FormatSubject(event),
				"text":          FormatMessage(event),
			},
		},
	}

	return postJSON(ctx, webhookURL, payload)
}

// DiscordNotifier sends notifications to Discord via webhooks.
type DiscordNotifier struct{}

func (n *DiscordNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	webhookURL := contact.Value

	color := 65280 // green
	if event.Type == EventDown {
		color = 16711680 // red
	} else if event.Type == EventDegraded {
		color = 16755200 // amber
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       FormatSubject(event),
				"description": FormatMessage(event),
				"color":       color,
				"footer": map[string]string{
					"text": "Uptimer Monitoring",
				},
			},
		},
	}

	return postJSON(ctx, webhookURL, payload)
}

// TelegramNotifier sends notifications via Telegram Bot API.
type TelegramNotifier struct{}

func (n *TelegramNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	chatID := contact.Value
	botToken, _ := contact.Config["bot_token"].(string)
	if botToken == "" {
		return fmt.Errorf("telegram bot_token not configured")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       FormatMessage(event),
		"parse_mode": "Markdown",
	}

	return postJSON(ctx, url, payload)
}

// PagerDutyNotifier sends notifications via PagerDuty Events API v2.
type PagerDutyNotifier struct{}

func (n *PagerDutyNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	routingKey := contact.Value

	severity := "info"
	action := "resolve"
	if event.Type == EventDown {
		severity = "critical"
		action = "trigger"
	} else if event.Type == EventDegraded {
		severity = "warning"
		action = "trigger"
	}

	payload := map[string]interface{}{
		"routing_key":  routingKey,
		"event_action": action,
		"dedup_key":    "uptimer-" + event.MonitorName,
		"payload": map[string]interface{}{
			"summary":  FormatSubject(event),
			"severity": severity,
			"source":   "Uptimer",
			"custom_details": map[string]string{
				"monitor_url":  event.MonitorURL,
				"monitor_type": event.MonitorType,
				"reason":       event.Reason,
			},
		},
	}

	return postJSON(ctx, "https://events.pagerduty.com/v2/enqueue", payload)
}

// GoogleChatNotifier sends notifications to Google Chat via webhooks.
type GoogleChatNotifier struct{}

func (n *GoogleChatNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	webhookURL := contact.Value

	payload := map[string]interface{}{
		"cards": []map[string]interface{}{
			{
				"header": map[string]string{
					"title": FormatSubject(event),
				},
				"sections": []map[string]interface{}{
					{
						"widgets": []map[string]interface{}{
							{
								"textParagraph": map[string]string{
									"text": FormatMessage(event),
								},
							},
						},
					},
				},
			},
		},
	}

	return postJSON(ctx, webhookURL, payload)
}

// PushbulletNotifier sends push notifications via Pushbullet.
type PushbulletNotifier struct{}

func (n *PushbulletNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	apiKey := contact.Value

	payload := map[string]interface{}{
		"type":  "note",
		"title": FormatSubject(event),
		"body":  FormatMessage(event),
	}

	data, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.pushbullet.com/v2/pushes", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Token", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("pushbullet returned status %d", resp.StatusCode)
	}
	return nil
}

// PushoverNotifier sends push notifications via Pushover.
type PushoverNotifier struct{}

func (n *PushoverNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	userKey := contact.Value
	appToken, _ := contact.Config["app_token"].(string)
	if appToken == "" {
		return fmt.Errorf("pushover app_token not configured")
	}

	priority := 0
	if event.Type == EventDown {
		priority = 1
	}

	payload := map[string]interface{}{
		"token":    appToken,
		"user":     userKey,
		"title":    FormatSubject(event),
		"message":  FormatMessage(event),
		"priority": priority,
	}

	return postJSON(ctx, "https://api.pushover.net/1/messages.json", payload)
}

// MattermostNotifier sends notifications to Mattermost via incoming webhooks.
type MattermostNotifier struct{}

func (n *MattermostNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	webhookURL := contact.Value

	color := "#36a64f"
	if event.Type == EventDown {
		color = "#ff0000"
	} else if event.Type == EventDegraded {
		color = "#ffaa00"
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  FormatSubject(event),
				"text":   FormatMessage(event),
				"footer": "Uptimer Monitoring",
			},
		},
	}

	return postJSON(ctx, webhookURL, payload)
}

// WebhookNotifier sends notifications to custom webhooks.
type WebhookNotifier struct{}

func (n *WebhookNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	webhookURL := contact.Value

	payload := map[string]interface{}{
		"event":        string(event.Type),
		"monitor_name": event.MonitorName,
		"monitor_url":  event.MonitorURL,
		"monitor_type": event.MonitorType,
		"reason":       event.Reason,
		"incident_id":  event.IncidentID,
		"started_at":   event.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		"response_ms":  event.ResponseMs,
	}

	return postJSON(ctx, webhookURL, payload)
}

// ZapierNotifier sends notifications to Zapier webhooks.
type ZapierNotifier struct{}

func (n *ZapierNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	webhookURL := contact.Value

	payload := map[string]interface{}{
		"event_type":   string(event.Type),
		"monitor_name": event.MonitorName,
		"monitor_url":  event.MonitorURL,
		"reason":       event.Reason,
		"message":      FormatMessage(event),
	}

	return postJSON(ctx, webhookURL, payload)
}

// postJSON sends a JSON POST request to the given URL.
func postJSON(ctx context.Context, url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
