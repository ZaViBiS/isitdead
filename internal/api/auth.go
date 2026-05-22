package api

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func (s *Server) authMiddleware(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing authorization header"})
	}

	userID, err := s.userIDFromJWT(strings.TrimPrefix(authHeader, "Bearer "))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	c.Locals("user_id", userID)

	return c.Next()
}

func (s *Server) userIDFromJWT(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.Config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user id claim")
	}

	return uint(userID), nil
}

func (s *Server) adminMiddleware(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing user context"})
	}

	user, err := s.DB.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	if s.isAdminEmail(user.Email) {
		return c.Next()
	}

	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Admin access required"})
}

func (s *Server) isAdminEmail(userEmail string) bool {
	for _, email := range strings.Split(s.Config.AdminEmails, ",") {
		if strings.EqualFold(strings.TrimSpace(email), userEmail) && strings.TrimSpace(email) != "" {
			return true
		}
	}
	return false
}
