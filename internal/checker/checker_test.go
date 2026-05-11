package checker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func setDefaultTransport(t *testing.T, rt http.RoundTripper) {
	prev := http.DefaultTransport
	http.DefaultTransport = rt
	t.Cleanup(func() {
		http.DefaultTransport = prev
	})
}

func httpResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
}

func TestCheck(t *testing.T) {
	t.Run("successful check", func(t *testing.T) {
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return httpResponse(http.StatusOK), nil
		}))
		status, latency := Check("http", "http://example.test")

		assert.Equal(t, "200 OK", status)
		assert.GreaterOrEqual(t, latency, int64(0))
	})

	t.Run("not found check", func(t *testing.T) {
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return httpResponse(http.StatusNotFound), nil
		}))
		status, _ := Check("http", "http://example.test")

		assert.Equal(t, "404 Not Found", status)
	})

	t.Run("server error check", func(t *testing.T) {
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return httpResponse(http.StatusInternalServerError), nil
		}))
		status, _ := Check("http", "http://example.test")

		assert.Equal(t, "500 Internal Server Error", status)
	})

	t.Run("invalid url", func(t *testing.T) {
		status, _ := Check("http", "://bad-url")

		assert.Contains(t, status, "missing protocol scheme")
	})

	t.Run("timeout", func(t *testing.T) {
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		}))
		status, _ := Check("http", "http://example.test")

		assert.Contains(t, status, "context deadline exceeded")
	})
}
