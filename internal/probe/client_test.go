package probe

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientCheckRegions(t *testing.T) {
	client := NewClient([]Target{{Region: "us", URL: "https://us.example.test"}}, "shared")
	client.httpClient = &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/probe/check", r.URL.Path)
		assert.Equal(t, "shared", r.Header.Get(SecretHeader))

		var req CheckRequest
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "http", req.CheckType)
		assert.Equal(t, "http://example.test", req.URL)
		assert.Equal(t, 10, req.Timeout)

		body, err := json.Marshal(CheckResponse{
			Region:  "us",
			Status:  "200 OK",
			Latency: 75,
		})
		assert.NoError(t, err)
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(string(body))),
			Header:     make(http.Header),
		}, nil
	})}

	results := client.CheckRegions(context.Background(), "http", "http://example.test", 10)

	assert.Len(t, results, 1)
	assert.Equal(t, "us", results[0].Region)
	assert.Equal(t, "200 OK", results[0].Status)
	assert.Equal(t, int64(75), results[0].Latency)
}

func TestClientReturnsProbeErrorAsRegionResult(t *testing.T) {
	client := NewClient([]Target{{Region: "us", URL: "https://us.example.test"}}, "shared")
	client.httpClient = &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Status:     "401 Unauthorized",
			Body:       io.NopCloser(strings.NewReader("unauthorized")),
			Header:     make(http.Header),
		}, nil
	})}

	results := client.CheckRegions(context.Background(), "http", "http://example.test", 10)

	assert.Len(t, results, 1)
	assert.Equal(t, "us", results[0].Region)
	assert.Contains(t, results[0].Status, "Probe HTTP error")
	assert.Contains(t, results[0].Status, "unauthorized")
}
