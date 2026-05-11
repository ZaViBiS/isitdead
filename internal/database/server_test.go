package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
	server, err := storage.AddServer(user.ID, "Google", "https://google.com", "http", 60)
	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, "Google", server.Name)
	assert.Equal(t, user.ID, server.UserID)
	assert.Equal(t, 60, server.CheckInterval)

	// 3. Тест: GetUserServers
	servers, err := storage.GetUserServers(user.ID)
	assert.NoError(t, err)
	assert.Len(t, servers, 1)
	assert.Equal(t, "https://google.com", servers[0].URL)

	// 4. Тест: UpdateServer
	updated, err := storage.UpdateServer(user.ID, server.ID, "Google Search", "https://google.com/search", "http", 60)
	assert.NoError(t, err)
	assert.Equal(t, "Google Search", updated.Name)
	assert.Equal(t, "https://google.com/search", updated.URL)

	// 5. Тест: UpdateServerStatus
	err = storage.UpdateServerStatus(server.ID, "200 OK", 150)
	assert.NoError(t, err)

	// Перевіряємо оновлення в базі
	servers, _ = storage.GetUserServers(user.ID)
	assert.Equal(t, "200 OK", servers[0].Status)
	assert.Equal(t, int64(150), servers[0].Latency)

	// 6. Тест: DeleteServer
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
	srv1, _ := storage.AddServer(user1.ID, "S1", "u1.com", "http", 60)

	// Тест: User 2 намагається оновити сервер User 1
	_, err = storage.UpdateServer(user2.ID, srv1.ID, "Hacked", "hacked.com", "http", 60)
	assert.Error(t, err, "User 2 should not be able to update User 1's server")

	// Тест: User 2 намагається видалити сервер User 1
	err = storage.DeleteServer(user2.ID, srv1.ID)
	// GORM Delete по ID і UserID просто не знайде запис, якщо їх немає, і не поверне помилку через специфіку роботи, 
	// але ми можемо перевірити чи сервер все ще на місці.
	assert.NoError(t, err)

	servers, _ := storage.GetUserServers(user1.ID)
	assert.Len(t, servers, 1, "Server should still exist after unauthorized delete attempt")
}
