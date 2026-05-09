package checker

import (
	"time"

	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/rs/zerolog/log"
)

// Scheduler керує циклічними перевірками серверів
type Scheduler struct {
	storage *database.Storage
}

// NewScheduler створює новий екземпляр планувальника
func NewScheduler(db *database.Storage) *Scheduler {
	return &Scheduler{
		storage: db,
	}
}

// Start запускає моніторинг для всіх серверів у базі
func (s *Scheduler) Start() error {
	servers, err := s.storage.GetAllServers()
	if err != nil {
		return err
	}

	log.Info().Int("count", len(servers)).Msg("Starting scheduler for servers")

	for _, server := range servers {
		s.RunServerMonitor(server)
	}

	return nil
}

// RunServerMonitor запускає окрему горутину для моніторингу конкретного сервера
func (s *Scheduler) RunServerMonitor(srv model.Server) {
	go func() {
		log.Info().Str("server", srv.Name).Str("url", srv.URL).Int("interval", srv.CheckInterval).Msg("Monitoring started")

		// Перша перевірка при запуску
		s.performCheck(srv)

		ticker := time.NewTicker(time.Duration(srv.CheckInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			s.performCheck(srv)
		}
	}()
}

func (s *Scheduler) performCheck(srv model.Server) {
	status, latency := Check(srv.URL)

	log.Debug().
		Str("server", srv.Name).
		Str("status", status).
		Int64("latency", latency).
		Msg("Check completed")

	// 1. Зберігаємо історію перевірок через канал БД
	err := s.storage.AddCheckResult(model.CheckResult{
		ServerID:  srv.ID,
		Status:    status,
		Latency:   latency,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Error().Err(err).Uint("server_id", srv.ID).Msg("Failed to save check result")
	}

	// 2. Оновлюємо поточний стан сервера
	err = s.storage.UpdateServerStatus(srv.ID, status, latency)
	if err != nil {
		log.Error().Err(err).Uint("server_id", srv.ID).Msg("Failed to update server status")
	}
}
