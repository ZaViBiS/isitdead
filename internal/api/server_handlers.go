package api

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
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

func (s *Server) handleGetDashboardServers(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	servers, err := s.DB.GetUserServers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch servers"})
	}

	now := time.Now().UTC()
	thirtyDaysAgo := now.Add(-720 * time.Hour)
	lastDayAgo := now.Add(-24 * time.Hour)
	response := make([]dashboardServerResponse, 0, len(servers))
	sslServerIDs := make([]uint, 0)
	for _, server := range servers {
		if server.SSLEnabled {
			sslServerIDs = append(sslServerIDs, server.ID)
		}
	}
	sslStatuses, err := s.DB.GetSSLCertificateStatuses(sslServerIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch SSL statuses"})
	}

	for _, server := range servers {
		summary, err := s.DB.GetHistorySummarySince(server.ID, thirtyDaysAgo)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch dashboard summary"})
		}

		recentHistory, err := s.DB.GetHistorySince(server.ID, lastDayAgo)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch dashboard history"})
		}

		currentStatus := "unknown"
		var currentLatency int64
		latest, err := s.DB.GetLatestCheckResult(server.ID)
		if err == nil {
			currentStatus = latest.Status
			currentLatency = latest.Latency
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch latest result"})
		}

		var uptime30d float64
		if summary.Total > 0 {
			uptime30d = float64(summary.Online) / float64(summary.Total) * 100
		}

		response = append(response, dashboardServerResponse{
			ID:             server.ID,
			Name:           server.Name,
			URL:            server.URL,
			CheckType:      server.CheckType,
			CheckInterval:  server.CheckInterval,
			Timeout:        server.Timeout,
			CheckCount30d:  summary.Total,
			Uptime30d:      uptime30d,
			AvgLatency30d:  int64(math.Round(summary.AvgLatency)),
			CurrentStatus:  currentStatus,
			CurrentLatency: currentLatency,
			HourlyBuckets:  buildHourlyBuckets(recentHistory, now, server.CheckType, server.SlowThreshold),
			SlowThreshold:  server.SlowThreshold,
			SSLEnabled:     server.SSLEnabled,
		})
		if server.SSLEnabled {
			if sslStatus, ok := sslStatuses[server.ID]; ok {
				expiresAt := ""
				if sslStatus.ExpiresAt != nil {
					expiresAt = sslStatus.ExpiresAt.UTC().Format(time.RFC3339)
				}
				response[len(response)-1].SSLStatus = &sslCertificateStatusResponse{
					Valid:         sslStatus.Valid,
					SelfSigned:    sslStatus.SelfSigned,
					ExpiresAt:     expiresAt,
					DaysRemaining: sslStatus.DaysRemaining,
					Issuer:        sslStatus.Issuer,
					LastError:     sslStatus.LastError,
					LastCheckedAt: sslStatus.LastCheckedAt.UTC().Format(time.RFC3339),
				}
			}
		}
	}

	return c.JSON(response)
}

func buildHourlyBuckets(history []model.CheckResult, now time.Time, checkType string, slowThreshold int) []string {
	buckets := make([]string, 24)
	for i := range buckets {
		buckets[i] = "empty"
	}

	windowStart := now.Add(-24 * time.Hour)
	for _, result := range history {
		if result.CreatedAt.Before(windowStart) || !result.CreatedAt.Before(now) {
			continue
		}

		index := int(result.CreatedAt.Sub(windowStart) / time.Hour)
		if index < 0 || index >= len(buckets) {
			continue
		}

		next := bucketStatus(result, checkType, slowThreshold)
		current := buckets[index]
		if current == "error" {
			continue
		}
		if next == "error" || current == "empty" {
			buckets[index] = next
			continue
		}
		if next == "slow" {
			buckets[index] = next
		}
	}

	return buckets
}

func bucketStatus(result model.CheckResult, checkType string, slowThreshold int) string {
	if !(strings.HasPrefix(result.Status, "2") || result.Status == "Connected") {
		return "error"
	}
	if result.Latency > int64(slowThreshold) {
		return "slow"
	}
	return "ok"
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
	if serverRequest.SlowThreshold <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Slow threshold is required"})
	}

	if serverRequest.SSLEnabled && !supportsSSLMonitoring(serverRequest.CheckType) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "SSL monitoring is only supported for HTTP and link monitors"})
	}

	server, err := s.DB.AddServer(userID, serverRequest.Name, serverRequest.URL, serverRequest.CheckType, serverRequest.CheckInterval, serverRequest.Timeout, serverRequest.SlowThreshold, serverRequest.SSLEnabled)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Public slug is already used"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create server"})
	}

	// Запускаємо моніторинг для нового сервера негайно
	if s.Scheduler != nil {
		s.Scheduler.RunServerMonitor(*server)
		if server.SSLEnabled {
			go s.Scheduler.RunSSLCheck(*server)
		}
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
	if req.SlowThreshold <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Slow threshold is required"})
	}

	if req.SSLEnabled && !supportsSSLMonitoring(req.CheckType) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "SSL monitoring is only supported for HTTP and link monitors"})
	}

	server, err := s.DB.UpdateServer(userID, serverID, req.Name, req.URL, req.CheckType, req.CheckInterval, req.Timeout, req.SlowThreshold, req.SSLEnabled)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Public slug is already used"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update server"})
	}

	// Оновлюємо планувальник, якщо він активний
	if s.Scheduler != nil {
		s.Scheduler.RunServerMonitor(*server)
		if server.SSLEnabled {
			go s.Scheduler.RunSSLCheck(*server)
		}
	}

	return c.JSON(server)
}

func supportsSSLMonitoring(checkType string) bool {
	return checkType == "http" || checkType == "links"
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
