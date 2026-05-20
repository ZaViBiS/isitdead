package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadProbeConfig(t *testing.T) {
	t.Setenv("INSTANCE_ROLE", "probe")
	t.Setenv("REGION", "eu")
	t.Setenv("PROBE_SECRET", "shared")
	t.Setenv("PROBE_REGIONS", "us=https://us.example.com, ap=https://ap.example.com/")

	cfg := Load()

	assert.Equal(t, RoleProbe, cfg.InstanceRole)
	assert.Equal(t, "eu", cfg.Region)
	assert.Equal(t, "shared", cfg.ProbeSecret)
	assert.Equal(t, []ProbeRegion{
		{Name: "us", URL: "https://us.example.com"},
		{Name: "ap", URL: "https://ap.example.com"},
	}, cfg.ProbeRegions)
}

func TestLoadAcceptsNodeRoleAlias(t *testing.T) {
	t.Setenv("INSTANCE_ROLE", "")
	t.Setenv("NODE_ROLE", "probe")

	cfg := Load()

	assert.Equal(t, RoleProbe, cfg.InstanceRole)
}
