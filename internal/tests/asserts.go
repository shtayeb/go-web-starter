package tests

import (
	"net/http"
	"strings"
	"testing"
)

func AssertRedirect(t *testing.T, statusCode int, headers http.Header, expectedLocation string) {
	t.Helper()

	if statusCode != http.StatusSeeOther && statusCode != http.StatusFound {
		t.Errorf("expected redirect status (303 or 302); got %d", statusCode)
	}

	location := headers.Get("Location")
	if location != expectedLocation {
		t.Errorf("expected redirect to %s; got %s", expectedLocation, location)
	}
}

func AssertContains(t *testing.T, body, expected string) {
	t.Helper()

	if !strings.Contains(body, expected) {
		t.Errorf("expected body to contain %q", expected)
	}
}

func AssertNotContains(t *testing.T, body, shouldNotContain string) {
	t.Helper()

	if strings.Contains(body, shouldNotContain) {
		t.Errorf("expected body to NOT contain %q", shouldNotContain)
	}
}

func AssertStatus(t *testing.T, actual, expected int) {
	t.Helper()

	if actual != expected {
		t.Errorf("expected status %d; got %d", expected, actual)
	}
}
