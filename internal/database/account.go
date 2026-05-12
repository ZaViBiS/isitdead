package database

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/ZaViBiS/isitdead/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GenerateToken створює випадковий токен
func GenerateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetVerificationByToken знаходить запис верифікації за токеном
func (s *Storage) GetVerificationByToken(token string) (*model.EmailVerification, error) {
	var v model.EmailVerification
	if err := s.DB.Where("token = ?", token).First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

// VerifyUser позначає користувача як верифікованого
func (s *Storage) VerifyUser(userID uint) error {
	return s.executeWrite(func(db *gorm.DB) error {
		return db.Transaction(func(tx *gorm.DB) error {
			// Оновлюємо статус користувача
			if err := tx.Model(&model.User{}).Where("id = ?", userID).Update("verified_email", true).Error; err != nil {
				return err
			}
			// Видаляємо токен
			if err := tx.Where("user_id = ?", userID).Delete(&model.EmailVerification{}).Error; err != nil {
				return err
			}
			return nil
		})
	})
}

// GetUserByEmail знаходить користувача за email
func (s *Storage) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID знаходить користувача за ID
func (s *Storage) GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	if err := s.DB.Where("id = ?", userID).First(&user).Error; err != nil {
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
		err := s.executeWrite(func(db *gorm.DB) error {
			return db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Model(existing).Updates(map[string]interface{}{
					"google_id":      googleID,
					"verified_email": true,
				}).Error; err != nil {
					return err
				}

				return tx.Where("user_id = ?", existing.ID).Delete(&model.EmailVerification{}).Error
			})
		})
		if err != nil {
			return nil, err
		}

		return s.GetUserByEmail(email)
	}

	user := &model.User{
		Username:      username,
		Email:         email,
		GoogleID:      &googleID,
		VerifiedEmail: true,
	}

	if err := s.executeWrite(func(db *gorm.DB) error {
		return db.Create(user).Error
	}); err != nil {
		return nil, err
	}

	return user, nil
}

// AddUser створює нового користувача з хешованим паролем та токеном верифікації
func (s *Storage) AddUser(username, email, password string) (*model.User, string, error) {
	// Хешуємо пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	token := GenerateToken()
	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	err = s.executeWrite(func(db *gorm.DB) error {
		return db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(user).Error; err != nil {
				return err
			}
			verification := &model.EmailVerification{
				UserID: user.ID,
				Token:  token,
			}
			return tx.Create(verification).Error
		})
	})

	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}
