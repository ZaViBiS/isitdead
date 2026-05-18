// Package config завантажує конфігурацію з змінних середовища.
package config

import "os"

type Config struct {
	Env          string // "dev" / "prod"
	Port         string
	Domain       string
	DBPath       string
	ResendAPIKey string
	ResendFrom   string
	ClientID     string
	ClientSecret string
	JWTSecret    string
	AdminEmails  string
	TelegramToken string
	BotAPIURL     string
}

func Load() *Config {
	return &Config{
		Env:           getEnv("ENV", "dev"),
		Port:          getEnv("PORT", "8080"),
		Domain:        getEnv("DOMAIN", "localhost"),
		DBPath:        getEnv("DB_PATH", "/tmp/isitdead.db"),
		ResendAPIKey:  getEnv("RESEND_API_KEY", ""),
		ResendFrom:    getEnv("RESEND_FROM", "no-reply@localhost"),
		ClientID:      getEnv("CLIENT_ID", ""),
		ClientSecret:  getEnv("CLIENT_SECRET", ""),
		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change-me"),
		TelegramToken: getEnv("TELEGRAM_TOKEN", ""),
		BotAPIURL:     getEnv("BOT_API_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
