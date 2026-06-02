package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserByEmail_NotFound(t *testing.T) {
	storage := newTestStorage(t)

	// Спроба знайти неіснуючого користувача має повернути помилку
	foundUser, err := storage.GetUserByEmail("missing@example.com")
	assert.Error(t, err)
	assert.Nil(t, foundUser)
}

func TestAddUser_DuplicateEmail(t *testing.T) {
	storage := newTestStorage(t)

	email := "duplicate@example.com"
	_, _, err := storage.AddUser("user1", email, "password")
	assert.NoError(t, err)

	// Спроба додати користувача з таким самим email має повернути помилку
	_, _, err = storage.AddUser("user2", email, "otherpassword")
	assert.Error(t, err, "Should fail when email is duplicated")
}

func TestAddGoogleUser_LinksExistingUnverifiedEmail(t *testing.T) {
	storage := newTestStorage(t)

	email := "google-link@example.com"
	user, token, err := storage.AddUser("googlelink", email, "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	linked, err := storage.AddGoogleUser("Google Link", email, "google-123")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, linked.ID)
	assert.True(t, linked.VerifiedEmail)
	assert.NotNil(t, linked.GoogleID)
	assert.Equal(t, "google-123", *linked.GoogleID)

	verification, err := storage.GetVerificationByToken(token)
	assert.Error(t, err)
	assert.Nil(t, verification)
}

func TestAddGoogleUser_CreatesVerifiedUser(t *testing.T) {
	storage := newTestStorage(t)

	user, err := storage.AddGoogleUser("Google User", "google-user@example.com", "google-456")
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.True(t, user.VerifiedEmail)
	assert.NotNil(t, user.GoogleID)
	assert.Equal(t, "google-456", *user.GoogleID)

	found, err := storage.GetUserByGoogleID("google-456")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}
