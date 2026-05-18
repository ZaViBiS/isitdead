package notify

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
)

type TelegramSender struct {
	token string
	client *http.Client
}

func NewTelegramSender(token string) *TelegramSender {
	return &TelegramSender{
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *TelegramSender) Channel() string {
	return model.NotificationChannelTelegram
}

func (s *TelegramSender) Send(ctx context.Context, message Message) error {
	if s.token == "" {
		return fmt.Errorf("telegram token is not configured")
	}

	chatID := message.Preference.Destination
	if chatID == "" {
		return fmt.Errorf("telegram chat ID is missing")
	}

	var text string
	if message.SSLStatus != nil {
		text = fmt.Sprintf(
			"🚨 *isitdead.cc: SSL Expiry*\n\n"+
				"*Server:* %s\n"+
				"*URL:* %s\n"+
				"*Expires in:* %d days\n"+
				"*Expires at:* %s\n"+
				"*Checked at:* %s",
			message.Server.Name,
			message.Server.URL,
			message.SSLStatus.DaysRemaining,
			message.SSLStatus.ExpiresAt.Format(time.RFC3339),
			message.CheckedAt.Format(time.RFC3339),
		)
	} else if message.Event == model.NotificationEventUp {
		text = fmt.Sprintf(
			"✅ *isitdead.cc: Recovered*\n\n"+
				"*Server:* %s\n"+
				"*URL:* %s\n"+
				"*Previous status:* %s\n"+
				"*Current status:* %s\n"+
				"*Latency:* %dms\n"+
				"*Checked at:* %s",
			message.Server.Name,
			message.Server.URL,
			message.PreviousState,
			message.CurrentState,
			message.Latency,
			message.CheckedAt.Format(time.RFC3339),
		)
	} else {
		text = fmt.Sprintf(
			"❌ *isitdead.cc: Down*\n\n"+
				"*Server:* %s\n"+
				"*URL:* %s\n"+
				"*Previous status:* %s\n"+
				"*Current status:* %s\n"+
				"*Latency:* %dms\n"+
				"*Checked at:* %s",
			message.Server.Name,
			message.Server.URL,
			message.PreviousState,
			message.CurrentState,
			message.Latency,
			message.CheckedAt.Format(time.RFC3339),
		)
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.token)
	payload := url.Values{}
	payload.Set("chat_id", chatID)
	payload.Set("text", text)
	payload.Set("parse_mode", "Markdown")

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth("", "") // Not needed for Telegram, but ensures correct header
	
	// Using form data for simplicity in Go
	resp, err := s.client.PostForm(apiURL, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api returned status %d", resp.StatusCode)
	}

	return nil
}
