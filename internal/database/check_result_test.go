package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ZaViBiS/isitdead/internal/model"
)

func TestGetIncidents(t *testing.T) {
	storage := newTestStorage(t)

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

func TestCheckResultRegionFiltering(t *testing.T) {
	storage := newTestStorage(t)

	serverID := uint(42)
	now := time.Now().UTC()
	legacyTime := now.Add(-time.Minute)
	results := []model.CheckResult{
		{ServerID: serverID, Region: model.CheckRegionGlobal, Status: "200 OK", Latency: 120, CreatedAt: legacyTime},
		{ServerID: serverID, Region: model.CheckRegionGlobal, Status: "200 OK", Latency: 100, CreatedAt: now},
		{ServerID: serverID, Region: "eu", Status: "500 Internal Server Error", Latency: 300, CreatedAt: now},
		{ServerID: serverID, Region: "us", Status: "200 OK", Latency: 80, CreatedAt: now},
	}
	for _, result := range results {
		assert.NoError(t, storage.AddCheckResult(result))
	}

	defaultHistory, err := storage.GetHistorySince(serverID, time.Time{})
	assert.NoError(t, err)
	assert.Len(t, defaultHistory, 2)
	assert.Equal(t, model.CheckRegionGlobal, defaultHistory[1].Region)
	assert.Equal(t, "200 OK", defaultHistory[1].Status)
	assert.Equal(t, int64(80), defaultHistory[1].Latency)
	assert.NotZero(t, defaultHistory[1].ID)

	euHistory, err := storage.GetHistorySinceForRegion(serverID, "eu", time.Time{})
	assert.NoError(t, err)
	assert.Len(t, euHistory, 1)
	assert.Equal(t, "500 Internal Server Error", euHistory[0].Status)

	allHistory, err := storage.GetHistorySinceForRegion(serverID, model.CheckRegionAll, time.Time{})
	assert.NoError(t, err)
	assert.Len(t, allHistory, 2)

	summary, err := storage.GetHistorySummarySince(serverID, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), summary.Total)
	assert.Equal(t, int64(2), summary.Online)

	allSummary, err := storage.GetHistorySummarySinceForRegion(serverID, model.CheckRegionAll, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), allSummary.Total)
	assert.Equal(t, int64(1), allSummary.Online)
}
