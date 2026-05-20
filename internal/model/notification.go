package model

import "time"

const (
	NotificationChannelEmail    = "email"
	NotificationChannelTelegram = "telegram"
	NotificationEventDown       = "down"
	NotificationEventUp         = "recovered"
	NotificationEventSSL30d     = "ssl_30d"
	NotificationEventSSL14d     = "ssl_14d"
	NotificationEventSSL7d      = "ssl_7d"
)

type NotificationPreference struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index;uniqueIndex:idx_notification_pref" json:"user_id"`
	ServerID    uint      `gorm:"not null;index;uniqueIndex:idx_notification_pref" json:"server_id"`
	Channel     string    `gorm:"not null;uniqueIndex:idx_notification_pref" json:"channel"`
	Event       string    `gorm:"not null;uniqueIndex:idx_notification_pref" json:"event"`
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`
	Destination string    `json:"destination"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
