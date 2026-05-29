package probe

import (
	"net"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/ZaViBiS/isitdead/internal/checker"
)

type Server struct {
	App    *fiber.App
	region string
	secret string
}

func NewServer(region, secret string) *Server {
	region = normalizeRegionName(region, "probe")

	s := &Server{
		App:    fiber.New(),
		region: region,
		secret: secret,
	}
	s.setupRoutes()
	return s
}

func (s *Server) Listen(addr string) error {
	return s.App.Listen(addr)
}

func (s *Server) Listener(ln net.Listener) error {
	return s.App.Listener(ln)
}

func (s *Server) setupRoutes() {
	api := s.App.Group("/api")
	api.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})
	api.Post("/probe/check", s.handleCheck)
}

func (s *Server) handleCheck(c fiber.Ctx) error {
	if s.secret != "" && c.Get(SecretHeader) != s.secret {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	var req CheckRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(CheckResponse{
			Region: s.region,
			Error:  "invalid request body",
		})
	}
	if strings.TrimSpace(req.URL) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(CheckResponse{
			Region: s.region,
			Error:  "url is required",
		})
	}
	if req.CheckType == "" {
		req.CheckType = "http"
	}
	target, err := checker.ValidateMonitorTarget(req.CheckType, req.URL)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(CheckResponse{
			Region: s.region,
			Error:  err.Error(),
		})
	}
	req.URL = target

	start := time.Now()
	status, latency := checker.Check(req.CheckType, req.URL, req.Timeout)
	log.Debug().
		Str("region", s.region).
		Str("target", req.URL).
		Str("status", status).
		Int64("latency", latency).
		Dur("duration", time.Since(start)).
		Msg("probe check completed")

	return c.JSON(CheckResponse{
		Region:  s.region,
		Status:  status,
		Latency: latency,
	})
}
