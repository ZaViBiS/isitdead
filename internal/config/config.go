// Package config завантажує конфігурацію з змінних середовища.
package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	RoleMain           = "main"
	RoleProbe          = "probe"
	DefaultLocalRegion = "de"
	DefaultJWTSecret   = "dev-secret-change-me"
)

type ProbeRegion struct {
	Name string
	URL  string
}

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
	StripeSecretKey       string
	StripeWebhookSecret   string
	StripeProPriceID      string
	StripeBusinessPriceID string
	InstanceRole          string
	Region                string
	ProbeSecret           string
	ProbeRegions          []ProbeRegion
}

func Load() *Config {
	role := strings.ToLower(strings.TrimSpace(getEnv("INSTANCE_ROLE", getEnv("NODE_ROLE", RoleMain))))
	region := strings.TrimSpace(getEnv("REGION", DefaultLocalRegion))

	return &Config{
		Env:                   getEnv("ENV", "dev"),
		Port:                  getEnv("PORT", "8080"),
		Domain:                getEnv("DOMAIN", "localhost"),
		DBPath:                getEnv("DB_PATH", "/tmp/isitdead.db"),
		ResendAPIKey:          getEnv("RESEND_API_KEY", ""),
		ResendFrom:            getEnv("RESEND_FROM", "no-reply@localhost"),
		ClientID:              getEnv("CLIENT_ID", ""),
		ClientSecret:          getEnv("CLIENT_SECRET", ""),
		JWTSecret:             getEnv("JWT_SECRET", DefaultJWTSecret),
		AdminEmails:           getEnv("ADMIN_EMAILS", ""),
		TelegramBotName:       getEnv("TELEGRAM_BOT_NAME", ""),
		TelegramAPIURL:        getEnv("TELEGRAM_API_URL", ""),
		TelegramAPISecret:     getEnv("TELEGRAM_API_SECRET", ""),
		StripeSecretKey:       getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:   getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeProPriceID:      getEnv("STRIPE_PRO_PRICE_ID", ""),
		StripeBusinessPriceID: getEnv("STRIPE_BUSINESS_PRICE_ID", ""),
		InstanceRole:          role,
		Region:                region,
		ProbeSecret:           getEnv("PROBE_SECRET", ""),
		ProbeRegions:          parseProbeRegions(getEnv("PROBE_REGIONS", "")),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseProbeRegions(raw string) []ProbeRegion {
	parts := strings.Split(raw, ",")
	regions := make([]ProbeRegion, 0, len(parts))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		name := ""
		url := part
		if before, after, ok := strings.Cut(part, "="); ok {
			name = strings.TrimSpace(before)
			url = strings.TrimSpace(after)
		}
		if name == "" {
			name = fmt.Sprintf("probe-%d", i+1)
		}
		if url == "" {
			continue
		}

		regions = append(regions, ProbeRegion{
			Name: name,
			URL:  strings.TrimRight(url, "/"),
		})
	}
	return regions
}
