package notify

import (
	"context"
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/rs/zerolog/log"
)

type Store interface {
	GetEnabledNotificationPreferences(serverID uint, event string) ([]model.NotificationPreference, error)
	GetUserByID(userID uint) (*model.User, error)
}

type Sender interface {
	Channel() string
	Send(ctx context.Context, message Message) error
}

type Message struct {
	Event         string
	Server        model.Server
	User          model.User
	Preference    model.NotificationPreference
	PreviousState string
	CurrentState  string
	Latency       int64
	CheckedAt     time.Time
}

type Service struct {
	store   Store
	senders map[string]Sender
}

func NewService(store Store, senders ...Sender) *Service {
	s := &Service{
		store:   store,
		senders: make(map[string]Sender, len(senders)),
	}
	for _, sender := range senders {
		s.senders[sender.Channel()] = sender
	}
	return s
}

func (s *Service) Notify(ctx context.Context, server model.Server, previousState, currentState string, latency int64) error {
	event, ok := transitionEvent(previousState, currentState)
	if !ok {
		return nil
	}

	prefs, err := s.store.GetEnabledNotificationPreferences(server.ID, event)
	if err != nil {
		return err
	}

	for _, pref := range prefs {
		sender, ok := s.senders[pref.Channel]
		if !ok {
			log.Warn().Str("channel", pref.Channel).Uint("server_id", server.ID).Msg("notification sender is not configured")
			continue
		}

		user, err := s.store.GetUserByID(pref.UserID)
		if err != nil {
			log.Error().Err(err).Uint("user_id", pref.UserID).Uint("server_id", server.ID).Msg("failed to load notification recipient")
			continue
		}

		message := Message{
			Event:         event,
			Server:        server,
			User:          *user,
			Preference:    pref,
			PreviousState: previousState,
			CurrentState:  currentState,
			Latency:       latency,
			CheckedAt:     time.Now().UTC(),
		}

		if err := sender.Send(ctx, message); err != nil {
			log.Error().Err(err).Str("channel", pref.Channel).Uint("server_id", server.ID).Msg("failed to send notification")
			continue
		}
	}

	return nil
}

func transitionEvent(previousState, currentState string) (string, bool) {
	previousKnown := previousState != "" && previousState != "unknown"
	if !previousKnown {
		return "", false
	}

	wasHealthy := isHealthy(previousState)
	isNowHealthy := isHealthy(currentState)

	if wasHealthy && !isNowHealthy {
		return model.NotificationEventDown, true
	}
	if !wasHealthy && isNowHealthy {
		return model.NotificationEventUp, true
	}
	return "", false
}

func isHealthy(status string) bool {
	return strings.HasPrefix(status, "2") || status == "Connected"
}
