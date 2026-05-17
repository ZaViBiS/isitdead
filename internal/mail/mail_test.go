package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ZaViBiS/isitdead/internal/config"
)

func TestSendHTMLUsesResendAPI(t *testing.T) {
	var got sendEmailRequest
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer re_test" {
			t.Fatalf("authorization header = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("User-Agent") != "isitdead/1.0" {
			t.Fatalf("user-agent header = %q", r.Header.Get("User-Agent"))
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		return response(http.StatusOK, ""), nil
	})}

	mailer := &Mailer{
		cfg: &config.Config{
			ResendAPIKey: "re_test",
			ResendFrom:   "Is It Dead <no-reply@isitdead.cc>",
		},
		client:   client,
		endpoint: "https://api.resend.test/emails",
	}

	if err := mailer.SendHTML("user@example.com", "Hello", "<p>Hi</p>"); err != nil {
		t.Fatalf("SendHTML() error = %v", err)
	}

	if got.From != "Is It Dead <no-reply@isitdead.cc>" {
		t.Fatalf("from = %q", got.From)
	}
	if len(got.To) != 1 || got.To[0] != "user@example.com" {
		t.Fatalf("to = %#v", got.To)
	}
	if got.Subject != "Hello" || got.HTML != "<p>Hi</p>" {
		t.Fatalf("payload = %#v", got)
	}
}

func TestSendHTMLReturnsResendErrors(t *testing.T) {
	mailer := &Mailer{
		cfg: &config.Config{
			ResendAPIKey: "re_test",
			ResendFrom:   "no-reply@isitdead.cc",
		},
		client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return response(http.StatusForbidden, `{"message":"domain is not verified"}`), nil
		})},
		endpoint: "https://api.resend.test/emails",
	}

	err := mailer.SendHTML("user@example.com", "Hello", "<p>Hi</p>")
	if err == nil || !strings.Contains(err.Error(), "403 Forbidden") {
		t.Fatalf("error = %v, want resend status", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}
