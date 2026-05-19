package api

import "github.com/gofiber/fiber/v3"

func (s *Server) handleTelegramNewUser(c fiber.Ctx) error {
	telegramChatID := c.Locals("user_id").(int)
	token := c.Locals("token").(string)

	userID, err := s.DB.GetUserIDByToken(token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add user"})
	}

	if err := s.DB.CreateTelegramAccount(userID, telegramChatID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add user"})
	}
	return c.SendStatus(fiber.StatusAccepted)
}
