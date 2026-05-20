package api

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

const telegramLinkTTL = 15 * time.Minute

func (s *Server) handleCreateTelegramLinkToken(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	token, err := s.DB.CreateTelegramLinkToken(userID, telegramLinkTTL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create Telegram link token"})
	}

	response := fiber.Map{
		"token":          token,
		"expires_in":     int(telegramLinkTTL.Seconds()),
		"link_available": false,
	}
	if s.Config.TelegramBotName != "" {
		botName := strings.TrimPrefix(s.Config.TelegramBotName, "@")
		response["bot_name"] = botName
		response["link_available"] = true
		response["url"] = "https://t.me/" + botName + "?start=" + token
	}

	return c.JSON(response)
}

func (s *Server) handleGetTelegramStatus(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	botName := strings.TrimPrefix(s.Config.TelegramBotName, "@")

	account, err := s.DB.GetTelegramAccountByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(fiber.Map{
				"linked":         false,
				"bot_name":       botName,
				"link_available": botName != "",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch Telegram status"})
	}

	return c.JSON(fiber.Map{
		"linked":         true,
		"linked_at":      account.LinkedAt,
		"bot_name":       botName,
		"link_available": botName != "",
	})
}

func (s *Server) handleTelegramNewUser(c fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		token = c.Query("token")
	}
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing token"})
	}

	chatIDValue := c.Params("chat_id")
	if chatIDValue == "" {
		chatIDValue = c.Query("chat_id")
	}
	telegramChatID, err := strconv.ParseInt(chatIDValue, 10, 64)
	if err != nil || telegramChatID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Telegram chat ID"})
	}

	userID, err := s.DB.LinkTelegramAccount(token, telegramChatID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrTelegramTokenInvalid):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invalid Telegram link token"})
		case errors.Is(err, database.ErrTelegramTokenExpired), errors.Is(err, database.ErrTelegramTokenUsed):
			return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "Telegram link token is no longer valid"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to link Telegram account"})
		}
	}

	if err := s.DB.EnsureTelegramNotificationPreferences(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to prepare Telegram preferences"})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Telegram account linked"})
}
