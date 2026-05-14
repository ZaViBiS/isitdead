package api

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func (s *Server) handleGetServers(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	servers, err := s.DB.GetUserServers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch servers"})
	}
	return c.JSON(servers)
}

func (s *Server) handleAddServer(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	var serverRequest serverRequest

	if err := c.Bind().Body(&serverRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if serverRequest.CheckType == "" {
		serverRequest.CheckType = "http"
	}

	if serverRequest.CheckInterval < 10 {
		serverRequest.CheckInterval = 300 // default
	}
	if serverRequest.Timeout <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Timeout is required"})
	}

	server, err := s.DB.AddServer(userID, serverRequest.Name, serverRequest.URL, serverRequest.CheckType, serverRequest.CheckInterval, serverRequest.Timeout)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Public slug is already used"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create server"})
	}

	// Запускаємо моніторинг для нового сервера негайно
	if s.Scheduler != nil {
		s.Scheduler.RunServerMonitor(*server)
	}

	return c.Status(fiber.StatusCreated).JSON(server)
}

func (s *Server) handleUpdateServer(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	serverID, err := parseServerID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	var req serverRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Timeout <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Timeout is required"})
	}

	server, err := s.DB.UpdateServer(userID, serverID, req.Name, req.URL, req.CheckType, req.CheckInterval, req.Timeout)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Public slug is already used"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update server"})
	}

	// Оновлюємо планувальник, якщо він активний
	if s.Scheduler != nil {
		s.Scheduler.RunServerMonitor(*server)
	}

	return c.JSON(server)
}

func (s *Server) handleDeleteServer(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	serverID, err := parseServerID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	err = s.DB.DeleteServer(userID, serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete server"})
	}

	if s.Scheduler != nil {
		s.Scheduler.StopServerMonitor(serverID)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
