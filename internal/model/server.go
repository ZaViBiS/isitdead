// Package model визначає GORM структури даних для бази.
package model

import (
	"time"

	"gorm.io/gorm"
)

type Server struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"not null" json:"name"`
	URL           string         `gorm:"not null" json:"url"`
	CheckType     string         `gorm:"not null;default:'http'" json:"check_type"` // 'http', 'ping', or 'links'
	CheckInterval int            `gorm:"not null" json:"check_interval"`
	Timeout       int            `gorm:"not null" json:"timeout"`
	SlowThreshold int            `gorm:"not null;default:300" json:"slow_threshold"`
	SSLEnabled    bool           `gorm:"not null;default:false" json:"ssl_enabled"`
	UserID        uint           `gorm:"not null" json:"user_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
