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
	Status        string         `json:"status"`
	Latency       int64          `json:"latency"`
	CheckInterval int            `gorm:"not null" json:"check_interval"`
	LastCheck     *time.Time     `json:"last_check"`
	UserID        uint           `gorm:"not null" json:"user_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
