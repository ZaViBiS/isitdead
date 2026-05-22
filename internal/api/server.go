package api

import (
	"embed"
	"io"
	"io/fs"
	"net"
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/checker"
	"github.com/ZaViBiS/isitdead/internal/config"
	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/rs/zerolog/log"
)

type VerificationMailer interface {
	SendVerificationEmail(to, token string) error
}

type Server struct {
	App       *fiber.App
	DB        *database.Storage
	Scheduler *checker.Scheduler
	Config    *config.Config
	Mailer    VerificationMailer
}

// New повертає готовий backend для сайту
func New(db *database.Storage, sched *checker.Scheduler, mailer VerificationMailer, staticFiles embed.FS) (*Server, error) {
	return NewWithConfig(db, sched, mailer, staticFiles, config.Load())
}

func NewWithConfig(db *database.Storage, sched *checker.Scheduler, mailer VerificationMailer, staticFiles embed.FS, cfg *config.Config) (*Server, error) {
	app := fiber.New()

	s := &Server{
		App:       app,
		DB:        db,
		Scheduler: sched,
		Config:    cfg,
		Mailer:    mailer,
	}

	// Логування запитів
	app.Use(func(c fiber.Ctx) error {
		c.Set("Referrer-Policy", "no-referrer")
		start := time.Now()
		err := c.Next()

		statusCode := c.Response().StatusCode()
		event := log.Info()
		if err != nil {
			event = log.Error().Err(err)
		}

		event.Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", statusCode).
			Dur("latency", time.Since(start)).
			Str("ip", c.IP()).
			Str("user_agent", string(c.Request().Header.UserAgent())).
			Msg("request")

		return err
	})

	// Налаштовуємо доступ до статичних файлів через embed
	distFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		return nil, err
	}

	// Реєстрація API маршрутів
	s.setupRoutes()

	// Hashed build assets can be cached aggressively, but HTML must stay fresh so
	// clients do not keep references to chunks from an older deployment.
	app.Use(func(c fiber.Ctx) error {
		err := c.Next()
		path := c.Path()
		contentType := string(c.Response().Header.ContentType())

		switch {
		case strings.HasPrefix(path, "/_app/immutable/"):
			c.Set("Cache-Control", "public, max-age=31536000, immutable")
		case strings.HasPrefix(contentType, "text/html"):
			c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}

		return err
	})

	// Обслуговування статичних файлів через middleware/static
	app.Get("/*", static.New("", static.Config{
		FS: distFS,
	}))

	// Fallback для SPA (SvelteKit)
	app.Use(func(c fiber.Ctx) error {
		// Використовуємо іншу перевірку, бо маршрути API вже зареєстровані
		path := c.Path()
		if path == "/api" || strings.HasPrefix(path, "/api/") {
			return c.Next()
		}
		if strings.HasPrefix(path, "/_app/") {
			return c.SendStatus(fiber.StatusNotFound)
		}
		if !isKnownSPARoute(path) {
			return c.Status(fiber.StatusNotFound).Type("html").SendString(s.siteNotFoundHTML(path))
		}

		index, err := distFS.Open("index.html")
		if err != nil {
			return c.Next()
		}
		defer index.Close()

		content, _ := io.ReadAll(index)
		c.Set("Content-Type", "text/html")
		return c.Send(content)
	})

	return s, nil
}

// Listen запускає сервер
func (s *Server) Listen(addr string) error {
	return s.App.Listen(addr)
}

// ListenMutualTLS запускає сервер з кастомним listener (для SSL)
func (s *Server) Listener(ln net.Listener) error {
	return s.App.Listener(ln)
}
