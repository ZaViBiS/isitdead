package database

import (
	"errors"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
	"gorm.io/gorm"
)

var (
	ErrDiscordTokenInvalid = errors.New("discord link token is invalid")
	ErrDiscordTokenExpired = errors.New("discord link token is expired")
	ErrDiscordTokenUsed    = errors.New("discord link token is already used")
)

func (s *Storage) CreateDiscordLinkToken(userID uint, ttl time.Duration) (string, error) {
	token := GenerateToken()
	expiresAt := time.Now().UTC().Add(ttl)

	err := s.executeWrite(func(db *gorm.DB) error {
		linkToken := &model.DiscordLinkToken{UserID: userID, Token: token, ExpiresAt: expiresAt}
		return db.Create(linkToken).Error
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Storage) LinkDiscordAccount(token, webhookURL string) (uint, error) {
	var userID uint
	err := s.executeWrite(func(db *gorm.DB) error {
		return db.Transaction(func(tx *gorm.DB) error {
			var linkToken model.DiscordLinkToken
			if err := tx.Where("token = ?", token).First(&linkToken).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrDiscordTokenInvalid
				}
				return err
			}
			if linkToken.UsedAt != nil {
				return ErrDiscordTokenUsed
			}
			if time.Now().After(linkToken.ExpiresAt) {
				return ErrDiscordTokenExpired
			}
			userID = linkToken.UserID
			if _, err := upsertDiscordAccountTx(tx, linkToken.UserID, webhookURL); err != nil {
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

func (s *Storage) GetDiscordAccountByUserID(userID uint) (*model.DiscordAccount, error) {
	var account model.DiscordAccount
	if err := s.DB.Where("user_id = ?", userID).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *Storage) EnsureDiscordNotificationPreferences(userID uint) error {
	var servers []model.Server
	if err := s.DB.Where("user_id = ?", userID).Find(&servers).Error; err != nil {
		return err
	}
	return s.executeWrite(func(db *gorm.DB) error {
		for _, server := range servers {
			if err := ensureDiscordNotificationPreferencesTx(db, userID, server.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

func upsertDiscordAccountTx(db *gorm.DB, userID uint, webhookURL string) (*model.DiscordAccount, error) {
	var existing model.DiscordAccount
	err := db.Where("user_id = ?", userID).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	now := time.Now().UTC()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		account := &model.DiscordAccount{UserID: userID, WebhookURL: webhookURL, LinkedAt: now}
		return account, db.Create(account).Error
	}
	err = db.Model(&existing).Updates(map[string]any{"webhook_url": webhookURL, "linked_at": now}).Error
	if err != nil {
		return nil, err
	}
	existing.WebhookURL = webhookURL
	existing.LinkedAt = now
	return &existing, nil
}

func ensureDiscordNotificationPreferencesTx(db *gorm.DB, userID, serverID uint) error {
	for _, event := range []string{model.NotificationEventDown, model.NotificationEventUp, model.NotificationEventSSL30d, model.NotificationEventSSL14d, model.NotificationEventSSL7d} {
		var existing model.NotificationPreference
		err := db.Where("user_id = ? AND server_id = ? AND channel = ? AND event = ?", userID, serverID, model.NotificationChannelDiscord, event).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := db.Create(&model.NotificationPreference{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelDiscord, Event: event, Enabled: false}).Error; err != nil {
			return err
		}
	}
	return nil
}
