package model

import (
	"time"

	"gorm.io/gorm"
)

type TelegramAccount struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;uniqueIndex" json:"user_id"`
	ChatID    int64          `gorm:"not null;uniqueIndex" json:"chat_id"`
	LinkedAt  time.Time      `json:"linked_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TelegramLinkToken цей токен створюєтся на backend і коли людина переходить на бота, телеграм бот присилає його через api
type TelegramLinkToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time
}
