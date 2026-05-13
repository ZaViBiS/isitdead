package api

import (
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type publicMonitorResponse struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	URL           string    `json:"url"`
	CheckType     string    `json:"check_type"`
	Status        string    `json:"status"`
	Latency       int64     `json:"latency"`
	CheckInterval int       `json:"check_interval"`
	PublicSlug    string    `json:"public_slug"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

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

	return s.respondWithServerResults(c, uint(serverID))
}

func (s *Server) handleGetPublicMonitor(c fiber.Ctx) error {
	server, err := s.DB.GetPublicServerBySlug(c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Public monitor not found"})
	}
	return c.JSON(toPublicMonitorResponse(*server))
}

func (s *Server) handleGetPublicMonitorResults(c fiber.Ctx) error {
	server, err := s.DB.GetPublicServerBySlug(c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Public monitor not found"})
	}
	return s.respondWithServerResults(c, server.ID)
}

func (s *Server) handleAdminGetPublicMonitors(c fiber.Ctx) error {
	servers, err := s.DB.GetPublicServers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch public monitors"})
	}
	return c.JSON(servers)
}

func (s *Server) handleAdminUpdatePublicMonitor(c fiber.Ctx) error {
	serverID, err := parseServerID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	var req struct {
		Public     bool   `json:"public"`
		PublicSlug string `json:"public_slug"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	server, err := s.DB.UpdatePublicServer(serverID, req.Public, req.PublicSlug)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Public slug is already used"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update public monitor"})
	}

	return c.JSON(server)
}

func (s *Server) handleGetNotificationPreferences(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	serverID, err := parseServerID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	if !s.userOwnsServer(userID, serverID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	if err := s.DB.EnsureDefaultNotificationPreferences(userID, serverID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not prepare notification preferences"})
	}

	prefs, err := s.DB.GetUserNotificationPreferences(userID, serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch notification preferences"})
	}

	return c.JSON(prefs)
}

func (s *Server) handleUpdateNotificationPreferences(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	serverID, err := parseServerID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	if !s.userOwnsServer(userID, serverID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	var req []struct {
		Channel     string `json:"channel"`
		Event       string `json:"event"`
		Enabled     bool   `json:"enabled"`
		Destination string `json:"destination"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	prefs := make([]model.NotificationPreference, 0, len(req))
	for _, item := range req {
		if !isAllowedNotificationPreference(item.Channel, item.Event) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Unsupported notification preference"})
		}
		prefs = append(prefs, model.NotificationPreference{
			Channel:     item.Channel,
			Event:       item.Event,
			Enabled:     item.Enabled,
			Destination: item.Destination,
		})
	}

	if err := s.DB.SaveUserNotificationPreferences(userID, serverID, prefs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not save notification preferences"})
	}

	saved, err := s.DB.GetUserNotificationPreferences(userID, serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch notification preferences"})
	}

	return c.JSON(saved)
}

func (s *Server) handleGetServers(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	servers, err := s.DB.GetUserServers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch servers"})
	}
	return c.JSON(servers)
}

func parseServerID(c fiber.Ctx) (uint, error) {
	serverID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(serverID), nil
}

func (s *Server) userOwnsServer(userID, serverID uint) bool {
	servers, err := s.DB.GetUserServers(userID)
	if err != nil {
		return false
	}
	for _, srv := range servers {
		if srv.ID == serverID {
			return true
		}
	}
	return false
}

func isAllowedNotificationPreference(channel, event string) bool {
	if channel != model.NotificationChannelEmail {
		return false
	}
	return event == model.NotificationEventDown || event == model.NotificationEventUp
}

func (s *Server) handleAddServer(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	var req struct {
		Name          string `json:"name"`
		URL           string `json:"url"`
		CheckType     string `json:"check_type"`
		CheckInterval int    `json:"check_interval"`
		Timeout       int    `json:"timeout"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.CheckType == "" {
		req.CheckType = "http"
	}

	if req.CheckInterval < 10 {
		req.CheckInterval = 300 // default
	}
	if req.Timeout <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Timeout is required"})
	}

	server, err := s.DB.AddServer(userID, req.Name, req.URL, req.CheckType, req.CheckInterval, req.Timeout, false, "")
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
		Timeout       int    `json:"timeout"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Timeout <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Timeout is required"})
	}

	server, err := s.DB.UpdateServer(userID, uint(serverID), req.Name, req.URL, req.CheckType, req.CheckInterval, req.Timeout, false, "")
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
	serverIDStr := c.Params("id")

	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	err = s.DB.DeleteServer(userID, uint(serverID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete server"})
	}

	if s.Scheduler != nil {
		s.Scheduler.StopServerMonitor(uint(serverID))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (s *Server) handlePing(c fiber.Ctx) error {
	return c.SendString("pong")
}

func (s *Server) handleSitemap(c fiber.Ctx) error {
	servers, err := s.DB.GetPublicServers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not build sitemap")
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, server := range servers {
		fmt.Fprintf(
			&b, "  <url><loc>%s</loc><lastmod>%s</lastmod></url>\n",
			html.EscapeString(s.publicStatusURL(server.PublicSlug)),
			server.UpdatedAt.UTC().Format("2006-01-02"),
		)
	}
	b.WriteString(`</urlset>` + "\n")

	c.Set("Content-Type", "application/xml; charset=utf-8")
	return c.SendString(b.String())
}

func (s *Server) respondWithServerResults(c fiber.Ctx, serverID uint) error {
	incidentsOnly := c.Query("incidents") == "true"
	limitStr := c.Query("limit")
	limit := 0
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	var (
		results []model.CheckResult
		err     error
	)

	if incidentsOnly {
		results, err = s.DB.GetIncidents(serverID, limit)
	} else {
		results, err = s.DB.GetHistory(serverID)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch results"})
	}

	hoursStr := c.Query("hours")
	if hoursStr != "" && !incidentsOnly {
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

	if !incidentsOnly && limit > 0 && hoursStr == "" && len(results) > limit {
		results = results[len(results)-limit:]
	}

	return c.JSON(results)
}

func toPublicMonitorResponse(server model.Server) publicMonitorResponse {
	return publicMonitorResponse{
		ID:            server.ID,
		Name:          server.Name,
		URL:           server.URL,
		CheckType:     server.CheckType,
		Status:        server.Status,
		Latency:       server.Latency,
		CheckInterval: server.CheckInterval,
		PublicSlug:    server.PublicSlug,
		CreatedAt:     server.CreatedAt,
		UpdatedAt:     server.UpdatedAt,
	}
}
