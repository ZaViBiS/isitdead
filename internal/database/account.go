package database

import (
	"github.com/ZaViBiS/isitdead/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GetUserByEmail знаходить користувача за email
func (s *Storage) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByGoogleID знаходить користувача за Google ID
func (s *Storage) GetUserByGoogleID(googleID string) (*model.User, error) {
	var user model.User
	if err := s.DB.Where("google_id = ?", googleID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// AddGoogleUser створює нового користувача через Google OAuth
func (s *Storage) AddGoogleUser(username, email, googleID string) (*model.User, error) {
	// Перевіряємо чи вже існує користувач з таким email
	existing, _ := s.GetUserByEmail(email)
	if existing != nil {
		// Якщо користувач існує, але не має GoogleID - оновлюємо його
		if existing.GoogleID == "" {
			err := s.executeWrite(func(db *gorm.DB) error {
				return db.Model(existing).Update("google_id", googleID).Error
			})
			return existing, err
		}
		return existing, nil
	}

	user := &model.User{
		Username: username,
		Email:    email,
		GoogleID: googleID,
	}

	if err := s.executeWrite(func(db *gorm.DB) error {
		return db.Create(user).Error
	}); err != nil {
		return nil, err
	}

	return user, nil
}

// AddUser створює нового користувача з хешованим паролем
func (s *Storage) AddUser(username, email, password string) (*model.User, error) {
	// Хешуємо пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err = s.executeWrite(func(db *gorm.DB) error {
		return db.Create(user).Error
	}); err != nil {
		return nil, err
	}

	return user, nil
}
