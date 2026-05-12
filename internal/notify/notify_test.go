package notify

import (
	"context"
	"testing"

	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	prefs []model.NotificationPreference
	user  model.User
}

func (s *fakeStore) GetEnabledNotificationPreferences(serverID uint, event string) ([]model.NotificationPreference, error) {
	prefs := []model.NotificationPreference{}
	for _, pref := range s.prefs {
		if pref.ServerID == serverID && pref.Event == event && pref.Enabled {
			prefs = append(prefs, pref)
		}
	}
	return prefs, nil
}

func (s *fakeStore) GetUserByID(userID uint) (*model.User, error) {
	return &s.user, nil
}

type fakeSender struct {
	channel  string
	messages []Message
}

func (s *fakeSender) Channel() string {
	return s.channel
}

func (s *fakeSender) Send(ctx context.Context, message Message) error {
	s.messages = append(s.messages, message)
	return nil
}

func TestServiceNotifySendsDownAndRecoveryEvents(t *testing.T) {
	store := &fakeStore{
		user: model.User{ID: 7, Email: "user@example.com"},
		prefs: []model.NotificationPreference{
			{UserID: 7, ServerID: 11, Channel: model.NotificationChannelEmail, Event: model.NotificationEventDown, Enabled: true},
			{UserID: 7, ServerID: 11, Channel: model.NotificationChannelEmail, Event: model.NotificationEventUp, Enabled: true},
		},
	}
	sender := &fakeSender{channel: model.NotificationChannelEmail}
	service := NewService(store, sender)
	server := model.Server{ID: 11, UserID: 7, Name: "API", URL: "https://example.com"}

	require.NoError(t, service.Notify(context.Background(), server, "200 OK", "500 Internal Server Error", 120))
	require.NoError(t, service.Notify(context.Background(), server, "500 Internal Server Error", "200 OK", 80))

	require.Len(t, sender.messages, 2)
	assert.Equal(t, model.NotificationEventDown, sender.messages[0].Event)
	assert.Equal(t, model.NotificationEventUp, sender.messages[1].Event)
}

func TestServiceNotifyIgnoresUnknownInitialState(t *testing.T) {
	store := &fakeStore{
		user:  model.User{ID: 7, Email: "user@example.com"},
		prefs: []model.NotificationPreference{{UserID: 7, ServerID: 11, Channel: model.NotificationChannelEmail, Event: model.NotificationEventDown, Enabled: true}},
	}
	sender := &fakeSender{channel: model.NotificationChannelEmail}
	service := NewService(store, sender)
	server := model.Server{ID: 11, UserID: 7, Name: "API", URL: "https://example.com"}

	require.NoError(t, service.Notify(context.Background(), server, "unknown", "500 Internal Server Error", 120))

	assert.Empty(t, sender.messages)
}
