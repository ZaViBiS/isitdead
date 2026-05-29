package database

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/ZaViBiS/isitdead/internal/model"
)

func TestServerCRUD(t *testing.T) {
	dbPath := "test_server.db"
	defer os.Remove(dbPath)

	storage, err := Init(dbPath)
	assert.NoError(t, err)
	defer storage.Close()

	// 1. Створюємо тестового користувача
	user, _, err := storage.AddUser("server_owner", "owner@example.com", "pass123")
	assert.NoError(t, err)

	// 2. Тест: AddServer
	server, err := storage.AddServer(user.ID, "Google", "https://google.com", "http", 300, 10, 300, true)
	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, "Google", server.Name)
	assert.Equal(t, user.ID, server.UserID)
	assert.Equal(t, 300, server.CheckInterval)
	assert.Equal(t, 10, server.Timeout)
	assert.Equal(t, 300, server.SlowThreshold)
	assert.True(t, server.SSLEnabled)

	// 3. Тест: GetUserServers
	servers, err := storage.GetUserServers(user.ID)
	assert.NoError(t, err)
	assert.Len(t, servers, 1)
	assert.Equal(t, "https://google.com", servers[0].URL)

	// 4. Тест: UpdateServer
	updated, err := storage.UpdateServer(user.ID, server.ID, "Google Search", "https://google.com/search", "http", 300, 15, 450, false)
	assert.NoError(t, err)
	assert.Equal(t, "Google Search", updated.Name)
	assert.Equal(t, "https://google.com/search", updated.URL)
	assert.Equal(t, 15, updated.Timeout)
	assert.Equal(t, 450, updated.SlowThreshold)
	assert.False(t, updated.SSLEnabled)

	// 5. Тест: DeleteServer
	err = storage.DeleteServer(user.ID, server.ID)
	assert.NoError(t, err)

	servers, _ = storage.GetUserServers(user.ID)
	assert.Len(t, servers, 0)
}

func TestServerSecurity(t *testing.T) {
	dbPath := "test_server_security.db"
	defer os.Remove(dbPath)

	storage, err := Init(dbPath)
	assert.NoError(t, err)
	defer storage.Close()

	user1, _, _ := storage.AddUser("u1", "u1@ex.com", "p")
	user2, _, _ := storage.AddUser("u2", "u2@ex.com", "p")

	// User 1 додає сервер
	srv1, _ := storage.AddServer(user1.ID, "S1", "u1.com", "http", 300, 10, 300, false)

	// Тест: User 2 намагається оновити сервер User 1
	_, err = storage.UpdateServer(user2.ID, srv1.ID, "Hacked", "hacked.com", "http", 300, 10, 300, false)
	assert.Error(t, err, "User 2 should not be able to update User 1's server")

	// Тест: User 2 намагається видалити сервер User 1
	err = storage.DeleteServer(user2.ID, srv1.ID)
	assert.Error(t, err, "User 2 should not be able to delete User 1's server")

	servers, _ := storage.GetUserServers(user1.ID)
	assert.Len(t, servers, 1, "Server should still exist after unauthorized delete attempt")
}

func TestDeleteServerDoesNotDeleteOtherUsersSSLStatus(t *testing.T) {
	dbPath := "test_server_delete_ssl_security.db"
	defer os.Remove(dbPath)

	storage, err := Init(dbPath)
	assert.NoError(t, err)
	defer storage.Close()

	user1, _, err := storage.AddUser("u1", "ssl-owner@example.com", "p")
	assert.NoError(t, err)
	user2, _, err := storage.AddUser("u2", "ssl-attacker@example.com", "p")
	assert.NoError(t, err)

	srv1, err := storage.AddServer(user1.ID, "S1", "https://example.com", "http", 300, 10, 300, true)
	assert.NoError(t, err)

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	assert.NoError(t, storage.UpsertSSLCertificateStatus(model.SSLCertificateStatus{
		ServerID:      srv1.ID,
		Valid:         true,
		ExpiresAt:     &expiresAt,
		DaysRemaining: 1,
		LastCheckedAt: time.Now().UTC(),
	}))

	err = storage.DeleteServer(user2.ID, srv1.ID)
	assert.Error(t, err)

	_, err = storage.GetSSLCertificateStatus(srv1.ID)
	assert.NoError(t, err, "unauthorized delete must not remove another user's SSL status")

	err = storage.DeleteServer(user1.ID, srv1.ID)
	assert.NoError(t, err)

	_, err = storage.GetSSLCertificateStatus(srv1.ID)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound), "owner delete should remove SSL status")
}
