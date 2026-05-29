package database

import (
	"gorm.io/gorm"

	"github.com/ZaViBiS/isitdead/internal/model"
)

func (s *Storage) UpsertSSLCertificateStatus(status model.SSLCertificateStatus) error {
	return s.executeWrite(func(db *gorm.DB) error {
		var existing model.SSLCertificateStatus
		err := db.Where("server_id = ?", status.ServerID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			return db.Create(&status).Error
		}
		if err != nil {
			return err
		}
		return db.Model(&existing).Updates(map[string]any{
			"valid":                   status.Valid,
			"self_signed":             status.SelfSigned,
			"expires_at":              status.ExpiresAt,
			"days_remaining":          status.DaysRemaining,
			"issuer":                  status.Issuer,
			"fingerprint":             status.Fingerprint,
			"last_error":              status.LastError,
			"last_notified_threshold": status.LastNotifiedThreshold,
			"last_checked_at":         status.LastCheckedAt,
		}).Error
	})
}

func (s *Storage) GetSSLCertificateStatus(serverID uint) (*model.SSLCertificateStatus, error) {
	var status model.SSLCertificateStatus
	if err := s.DB.Where("server_id = ?", serverID).First(&status).Error; err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *Storage) GetSSLCertificateStatuses(serverIDs []uint) (map[uint]model.SSLCertificateStatus, error) {
	statuses := make(map[uint]model.SSLCertificateStatus, len(serverIDs))
	if len(serverIDs) == 0 {
		return statuses, nil
	}

	var rows []model.SSLCertificateStatus
	if err := s.DB.Where("server_id IN ?", serverIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, status := range rows {
		statuses[status.ServerID] = status
	}
	return statuses, nil
}

func (s *Storage) GetSSLEnabledServers() ([]model.Server, error) {
	var servers []model.Server
	if err := s.DB.Where("ssl_enabled = ?", true).Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}
