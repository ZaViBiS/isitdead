// Package app ініціалізує та запускає основні компоненти застосунку.
package app

import (
	"crypto/tls"
	"embed"
	"fmt"
	"net/http"

	"github.com/ZaViBiS/isitdead/internal/api"
	"github.com/ZaViBiS/isitdead/internal/checker"
	"github.com/ZaViBiS/isitdead/internal/config"
	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/ZaViBiS/isitdead/internal/mail"
	"github.com/ZaViBiS/isitdead/internal/notify"
	"golang.org/x/crypto/acme/autocert"
)

type App struct {
	server    *api.Server
	scheduler *checker.Scheduler
	config    *config.Config
	mailer    *mail.Mailer
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

	// Поштовий сервіс
	mailer := mail.New(cfg)
	notifier := notify.NewService(db, notify.NewEmailSender(mailer))
	sched.SetNotifier(notifier)

	// Backend + Frontend (embed)
	server, err := api.New(db, sched, mailer, staticFiles)
	if err != nil {
		return nil, err
	}

	return &App{
		server:    server,
		scheduler: sched,
		config:    cfg,
		mailer:    mailer,
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
		return a.server.Listen(fmt.Sprintf(":%s", a.config.Port))
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
