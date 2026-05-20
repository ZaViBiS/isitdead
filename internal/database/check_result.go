package database

import (
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
	"gorm.io/gorm"
)

type HistorySummary struct {
	Total      int64
	Online     int64
	AvgLatency float64
}

// AddCheckResult додає новий результат перевірки в базу даних через канал запису
func (s *Storage) AddCheckResult(result model.CheckResult) error {
	normalizeCheckResultRegion(&result)
	return s.executeWrite(func(db *gorm.DB) error {
		return db.Create(&result).Error
	})
}

// GetHistorySince повертає історію результатів перевірки для сервера після заданого часу.
func (s *Storage) GetHistorySince(serverID uint, since time.Time) ([]model.CheckResult, error) {
	return s.GetHistorySinceForRegion(serverID, model.CheckRegionGlobal, since)
}

func (s *Storage) GetHistorySinceForRegion(serverID uint, region string, since time.Time) ([]model.CheckResult, error) {
	var results []model.CheckResult
	query := s.DB.Where("server_id = ? AND created_at > ?", serverID, since)
	err := filterCheckRegion(query, region).
		Order("created_at asc").
		Find(&results).Error
	return results, err
}

// GetRecentHistory повертає останні limit результатів перевірки в хронологічному порядку.
func (s *Storage) GetRecentHistory(serverID uint, limit int) ([]model.CheckResult, error) {
	return s.GetRecentHistoryForRegion(serverID, model.CheckRegionGlobal, limit)
}

func (s *Storage) GetRecentHistoryForRegion(serverID uint, region string, limit int) ([]model.CheckResult, error) {
	var results []model.CheckResult
	query := s.DB.Where("server_id = ?", serverID)
	err := filterCheckRegion(query, region).
		Order("created_at desc").
		Limit(limit).
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	for left, right := 0, len(results)-1; left < right; left, right = left+1, right-1 {
		results[left], results[right] = results[right], results[left]
	}

	return results, nil
}

func (s *Storage) GetHistorySummarySince(serverID uint, since time.Time) (HistorySummary, error) {
	return s.GetHistorySummarySinceForRegion(serverID, model.CheckRegionGlobal, since)
}

func (s *Storage) GetHistorySummarySinceForRegion(serverID uint, region string, since time.Time) (HistorySummary, error) {
	var summary HistorySummary
	query := s.DB.Model(&model.CheckResult{}).
		Select(`
			COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN status LIKE '2%' OR status = 'Connected' THEN 1 ELSE 0 END), 0) AS online,
			COALESCE(AVG(latency), 0) AS avg_latency
		`).
		Where("server_id = ? AND created_at > ?", serverID, since)
	err := filterCheckRegion(query, region).
		Scan(&summary).Error
	return summary, err
}

func (s *Storage) GetLatestCheckResult(serverID uint) (*model.CheckResult, error) {
	return s.GetLatestCheckResultForRegion(serverID, model.CheckRegionGlobal)
}

func (s *Storage) GetLatestCheckResultForRegion(serverID uint, region string) (*model.CheckResult, error) {
	var result model.CheckResult
	query := s.DB.Where("server_id = ?", serverID)
	err := filterCheckRegion(query, region).
		Order("created_at desc").
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetIncidents повертає тільки результати з помилками для сервера
func (s *Storage) GetIncidents(serverID uint, limit int) ([]model.CheckResult, error) {
	return s.GetIncidentsForRegion(serverID, model.CheckRegionGlobal, limit)
}

func (s *Storage) GetIncidentsForRegion(serverID uint, region string, limit int) ([]model.CheckResult, error) {
	var results []model.CheckResult
	query := s.DB.Where("server_id = ?", serverID).
		Where("status NOT LIKE '2%' AND status != 'Connected'").
		Order("created_at desc")
	query = filterCheckRegion(query, region)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&results).Error
	return results, err
}

func normalizeCheckResultRegion(result *model.CheckResult) {
	result.Region = normalizeRegion(result.Region)
}

func normalizeRegion(region string) string {
	region = strings.TrimSpace(region)
	if region == "" {
		return model.CheckRegionGlobal
	}
	return region
}

func filterCheckRegion(query *gorm.DB, region string) *gorm.DB {
	region = normalizeRegion(region)
	if region == model.CheckRegionAll {
		return query
	}
	return query.Where("region = ?", region)
}
