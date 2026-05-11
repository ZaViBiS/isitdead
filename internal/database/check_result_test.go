package database

import (
	"os"
	"testing"

	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestGetIncidents(t *testing.T) {
	dbPath := "test_incidents.db"
	defer os.Remove(dbPath)

	storage, err := Init(dbPath)
	assert.NoError(t, err)
	defer storage.Close()

	serverID := uint(1)

	// Add some mixed results
	results := []model.CheckResult{
		{ServerID: serverID, Status: "200 OK", Latency: 100},
		{ServerID: serverID, Status: "Connected", Latency: 50},
		{ServerID: serverID, Status: "500 Internal Server Error", Latency: 200},
		{ServerID: serverID, Status: "Timeout", Latency: 5000},
		{ServerID: serverID, Status: "201 Created", Latency: 150},
	}

	for _, r := range results {
		err := storage.AddCheckResult(r)
		assert.NoError(t, err)
	}

	incidents, err := storage.GetIncidents(serverID, 0)
	assert.NoError(t, err)
	assert.Len(t, incidents, 2)

	// Verify they are the correct ones
	for _, r := range incidents {
		assert.NotContains(t, r.Status, "200 OK")
		assert.NotContains(t, r.Status, "Connected")
		assert.NotContains(t, r.Status, "201 Created")
	}

	// Test limit
	incidentsLimit, err := storage.GetIncidents(serverID, 1)
	assert.NoError(t, err)
	assert.Len(t, incidentsLimit, 1)
}
