package app

import (
	"embed"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppNew(t *testing.T) {
	t.Setenv("INSTANCE_ROLE", "main")
	// Clean up after test if Init creates a file
	defer os.Remove("/tmp/isitdead.db")

	a, err := New(embed.FS{})
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.NotNil(t, a.server)
	assert.NotNil(t, a.scheduler)
}

func TestAppNewProbeRole(t *testing.T) {
	t.Setenv("INSTANCE_ROLE", "probe")
	t.Setenv("REGION", "eu")
	t.Setenv("DB_PATH", "/tmp/isitdead-probe-test.db")
	defer os.Remove("/tmp/isitdead-probe-test.db")

	a, err := New(embed.FS{})

	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.Nil(t, a.server)
	assert.Nil(t, a.scheduler)
	assert.NotNil(t, a.probeServer)
	assert.NoFileExists(t, "/tmp/isitdead-probe-test.db")
}
