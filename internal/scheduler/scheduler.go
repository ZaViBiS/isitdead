// Package scheduler відповідає за планування перевірок серверів.
package scheduler

import (
	"time"

	"github.com/ZaViBiS/isitdead/internal/checker"
	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/rs/zerolog/log"
)

type Scheduler struct {
	DB *database.Storage
}

func New(db *database.Storage) *Scheduler {
	return &Scheduler{
		DB: db,
	}
}

func (s *Scheduler) Run() error {
	servers, err := s.DB.GetAllServers()
	if err != nil {
		return err
	}

	for _, server := range servers {
		go func() {
			ticker := time.NewTicker(time.Duration(server.CheckInterval))
			for range ticker.C {
				status, latency := checker.Check(server.URL)
				log.Info().Str("status", status).Int64("latency", latency).Msg("result")
			}
		}()
	}

	return nil
}
