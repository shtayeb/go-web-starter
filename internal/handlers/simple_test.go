package handlers_test

import (
	"net/http"
	"testing"

	"go-web-starter/internal/tests"
)

func TestSimpleHealthCheck(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	status, _, body := ts.Get(t, "/health")

	if status != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, status)
	}

	if body == "" {
		t.Error("expected non-empty response body")
	}
}

func TestSimpleLandingPage(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	status, _, body := ts.Get(t, "/")

	if status != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, status)
	}

	if len(body) == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestSimpleLogin(t *testing.T) {
	ts := tests.NewTestServerNoCSRF(t)
	defer ts.Close()

	// Create a test user
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")

	// Test login
	client := ts.LoginUserSimple(t, "test@example.com", "password123")

	// Try to access protected route
	status, _, _ := ts.GetWithClient(t, client, "/dashboard")

	if status != http.StatusOK {
		t.Errorf("expected to access dashboard after login; got status %d", status)
	}
}

func TestSimpleProtectedRoute(t *testing.T) {
	ts := tests.NewTestServerNoCSRF(t)
	defer ts.Close()

	// Test without authentication
	status, headers, _ := ts.Get(t, "/dashboard")

	if status != http.StatusSeeOther {
		t.Errorf("expected redirect status %d; got %d", http.StatusSeeOther, status)
	}

	location := headers.Get("Location")
	if location != "/login?next=/dashboard" {
		t.Errorf("expected redirect to login; got %s", location)
	}
}

func TestSimpleSignup(t *testing.T) {
	ts := tests.NewTestServerNoCSRF(t)
	defer ts.Close()

	formData := map[string]string{
		"name":             "New User",
		"email":            "newuser@example.com",
		"password":         "Password123!",
		"confirm_password": "Password123!",
	}

	status, headers, _ := ts.PostFormSimple(t, nil, "/signup", formData)

	if status != http.StatusSeeOther {
		t.Errorf("expected redirect after signup; got status %d", status)
	}

	location := headers.Get("Location")
	if location != "/login" {
		t.Errorf("expected redirect to /login; got %s", location)
	}

}

func TestSimpleLogout(t *testing.T) {
	ts := tests.NewTestServerNoCSRF(t)
	defer ts.Close()

	// Create and login a user
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")
	client := ts.LoginUserSimple(t, "test@example.com", "password123")

	// Logout
	status, headers, _ := ts.PostFormSimple(t, client, "/logout", nil)

	if status != http.StatusSeeOther {
		t.Errorf("expected redirect after logout; got status %d", status)
	}

	location := headers.Get("Location")
	if location != "/" {
		t.Errorf("expected redirect to home; got %s", location)
	}

	// Try to access protected route after logout
	status, headers, _ = ts.GetWithClient(t, client, "/dashboard")

	if status != http.StatusSeeOther {
		t.Errorf("expected redirect after logout; got status %d", status)
	}

	location = headers.Get("Location")
	if location != "/login?next=/dashboard" {
		t.Errorf("expected redirect to login; got %s", location)
	}
}

func TestSimpleProfile(t *testing.T) {
	ts := tests.NewTestServerNoCSRF(t)
	defer ts.Close()

	// Create and login a user
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")
	client := ts.LoginUserSimple(t, "test@example.com", "password123")

	// Access profile
	status, _, body := ts.GetWithClient(t, client, "/profile")

	if status != http.StatusOK {
		t.Errorf("expected to access profile; got status %d", status)
	}

	if len(body) == 0 {
		t.Error("expected profile page content")
	}

	tests.AssertContains(t, string(body), "Update your account settings")
	tests.AssertContains(t, string(body), "Change your password here. After saving, you will be logged out.")
	tests.AssertContains(t, string(body), "Danger Zone")
}

func TestSimpleUpdateProfile(t *testing.T) {
	ts := tests.NewTestServerNoCSRF(t)
	defer ts.Close()

	// Create and login a user
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")
	client := ts.LoginUserSimple(t, "test@example.com", "password123")

	// Update profile
	formData := map[string]string{
		"name":  "Updated Name",
		"image": "https://example.com/avatar.jpg",
	}

	status, _, _ := ts.PostFormSimple(t, client, "/profile/update", formData)

	if status != http.StatusOK {
		t.Errorf("expected successful profile update; got status %d", status)
	}
}

func TestSimplePasswordUpdate(t *testing.T) {
	ts := tests.NewTestServerNoCSRF(t)
	defer ts.Close()

	// Create and login a user
	ts.CreateTestUser(t, "Test User", "test@example.com", "OldPassword123!")
	client := ts.LoginUserSimple(t, "test@example.com", "OldPassword123!")

	// Update password
	formData := map[string]string{
		"current_password": "OldPassword123!",
		"new_password":     "NewPassword123!",
		"confirm_password": "NewPassword123!",
	}

	status, _, _ := ts.PostFormSimple(t, client, "/profile/update-password", formData)

	if status != http.StatusOK {
		t.Errorf("expected successful password update; got status %d", status)
	}

	// Try to login with new password
	client2 := ts.LoginUserSimple(t, "test@example.com", "NewPassword123!")

	// Access dashboard with new login
	status, _, _ = ts.GetWithClient(t, client2, "/dashboard")

	if status != http.StatusOK {
		t.Error("expected to login with new password")
	}
}

func TestSimpleMultipleUsers(t *testing.T) {
	ts := tests.NewTestServerNoCSRF(t)
	defer ts.Close()

	// Create multiple users
	ts.CreateTestUser(t, "User One", "user1@example.com", "password1")
	ts.CreateTestUser(t, "User Two", "user2@example.com", "password2")
	ts.CreateTestUser(t, "User Three", "user3@example.com", "password3")

	// Login as different users
	client1 := ts.LoginUserSimple(t, "user1@example.com", "password1")
	client2 := ts.LoginUserSimple(t, "user2@example.com", "password2")
	client3 := ts.LoginUserSimple(t, "user3@example.com", "password3")

	// All should be able to access dashboard
	status1, _, _ := ts.GetWithClient(t, client1, "/dashboard")
	status2, _, _ := ts.GetWithClient(t, client2, "/dashboard")
	status3, _, _ := ts.GetWithClient(t, client3, "/dashboard")

	if status1 != http.StatusOK {
		t.Errorf("user1 cannot access dashboard; got status %d", status1)
	}

	if status2 != http.StatusOK {
		t.Errorf("user2 cannot access dashboard; got status %d", status2)
	}

	if status3 != http.StatusOK {
		t.Errorf("user3 cannot access dashboard; got status %d", status3)
	}
}
