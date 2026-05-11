package mail

import (
	"fmt"
	"net/smtp"

	"github.com/ZaViBiS/isitdead/internal/config"
)

type Mailer struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) SendVerificationEmail(to, token string) error {
	subject := "Subject: Confirm your email for IsItDead\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	
	confirmURL := fmt.Sprintf("https://%s/api/auth/confirm?token=%s", m.cfg.Domain, token)
	if m.cfg.Env == "dev" {
		confirmURL = fmt.Sprintf("http://localhost:%s/api/auth/confirm?token=%s", m.cfg.Port, token)
	}

	body := fmt.Sprintf("<html><body><h1>Welcome!</h1><p>Please click the link below to confirm your email:</p><a href=\"%s\">%s</a></body></html>", confirmURL, confirmURL)
	
	msg := []byte(subject + mime + body)
	addr := fmt.Sprintf("%s:%s", m.cfg.SMTPHost, m.cfg.SMTPPort)
	
	auth := smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPass, m.cfg.SMTPHost)
	
	return smtp.SendMail(addr, auth, m.cfg.SMTPFrom, []string{to}, msg)
}
