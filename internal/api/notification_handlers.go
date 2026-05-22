package api

import (
	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/gofiber/fiber/v3"
)

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

func isAllowedNotificationPreference(channel, event string) bool {
	if channel != model.NotificationChannelEmail && channel != model.NotificationChannelTelegram && channel != model.NotificationChannelDiscord {
		return false
	}
	return event == model.NotificationEventDown ||
		event == model.NotificationEventUp ||
		event == model.NotificationEventSSL30d ||
		event == model.NotificationEventSSL14d ||
		event == model.NotificationEventSSL7d
}
