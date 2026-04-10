package notification

import (
	"context"
	"fmt"
	"net/smtp"
	"os"

	"github.com/albal/uptimer/internal/models"
)

// EmailNotifier sends notifications via SMTP email.
type EmailNotifier struct{}

func (n *EmailNotifier) Notify(ctx context.Context, contact models.AlertContact, event Event) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")

	if host == "" {
		return fmt.Errorf("SMTP not configured")
	}
	if port == "" {
		port = "587"
	}
	if from == "" {
		from = "noreply@uptimer.local"
	}

	to := contact.Value
	subject := FormatSubject(event)
	body := FormatMessage(event)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	addr := host + ":" + port
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
