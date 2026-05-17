package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/config"
)

const resendEmailsEndpoint = "https://api.resend.com/emails"

type Mailer struct {
	cfg      *config.Config
	client   *http.Client
	endpoint string
}

type sendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func New(cfg *config.Config) *Mailer {
	return &Mailer{
		cfg:      cfg,
		client:   &http.Client{Timeout: 10 * time.Second},
		endpoint: resendEmailsEndpoint,
	}
}

func (m *Mailer) SendVerificationEmail(to, token string) error {
	confirmURL := fmt.Sprintf("https://%s/api/auth/confirm?token=%s", m.cfg.Domain, token)
	if m.cfg.Env == "dev" {
		confirmURL = fmt.Sprintf("http://localhost:%s/api/auth/confirm?token=%s", m.cfg.Port, token)
	}

	body := fmt.Sprintf("<html><body><h1>Welcome!</h1><p>Please click the link below to confirm your email:</p><a href=\"%s\">%s</a></body></html>", confirmURL, confirmURL)

	return m.SendHTML(to, "Confirm your email for isitdead.cc", body)
}

func (m *Mailer) SendHTML(to, subject, body string) error {
	if strings.TrimSpace(m.cfg.ResendAPIKey) == "" {
		return fmt.Errorf("resend api key is empty")
	}
	if strings.TrimSpace(m.cfg.ResendFrom) == "" {
		return fmt.Errorf("resend from address is empty")
	}

	payload, err := json.Marshal(sendEmailRequest{
		From:    m.cfg.ResendFrom,
		To:      []string{to},
		Subject: subject,
		HTML:    body,
	})
	if err != nil {
		return fmt.Errorf("marshal resend payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, m.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "isitdead/1.0")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("send email via resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("resend returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return nil
}
