package database

import (
	"github.com/ZaViBiS/isitdead/internal/model"
	"gorm.io/gorm"
)

// AddServer додає новий сервер для моніторингу
func (s *Storage) AddServer(userID uint, name string, url string, checkType string, checkInterval int, timeout int) (*model.Server, error) {
	server := &model.Server{
		UserID:        userID,
		Name:          name,
		URL:           url,
		CheckType:     checkType,
		CheckInterval: checkInterval,
		Timeout:       timeout,
	}

	if err := s.executeWrite(func(db *gorm.DB) error {
		return db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(server).Error; err != nil {
				return err
			}
			for _, pref := range DefaultNotificationPreferences(userID, server.ID) {
				if err := tx.Create(&pref).Error; err != nil {
					return err
				}
			}
			return nil
		})
	}); err != nil {
		return nil, err
	}

	return server, nil
}

// GetUserServers повертає всі сервери певного користувача
func (s *Storage) GetUserServers(userID uint) ([]model.Server, error) {
	var servers []model.Server
	if err := s.DB.Where("user_id = ?", userID).Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (s *Storage) GetServerByID(serverID uint) (*model.Server, error) {
	var server model.Server
	if err := s.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

// UpdateServer оновлює дані сервера
func (s *Storage) UpdateServer(userID, serverID uint, name, url, checkType string, interval int, timeout int) (*model.Server, error) {
	var server model.Server
	// Перевіряємо, що сервер належить саме цьому користувачу
	if err := s.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server).Error; err != nil {
		return nil, err
	}

	server.Name = name
	server.URL = url
	server.CheckType = checkType
	server.CheckInterval = interval
	server.Timeout = timeout

	if err := s.executeWrite(func(db *gorm.DB) error {
		return db.Save(&server).Error
	}); err != nil {
		return nil, err
	}

	return &server, nil
}

// UpdateServerStatus оновлює тільки статус та затримку
func (s *Storage) UpdateServerStatus(serverID uint, status string, latency int64) error {
	return s.executeWrite(func(db *gorm.DB) error {
		return db.Model(&model.Server{}).Where("id = ?", serverID).Updates(map[string]any{
			"status":  status,
			"latency": latency,
		}).Error
	})
}

// DeleteServer видаляє сервер з бази даних
func (s *Storage) DeleteServer(userID, serverID uint) error {
	return s.executeWrite(func(db *gorm.DB) error {
		return db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("id = ? AND user_id = ?", serverID, userID).Delete(&model.Server{}).Error; err != nil {
				return err
			}
			return tx.Where("server_id = ? AND user_id = ?", serverID, userID).Delete(&model.NotificationPreference{}).Error
		})
	})
}

// GetAllServers повертає абсолютно всі сервери з бази (для планувальника/checker)
func (s *Storage) GetAllServers() ([]model.Server, error) {
	var servers []model.Server
	if err := s.DB.Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}
