package app

import (
	"embed"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ZaViBiS/isitdead/internal/config"
)

func TestAppNew(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL app tests")
	}

	t.Setenv("INSTANCE_ROLE", "main")
	t.Setenv("DATABASE_URL", databaseURL)

	a, err := New(embed.FS{})
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.NotNil(t, a.server)
	assert.NotNil(t, a.scheduler)
}

func TestAppNewProbeRole(t *testing.T) {
	t.Setenv("INSTANCE_ROLE", "probe")
	t.Setenv("REGION", "eu")
	t.Setenv("PROBE_SECRET", "shared")

	a, err := New(embed.FS{})

	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.Nil(t, a.server)
	assert.Nil(t, a.scheduler)
	assert.NotNil(t, a.probeServer)
}

func TestValidateConfigRejectsDefaultJWTSecretInProduction(t *testing.T) {
	err := validateConfig(&config.Config{
		Env:          "prod",
		InstanceRole: config.RoleMain,
		JWTSecret:    config.DefaultJWTSecret,
	})
	assert.Error(t, err)
}

func TestValidateConfigRejectsProbeWithoutSecret(t *testing.T) {
	err := validateConfig(&config.Config{
		Env:          "dev",
		InstanceRole: config.RoleProbe,
		JWTSecret:    config.DefaultJWTSecret,
	})
	assert.Error(t, err)
}

func TestValidateConfigRejectsProbeRegionsWithoutSecret(t *testing.T) {
	err := validateConfig(&config.Config{
		Env:          "dev",
		InstanceRole: config.RoleMain,
		JWTSecret:    config.DefaultJWTSecret,
		ProbeRegions: []config.ProbeRegion{{Name: "us", URL: "https://probe.example.com"}},
	})
	assert.Error(t, err)
}
