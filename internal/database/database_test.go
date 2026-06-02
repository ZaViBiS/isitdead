package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUserFlow(t *testing.T) {
	storage := newTestStorage(t)

	// Тест: створення користувача
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	user, _, err := storage.AddUser(username, email, password)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)

	// Перевірка хешування пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	assert.NoError(t, err, "Password hash should be valid")

	// Тест: отримання користувача
	foundUser, err := storage.GetUserByEmail(email)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, username, foundUser.Username)
}
