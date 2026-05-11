package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserByEmail_NotFound(t *testing.T) {
	dbPath := "test_account_notfound.db"
	defer os.Remove(dbPath)

	storage, err := Init(dbPath)
	assert.NoError(t, err)
	defer storage.Close()

	// Спроба знайти неіснуючого користувача має повернути помилку
	foundUser, err := storage.GetUserByEmail("missing@example.com")
	assert.Error(t, err)
	assert.Nil(t, foundUser)
}

func TestAddUser_DuplicateEmail(t *testing.T) {
	dbPath := "test_account_dup.db"
	defer os.Remove(dbPath)

	storage, err := Init(dbPath)
	assert.NoError(t, err)
	defer storage.Close()

	email := "duplicate@example.com"
	_, _, err = storage.AddUser("user1", email, "password")
	assert.NoError(t, err)

	// Спроба додати користувача з таким самим email має повернути помилку
	_, _, err = storage.AddUser("user2", email, "otherpassword")
	assert.Error(t, err, "Should fail when email is duplicated")
}
