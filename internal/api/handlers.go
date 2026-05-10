package api

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func (s *Server) handleGetServerResults(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	serverIDStr := c.Params("id")

	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	// Перевіряємо приналежність сервера користувачу перед видачею результатів
	servers, err := s.DB.GetUserServers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	owned := false
	for _, srv := range servers {
		if srv.ID == uint(serverID) {
			owned = true
			break
		}
	}

	if !owned {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	results, err := s.DB.GetHistory(uint(serverID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch results"})
	}

	limitStr := c.Query("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 && limit < len(results) {
			// Отримуємо останні 'limit' результатів (кінець масиву)
			results = results[len(results)-limit:]
		}
	}

	return c.JSON(results)
}

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
	var req struct {
		Name          string `json:"name"`
		URL           string `json:"url"`
		CheckInterval int    `json:"check_interval"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.CheckInterval < 10 {
		req.CheckInterval = 60 // default
	}

	server, err := s.DB.AddServer(userID, req.Name, req.URL, req.CheckInterval)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not add server"})
	}

	// Запускаємо моніторинг для нового сервера негайно
	if s.scheduler != nil {
		s.scheduler.RunServerMonitor(*server)
	}

	return c.Status(fiber.StatusCreated).JSON(server)
}

func (s *Server) handleDeleteServer(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	serverIDStr := c.Params("id")

	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	err = s.DB.DeleteServer(userID, uint(serverID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not delete server"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (s *Server) handlePing(c fiber.Ctx) error {
	return c.SendString("pong")
}
