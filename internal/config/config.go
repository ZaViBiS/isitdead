// Package config завантажує конфігурацію з змінних середовища.
package config

import "os"

type Config struct {
	Env    string // "dev" / "prod"
	Port   string
	Domain string
	DBPath string
}

func Load() *Config {
	return &Config{
		Env:    getEnv("ENV", "dev"),
		Port:   getEnv("PORT", "8080"),
		Domain: getEnv("DOMAIN", "localhost"),
		DBPath: getEnv("DB_PATH", "/tmp/isitdead.db"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
