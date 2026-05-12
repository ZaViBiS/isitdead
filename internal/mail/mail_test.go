package mail

import (
	"strings"
	"testing"
)

func TestBuildVerificationMessageIncludesRFC5322Headers(t *testing.T) {
	msg := string(buildVerificationMessage(
		"no-reply@isitdead.cc",
		"user@example.com",
		"<html><body>Confirm</body></html>",
	))

	for _, header := range []string{
		"From: no-reply@isitdead.cc\r\n",
		"To: user@example.com\r\n",
		"Subject: Confirm your email for isitdead.cc\r\n",
		"Date: ",
		"MIME-Version: 1.0\r\n",
		"Content-Type: text/html; charset=UTF-8\r\n",
	} {
		if !strings.Contains(msg, header) {
			t.Fatalf("message is missing header %q:\n%s", header, msg)
		}
	}

	if !strings.Contains(msg, "\r\n\r\n<html><body>Confirm</body></html>") {
		t.Fatalf("message does not separate headers and body with CRLF CRLF:\n%s", msg)
	}
}
