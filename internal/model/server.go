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
	Public        bool           `gorm:"not null;default:false" json:"public"`
	PublicSlug    string         `gorm:"index" json:"public_slug"`
	Status        string         `json:"status"`
	Latency       int64          `json:"latency"`
	CheckInterval int            `gorm:"not null" json:"check_interval"`
	Timeout       int            `gorm:"not null;default:10" json:"timeout"`
	LastCheck     *time.Time     `json:"last_check"`
	UserID        uint           `gorm:"not null" json:"user_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
