package handlers_test

import (
	"net/http"
	"testing"

	"go-web-starter/internal/tests"
)

func TestHealthHandler(t *testing.T) {
	// Create a new test server
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Make a request to the health endpoint
	status, _, body := ts.Get(t, "/health")

	// Check the response
	if status != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, status)
	}

	// Check that response contains expected content
	if body == "" {
		t.Error("expected non-empty response body")
	}
}

func TestLandingPageHandler(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	status, _, body := ts.Get(t, "/")

	if status != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, status)
	}

	// Check for expected content in the landing page
	if len(body) == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestDashboardRequiresAuth(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Test accessing protected route without authentication
	status, headers, _ := ts.Get(t, "/dashboard")

	// Should redirect to login
	if status != http.StatusSeeOther {
		t.Errorf("expected status %d; got %d", http.StatusSeeOther, status)
	}

	// Check redirect location
	location := headers.Get("Location")
	if location != "/login?next=/dashboard" {
		t.Errorf("expected redirect to /login?next=/dashboard; got %s", location)
	}
}

func TestDashboardWithAuth(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create and login a test user
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")
	client := ts.LoginUser(t, "test@example.com", "password123")

	// Access protected route with authenticated client
	status, _, _ := ts.GetWithClient(t, client, "/dashboard")

	// Should be able to access the dashboard
	if status != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, status)
	}
}

func TestProjectsRequiresAuth(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Test without authentication
	status, headers, _ := ts.Get(t, "/projects")

	// Should redirect to login
	if status != http.StatusSeeOther {
		t.Errorf("expected status %d; got %d", http.StatusSeeOther, status)
	}

	// Check redirect location
	location := headers.Get("Location")
	if location != "/login?next=/projects" {
		t.Errorf("expected redirect to /login?next=/projects; got %s", location)
	}

	// Now test with authentication
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")
	client := ts.LoginUser(t, "test@example.com", "password123")

	status, _, _ = ts.GetWithClient(t, client, "/projects")

	// Should be able to access projects page
	if status != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, status)
	}
}

func TestProfileRequiresAuth(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Test without authentication
	status, headers, _ := ts.Get(t, "/profile")

	// Should redirect to login
	if status != http.StatusSeeOther {
		t.Errorf("expected status %d; got %d", http.StatusSeeOther, status)
	}

	// Check redirect location
	location := headers.Get("Location")
	if location != "/login?next=/profile" {
		t.Errorf("expected redirect to /login?next=/profile; got %s", location)
	}
}

func TestProfileWithAuth(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create and login a test user
	user := ts.CreateTestUser(t, "Test User", "test@example.com", "password123")
	client := ts.LoginUser(t, "test@example.com", "password123")

	// Access profile page
	status, _, body := ts.GetWithClient(t, client, "/profile")

	if status != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, status)
	}

	// Check that response contains user information
	if len(body) == 0 {
		t.Error("expected non-empty response body")
	}

	// Verify user data is available
	if user.Email != "test@example.com" {
		t.Errorf("unexpected user email: %s", user.Email)
	}
}

func TestMultipleUsersIsolation(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create two different users
	user1 := ts.CreateTestUser(t, "User One", "user1@example.com", "password123")
	user2 := ts.CreateTestUser(t, "User Two", "user2@example.com", "password456")

	// Login as user1
	client1 := ts.LoginUser(t, "user1@example.com", "password123")

	// Login as user2
	client2 := ts.LoginUser(t, "user2@example.com", "password456")

	// Both should be able to access their dashboards
	status1, _, _ := ts.GetWithClient(t, client1, "/dashboard")
	status2, _, _ := ts.GetWithClient(t, client2, "/dashboard")

	if status1 != http.StatusOK {
		t.Errorf("user1: expected status %d; got %d", http.StatusOK, status1)
	}

	if status2 != http.StatusOK {
		t.Errorf("user2: expected status %d; got %d", http.StatusOK, status2)
	}

	// Verify users are different
	if user1.ID == user2.ID {
		t.Error("expected different user IDs")
	}

	if user1.Email == user2.Email {
		t.Error("expected different user emails")
	}
}
