package mail

import (
	"crypto/tls"
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
	if m.cfg.SMTPPort == "465" {
		return sendMailImplicitTLS(addr, m.cfg.SMTPHost, auth, m.cfg.SMTPFrom, to, msg)
	}

	return smtp.SendMail(addr, auth, m.cfg.SMTPFrom, []string{to}, msg)
}

func sendMailImplicitTLS(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return client.Quit()
}
