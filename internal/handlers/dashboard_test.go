package handlers

import (
	"net/http"
	"testing"

	"go-web-starter/internal/testhelpers"

	"github.com/stretchr/testify/assert"
)

// TestDashboardRouteUnauthenticated tests dashboard access without authentication
func TestDashboardRouteUnauthenticated(t *testing.T) {
	testhelpers.WithTestServer(t, func(s *testhelpers.TestServer) {
		// Test GET request to dashboard without authentication
		res := testhelpers.PerformRequest(t, s, "GET", "/dashboard", nil, nil)

		// Should redirect to login page since it's a protected route
		testhelpers.AssertStatusCode(t, res, http.StatusSeeOther)

		// Check redirect location
		location := res.Result().Header.Get("Location")
		assert.Contains(t, location, "/login", "Should redirect to login page")
		assert.Contains(t, location, "next=/dashboard", "Should include next parameter")
	})
}

// TestDashboardRouteAuthenticated tests dashboard access with authentication
func TestDashboardRouteAuthenticated(t *testing.T) {
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

		// Now test dashboard with authenticated user
		headers := map[string]string{}
		if sessionCookie != nil {
			headers["Cookie"] = sessionCookie.String()
		}

		res = testhelpers.PerformRequest(t, s, "GET", "/dashboard", nil, headers)

		// Should return 200 OK since user is authenticated
		testhelpers.AssertStatusCode(t, res, http.StatusOK)

		// Verify response contains expected content
		responseBody := testhelpers.GetResponseBody(res)
		assert.Contains(t, responseBody, "Dashboard", "Dashboard page should contain dashboard content")
		assert.Contains(t, responseBody, "html", "Response should be HTML content")
	})
}

// TestDashboardRouteDifferentMethods tests dashboard with different HTTP methods
func TestDashboardRouteDifferentMethods(t *testing.T) {
	testhelpers.WithTestServer(t, func(s *testhelpers.TestServer) {
		// Test unauthenticated requests
		testCases := []struct {
			method   string
			expected int
		}{
			{"GET", http.StatusSeeOther},    // Should redirect to login
			{"POST", http.StatusSeeOther},   // Should redirect to login
			{"PUT", http.StatusSeeOther},    // Should redirect to login
			{"DELETE", http.StatusSeeOther}, // Should redirect to login
		}

		for _, tc := range testCases {
			t.Run("Unauthenticated_"+tc.method, func(t *testing.T) {
				res := testhelpers.PerformRequest(t, s, tc.method, "/dashboard", nil, nil)
				testhelpers.AssertStatusCode(t, res, tc.expected)

				// Verify redirect to login
				location := res.Result().Header.Get("Location")
				assert.Contains(t, location, "/login", "Should redirect to login page")
			})
		}
	})
}

// TestDashboardRouteWithInvalidSession tests dashboard with invalid session
func TestDashboardRouteWithInvalidSession(t *testing.T) {
	testhelpers.WithTestServer(t, func(s *testhelpers.TestServer) {
		// Test with invalid session cookie
		headers := map[string]string{
			"Cookie": "session=invalid-session-id",
		}

		res := testhelpers.PerformRequest(t, s, "GET", "/dashboard", nil, headers)

		// Should redirect to login since session is invalid
		testhelpers.AssertStatusCode(t, res, http.StatusSeeOther)

		location := res.Result().Header.Get("Location")
		assert.Contains(t, location, "/login", "Should redirect to login page")
	})
}

// TestDashboardRouteHeaders tests dashboard response headers
func TestDashboardRouteHeaders(t *testing.T) {
	testhelpers.WithTestServer(t, func(s *testhelpers.TestServer) {
		// First authenticate a user
		username := testhelpers.RandomUsername(t)
		email := testhelpers.RandomEmail(t)
		password := testhelpers.RandomPassword(t)

		// Register and login
		registerPayload := testhelpers.GenericPayload{
			"username": username,
			"email":    email,
			"password": password,
		}

		res := testhelpers.PerformRequest(t, s, "POST", "/signup", registerPayload, nil)
		testhelpers.AssertStatusCode(t, res, http.StatusSeeOther)

		loginPayload := testhelpers.GenericPayload{
			"email":    email,
			"password": password,
		}

		res = testhelpers.PerformRequest(t, s, "POST", "/login", loginPayload, nil)
		testhelpers.AssertStatusCode(t, res, http.StatusSeeOther)

		// Extract session cookie
		cookies := res.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "session" {
				sessionCookie = cookie
				break
			}
		}

		// Test dashboard with authentication
		headers := map[string]string{}
		if sessionCookie != nil {
			headers["Cookie"] = sessionCookie.String()
		}

		res = testhelpers.PerformRequest(t, s, "GET", "/dashboard", nil, headers)

		// Check for expected headers
		assert.Equal(t, "text/html; charset=utf-8", res.Result().Header.Get("Content-Type"))
		assert.Contains(t, res.Result().Header.Get("Cache-Control"), "no-store", "Protected routes should have no-store cache control")
	})
}

// TestDashboardRouteAfterLogout tests dashboard access after logout
func TestDashboardRouteAfterLogout(t *testing.T) {
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
		testhelpers.AssertStatusCode(t, res, http.StatusSeeOther)

		// Login user
		loginPayload := testhelpers.GenericPayload{
			"email":    email,
			"password": password,
		}

		res = testhelpers.PerformRequest(t, s, "POST", "/login", loginPayload, nil)
		testhelpers.AssertStatusCode(t, res, http.StatusSeeOther)

		// Extract session cookie
		cookies := res.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "session" {
				sessionCookie = cookie
				break
			}
		}

		// Logout user
		headers := map[string]string{}
		if sessionCookie != nil {
			headers["Cookie"] = sessionCookie.String()
		}

		res = testhelpers.PerformRequest(t, s, "POST", "/logout", nil, headers)
		testhelpers.AssertStatusCode(t, res, http.StatusSeeOther) // Should redirect after logout

		// Extract new session cookie after logout
		cookies = res.Result().Cookies()
		var newSessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "session" {
				newSessionCookie = cookie
				break
			}
		}

		// Try to access dashboard with the new session (should be logged out)
		headers = map[string]string{}
		if newSessionCookie != nil {
			headers["Cookie"] = newSessionCookie.String()
		}

		res = testhelpers.PerformRequest(t, s, "GET", "/dashboard", nil, headers)

		// Should redirect to login since user is logged out
		testhelpers.AssertStatusCode(t, res, http.StatusSeeOther)

		location := res.Result().Header.Get("Location")
		assert.Contains(t, location, "/login", "Should redirect to login page after logout")
	})
}
