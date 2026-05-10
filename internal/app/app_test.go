package app

import (
	"embed"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppNew(t *testing.T) {
	// Clean up after test if Init creates a file
	defer os.Remove("/tmp/isitdead.db")

	a, err := New(embed.FS{})
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.NotNil(t, a.server)
	assert.NotNil(t, a.scheduler)
}
