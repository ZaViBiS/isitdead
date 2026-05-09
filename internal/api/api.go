// Package api надає ендпоінти API для керування серверами.
package api

import (
	"embed"
	"io"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/checker"
	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// JWTSecret повинен бути у файлі конфігурації або змінних оточення
var JWTSecret = []byte("your-very-secret-key")

type Server struct {
	App       *fiber.App
	DB        *database.Storage
	scheduler *checker.Scheduler
}

// New повертає готовий backend для сайту
func New(db *database.Storage, sched *checker.Scheduler, staticFiles embed.FS) (*Server, error) {
	app := fiber.New()

	s := &Server{
		App:       app,
		DB:        db,
		scheduler: sched,
	}

	// Логування запитів
	app.Use(func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		statusCode := c.Response().StatusCode()
		event := log.Info()
		if err != nil {
			event = log.Error().Err(err)
		}

		event.Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", statusCode).
			Dur("latency", time.Since(start)).
			Str("ip", c.IP()).
			Str("user_agent", string(c.Request().Header.UserAgent())).
			Msg("request")

		return err
	})

	// Налаштовуємо доступ до статичних файлів через embed
	distFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		return nil, err
	}

	// Реєстрація API маршрутів
	s.setupRoutes()

	// Обслуговування статичних файлів через middleware/static
	app.Get("/*", static.New("", static.Config{
		FS: distFS,
	}))

	// Fallback для SPA (SvelteKit)
	app.Use(func(c fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/api") {
			return c.Next()
		}

		index, err := distFS.Open("index.html")
		if err != nil {
			return c.Next()
		}
		defer index.Close()

		content, _ := io.ReadAll(index)
		c.Set("Content-Type", "text/html")
		return c.Send(content)
	})

	return s, nil
}

func (s *Server) setupRoutes() {
	api := s.App.Group("/api")

	api.Get("/ping", s.handlePing)
	api.Post("/register", s.handleRegister)
	api.Post("/login", s.handleLogin)

	// Protected routes
	api.Use(s.authMiddleware)
	api.Get("/servers", s.handleGetServers)
	api.Post("/servers", s.handleAddServer)
	api.Delete("/servers/:id", s.handleDeleteServer)
}

func (s *Server) authMiddleware(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing authorization header"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	userID := uint(claims["user_id"].(float64))
	c.Locals("user_id", userID)

	return c.Next()
}

// Handlers for servers
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

// Handlers
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

	// Створюємо користувача (пароль хешується всередині AddUser)
	user, err := s.DB.AddUser(req.Username, req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create user"})
	}

	// Генеруємо JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := token.SignedString(JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"token":   t,
		"user":    fiber.Map{"id": user.ID, "username": user.Username, "email": user.Email},
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

	// Перевіряємо пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Генеруємо JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := token.SignedString(JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{
		"token": t,
		"user":  fiber.Map{"id": user.ID, "username": user.Username, "email": user.Email},
	})
}

func (s *Server) handlePing(c fiber.Ctx) error {
	return c.SendString("pong")
}

// Listen запускає сервер
func (s *Server) Listen(addr string) error {
	return s.App.Listen(addr)
}
