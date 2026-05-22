package api

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ZaViBiS/isitdead/internal/checker"
	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBuildHourlyBucketsUsesMonitorSlowThreshold(t *testing.T) {
	now := time.Date(2026, time.May, 16, 12, 0, 0, 0, time.UTC)
	history := []model.CheckResult{
		{Status: "200 OK", Latency: 350, CreatedAt: now.Add(-30 * time.Minute)},
	}

	assert.Equal(t, "slow", buildHourlyBuckets(history, now, "http", 300)[23])
	assert.Equal(t, "ok", buildHourlyBuckets(history, now, "http", 400)[23])
}

type stubMailer struct {
	lastTo    string
	lastToken string
	err       error
}

func (m *stubMailer) SendVerificationEmail(to, token string) error {
	if m.err != nil {
		return m.err
	}
	m.lastTo = to
	m.lastToken = token
	return nil
}

func TestAPI(t *testing.T) {
	dbPath := "test_api.db"
	storage, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer func() {
		storage.Close()
		os.Remove(dbPath)
	}()

	sched := checker.NewScheduler(storage)
	defer sched.Stop()
	mailer := &stubMailer{}
	server, err := New(storage, sched, mailer, embed.FS{})
	assert.NoError(t, err)
	verificationToken := ""

	t.Run("Ping", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/ping", nil)
		resp, _ := server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Register and Verify", func(t *testing.T) {
		payload := map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := server.App.Test(req)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, "test@example.com", mailer.lastTo)
		assert.NotEmpty(t, mailer.lastToken)
		verificationToken = mailer.lastToken

		loginPayload := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		loginBody, _ := json.Marshal(loginPayload)
		loginReq := httptest.NewRequest("POST", "/api/login", bytes.NewReader(loginBody))
		loginReq.Header.Set("Content-Type", "application/json")

		loginResp, _ := server.App.Test(loginReq)
		assert.Equal(t, http.StatusForbidden, loginResp.StatusCode)
	})

	t.Run("Confirm Email", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/auth/confirm?token="+verificationToken, nil)
		resp, _ := server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Login and Protected Route", func(t *testing.T) {
		// Login
		payload := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var loginResp struct {
			Token string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&loginResp)
		token := loginResp.Token
		assert.NotEmpty(t, token)

		// Get Servers (Protected)
		req = httptest.NewRequest("GET", "/api/servers", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var servers []interface{}
		json.NewDecoder(resp.Body).Decode(&servers)
		assert.Equal(t, 0, len(servers))
	})

	t.Run("Register Fails When Mailer Fails", func(t *testing.T) {
		mailer.err = errors.New("smtp failed")
		t.Cleanup(func() { mailer.err = nil })

		payload := map[string]string{
			"username": "mail-failure",
			"email":    "mail-failure@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := server.App.Test(req)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Add, Update and Delete Server", func(t *testing.T) {
		// Get Token
		payload := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := server.App.Test(req)
		var loginResp struct {
			Token string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&loginResp)
		token := loginResp.Token

		// Timeout is required for monitor create/update payloads.
		missingTimeoutPayload := map[string]interface{}{
			"name":           "Missing Timeout",
			"url":            "http://example.com",
			"check_interval": 300,
		}
		body, _ = json.Marshal(missingTimeoutPayload)
		req = httptest.NewRequest("POST", "/api/servers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Add Server
		srvPayload := map[string]interface{}{
			"name":           "Test Server",
			"url":            "http://example.com",
			"check_interval": 300,
			"timeout":        10,
			"slow_threshold": 300,
		}
		body, _ = json.Marshal(srvPayload)
		req = httptest.NewRequest("POST", "/api/servers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var srv struct {
			ID            uint `json:"id"`
			Timeout       int  `json:"timeout"`
			SlowThreshold int  `json:"slow_threshold"`
		}
		json.NewDecoder(resp.Body).Decode(&srv)
		assert.Equal(t, 10, srv.Timeout)
		assert.Equal(t, 300, srv.SlowThreshold)
		serverIDStr := strconv.Itoa(int(srv.ID))

		// Update through the regular server API.
		userUpdate := map[string]interface{}{
			"name":           "Test Server",
			"url":            "http://example.com",
			"check_interval": 300,
			"check_type":     "http",
			"timeout":        10,
			"slow_threshold": 300,
		}
		body, _ = json.Marshal(userUpdate)
		req = httptest.NewRequest("PUT", "/api/servers/"+serverIDStr, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Notification preferences
		req = httptest.NewRequest("GET", "/api/servers/"+serverIDStr+"/notifications", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var prefs []struct {
			Channel string `json:"channel"`
			Event   string `json:"event"`
			Enabled bool   `json:"enabled"`
		}
		json.NewDecoder(resp.Body).Decode(&prefs)
		assert.Len(t, prefs, 10)

		updatePrefsPayload := []map[string]interface{}{
			{"channel": "email", "event": "down", "enabled": false},
			{"channel": "email", "event": "recovered", "enabled": true},
			{"channel": "telegram", "event": "down", "enabled": true},
		}
		body, _ = json.Marshal(updatePrefsPayload)
		req = httptest.NewRequest("PUT", "/api/servers/"+serverIDStr+"/notifications", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		req = httptest.NewRequest("GET", "/api/telegram/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var telegramStatus struct {
			Linked bool `json:"linked"`
		}
		json.NewDecoder(resp.Body).Decode(&telegramStatus)
		assert.False(t, telegramStatus.Linked)

		req = httptest.NewRequest("POST", "/api/telegram/link-token", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var linkResp struct {
			Token string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&linkResp)
		assert.NotEmpty(t, linkResp.Token)

		req = httptest.NewRequest("GET", "/api/telegram/token/123456/"+linkResp.Token, nil)
		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		req = httptest.NewRequest("GET", "/api/telegram/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		json.NewDecoder(resp.Body).Decode(&telegramStatus)
		assert.True(t, telegramStatus.Linked)

		req = httptest.NewRequest("GET", "/api/servers/not-a-number/results?limit=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Update Server
		updatePayload := map[string]interface{}{
			"name":           "Updated Name",
			"url":            "http://example.com/updated",
			"check_interval": 300,
			"check_type":     "http",
			"timeout":        15,
			"slow_threshold": 450,
		}
		body, _ = json.Marshal(updatePayload)
		req = httptest.NewRequest("PUT", "/api/servers/"+serverIDStr, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		missingUpdateTimeoutPayload := map[string]interface{}{
			"name":           "Updated Name",
			"url":            "http://example.com/updated",
			"check_interval": 120,
			"check_type":     "http",
		}
		body, _ = json.Marshal(missingUpdateTimeoutPayload)
		req = httptest.NewRequest("PUT", "/api/servers/"+serverIDStr, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Delete Server
		req = httptest.NewRequest("DELETE", "/api/servers/"+serverIDStr, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
