package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
)

type DiscordAccountStore interface {
	GetDiscordAccountByUserID(userID uint) (*model.DiscordAccount, error)
}

type DiscordSender struct {
	store      DiscordAccountStore
	endpoint   string
	secret     string
	httpClient *http.Client
}

func NewDiscordSender(store DiscordAccountStore, endpoint, secret string) *DiscordSender {
	return &DiscordSender{
		store:    store,
		endpoint: strings.TrimRight(endpoint, "/"),
		secret:   secret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *DiscordSender) Channel() string {
	return model.NotificationChannelDiscord
}

func (s *DiscordSender) Send(ctx context.Context, message Message) error {
	if s.endpoint == "" {
		return fmt.Errorf("discord integration endpoint is not configured")
	}

	account, err := s.store.GetDiscordAccountByUserID(message.User.ID)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]any{
		"channel_id": account.ChannelID,
		"text":       discordText(message),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint+"/api/messages", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.secret != "" {
		req.Header.Set("Authorization", "Bearer "+s.secret)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord integration returned %s", resp.Status)
	}
	return nil
}

func discordText(message Message) string {
	if message.SSLStatus != nil {
		expiresAt := "unknown"
		if message.SSLStatus.ExpiresAt != nil {
			expiresAt = message.SSLStatus.ExpiresAt.Format(time.RFC3339)
		}
		return fmt.Sprintf(
			"isitdead.cc: Discord SSL reminder\n\n%s\nURL: %s\nExpires in: %d days\nExpires at: %s\nChecked at: %s",
			message.Server.Name,
			message.Server.URL,
			message.SSLStatus.DaysRemaining,
			expiresAt,
			message.CheckedAt.Format(time.RFC3339),
		)
	}

	headline := "Monitor is down"
	if message.Event == model.NotificationEventUp {
		headline = "Monitor recovered"
	}

	return fmt.Sprintf(
		"isitdead.cc: %s\n\n%s\nURL: %s\nPrevious status: %s\nCurrent status: %s\nLatency: %dms\nChecked at: %s",
		headline,
		message.Server.Name,
		message.Server.URL,
		message.PreviousState,
		message.CurrentState,
		message.Latency,
		message.CheckedAt.Format(time.RFC3339),
	)
}
