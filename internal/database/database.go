package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

func New(databaseURL string) (*Storage, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Migrate() error {
	return s.db.AutoMigrate(
		&User{},
		&EmailVerification{},
		&Server{},
		&CheckResult{},
		&TelegramAccount{},
		&TelegramLinkToken{},
	)
}
