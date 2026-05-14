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

// GetHistory повертає всю історію результатів перевірки для сервера
// TODO: додати ліміт
func (s *Storage) GetHistory(serverID uint) ([]model.CheckResult, error) {
	var results []model.CheckResult
	err := s.DB.Where("server_id = ?", serverID).
		Order("created_at asc").
		Find(&results).Error
	return results, err
}

// GetIncidents повертає тільки результати з помилками для сервера
func (s *Storage) GetIncidents(serverID uint, limit int) ([]model.CheckResult, error) {
	var results []model.CheckResult
	query := s.DB.Where("server_id = ?", serverID).
		Where("status NOT LIKE '2%' AND status != 'Connected'").
		Order("created_at desc")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&results).Error
	return results, err
}
