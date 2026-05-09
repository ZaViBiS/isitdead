package database

import (
	"github.com/ZaViBiS/isitdead/internal/model"
	"gorm.io/gorm"
)

// AddServer додає новий сервер для моніторингу
func (s *Storage) AddServer(userID uint, name, url string, checkInterval int) (*model.Server, error) {
	server := &model.Server{
		UserID:        userID,
		Name:          name,
		URL:           url,
		Status:        "unknown",
		CheckInterval: checkInterval,
	}

	err := s.executeWrite(func(db *gorm.DB) error {
		return db.Create(server).Error
	})
	if err != nil {
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

// UpdateServer оновлює дані сервера (ім'я або URL)
func (s *Storage) UpdateServer(userID, serverID uint, name, url string) (*model.Server, error) {
	var server model.Server
	// Перевіряємо, що сервер належить саме цьому користувачу
	if err := s.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server).Error; err != nil {
		return nil, err
	}

	server.Name = name
	server.URL = url

	err := s.executeWrite(func(db *gorm.DB) error {
		return db.Save(&server).Error
	})
	if err != nil {
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
		return db.Where("id = ? AND user_id = ?", serverID, userID).Delete(&model.Server{}).Error
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
