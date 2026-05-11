package checker

import (
	"context"
	"sync"
	"time"

	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/rs/zerolog/log"
)

// Scheduler керує циклічними перевірками серверів
type Scheduler struct {
	storage  *database.Storage
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.Mutex
	monitors map[uint]monitorControl
}

type monitorControl struct {
	cancel context.CancelFunc
	done   chan struct{}
}

// NewScheduler створює новий екземпляр планувальника
func NewScheduler(db *database.Storage) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		storage:  db,
		ctx:      ctx,
		cancel:   cancel,
		monitors: make(map[uint]monitorControl),
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

// Stop зупиняє всі процеси моніторингу
func (s *Scheduler) Stop() {
	s.cancel()
	s.mu.Lock()
	for id, monitor := range s.monitors {
		monitor.cancel()
		delete(s.monitors, id)
	}
	s.mu.Unlock()
	s.wg.Wait()
}

// RunServerMonitor запускає окрему горутину для моніторингу конкретного сервера
func (s *Scheduler) RunServerMonitor(srv model.Server) {
	s.StopServerMonitor(srv.ID)

	ctx, cancel := context.WithCancel(s.ctx)
	done := make(chan struct{})

	s.mu.Lock()
	s.monitors[srv.ID] = monitorControl{cancel: cancel, done: done}
	s.wg.Add(1)
	s.mu.Unlock()

	go func() {
		defer func() {
			close(done)
			s.mu.Lock()
			if current, ok := s.monitors[srv.ID]; ok && current.done == done {
				delete(s.monitors, srv.ID)
			}
			s.mu.Unlock()
			s.wg.Done()
		}()

		log.Info().Str("server", srv.Name).Str("url", srv.URL).Int("interval", srv.CheckInterval).Msg("Monitoring started")

		// Перша перевірка при запуску
		s.performCheck(srv)

		ticker := time.NewTicker(time.Duration(srv.CheckInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info().Str("server", srv.Name).Msg("Monitoring stopped")
				return
			case <-ticker.C:
				s.performCheck(srv)
			}
		}
	}()
}

// StopServerMonitor зупиняє моніторинг конкретного сервера.
func (s *Scheduler) StopServerMonitor(serverID uint) {
	s.mu.Lock()
	monitor, ok := s.monitors[serverID]
	if ok {
		delete(s.monitors, serverID)
	}
	s.mu.Unlock()

	if ok {
		monitor.cancel()
	}
}

func (s *Scheduler) performCheck(srv model.Server) {
	status, latency := Check(srv.CheckType, srv.URL)

	log.Debug().
		Str("server", srv.URL).
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
