// Package model визначає GORM структури даних для бази.
package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Username      string         `gorm:"uniqueIndex;not null" json:"username"`
	Email         string         `gorm:"uniqueIndex;not null" json:"email"`
	VerifiedEmail bool           `gorm:"not null" json:"-"`
	PasswordHash  string         `json:"-"`
	GoogleID      *string        `gorm:"uniqueIndex" json:"-"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type EmailVerification struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	Token     string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `json:"created_at"`
}
