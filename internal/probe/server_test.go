package probe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func stubHTTPStatusTransport(t *testing.T, statusCode int) {
	prev := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: statusCode,
			Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = prev
	})
}

func TestServerHandleCheck(t *testing.T) {
	stubHTTPStatusTransport(t, http.StatusOK)

	server := NewServer("eu", "shared")
	payload, err := json.Marshal(CheckRequest{
		CheckType: "http",
		URL:       "http://example.test",
		Timeout:   10,
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/probe/check", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := server.App.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/probe/check", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(SecretHeader, "shared")
	resp, err = server.App.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body CheckResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "eu", body.Region)
	assert.Equal(t, "200 OK", body.Status)
	assert.GreaterOrEqual(t, body.Latency, int64(0))
}
