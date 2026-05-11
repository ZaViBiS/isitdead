// Package app ініціалізує та запускає основні компоненти застосунку.
package app

import (
	"crypto/tls"
	"embed"
	"net/http"

	"github.com/ZaViBiS/isitdead/internal/api"
	"github.com/ZaViBiS/isitdead/internal/checker"
	"github.com/ZaViBiS/isitdead/internal/config"
	"github.com/ZaViBiS/isitdead/internal/database"
	"golang.org/x/crypto/acme/autocert"
)

type App struct {
	server    *api.Server
	scheduler *checker.Scheduler
	config    *config.Config
}

func New(staticFiles embed.FS) (*App, error) {
	cfg := config.Load()

	// БД
	db, err := database.Init(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	// Планувальник перевірок
	sched := checker.NewScheduler(db)

	// Backend + Frontend (embed)
	server, err := api.New(db, sched, staticFiles)
	if err != nil {
		return nil, err
	}

	return &App{
		server:    server,
		scheduler: sched,
		config:    cfg,
	}, nil
}

func (a *App) Run() error {
	// Запускаємо моніторинг у фоні
	if err := a.scheduler.Start(); err != nil {
		return err
	}

	// локальна розробка — без SSL
	if a.config.Env == "dev" {
		// Запускаємо HTTP сервер
		return a.server.Listen(":8080")
	}

	// Продакшен — автоматичний SSL через Let's Encrypt
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(a.config.Domain),
		Cache:      autocert.DirCache("./certs"),
	}

	tlsCfg := &tls.Config{
		GetCertificate: m.GetCertificate,
		MinVersion:     tls.VersionTLS12,
	}

	ln, err := tls.Listen("tcp", ":443", tlsCfg)
	if err != nil {
		return err
	}

	// Редирект HTTP → HTTPS
	go http.ListenAndServe(":80", m.HTTPHandler(nil))

	return a.server.Listener(ln)
}
