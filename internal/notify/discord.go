package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
)

type DiscordAccountStore interface {
	GetDiscordAccountByUserID(userID uint) (*model.DiscordAccount, error)
}

type DiscordSender struct {
	store      DiscordAccountStore
	httpClient *http.Client
}

func NewDiscordSender(store DiscordAccountStore) *DiscordSender {
	return &DiscordSender{store: store, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (s *DiscordSender) Channel() string { return model.NotificationChannelDiscord }

func (s *DiscordSender) Send(ctx context.Context, message Message) error {
	account, err := s.store.GetDiscordAccountByUserID(message.User.ID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]any{"content": discordText(message)})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, account.WebhookURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned %s", resp.Status)
	}
	return nil
}

func discordText(message Message) string { return telegramText(message) }
