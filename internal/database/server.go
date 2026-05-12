package database

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/ZaViBiS/isitdead/internal/model"
	"gorm.io/gorm"
)

// AddServer додає новий сервер для моніторингу
func (s *Storage) AddServer(userID uint, name, url, checkType string, checkInterval int, public bool, publicSlug string) (*model.Server, error) {
	slug, err := s.preparePublicSlug(0, name, public, publicSlug)
	if err != nil {
		return nil, err
	}

	server := &model.Server{
		UserID:        userID,
		Name:          name,
		URL:           url,
		CheckType:     checkType,
		Public:        public,
		PublicSlug:    slug,
		Status:        "unknown",
		CheckInterval: checkInterval,
	}

	err = s.executeWrite(func(db *gorm.DB) error {
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

// UpdateServer оновлює дані сервера
func (s *Storage) UpdateServer(userID, serverID uint, name, url, checkType string, interval int, public bool, publicSlug string) (*model.Server, error) {
	var server model.Server
	// Перевіряємо, що сервер належить саме цьому користувачу
	if err := s.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server).Error; err != nil {
		return nil, err
	}

	slug, err := s.preparePublicSlug(serverID, name, public, publicSlug)
	if err != nil {
		return nil, err
	}

	server.Name = name
	server.URL = url
	server.CheckType = checkType
	server.CheckInterval = interval
	server.Public = public
	server.PublicSlug = slug

	err = s.executeWrite(func(db *gorm.DB) error {
		return db.Save(&server).Error
	})
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (s *Storage) GetPublicServerBySlug(slug string) (*model.Server, error) {
	var server model.Server
	if err := s.DB.Where("public = ? AND public_slug = ?", true, slug).First(&server).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *Storage) GetPublicServers() ([]model.Server, error) {
	var servers []model.Server
	if err := s.DB.Where("public = ?", true).Order("public_slug asc").Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (s *Storage) UpdatePublicServer(serverID uint, public bool, publicSlug string) (*model.Server, error) {
	var server model.Server
	if err := s.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return nil, err
	}

	slug, err := s.preparePublicSlug(serverID, server.Name, public, publicSlug)
	if err != nil {
		return nil, err
	}

	server.Public = public
	server.PublicSlug = slug

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

func (s *Storage) preparePublicSlug(serverID uint, name string, public bool, requested string) (string, error) {
	if !public {
		return "", nil
	}

	base := slugify(requested)
	if base == "" {
		base = slugify(name)
	}
	if base == "" {
		base = "monitor"
	}

	slug := base
	for i := 2; ; i++ {
		var count int64
		query := s.DB.Model(&model.Server{}).Where("public_slug = ?", slug)
		if serverID != 0 {
			query = query.Where("id != ?", serverID)
		}
		if err := query.Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}
		if requested != "" {
			return "", gorm.ErrDuplicatedKey
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false

	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if r <= unicode.MaxASCII {
				b.WriteRune(r)
				lastDash = false
			}
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(b.String(), "-")
}
