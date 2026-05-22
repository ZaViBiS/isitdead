// Package config завантажує конфігурацію з змінних середовища.
package config

import "os"

type Config struct {
	Env                   string // "dev" / "prod"
	Port                  string
	Domain                string
	DBPath                string
	ResendAPIKey          string
	ResendFrom            string
	ClientID              string
	ClientSecret          string
	JWTSecret             string
	AdminEmails           string
	TelegramBotName       string
	TelegramAPIURL        string
	TelegramAPISecret     string
	DiscordWebhookEnabled string
}

func Load() *Config {
	return &Config{
		Env:                   getEnv("ENV", "dev"),
		Port:                  getEnv("PORT", "8080"),
		Domain:                getEnv("DOMAIN", "localhost"),
		DBPath:                getEnv("DB_PATH", "/tmp/isitdead.db"),
		ResendAPIKey:          getEnv("RESEND_API_KEY", ""),
		ResendFrom:            getEnv("RESEND_FROM", "no-reply@localhost"),
		ClientID:              getEnv("CLIENT_ID", ""),
		ClientSecret:          getEnv("CLIENT_SECRET", ""),
		JWTSecret:             getEnv("JWT_SECRET", "dev-secret-change-me"),
		AdminEmails:           getEnv("ADMIN_EMAILS", ""),
		TelegramBotName:       getEnv("TELEGRAM_BOT_NAME", ""),
		TelegramAPIURL:        getEnv("TELEGRAM_API_URL", ""),
		TelegramAPISecret:     getEnv("TELEGRAM_API_SECRET", ""),
		DiscordWebhookEnabled: getEnv("DISCORD_WEBHOOK_ENABLED", "1"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
