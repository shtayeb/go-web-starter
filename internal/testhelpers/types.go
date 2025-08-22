package testhelpers

import (
	"net/http/httptest"
)

// TestingT is an interface that covers both *testing.T and *testing.B
type TestingT interface {
	Cleanup(func())
	Error(args ...any)
	Errorf(format string, args ...any)
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Helper()
	Log(args ...any)
	Logf(format string, args ...any)
	Name() string
	Skip(args ...any)
	SkipNow()
	Skipf(format string, args ...any)
	Skipped() bool
	TempDir() string
}

// GenericPayload represents a generic JSON payload for HTTP requests
type GenericPayload map[string]interface{}

// HTTPResponse wraps httptest.ResponseRecorder with additional utilities
type HTTPResponse struct {
	*httptest.ResponseRecorder
}

// NewHTTPResponse creates a new HTTPResponse wrapper
func NewHTTPResponse(recorder *httptest.ResponseRecorder) *HTTPResponse {
	return &HTTPResponse{recorder}
}
