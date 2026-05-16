package checker

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/stretchr/testify/assert"
)

func stubHTTP200Transport(t *testing.T) {
	prev := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = prev
	})
}

func TestScheduler(t *testing.T) {
	dbPath := "test_scheduler.db"
	storage, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer func() {
		storage.Close()
		os.Remove(dbPath)
	}()

	// Create a user first
	user, _, err := storage.AddUser("testuser", "test@example.com", "password")
	assert.NoError(t, err)

	stubHTTP200Transport(t)

	// Add server to DB
	srv, err := storage.AddServer(user.ID, "Test Server", "http://example.test", "http", 1, 10, 300, false) // 1 second interval
	assert.NoError(t, err)

	scheduler := NewScheduler(storage)
	defer scheduler.Stop()

	t.Run("Start Scheduler", func(t *testing.T) {
		err := scheduler.Start()
		assert.NoError(t, err)

		// Wait for at least one check to complete
		time.Sleep(1500 * time.Millisecond)

		// Check if history was populated
		history, err := storage.GetHistorySince(srv.ID, time.Time{})
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(history), 1)
		assert.Equal(t, "200 OK", history[0].Status)

		assert.GreaterOrEqual(t, history[0].Latency, int64(0))
	})

	t.Run("Stop Server Monitor", func(t *testing.T) {
		historyBeforeStop, err := storage.GetHistorySince(srv.ID, time.Time{})
		assert.NoError(t, err)
		assert.NotEmpty(t, historyBeforeStop)

		scheduler.StopServerMonitor(srv.ID)
		time.Sleep(1500 * time.Millisecond)

		historyAfterStop, err := storage.GetHistorySince(srv.ID, time.Time{})
		assert.NoError(t, err)
		assert.Len(t, historyAfterStop, len(historyBeforeStop))
	})
}

func TestSSLReminder(t *testing.T) {
	tests := []struct {
		name          string
		daysRemaining int
		lastThreshold int
		wantEvent     string
		wantThreshold int
		wantOK        bool
	}{
		{name: "thirty day reminder", daysRemaining: 29, wantEvent: model.NotificationEventSSL30d, wantThreshold: 30, wantOK: true},
		{name: "fourteen day reminder", daysRemaining: 14, lastThreshold: 30, wantEvent: model.NotificationEventSSL14d, wantThreshold: 14, wantOK: true},
		{name: "seven day reminder", daysRemaining: 6, lastThreshold: 14, wantEvent: model.NotificationEventSSL7d, wantThreshold: 7, wantOK: true},
		{name: "deduplicates current reminder", daysRemaining: 6, lastThreshold: 7, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, threshold, ok := sslReminder(tt.daysRemaining, tt.lastThreshold)
			assert.Equal(t, tt.wantEvent, event)
			assert.Equal(t, tt.wantThreshold, threshold)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestRetainedSSLNotificationThreshold(t *testing.T) {
	expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour)
	previous := &model.SSLCertificateStatus{
		ExpiresAt:             &expiresAt,
		Fingerprint:           "same-cert",
		LastNotifiedThreshold: 30,
	}

	assert.Equal(t, 30, retainedSSLNotificationThreshold(previous, SSLCertificateInfo{
		ExpiresAt:   &expiresAt,
		Fingerprint: "same-cert",
	}))
	assert.Equal(t, 0, retainedSSLNotificationThreshold(previous, SSLCertificateInfo{
		ExpiresAt:   &expiresAt,
		Fingerprint: "new-cert",
	}))

	renewedExpiry := expiresAt.Add(90 * 24 * time.Hour)
	assert.Equal(t, 0, retainedSSLNotificationThreshold(previous, SSLCertificateInfo{
		ExpiresAt:   &renewedExpiry,
		Fingerprint: "same-cert",
	}))
}
