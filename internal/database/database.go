// Package database забезпечує роботу з базою даних SQLite.
package database

import (
	"github.com/ZaViBiS/isitdead/internal/database/model"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NOTE: можливо є сенс змінити назву усього пакету на storage, або змінити назву структури на database
// dbTask представляє одиницю роботи для запису в базу даних
type dbTask struct {
	action  func(db *gorm.DB) error
	errChan chan error
}

// Storage забезпечує доступ до бази даних
type Storage struct {
	DB         *gorm.DB
	writerChan chan dbTask
}

// Init ініціалізує підключення до бази даних SQLite, повертає структуру Storage і запускає writeWorker для запису з каналу writerChan
func Init(dbPath string) (*Storage, error) {
	log.Info().Str("path", dbPath).Msg("Connecting to database")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	// автоматично створює/оновлює таблиці
	if err := db.AutoMigrate(&model.User{}, &model.Server{}, &model.CheckResult{}); err != nil {
		panic(err)
	}

	s := &Storage{
		DB:         db,
		writerChan: make(chan dbTask, 100), // Буфер для запитів на запис
	}

	// Запуск воркера для обробки записів
	go s.runWorker()

	return s, nil
}

// runWorker послідовно виконує завдання на запис із каналу
func (s *Storage) runWorker() {
	log.Info().Msg("Database writer worker started")
	for task := range s.writerChan {
		err := task.action(s.DB)
		task.errChan <- err
	}
	log.Info().Msg("Database writer worker stopped")
}

// executeWrite відправляє завдання на запис у канал і чекає на результат
func (s *Storage) executeWrite(action func(db *gorm.DB) error) error {
	errChan := make(chan error, 1)
	s.writerChan <- dbTask{
		action:  action,
		errChan: errChan,
	}
	return <-errChan
}

// Close закриває підключення до бази даних
func (s *Storage) Close() error {
	close(s.writerChan)
	sqlDB, err := s.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
