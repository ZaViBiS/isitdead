package database

import (
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
	"gorm.io/gorm"
)

func (s *Storage) CreateTelegramLinkToken(userID uint, ttl time.Duration) (string, error) {
	token := GenerateToken()
	expiresAt := time.Now().Add(ttl)

	err := s.executeWrite(func(db *gorm.DB) error {
		linkToken := &model.TelegramLinkToken{
			UserID:    userID,
			Token:     token,
			ExpiresAt: expiresAt,
		}

		return db.Create(linkToken).Error
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Storage) GetUserIDByToken(token string) (uint, error) {
	var telegramLinkToken model.TelegramLinkToken
	if err := s.DB.Where("token = ?", token).Find(&telegramLinkToken).Error; err != nil {
		return 0, err
	}
	return telegramLinkToken.UserID, nil
}

func (s *Storage) CreateTelegramAccount(userID uint, chatID int) error {
	telegramAccount := model.TelegramAccount{
		UserID: userID,
		ChatID: int64(chatID),
	}

	if err := s.executeWrite(func(db *gorm.DB) error {
		return db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(telegramAccount).Error; err != nil {
				return err
			}
			return nil
		})
	}); err != nil {
		return err
	}

	return nil
}
