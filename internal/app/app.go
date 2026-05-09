// Package app ініціалізує та запускає основні компоненти застосунку.
package app

import (
	"embed"

	"github.com/ZaViBiS/isitdead/internal/api"
	"github.com/ZaViBiS/isitdead/internal/checker"
	"github.com/ZaViBiS/isitdead/internal/database"
)

type App struct {
	server    *api.Server
	scheduler *checker.Scheduler
}

func New(staticFiles embed.FS) (*App, error) {
	// БД
	db, err := database.Init("/tmp/isitdead.db")
	if err != nil {
		return nil, err
	}

	// Backend + Frontend (embed)
	server, err := api.New(db, staticFiles)
	if err != nil {
		return nil, err
	}

	// Планувальник перевірок
	sched := checker.NewScheduler(db)

	return &App{
		server:    server,
		scheduler: sched,
	}, nil
}

func (a *App) Run() error {
	// Запускаємо моніторинг у фоні
	if err := a.scheduler.Start(); err != nil {
		return err
	}

	// Запускаємо HTTP сервер
	return a.server.Listen(":8080")
}
