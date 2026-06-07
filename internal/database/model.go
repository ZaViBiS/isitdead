package database

import (
	"time"

	"gorm.io/gorm"
)

// Account related structs

type User struct {
	ID                       uint           `gorm:"primaryKey" json:"id"`
	Username                 string         `gorm:"uniqueIndex;not null" json:"username"`
	Email                    string         `gorm:"uniqueIndex;not null" json:"email"`
	VerifiedEmail            bool           `gorm:"not null" json:"-"`
	PasswordHash             string         `json:"-"`
	GoogleID                 *string        `gorm:"uniqueIndex" json:"-"`
	Plan                     string         `gorm:"not null;default:'free'" json:"plan"`
	StripeCustomerID         string         `gorm:"index" json:"-"`
	StripeSubscriptionID     string         `gorm:"index" json:"-"`
	StripeSubscriptionStatus string         `json:"stripe_subscription_status"`
	StripePriceID            string         `json:"-"`
	PlanCurrentPeriodEnd     *time.Time     `json:"plan_current_period_end,omitempty"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`
}

type EmailVerification struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	Token     string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `json:"created_at"`
}

// Server

type Server struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"not null" json:"name"`
	URL           string         `gorm:"not null" json:"url"`
	CheckType     string         `gorm:"not null" json:"check_type"` // 'http' or 'ping'
	CheckInterval int            `gorm:"not null" json:"check_interval"`
	Timeout       int            `gorm:"not null" json:"timeout"`
	SlowThreshold int            `gorm:"not null" json:"slow_threshold"`
	SSLEnabled    bool           `gorm:"not null" json:"ssl_enabled"`
	UserID        uint           `gorm:"not null" json:"user_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// Check result

type CheckResult struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ServerID  uint      `gorm:"not null;index" json:"server_id"`
	Region    string    `gorm:"not null;index" json:"region"`
	Status    string    `gorm:"not null" json:"status"`
	Latency   int64     `json:"latency"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// Telegram

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
