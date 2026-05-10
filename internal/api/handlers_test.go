package api

import (
	"bytes"
	"embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ZaViBiS/isitdead/internal/checker"
	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/stretchr/testify/assert"
)

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
	server, _ := New(storage, sched, embed.FS{})

	t.Run("Ping", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/ping", nil)
		resp, _ := server.App.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Register", func(t *testing.T) {
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

	t.Run("Add and Delete Server", func(t *testing.T) {
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

		// Add Server
		srvPayload := map[string]interface{}{
			"name":           "Test Server",
			"url":            "http://example.com",
			"check_interval": 60,
		}
		body, _ = json.Marshal(srvPayload)
		req = httptest.NewRequest("POST", "/api/servers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var srv struct {
			ID uint `json:"id"`
		}
		json.NewDecoder(resp.Body).Decode(&srv)

		// Delete Server
		req = httptest.NewRequest("DELETE", "/api/servers/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ = server.App.Test(req)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
