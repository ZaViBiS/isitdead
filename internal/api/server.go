package api

import (
	"embed"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/checker"
	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/rs/zerolog/log"
)

type Server struct {
	App       *fiber.App
	DB        *database.Storage
	scheduler *checker.Scheduler
}

// New повертає готовий backend для сайту
func New(db *database.Storage, sched *checker.Scheduler, staticFiles embed.FS) (*Server, error) {
	app := fiber.New()

	s := &Server{
		App:       app,
		DB:        db,
		scheduler: sched,
	}

	// Логування запитів
	app.Use(func(c fiber.Ctx) error {
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
