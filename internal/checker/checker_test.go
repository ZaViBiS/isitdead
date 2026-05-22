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

func assertMonitorUserAgent(t *testing.T, req *http.Request) {
	t.Helper()
	assert.Equal(t, monitorUserAgent, req.Header.Get("User-Agent"))
}

func htmlResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"text/html"}},
	}
}

func TestCheck(t *testing.T) {
	t.Run("successful check", func(t *testing.T) {
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			assertMonitorUserAgent(t, req)
			return httpResponse(http.StatusOK), nil
		}))
		status, latency := Check("http", "http://example.test", 10)

		assert.Equal(t, "200 OK", status)
		assert.GreaterOrEqual(t, latency, int64(0))
	})

	t.Run("not found check", func(t *testing.T) {
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return httpResponse(http.StatusNotFound), nil
		}))
		status, _ := Check("http", "http://example.test", 10)

		assert.Equal(t, "404 Not Found", status)
	})

	t.Run("server error check", func(t *testing.T) {
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return httpResponse(http.StatusInternalServerError), nil
		}))
		status, _ := Check("http", "http://example.test", 10)

		assert.Equal(t, "500 Internal Server Error", status)
	})

	t.Run("invalid url", func(t *testing.T) {
		status, _ := Check("http", "://bad-url", 10)

		assert.Contains(t, status, "missing protocol scheme")
	})

	t.Run("blocks private http target", func(t *testing.T) {
		status, _ := Check("http", "http://127.0.0.1/admin", 10)

		assert.Contains(t, status, "private or internal address")
	})

	t.Run("timeout", func(t *testing.T) {
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		}))
		status, _ := Check("http", "http://example.test", 10)

		assert.Contains(t, status, "context deadline exceeded")
	})
}

func TestValidateMonitorTarget(t *testing.T) {
	tests := []struct {
		name      string
		checkType string
		target    string
		want      string
		wantErr   string
	}{
		{name: "http url", checkType: "http", target: " https://example.com/path#frag ", want: "https://example.com/path"},
		{name: "links url", checkType: "links", target: "http://example.com", want: "http://example.com"},
		{name: "reject javascript", checkType: "http", target: "javascript:alert(1)", wantErr: "unsupported URL scheme"},
		{name: "reject credentials", checkType: "http", target: "https://user:pass@example.com", wantErr: "credentials"},
		{name: "reject private literal", checkType: "http", target: "http://10.0.0.1", wantErr: "private or internal address"},
		{name: "ping host", checkType: "ping", target: "example.com:443", want: "example.com:443"},
		{name: "reject ping url", checkType: "ping", target: "https://example.com", wantErr: "host[:port]"},
		{name: "reject ping bad port", checkType: "ping", target: "example.com:99999", wantErr: "invalid target port"},
		{name: "reject unknown type", checkType: "smtp", target: "example.com", wantErr: "unsupported check type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateMonitorTarget(tt.checkType, tt.target)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLinkCheck(t *testing.T) {
	t.Run("healthy page and references", func(t *testing.T) {
		baseURL := "https://example.test"
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			assertMonitorUserAgent(t, req)
			path := req.URL.Path
			if path == "" {
				path = "/"
			}
			switch path {
			case "/":
				return htmlResponse(`<html><body><a href="/about">About</a><img src="/logo.png"></body></html>`), nil
			case "/about":
				return htmlResponse(`<html><body>About</body></html>`), nil
			case "/logo.png":
				return httpResponse(http.StatusOK), nil
			default:
				return httpResponse(http.StatusNotFound), nil
			}
		}))

		status, latency := LinkCheck(baseURL, defaultConnectionTimeout)

		assert.Equal(t, "200 OK", status)
		assert.GreaterOrEqual(t, latency, int64(0))
	})

	t.Run("reports broken references with source page", func(t *testing.T) {
		baseURL := "https://example.test"
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			path := req.URL.Path
			if path == "" {
				path = "/"
			}
			switch path {
			case "/":
				return htmlResponse(`<html><body><a href="/missing">Missing</a><img src="/broken.png"></body></html>`), nil
			case "/missing":
				return httpResponse(http.StatusNotFound), nil
			case "/broken.png":
				return httpResponse(http.StatusInternalServerError), nil
			default:
				return httpResponse(http.StatusNotFound), nil
			}
		}))

		status, _ := LinkCheck(baseURL, defaultConnectionTimeout)

		assert.Contains(t, status, "Broken links: 2")
		assert.Contains(t, status, "404 Not Found "+baseURL+"/missing from "+baseURL)
		assert.Contains(t, status, "500 Internal Server Error "+baseURL+"/broken.png from "+baseURL)
	})

	t.Run("crawls internal links", func(t *testing.T) {
		baseURL := "https://example.test"
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			path := req.URL.Path
			if path == "" {
				path = "/"
			}
			switch path {
			case "/":
				return htmlResponse(`<html><body><a href="/child">Child</a></body></html>`), nil
			case "/child":
				return htmlResponse(`<html><body><a href="/gone">Gone</a></body></html>`), nil
			case "/gone":
				return httpResponse(http.StatusNotFound), nil
			default:
				return httpResponse(http.StatusNotFound), nil
			}
		}))

		status, _ := LinkCheck(baseURL, defaultConnectionTimeout)

		assert.Contains(t, status, "Broken links: 1")
		assert.Contains(t, status, "404 Not Found "+baseURL+"/gone from "+baseURL+"/child")
	})

	t.Run("blocks redirect to private target", func(t *testing.T) {
		setDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Host == "example.test" {
				resp := httpResponse(http.StatusFound)
				resp.Header.Set("Location", "http://127.0.0.1/admin")
				return resp, nil
			}
			return httpResponse(http.StatusOK), nil
		}))

		status, _ := HttpCheck("https://example.test", defaultConnectionTimeout)

		assert.Contains(t, status, "private or internal address")
	})
}

func TestParseTLSTarget(t *testing.T) {
	t.Run("parses bare host and explicit port", func(t *testing.T) {
		address, serverName, err := parseTLSTarget("example.com:8443")

		assert.NoError(t, err)
		assert.Equal(t, "example.com:8443", address)
		assert.Equal(t, "example.com", serverName)
	})

	t.Run("uses default HTTPS port for URLs", func(t *testing.T) {
		address, serverName, err := parseTLSTarget("https://example.com/path")

		assert.NoError(t, err)
		assert.Equal(t, "example.com:443", address)
		assert.Equal(t, "example.com", serverName)
	})
}
