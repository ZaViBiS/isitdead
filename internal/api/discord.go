package api

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"github.com/ZaViBiS/isitdead/internal/database"
)

const discordLinkTTL = 15 * time.Minute

func (s *Server) handleCreateDiscordLinkToken(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	token, err := s.DB.CreateDiscordLinkToken(userID, discordLinkTTL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create Discord link token"})
	}
	return c.JSON(fiber.Map{"token": token, "expires_in": int(discordLinkTTL.Seconds())})
}

func (s *Server) handleGetDiscordStatus(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	account, err := s.DB.GetDiscordAccountByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(fiber.Map{"linked": false})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch Discord status"})
	}
	return c.JSON(fiber.Map{"linked": true, "linked_at": account.LinkedAt})
}

func (s *Server) handleDiscordNewUser(c fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		token = c.Query("token")
	}
	webhookURL := strings.TrimSpace(c.Query("webhook_url"))
	if webhookURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing webhook_url"})
	}
	if !strings.HasPrefix(webhookURL, "https://discord.com/api/webhooks/") && !strings.HasPrefix(webhookURL, "https://discordapp.com/api/webhooks/") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Discord webhook URL"})
	}
	userID, err := s.DB.LinkDiscordAccount(token, webhookURL)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrDiscordTokenInvalid):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invalid Discord link token"})
		case errors.Is(err, database.ErrDiscordTokenExpired), errors.Is(err, database.ErrDiscordTokenUsed):
			return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "Discord link token is no longer valid"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to link Discord account"})
		}
	}
	if err := s.DB.EnsureDiscordNotificationPreferences(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to prepare Discord preferences"})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Discord account linked"})
}
