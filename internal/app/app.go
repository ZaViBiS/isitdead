// Package app ініціалізує та запускає основні компоненти застосунку.
package app

import (
	"crypto/tls"
	"embed"
	"fmt"
	"net"
	"net/http"

	"github.com/ZaViBiS/isitdead/internal/api"
	"github.com/ZaViBiS/isitdead/internal/checker"
	"github.com/ZaViBiS/isitdead/internal/config"
	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/ZaViBiS/isitdead/internal/mail"
	"github.com/ZaViBiS/isitdead/internal/notify"
	"github.com/ZaViBiS/isitdead/internal/probe"
	"golang.org/x/crypto/acme/autocert"
)

type runnableServer interface {
	Listen(addr string) error
	Listener(ln net.Listener) error
}

type App struct {
	server      *api.Server
	probeServer *probe.Server
	scheduler   *checker.Scheduler
	config      *config.Config
	mailer      *mail.Mailer
}

func New(staticFiles embed.FS) (*App, error) {
	cfg := config.Load()
	if cfg.InstanceRole == config.RoleProbe {
		return &App{
			probeServer: probe.NewServer(cfg.Region, cfg.ProbeSecret),
			config:      cfg,
		}, nil
	}
	if cfg.InstanceRole != config.RoleMain {
		return nil, fmt.Errorf("unsupported INSTANCE_ROLE %q", cfg.InstanceRole)
	}

	// БД
	db, err := database.Init(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	// Планувальник перевірок
	sched := checker.NewScheduler(db)
	sched.SetLocalRegion(cfg.Region)
	if len(cfg.ProbeRegions) > 0 {
		sched.SetRegionalChecker(probe.NewClient(probeTargets(cfg.ProbeRegions), cfg.ProbeSecret))
	}

	// Поштовий сервіс
	mailer := mail.New(cfg)

	// сервіс нотифікацій
	senders := []notify.Sender{notify.NewEmailSender(mailer)}
	if cfg.TelegramAPIURL != "" {
		senders = append(senders, notify.NewTelegramSender(db, cfg.TelegramAPIURL, cfg.TelegramAPISecret))
	}
	notifier := notify.NewService(db, senders...)
	sched.SetNotifier(notifier)

	// Backend + Frontend (embed)
	server, err := api.NewWithConfig(db, sched, mailer, staticFiles, cfg)
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
	server := a.activeServer()

	if a.config.InstanceRole == config.RoleMain {
		// Запускаємо моніторинг у фоні
		if err := a.scheduler.Start(); err != nil {
			return err
		}
	}

	// локальна розробка — без SSL
	if a.config.Env == "dev" {
		// Запускаємо HTTP сервер
		return server.Listen(fmt.Sprintf(":%s", a.config.Port))
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

	return server.Listener(ln)
}

func (a *App) activeServer() runnableServer {
	if a.probeServer != nil {
		return a.probeServer
	}
	return a.server
}

func probeTargets(regions []config.ProbeRegion) []probe.Target {
	targets := make([]probe.Target, 0, len(regions))
	for _, region := range regions {
		targets = append(targets, probe.Target{
			Region: region.Name,
			URL:    region.URL,
		})
	}
	return targets
}
