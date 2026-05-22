package database

import (
	"github.com/ZaViBiS/isitdead/internal/model"
	"gorm.io/gorm"
)

func DefaultNotificationPreferences(userID, serverID uint) []model.NotificationPreference {
	return []model.NotificationPreference{
		{
			UserID:   userID,
			ServerID: serverID,
			Channel:  model.NotificationChannelEmail,
			Event:    model.NotificationEventDown,
			Enabled:  true,
		},
		{
			UserID:   userID,
			ServerID: serverID,
			Channel:  model.NotificationChannelEmail,
			Event:    model.NotificationEventUp,
			Enabled:  true,
		},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelEmail, Event: model.NotificationEventSSL30d, Enabled: true},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelEmail, Event: model.NotificationEventSSL14d, Enabled: true},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelEmail, Event: model.NotificationEventSSL7d, Enabled: true},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelTelegram, Event: model.NotificationEventDown, Enabled: false},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelTelegram, Event: model.NotificationEventUp, Enabled: false},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelTelegram, Event: model.NotificationEventSSL30d, Enabled: false},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelTelegram, Event: model.NotificationEventSSL14d, Enabled: false},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelTelegram, Event: model.NotificationEventSSL7d, Enabled: false},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelDiscord, Event: model.NotificationEventDown, Enabled: false},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelDiscord, Event: model.NotificationEventUp, Enabled: false},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelDiscord, Event: model.NotificationEventSSL30d, Enabled: false},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelDiscord, Event: model.NotificationEventSSL14d, Enabled: false},
		{UserID: userID, ServerID: serverID, Channel: model.NotificationChannelDiscord, Event: model.NotificationEventSSL7d, Enabled: false},
	}
}

func (s *Storage) EnsureDefaultNotificationPreferences(userID, serverID uint) error {
	return s.executeWrite(func(db *gorm.DB) error {
		for _, pref := range DefaultNotificationPreferences(userID, serverID) {
			var existing model.NotificationPreference
			err := db.Where("user_id = ? AND server_id = ? AND channel = ? AND event = ?", userID, serverID, pref.Channel, pref.Event).
				First(&existing).Error
			if err == nil {
				continue
			}
			if err != gorm.ErrRecordNotFound {
				return err
			}
			if err := db.Create(&pref).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Storage) GetUserNotificationPreferences(userID, serverID uint) ([]model.NotificationPreference, error) {
	var prefs []model.NotificationPreference
	if err := s.DB.Where("user_id = ? AND server_id = ?", userID, serverID).
		Order("channel asc, event asc").
		Find(&prefs).Error; err != nil {
		return nil, err
	}
	return prefs, nil
}

func (s *Storage) GetEnabledNotificationPreferences(serverID uint, event string) ([]model.NotificationPreference, error) {
	var prefs []model.NotificationPreference
	if err := s.DB.Where("server_id = ? AND event = ? AND enabled = ?", serverID, event, true).
		Order("channel asc").
		Find(&prefs).Error; err != nil {
		return nil, err
	}
	return prefs, nil
}

func (s *Storage) SaveUserNotificationPreferences(userID, serverID uint, prefs []model.NotificationPreference) error {
	return s.executeWrite(func(db *gorm.DB) error {
		return db.Transaction(func(tx *gorm.DB) error {
			for _, pref := range prefs {
				pref.UserID = userID
				pref.ServerID = serverID

				var existing model.NotificationPreference
				err := tx.Where("user_id = ? AND server_id = ? AND channel = ? AND event = ?", userID, serverID, pref.Channel, pref.Event).
					First(&existing).Error
				if err != nil && err != gorm.ErrRecordNotFound {
					return err
				}

				if err == gorm.ErrRecordNotFound {
					if err := tx.Create(&pref).Error; err != nil {
						return err
					}
					continue
				}

				if err := tx.Model(&existing).Updates(map[string]any{
					"enabled":     pref.Enabled,
					"destination": pref.Destination,
				}).Error; err != nil {
					return err
				}
			}
			return nil
		})
	})
}
