// Package config завантажує конфігурацію з змінних середовища.
package config

import "os"

type Config struct {
	Env          string // "dev" / "prod"
	Port         string
	Domain       string
	DBPath       string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPass     string
	SMTPFrom     string
	ClientID     string
	ClientSecret string
	JWTSecret    string
	AdminEmails  string
}

func Load() *Config {
	return &Config{
		Env:          getEnv("ENV", "dev"),
		Port:         getEnv("PORT", "8080"),
		Domain:       getEnv("DOMAIN", "localhost"),
		DBPath:       getEnv("DB_PATH", "/tmp/isitdead.db"),
		SMTPHost:     getEnv("SMTP_HOST", "localhost"),
		SMTPPort:     getEnv("SMTP_PORT", "465"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPass:     getEnv("SMTP_PASS", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "no-reply@localhost"),
		ClientID:     getEnv("CLIENT_ID", ""),
		ClientSecret: getEnv("CLIENT_SECRET", ""),
		JWTSecret:    getEnv("JWT_SECRET", "dev-secret-change-me"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
