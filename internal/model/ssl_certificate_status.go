package model

import "time"

type SSLCertificateStatus struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	ServerID              uint       `gorm:"not null;uniqueIndex" json:"server_id"`
	Valid                 bool       `gorm:"not null" json:"valid"`
	SelfSigned            bool       `gorm:"not null" json:"self_signed"`
	ExpiresAt             *time.Time `json:"expires_at"`
	DaysRemaining         int        `json:"days_remaining"`
	Issuer                string     `json:"issuer"`
	Fingerprint           string     `json:"fingerprint"`
	LastError             string     `json:"last_error"`
	LastNotifiedThreshold int        `json:"last_notified_threshold"`
	LastCheckedAt         time.Time  `json:"last_checked_at"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
