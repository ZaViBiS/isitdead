package database

import (
	"sync"

	"github.com/ZaViBiS/isitdead/internal/database/model"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NOTE: можливо є сенс змінити назву усього пакету на storage, або змінити назву структури на database
type Storage struct {
	DB *gorm.DB
	mu *sync.Mutex
}

// Init ініціалізує підключення до бази даних SQLite і повертає структуру Storage
func Init(dbPath string) (*Storage, error) {
	log.Info().Str("path", dbPath).Msg("Connecting to database")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		// Використовуємо стандартний логер GORM, але ви можете налаштувати кастомний,
		// якщо захочете повну інтеграцію zerolog всередині запитів.
		// Наразі для початкових функцій залишаємо базовий конфіг.
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	// автоматично створює/оновлює таблиці
	if err := db.AutoMigrate(&model.User{}, &model.Server{}); err != nil {
		// HACK: нормальна обробка
		panic(err)
	}

	return &Storage{
		DB: db,
		mu: &sync.Mutex{},
	}, nil
}

// Close закриває підключення до бази даних
func (s *Storage) Close() error {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
