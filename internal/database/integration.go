package database

import (
	"errors"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
	"gorm.io/gorm"
)

var (
	ErrTelegramTokenInvalid = errors.New("telegram link token is invalid")
	ErrTelegramTokenExpired = errors.New("telegram link token is expired")
	ErrTelegramTokenUsed    = errors.New("telegram link token is already used")
)

func (s *Storage) CreateTelegramLinkToken(userID uint, ttl time.Duration) (string, error) {
	token := GenerateToken()
	expiresAt := time.Now().UTC().Add(ttl)

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
	if err := s.DB.Where("token = ?", token).First(&telegramLinkToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrTelegramTokenInvalid
		}
		return 0, err
	}
	if telegramLinkToken.UsedAt != nil {
		return 0, ErrTelegramTokenUsed
	}
	if time.Now().After(telegramLinkToken.ExpiresAt) {
		return 0, ErrTelegramTokenExpired
	}
	return telegramLinkToken.UserID, nil
}

func (s *Storage) CreateTelegramAccount(userID uint, chatID int64) error {
	_, err := s.upsertTelegramAccount(userID, chatID, nil)
	return err
}

func (s *Storage) LinkTelegramAccount(token string, chatID int64) (uint, error) {
	var userID uint
	err := s.executeWrite(func(db *gorm.DB) error {
		return db.Transaction(func(tx *gorm.DB) error {
			var linkToken model.TelegramLinkToken
			if err := tx.Where("token = ?", token).First(&linkToken).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrTelegramTokenInvalid
				}
				return err
			}
			if linkToken.UsedAt != nil {
				return ErrTelegramTokenUsed
			}
			if time.Now().After(linkToken.ExpiresAt) {
				return ErrTelegramTokenExpired
			}

			userID = linkToken.UserID
			if _, err := upsertTelegramAccountTx(tx, linkToken.UserID, chatID); err != nil {
				return err
			}

			usedAt := time.Now().UTC()
			return tx.Model(&linkToken).Update("used_at", &usedAt).Error
		})
	})
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *Storage) GetTelegramAccountByUserID(userID uint) (*model.TelegramAccount, error) {
	var account model.TelegramAccount
	if err := s.DB.Where("user_id = ?", userID).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *Storage) EnsureTelegramNotificationPreferences(userID uint) error {
	var servers []model.Server
	if err := s.DB.Where("user_id = ?", userID).Find(&servers).Error; err != nil {
		return err
	}

	return s.executeWrite(func(db *gorm.DB) error {
		for _, server := range servers {
			if err := ensureTelegramNotificationPreferencesTx(db, userID, server.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Storage) upsertTelegramAccount(userID uint, chatID int64, tx *gorm.DB) (*model.TelegramAccount, error) {
	if tx != nil {
		return upsertTelegramAccountTx(tx, userID, chatID)
	}

	var account *model.TelegramAccount
	err := s.executeWrite(func(db *gorm.DB) error {
		var err error
		account, err = upsertTelegramAccountTx(db, userID, chatID)
		return err
	})
	return account, err
}

func upsertTelegramAccountTx(db *gorm.DB, userID uint, chatID int64) (*model.TelegramAccount, error) {
	var existing model.TelegramAccount
	err := db.Where("user_id = ?", userID).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now().UTC()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		account := &model.TelegramAccount{
			UserID:   userID,
			ChatID:   chatID,
			LinkedAt: now,
		}
		return account, db.Create(account).Error
	}

	err = db.Model(&existing).Updates(map[string]any{
		"chat_id":   chatID,
		"linked_at": now,
	}).Error
	if err != nil {
		return nil, err
	}
	existing.ChatID = chatID
	existing.LinkedAt = now
	return &existing, nil
}

func ensureTelegramNotificationPreferencesTx(db *gorm.DB, userID, serverID uint) error {
	for _, event := range []string{
		model.NotificationEventDown,
		model.NotificationEventUp,
		model.NotificationEventSSL30d,
		model.NotificationEventSSL14d,
		model.NotificationEventSSL7d,
	} {
		var existing model.NotificationPreference
		err := db.Where("user_id = ? AND server_id = ? AND channel = ? AND event = ?", userID, serverID, model.NotificationChannelTelegram, event).
			First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := db.Create(&model.NotificationPreference{
			UserID:   userID,
			ServerID: serverID,
			Channel:  model.NotificationChannelTelegram,
			Event:    event,
			Enabled:  false,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
