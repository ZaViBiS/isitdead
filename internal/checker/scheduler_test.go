package checker

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/stretchr/testify/assert"
)

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
	user, err := storage.AddUser("testuser", "test@example.com", "pass")
	assert.NoError(t, err)

	// Create a test server to monitor
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Add server to DB
	srv, err := storage.AddServer(user.ID, "Test Server", ts.URL, "http", 1) // 1 second interval
	assert.NoError(t, err)

	scheduler := NewScheduler(storage)
	defer scheduler.Stop()

	t.Run("Start Scheduler", func(t *testing.T) {
		err := scheduler.Start()
		assert.NoError(t, err)

		// Wait for at least one check to complete
		time.Sleep(1500 * time.Millisecond)

		// Check if history was populated
		history, err := storage.GetHistory(srv.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(history), 1)
		assert.Equal(t, "200 OK", history[0].Status)

		// Check if server status was updated
		servers, err := storage.GetUserServers(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, "200 OK", servers[0].Status)
	})
}
