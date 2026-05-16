package notify

import (
	"context"
	"fmt"
	"html"
	"time"

	"github.com/ZaViBiS/isitdead/internal/mail"
	"github.com/ZaViBiS/isitdead/internal/model"
)

type EmailSender struct {
	mailer *mail.Mailer
}

func NewEmailSender(mailer *mail.Mailer) *EmailSender {
	return &EmailSender{mailer: mailer}
}

func (s *EmailSender) Channel() string {
	return model.NotificationChannelEmail
}

func (s *EmailSender) Send(ctx context.Context, message Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	to := message.Preference.Destination
	if to == "" {
		to = message.User.Email
	}

	subject := fmt.Sprintf("isitdead.cc: %s is down", message.Server.Name)
	headline := "Monitor is down"
	if message.SSLStatus != nil {
		subject = fmt.Sprintf("isitdead.cc: %s SSL certificate expires in %d days", message.Server.Name, message.SSLStatus.DaysRemaining)
		headline = "SSL certificate expiry reminder"
		body := fmt.Sprintf(
			`<html><body><h1>%s</h1><p><strong>%s</strong></p><p>URL: <a href="%s">%s</a></p><p>Certificate expires in: %d days</p><p>Expires at: %s</p><p>Checked at: %s</p></body></html>`,
			html.EscapeString(headline),
			html.EscapeString(message.Server.Name),
			html.EscapeString(message.Server.URL),
			html.EscapeString(message.Server.URL),
			message.SSLStatus.DaysRemaining,
			message.SSLStatus.ExpiresAt.Format(time.RFC3339),
			message.CheckedAt.Format(time.RFC3339),
		)
		return s.mailer.SendHTML(to, subject, body)
	}
	if message.Event == model.NotificationEventUp {
		subject = fmt.Sprintf("isitdead.cc: %s recovered", message.Server.Name)
		headline = "Monitor recovered"
	}

	body := fmt.Sprintf(
		`<html><body><h1>%s</h1><p><strong>%s</strong></p><p>URL: <a href="%s">%s</a></p><p>Previous status: %s</p><p>Current status: %s</p><p>Latency: %dms</p><p>Checked at: %s</p></body></html>`,
		html.EscapeString(headline),
		html.EscapeString(message.Server.Name),
		html.EscapeString(message.Server.URL),
		html.EscapeString(message.Server.URL),
		html.EscapeString(message.PreviousState),
		html.EscapeString(message.CurrentState),
		message.Latency,
		message.CheckedAt.Format(time.RFC3339),
	)

	return s.mailer.SendHTML(to, subject, body)
}
