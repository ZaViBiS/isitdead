package api

import (
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) handleRegister(c fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Перевіряємо чи користувач вже існує
	_, err := s.DB.GetUserByEmail(req.Email)
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User with this email already exists"})
	}

	// Створюємо користувача та токен
	user, token, err := s.DB.AddUser(req.Username, req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create user"})
	}

	// Відправляємо лист для підтвердження
	if s.Mailer == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Mailer is not configured"})
	}

	if err := s.Mailer.SendVerificationEmail(user.Email, token); err != nil {
		log.Error().Err(err).Msg("Failed to send verification email")
		// Можна повернути помилку або просто залогувати,
		// але краще повідомити користувача, що щось не так з поштою
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send verification email"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully. Please check your email to confirm your account.",
	})
}

func (s *Server) handleLogin(c fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Знаходимо користувача
	user, err := s.DB.GetUserByEmail(req.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Перевіряємо чи підтверджена пошта
	if !user.VerifiedEmail {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Please verify your email before logging in"})
	}

	// Перевіряємо пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Генеруємо JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := token.SignedString([]byte(s.Config.JWTSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{
		"token": t,
		"user":  fiber.Map{"id": user.ID, "username": user.Username, "email": user.Email, "is_admin": s.isAdminEmail(user.Email)},
	})
}

func (s *Server) handleGetMe(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing user context"})
	}

	user, err := s.DB.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	return c.JSON(fiber.Map{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"is_admin": s.isAdminEmail(user.Email),
	})
}

func (s *Server) handleConfirmEmail(c fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Verification token is required"})
	}

	verification, err := s.DB.GetVerificationByToken(token)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invalid or expired verification token"})
	}

	if err := s.DB.VerifyUser(verification.UserID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify email"})
	}

	// Після підтвердження можна або редиректнути на логін, або віддати JSON
	// Оскільки це API, віддаємо JSON або редиректимо на сторінку логіну фронтенда
	return c.SendString("Email verified successfully! You can now log in.")
}
