// Package model визначає GORM структури даних для бази.
package model

import (
	"time"
)

const (
	CheckRegionGlobal = "global"
	CheckRegionAll    = "all"
)

// CheckResult представляє результати перевірки сервера.
type CheckResult struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ServerID  uint      `gorm:"not null;index" json:"server_id"`
	Region    string    `gorm:"not null;default:'global';index" json:"region"`
	Status    string    `gorm:"not null" json:"status"`
	Latency   int64     `json:"latency"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
