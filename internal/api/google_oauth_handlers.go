package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

func (s *Server) getGoogleOauthConfig() *oauth2.Config {
	redirectURL := fmt.Sprintf("https://%s/api/auth/google/callback", s.Config.Domain)
	if s.Config.Env == "dev" {
		redirectURL = fmt.Sprintf("http://localhost:%s/api/auth/google/callback", s.Config.Port)
	}

	return &oauth2.Config{
		RedirectURL:  redirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		ClientID:     s.Config.ClientID,
		ClientSecret: s.Config.ClientSecret,
		Endpoint:     google.Endpoint,
	}
}

func (s *Server) handleGoogleLogin(c fiber.Ctx) error {
	url := s.getGoogleOauthConfig().AuthCodeURL("state-token")
	return c.Redirect().To(url)
}

func (s *Server) handleGoogleCallback(c fiber.Ctx) error {
	code := c.Query("code")
	token, err := s.getGoogleOauthConfig().Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to exchange token"})
	}

	oauth2Service, err := oauth2api.NewService(context.Background(), option.WithTokenSource(s.getGoogleOauthConfig().TokenSource(context.Background(), token)))
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
		// Якщо Google ID ще не прив'язаний, створюємо користувача або прив'язуємо
		// існуючий email-акаунт до Google.
		user, err = s.DB.AddGoogleUser(userInfo.Name, userInfo.Email, userInfo.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
		}
	}

	// Генерируємо JWT
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := jwtToken.SignedString([]byte(s.Config.JWTSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	// Перенаправляємо на фронтенд з токеном
	// В SPA краще передати через кукі або редірект з параметром, а потім зберегти в localStorage
	return c.Redirect().To("/login?token=" + t + "&user=" + userInfo.Name)
}
