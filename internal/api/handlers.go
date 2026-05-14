package api

import (
	"github.com/gofiber/fiber/v3"
)

func (s *Server) handlePing(c fiber.Ctx) error {
	return c.SendString("pong")
}
