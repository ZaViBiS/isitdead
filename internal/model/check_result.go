// Package model визначає GORM структури даних для бази.
package model

import (
	"time"
)

// CheckResult представляє результати перевірки сервера.
type CheckResult struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ServerID  uint      `gorm:"not null;index" json:"server_id"`
	Status    string    `gorm:"not null" json:"status"`
	Latency   int64     `json:"latency"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
