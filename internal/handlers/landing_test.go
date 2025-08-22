package handlers

import (
	"net/http"
	"testing"

	"go-web-starter/internal/testhelpers"

	"github.com/stretchr/testify/assert"
)

// TestLandingRoute tests the landing page route
func TestLandingRoute(t *testing.T) {
	testhelpers.WithTestServer(t, func(s *testhelpers.TestServer) {
		// Test GET request to landing page (root path)
		res := testhelpers.PerformRequest(t, s, "GET", "/", nil, nil)

		// Should return 200 OK since it's a public route
		testhelpers.AssertStatusCode(t, res, http.StatusOK)

		// Verify response contains expected content
		responseBody := testhelpers.GetResponseBody(res)
		assert.Contains(t, responseBody, "Welcome", "Landing page should contain welcome content")
		assert.Contains(t, responseBody, "html", "Response should be HTML content")
	})
}

// TestLandingRouteWithAuthenticatedUser tests landing page with authenticated user
func TestLandingRouteWithAuthenticatedUser(t *testing.T) {
	testhelpers.WithTestServer(t, func(s *testhelpers.TestServer) {
		// First, create and authenticate a user
		username := testhelpers.RandomUsername(t)
		email := testhelpers.RandomEmail(t)
		password := testhelpers.RandomPassword(t)

		// Register user
		registerPayload := testhelpers.GenericPayload{
			"username": username,
			"email":    email,
			"password": password,
		}

		res := testhelpers.PerformRequest(t, s, "POST", "/signup", registerPayload, nil)
		testhelpers.AssertStatusCode(t, res, http.StatusSeeOther) // Should redirect after registration

		// Login user
		loginPayload := testhelpers.GenericPayload{
			"email":    email,
			"password": password,
		}

		res = testhelpers.PerformRequest(t, s, "POST", "/login", loginPayload, nil)
		testhelpers.AssertStatusCode(t, res, http.StatusSeeOther) // Should redirect after login

		// Extract session cookie from login response
		cookies := res.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "session" {
				sessionCookie = cookie
				break
			}
		}

		// Now test landing page with authenticated user
		headers := map[string]string{}
		if sessionCookie != nil {
			headers["Cookie"] = sessionCookie.String()
		}

		res = testhelpers.PerformRequest(t, s, "GET", "/", nil, headers)

		// Should still return 200 OK
		testhelpers.AssertStatusCode(t, res, http.StatusOK)

		// Verify response contains expected content
		responseBody := testhelpers.GetResponseBody(res)
		assert.Contains(t, responseBody, "Welcome", "Landing page should contain welcome content")
		assert.Contains(t, responseBody, "html", "Response should be HTML content")
	})
}

// TestLandingRouteDifferentMethods tests landing page with different HTTP methods
func TestLandingRouteDifferentMethods(t *testing.T) {
	testhelpers.WithTestServer(t, func(s *testhelpers.TestServer) {
		testCases := []struct {
			method   string
			expected int
		}{
			{"GET", http.StatusOK},
			{"POST", http.StatusMethodNotAllowed},
			{"PUT", http.StatusMethodNotAllowed},
			{"DELETE", http.StatusMethodNotAllowed},
		}

		for _, tc := range testCases {
			t.Run(tc.method, func(t *testing.T) {
				res := testhelpers.PerformRequest(t, s, tc.method, "/", nil, nil)
				testhelpers.AssertStatusCode(t, res, tc.expected)
			})
		}
	})
}

// TestLandingRouteHeaders tests landing page response headers
func TestLandingRouteHeaders(t *testing.T) {
	testhelpers.WithTestServer(t, func(s *testhelpers.TestServer) {
		res := testhelpers.PerformRequest(t, s, "GET", "/", nil, nil)

		// Check for expected headers
		assert.Equal(t, "text/html; charset=utf-8", res.Result().Header.Get("Content-Type"))
		assert.Contains(t, res.Result().Header.Get("Cache-Control"), "no-store")
	})
}
