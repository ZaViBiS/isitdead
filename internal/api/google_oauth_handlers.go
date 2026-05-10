package api

import (
	"context"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/api/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

func (s *Server) handleGoogleLogin(c fiber.Ctx) error {
	url := googleOauthConfig.AuthCodeURL("state-token")
	return c.Redirect().To(url)
}

func (s *Server) handleGoogleCallback(c fiber.Ctx) error {
	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to exchange token"})
	}

	oauth2Service, err := oauth2api.NewService(context.Background(), option.WithTokenSource(googleOauthConfig.TokenSource(context.Background(), token)))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create oauth2 service"})
	}

	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user info"})
	}

	// Шукаємо користувача за Google ID
	user, err := s.DB.GetUserByGoogleID(userInfo.Id)
	if err != nil {
		// Якщо не знайшли за Google ID, спробуємо за Email (якщо користувач раніше реєструвався звичайно)
		user, err = s.DB.GetUserByEmail(userInfo.Email)
		if err == nil {
			// Користувач існує, прив'язуємо Google ID (якщо потрібно)
			// Для спрощення просто продовжуємо, але в ідеалі треба оновити запис
		} else {
			// Створюємо нового користувача
			user, err = s.DB.AddGoogleUser(userInfo.Name, userInfo.Email, userInfo.Id)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
			}
		}
	}

	// Генерируємо JWT
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := jwtToken.SignedString(JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	// Перенаправляємо на фронтенд з токеном
	// В SPA краще передати через кукі або редірект з параметром, а потім зберегти в localStorage
	return c.Redirect().To("/login?token=" + t + "&user=" + userInfo.Name)
}
