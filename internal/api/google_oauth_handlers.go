package api

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

const (
	googleOAuthStateCookie   = "google_oauth_state"
	googleOAuthSessionCookie = "google_oauth_session"
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
	state, err := generateOAuthState()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start OAuth flow"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     googleOAuthStateCookie,
		Value:    state,
		Path:     "/api/auth/google",
		MaxAge:   10 * 60,
		Secure:   s.Config.Env != "dev",
		HTTPOnly: true,
		SameSite: "Lax",
	})

	url := s.getGoogleOauthConfig().AuthCodeURL(state)
	return c.Redirect().To(url)
}

func (s *Server) handleGoogleCallback(c fiber.Ctx) error {
	if !validOAuthState(c.Cookies(googleOAuthStateCookie), c.Query("state")) {
		clearGoogleOAuthState(c, s.Config.Env != "dev")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid OAuth state"})
	}
	clearGoogleOAuthState(c, s.Config.Env != "dev")

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

	c.Cookie(&fiber.Cookie{
		Name:     googleOAuthSessionCookie,
		Value:    t,
		Path:     "/api/auth/session",
		MaxAge:   60,
		Secure:   s.Config.Env != "dev",
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return c.Redirect().To("/login?oauth=success")
}

func (s *Server) handleGoogleSession(c fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	tokenString := c.Cookies(googleOAuthSessionCookie)
	clearGoogleOAuthSession(c, s.Config.Env != "dev")
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing OAuth session"})
	}

	userID, err := s.userIDFromJWT(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired OAuth session"})
	}

	user, err := s.DB.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	return c.JSON(fiber.Map{
		"token": tokenString,
		"user":  fiber.Map{"id": user.ID, "username": user.Username, "email": user.Email, "is_admin": s.isAdminEmail(user.Email)},
	})
}

func generateOAuthState() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func validOAuthState(cookieState, queryState string) bool {
	if cookieState == "" || queryState == "" || len(cookieState) != len(queryState) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookieState), []byte(queryState)) == 1
}

func clearGoogleOAuthState(c fiber.Ctx, secure bool) {
	c.Cookie(&fiber.Cookie{
		Name:     googleOAuthStateCookie,
		Value:    "",
		Path:     "/api/auth/google",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		Secure:   secure,
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

func clearGoogleOAuthSession(c fiber.Ctx, secure bool) {
	c.Cookie(&fiber.Cookie{
		Name:     googleOAuthSessionCookie,
		Value:    "",
		Path:     "/api/auth/session",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		Secure:   secure,
		HTTPOnly: true,
		SameSite: "Lax",
	})
}
