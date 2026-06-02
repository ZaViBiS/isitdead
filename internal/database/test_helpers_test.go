package database

import (
	"os"
	"testing"
)

func newTestStorage(t *testing.T) *Storage {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL database tests")
	}

	storage, err := Init(databaseURL)
	if err != nil {
		t.Fatalf("init test database: %v", err)
	}

	if err := storage.ResetForTest(); err != nil {
		storage.Close()
		t.Fatalf("reset test database: %v", err)
	}

	t.Cleanup(func() {
		_ = storage.ResetForTest()
		_ = storage.Close()
	})

	return storage
}

func (s *Storage) ResetForTest() error {
	return s.DB.Exec(`
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
	`).Error
}
