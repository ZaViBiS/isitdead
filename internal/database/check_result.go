package database

import (
	"github.com/ZaViBiS/isitdead/internal/model"
	"gorm.io/gorm"
)

// AddCheckResult додає новий результат перевірки в базу даних через канал запису
func (s *Storage) AddCheckResult(result model.CheckResult) error {
	return s.executeWrite(func(db *gorm.DB) error {
		return db.Create(&result).Error
	})
}

// GetLatestResults повертає останні n результатів для сервера
func (s *Storage) GetLatestResults(serverID uint, limit int) ([]model.CheckResult, error) {
	var results []model.CheckResult
	err := s.DB.Where("server_id = ?", serverID).
		Order("created_at desc").
		Limit(limit).
		Find(&results).Error
	return results, err
}
