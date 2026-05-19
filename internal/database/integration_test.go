package database

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserIDByToken(t *testing.T) {
	dbPath := "test_get_userID_by_token.db"
	defer os.Remove(dbPath)

	storage, err := Init(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	token, err := storage.CreateTelegramLinkToken(1, 15*time.Minute)
	require.NoError(t, err)

	userID, err := storage.GetUserIDByToken(token)
	require.NoError(t, err)
	require.Equal(t, uint(1), userID)
}

func TestCreateTelegramLinkToken(t *testing.T) {
	dbPath := "test_telegram_link_token.db"
	defer os.Remove(dbPath)

	storage, err := Init(dbPath)
	assert.NoError(t, err)
	defer storage.Close()

	user, _, err := storage.AddUser("telegramuser", "telegram@example.com", "password")
	assert.NoError(t, err)

	token, err := storage.CreateTelegramLinkToken(user.ID, 15*time.Minute)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	userID, err := storage.GetUserIDByToken(token)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userID)
}
