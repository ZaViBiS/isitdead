package checker

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	t.Run("successful check", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		status, latency := Check(ts.URL)

		assert.Equal(t, "200 OK", status)
		assert.GreaterOrEqual(t, latency, int64(0))
	})

	t.Run("not found check", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		status, _ := Check(ts.URL)

		assert.Equal(t, "404 Not Found", status)
	})

	t.Run("server error check", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		status, _ := Check(ts.URL)

		assert.Equal(t, "500 Internal Server Error", status)
	})

	t.Run("invalid url", func(t *testing.T) {
		status, _ := Check("http://invalid.url.that.does.not.exist")

		assert.Contains(t, status, "no such host")
	})

	t.Run("timeout", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(11 * time.Second) // Check timeout is 10s
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		status, _ := Check(ts.URL)

		assert.Contains(t, status, "Client.Timeout exceeded")
	})
}
