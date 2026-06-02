package checker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/ZaViBiS/isitdead/internal/model"
)

type fakeRegionalChecker struct {
	results []RegionResult
}

func (f fakeRegionalChecker) CheckRegions(context.Context, string, string, int) []RegionResult {
	return f.results
}

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

func stubHTTPStatusTransport(t *testing.T, statusCode int) {
	prev := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: statusCode,
			Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = prev
	})
}

func TestScheduler(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

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

func TestSchedulerMultiRegionCheckStoresAggregateAndRegions(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	user, _, err := storage.AddUser("regionuser", "region@example.com", "password")
	assert.NoError(t, err)
	srv, err := storage.AddServer(user.ID, "Regional Server", "http://example.test", "http", 300, 10, 300, false)
	assert.NoError(t, err)

	stubHTTPStatusTransport(t, http.StatusInternalServerError)

	scheduler := NewScheduler(storage)
	scheduler.SetLocalRegion("eu")
	scheduler.SetRegionalChecker(fakeRegionalChecker{
		results: []RegionResult{
			{Region: "us", Status: "200 OK", Latency: 80},
		},
	})
	defer scheduler.Stop()

	last := scheduler.performCheck(*srv, lastResult{})
	assert.Equal(t, "200 OK", last.Status)
	assert.Equal(t, int64(80), last.Latency)

	globalHistory, err := storage.GetHistorySince(srv.ID, time.Time{})
	assert.NoError(t, err)
	assert.Len(t, globalHistory, 1)
	assert.Equal(t, model.CheckRegionGlobal, globalHistory[0].Region)
	assert.Equal(t, "200 OK", globalHistory[0].Status)

	allHistory, err := storage.GetHistorySinceForRegion(srv.ID, model.CheckRegionAll, time.Time{})
	assert.NoError(t, err)
	assert.Len(t, allHistory, 2)
	for _, result := range allHistory {
		assert.NotEqual(t, model.CheckRegionGlobal, result.Region)
	}

	localHistory, err := storage.GetHistorySinceForRegion(srv.ID, "eu", time.Time{})
	assert.NoError(t, err)
	assert.Len(t, localHistory, 1)
	assert.Equal(t, "500 Internal Server Error", localHistory[0].Status)

	remoteHistory, err := storage.GetHistorySinceForRegion(srv.ID, "us", time.Time{})
	assert.NoError(t, err)
	assert.Len(t, remoteHistory, 1)
	assert.Equal(t, "200 OK", remoteHistory[0].Status)
}

func newTestStorage(t *testing.T) *database.Storage {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL checker tests")
	}

	storage, err := database.Init(databaseURL)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}

	resetTestDatabase(t, storage)
	t.Cleanup(func() {
		resetTestDatabase(t, storage)
	})

	return storage
}

func resetTestDatabase(t *testing.T, storage *database.Storage) {
	t.Helper()

	if err := storage.DB.Exec(`
		TRUNCATE TABLE
			notification_preferences,
			ssl_certificate_statuses,
			check_results,
			servers,
			email_verifications,
			telegram_accounts,
			telegram_link_tokens,
			discord_accounts,
			discord_link_tokens,
			users
		RESTART IDENTITY CASCADE
	`).Error; err != nil {
		t.Fatalf("reset test database: %v", err)
	}
}

func TestAggregateRegionResults(t *testing.T) {
	t.Run("healthy when any region succeeds", func(t *testing.T) {
		result := aggregateRegionResults([]RegionResult{
			{Region: "eu", Status: "500 Internal Server Error", Latency: 40},
			{Region: "us", Status: "200 OK", Latency: 80},
			{Region: "ap", Status: "204 No Content", Latency: 120},
		})

		assert.Equal(t, model.CheckRegionGlobal, result.Region)
		assert.Equal(t, "200 OK", result.Status)
		assert.Equal(t, int64(100), result.Latency)
	})

	t.Run("down when all regions fail", func(t *testing.T) {
		result := aggregateRegionResults([]RegionResult{
			{Region: "eu", Status: "500 Internal Server Error", Latency: 40},
			{Region: "us", Status: "Probe request error: timeout", Latency: 1200},
		})

		assert.Equal(t, model.CheckRegionGlobal, result.Region)
		assert.Contains(t, result.Status, "All regions failed")
		assert.Contains(t, result.Status, "eu: 500 Internal Server Error")
		assert.Contains(t, result.Status, "us: Probe request error: timeout")
		assert.Equal(t, int64(1200), result.Latency)
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
