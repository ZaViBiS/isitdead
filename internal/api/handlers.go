package api

import (
	"strconv"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
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

	// Фільтрація за часом (наприклад, за останні N годин)
	hoursStr := c.Query("hours")
	if hoursStr != "" {
		hours, err := strconv.Atoi(hoursStr)
		if err == nil && hours > 0 {
			since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
			filtered := []model.CheckResult{}
			for _, r := range results {
				if r.CreatedAt.UTC().After(since) {
					filtered = append(filtered, r)
				}
			}
			results = filtered
		}
	}

	limitStr := c.Query("limit")
	if limitStr != "" && hoursStr == "" { // limit працює тільки якщо не вказано години
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 && limit < len(results) {
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
		CheckType     string `json:"check_type"`
		CheckInterval int    `json:"check_interval"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.CheckType == "" {
		req.CheckType = "http"
	}

	if req.CheckInterval < 10 {
		req.CheckInterval = 60 // default
	}

	server, err := s.DB.AddServer(userID, req.Name, req.URL, req.CheckType, req.CheckInterval)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create server"})
	}

	// Запускаємо моніторинг для нового сервера негайно
	if s.scheduler != nil {
		s.scheduler.RunServerMonitor(*server)
	}

	return c.Status(fiber.StatusCreated).JSON(server)
}

func (s *Server) handleUpdateServer(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	serverIDStr := c.Params("id")

	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	var req struct {
		Name          string `json:"name"`
		URL           string `json:"url"`
		CheckType     string `json:"check_type"`
		CheckInterval int    `json:"check_interval"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	server, err := s.DB.UpdateServer(userID, uint(serverID), req.Name, req.URL, req.CheckType, req.CheckInterval)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update server"})
	}

	// Оновлюємо планувальник, якщо він активний
	if s.scheduler != nil {
		s.scheduler.RunServerMonitor(*server)
	}

	return c.JSON(server)
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete server"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (s *Server) handlePing(c fiber.Ctx) error {
	return c.SendString("pong")
}
