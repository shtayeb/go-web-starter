package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PerformRequest performs an HTTP request against the test server
func PerformRequest(t TestingT, server *TestServer, method, path string, payload interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var body io.Reader

	if payload != nil {
		jsonData, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, server.URL()+path, body)
	require.NoError(t, err)

	// Set default headers
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Get the handler from the server
	handler := server.httpServer.Config.Handler

	// Serve the request
	handler.ServeHTTP(rr, req)

	return rr
}

// PerformFormRequest performs a form-encoded HTTP request
func PerformFormRequest(t TestingT, server *TestServer, method, path string, formData url.Values, headers map[string]string) *httptest.ResponseRecorder {
	var body io.Reader
	if formData != nil {
		body = strings.NewReader(formData.Encode())
	}

	req, err := http.NewRequest(method, server.URL()+path, body)
	require.NoError(t, err)

	// Set form content type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rr := httptest.NewRecorder()
	handler := server.httpServer.Config.Handler
	handler.ServeHTTP(rr, req)

	return rr
}

// ParseResponseAndValidate parses JSON response and validates status code
func ParseResponseAndValidate(t TestingT, response *httptest.ResponseRecorder, target interface{}) {
	require.Equal(t, http.StatusOK, response.Code, "Expected status OK, got %d. Response: %s", response.Code, response.Body.String())

	err := json.Unmarshal(response.Body.Bytes(), target)
	require.NoError(t, err, "Failed to parse response JSON: %s", response.Body.String())
}

// AssertStatusCode asserts the response status code
func AssertStatusCode(t TestingT, response *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(t, expectedStatus, response.Code, "Expected status %d, got %d. Response: %s", expectedStatus, response.Code, response.Body.String())
}

// AssertJSONResponse asserts that the response is valid JSON and matches expected structure
func AssertJSONResponse(t TestingT, response *httptest.ResponseRecorder, expectedStatus int) {
	AssertStatusCode(t, response, expectedStatus)

	var jsonResponse interface{}
	err := json.Unmarshal(response.Body.Bytes(), &jsonResponse)
	require.NoError(t, err, "Response is not valid JSON: %s", response.Body.String())
}

// AssertErrorResponse asserts that the response contains an error
func AssertErrorResponse(t TestingT, response *httptest.ResponseRecorder, expectedStatus int, expectedError string) {
	AssertStatusCode(t, response, expectedStatus)

	var errorResponse map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &errorResponse)
	require.NoError(t, err, "Response is not valid JSON: %s", response.Body.String())

	if expectedError != "" {
		errorMsg, exists := errorResponse["error"]
		require.True(t, exists, "Expected error field in response")
		assert.Contains(t, fmt.Sprintf("%v", errorMsg), expectedError)
	}
}

// AssertSuccessResponse asserts that the response indicates success
func AssertSuccessResponse(t TestingT, response *httptest.ResponseRecorder) {
	AssertStatusCode(t, response, http.StatusOK)

	var successResponse map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &successResponse)
	require.NoError(t, err, "Response is not valid JSON: %s", response.Body.String())

	// Check for common success indicators
	if success, exists := successResponse["success"]; exists {
		assert.True(t, success.(bool), "Expected success to be true")
	}
}

// GetResponseBody returns the response body as a string
func GetResponseBody(response *httptest.ResponseRecorder) string {
	return response.Body.String()
}

// GetResponseJSON parses and returns the response as a map
func GetResponseJSON(t TestingT, response *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &result)
	require.NoError(t, err, "Failed to parse response as JSON: %s", response.Body.String())
	return result
}

// AssertResponseContains asserts that the response body contains specific text
func AssertResponseContains(t TestingT, response *httptest.ResponseRecorder, expectedText string) {
	body := GetResponseBody(response)
	assert.Contains(t, body, expectedText, "Response body does not contain expected text")
}

// AssertResponseNotContains asserts that the response body does not contain specific text
func AssertResponseNotContains(t TestingT, response *httptest.ResponseRecorder, unexpectedText string) {
	body := GetResponseBody(response)
	assert.NotContains(t, body, unexpectedText, "Response body contains unexpected text")
}

// CreateAuthHeaders creates headers with authentication token
func CreateAuthHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
}

// CreateSessionHeaders creates headers with session cookie
func CreateSessionHeaders(sessionID string) map[string]string {
	return map[string]string{
		"Cookie": fmt.Sprintf("session=%s", sessionID),
	}
}
