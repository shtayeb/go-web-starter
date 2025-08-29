package auth_test

import (
	"net/http"
	"strings"
	"testing"

	"go-web-starter/internal/tests"
)

func TestSignUpHandler_GET(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	status, _, body := ts.Get(t, "/signup")

	if status != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, status)
	}

	// Check that signup form is rendered
	if !strings.Contains(body, "signup") || !strings.Contains(body, "Sign") {
		t.Error("expected signup form in response")
	}
}

func TestSignUpHandler_POST(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	testCases := []struct {
		name           string
		formData       map[string]string
		expectedStatus int
		expectRedirect bool
		expectEmail    bool
	}{
		{
			name: "valid signup",
			formData: map[string]string{
				"name":             "John Doe",
				"email":            "john@example.com",
				"password":         "Password123!",
				"confirm_password": "Password123!",
			},
			expectedStatus: http.StatusSeeOther,
			expectRedirect: true,
			expectEmail:    true,
		},
		{
			name: "password mismatch",
			formData: map[string]string{
				"name":             "Jane Doe",
				"email":            "jane@example.com",
				"password":         "Password123!",
				"confirm_password": "DifferentPassword!",
			},
			expectedStatus: http.StatusOK,
			expectRedirect: false,
			expectEmail:    false,
		},
		{
			name: "invalid email",
			formData: map[string]string{
				"name":             "Invalid User",
				"email":            "not-an-email",
				"password":         "Password123!",
				"confirm_password": "Password123!",
			},
			expectedStatus: http.StatusOK,
			expectRedirect: false,
			expectEmail:    false,
		},
		{
			name: "short password",
			formData: map[string]string{
				"name":             "Short Pass",
				"email":            "short@example.com",
				"password":         "Pass1!",
				"confirm_password": "Pass1!",
			},
			expectedStatus: http.StatusOK,
			expectRedirect: false,
			expectEmail:    false,
		},
		{
			name: "empty name",
			formData: map[string]string{
				"name":             "",
				"email":            "noname@example.com",
				"password":         "Password123!",
				"confirm_password": "Password123!",
			},
			expectedStatus: http.StatusOK,
			expectRedirect: false,
			expectEmail:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear emails before each test
			ts.Mailer.Clear()

			status, headers, _ := ts.PostForm(t, "/signup", tc.formData)

			if status != tc.expectedStatus {
				t.Errorf("expected status %d; got %d", tc.expectedStatus, status)
			}

			if tc.expectRedirect {
				location := headers.Get("Location")
				if location != "/login" {
					t.Errorf("expected redirect to /login; got %s", location)
				}
			}

			if tc.expectEmail {
				emails := ts.Mailer.GetSentEmails()
				if len(emails) == 0 {
					t.Errorf("sent emails:%v = expected welcome email to be sent", emails)
				} else {
					lastEmail := emails[0]
					if lastEmail.Recipient != tc.formData["email"] {
						t.Errorf("email sent to wrong recipient: %s", lastEmail.Recipient)
					}
					if lastEmail.TemplateFile != "user_welcome.tmpl" {
						t.Errorf("wrong email template: %s", lastEmail.TemplateFile)
					}
				}
			} else {
				if ts.Mailer.EmailCount() > 0 {
					t.Error("no email should be sent for invalid signup")
				}
			}
		})
	}
}

func TestLoginHandler_GET(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	status, _, body := ts.Get(t, "/login")

	if status != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, status)
	}

	// Check that login form is rendered
	if !strings.Contains(body, "login") || !strings.Contains(body, "Login") {
		t.Error("expected login form in response")
	}
}

func TestLoginHandler_POST(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create a test user
	ts.CreateTestUser(t, "Test User", "test@example.com", "SecurePassword123!")

	testCases := []struct {
		name             string
		formData         map[string]string
		expectedStatus   int
		expectedRedirect string
	}{
		{
			name: "valid login",
			formData: map[string]string{
				"email":    "test@example.com",
				"password": "SecurePassword123!",
			},
			expectedStatus:   http.StatusSeeOther,
			expectedRedirect: "/dashboard",
		},
		{
			name: "wrong password",
			formData: map[string]string{
				"email":    "test@example.com",
				"password": "WrongPassword123!",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "non-existent user",
			formData: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "Password123!",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty email",
			formData: map[string]string{
				"email":    "",
				"password": "SecurePassword123!",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty password",
			formData: map[string]string{
				"email":    "test@example.com",
				"password": "",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid email format",
			formData: map[string]string{
				"email":    "not-an-email",
				"password": "SecurePassword123!",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, headers, _ := ts.PostForm(t, "/login", tc.formData)

			if status != tc.expectedStatus {
				t.Errorf("expected status %d; got %d", tc.expectedStatus, status)
			}

			if tc.expectedStatus == http.StatusSeeOther {
				location := headers.Get("Location")
				if location != tc.expectedRedirect {
					t.Errorf("expected redirect to %s; got %s", tc.expectedRedirect, location)
				}
			}
		})
	}
}

func TestLoginHandler_WithNextParameter(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create a test user
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")

	// Login with next parameter
	formData := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}

	status, headers, _ := ts.PostForm(t, "/login?next=/projects", formData)

	if status != http.StatusSeeOther {
		t.Errorf("expected status %d; got %d", http.StatusSeeOther, status)
	}

	// Should redirect to the next URL
	location := headers.Get("Location")
	if location != "/projects" {
		t.Errorf("expected redirect to /projects; got %s", location)
	}
}

func TestLogoutHandler(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create and login a user
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")
	client := ts.LoginUser(t, "test@example.com", "password123")

	// Perform logout
	status, headers, _ := ts.PostFormWithClient(t, client, "/logout", nil)

	if status != http.StatusSeeOther {
		t.Errorf("expected status %d; got %d", http.StatusSeeOther, status)
	}

	location := headers.Get("Location")
	if location != "/" {
		t.Errorf("expected redirect to /; got %s", location)
	}

	// Try to access protected route after logout
	status, headers, _ = ts.GetWithClient(t, client, "/dashboard")

	// Should be redirected to login
	if status != http.StatusSeeOther {
		t.Errorf("expected redirect status; got %d", status)
	}

	location = headers.Get("Location")
	if location != "/login?next=/dashboard" {
		t.Errorf("expected redirect to login; got %s", location)
	}
}

func TestForgotPasswordHandler_GET(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	status, _, body := ts.Get(t, "/forgot-password")

	if status != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, status)
	}

	// Check that forgot password form is rendered
	if !strings.Contains(body, "Forgot") || !strings.Contains(body, "Password") {
		t.Error("expected forgot password form in response")
	}
}

func TestForgotPasswordHandler_POST(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create a test user
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")

	testCases := []struct {
		name        string
		formData    map[string]string
		expectEmail bool
	}{
		{
			name: "existing user",
			formData: map[string]string{
				"email": "test@example.com",
			},
			expectEmail: true,
		},
		{
			name: "non-existent user",
			formData: map[string]string{
				"email": "nonexistent@example.com",
			},
			expectEmail: false,
		},
		{
			name: "invalid email format",
			formData: map[string]string{
				"email": "not-an-email",
			},
			expectEmail: false,
		},
		{
			name: "empty email",
			formData: map[string]string{
				"email": "",
			},
			expectEmail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts.Mailer.Clear()

			status, _, _ := ts.PostForm(t, "/forgot-password", tc.formData)

			if status != http.StatusOK {
				t.Errorf("expected status %d; got %d", http.StatusOK, status)
			}

			if tc.expectEmail {
				emails := ts.Mailer.GetSentEmails()
				if len(emails) == 0 {
					t.Error("expected password reset email to be sent")
				} else {
					lastEmail := emails[0]
					if lastEmail.Recipient != tc.formData["email"] {
						t.Errorf("email sent to wrong recipient: %s", lastEmail.Recipient)
					}
					if lastEmail.TemplateFile != "reset_password.tmpl" {
						t.Errorf("wrong email template: %s", lastEmail.TemplateFile)
					}
				}
			} else {
				if ts.Mailer.EmailCount() > 0 {
					t.Error("no email should be sent for invalid request")
				}
			}
		})
	}
}

func TestProfileHandler_GET(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Test without authentication
	status, headers, _ := ts.Get(t, "/profile")

	if status != http.StatusSeeOther {
		t.Errorf("expected status %d; got %d", http.StatusSeeOther, status)
	}

	location := headers.Get("Location")
	if location != "/login?next=/profile" {
		t.Errorf("expected redirect to login; got %s", location)
	}

	// Test with authentication
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")
	client := ts.LoginUser(t, "test@example.com", "password123")

	status, _, body := ts.GetWithClient(t, client, "/profile")

	if status != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, status)
	}

	// Check that profile page is rendered
	if !strings.Contains(body, "Profile") || !strings.Contains(body, "Update your account settings") {
		t.Error("expected profile page with user data")
	}
}

func TestUpdateUserNameAndImageHandler(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create and login a user
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")
	client := ts.LoginUser(t, "test@example.com", "password123")

	testCases := []struct {
		name           string
		formData       map[string]string
		expectedStatus int
	}{
		{
			name: "valid update",
			formData: map[string]string{
				"name":  "Updated Name",
				"image": "https://example.com/avatar.jpg",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty name",
			formData: map[string]string{
				"name":  "",
				"image": "https://example.com/avatar.jpg",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty image",
			formData: map[string]string{
				"name":  "Updated Name",
				"image": "",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, _, _ := ts.PostFormWithClient(t, client, "/profile/update", tc.formData)

			if status != tc.expectedStatus {
				t.Errorf("expected status %d; got %d", tc.expectedStatus, status)
			}
		})
	}
}

func TestUpdateAccountPasswordHandler(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create and login a user
	ts.CreateTestUser(t, "Test User", "test@example.com", "OldPassword123!")
	client := ts.LoginUser(t, "test@example.com", "OldPassword123!")

	testCases := []struct {
		name           string
		formData       map[string]string
		expectedStatus int
	}{
		{
			name: "valid password change",
			formData: map[string]string{
				"current_password": "OldPassword123!",
				"new_password":     "NewPassword123!",
				"confirm_password": "NewPassword123!",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "wrong current password",
			formData: map[string]string{
				"current_password": "WrongPassword123!",
				"new_password":     "NewPassword123!",
				"confirm_password": "NewPassword123!",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty current password",
			formData: map[string]string{
				"current_password": "",
				"new_password":     "NewPassword123!",
				"confirm_password": "NewPassword123!",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty new password",
			formData: map[string]string{
				"current_password": "OldPassword123!",
				"new_password":     "",
				"confirm_password": "",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, _, _ := ts.PostFormWithClient(t, client, "/profile/update-password", tc.formData)

			if status != tc.expectedStatus {
				t.Errorf("expected status %d; got %d", tc.expectedStatus, status)
			}
		})
	}
}

func TestDeleteAccountHandler(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create and login a user
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")
	client := ts.LoginUser(t, "test@example.com", "password123")

	// Test with wrong password
	formData := map[string]string{
		"password": "wrongpassword",
	}

	status, _, _ := ts.PostFormWithClient(t, client, "/profile/delete-account", formData)

	if status != http.StatusOK {
		t.Errorf("expected status %d for wrong password; got %d", http.StatusOK, status)
	}

	// Test with correct password
	formData["password"] = "password123"

	status, headers, _ := ts.PostFormWithClient(t, client, "/profile/delete-account", formData)

	// Should redirect to home after deletion
	if status != http.StatusSeeOther {
		t.Errorf("expected redirect status; got %d", status)
	}

	location := headers.Get("Location")
	if location != "/" {
		t.Errorf("expected redirect to /; got %s", location)
	}
}

func TestRequireNoAuthMiddleware(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create and login a user
	ts.CreateTestUser(t, "Test User", "test@example.com", "password123")
	client := ts.LoginUser(t, "test@example.com", "password123")

	// Try to access login page while authenticated
	status, headers, _ := ts.GetWithClient(t, client, "/login")

	// Should redirect away from login page
	if status != http.StatusSeeOther {
		t.Errorf("expected redirect status; got %d", status)
	}

	// Should redirect to referer or home
	location := headers.Get("Location")
	if location != "/" && !strings.Contains(location, "/") {
		t.Errorf("expected redirect to home or referer; got %s", location)
	}

	// Same for signup page
	status, _, _ = ts.GetWithClient(t, client, "/signup")

	if status != http.StatusSeeOther {
		t.Errorf("expected redirect status for signup; got %d", status)
	}
}

func TestDuplicateUserSignup(t *testing.T) {
	ts := tests.NewTestServer(t)
	defer ts.Close()

	// Create first user
	ts.CreateTestUser(t, "First User", "duplicate@example.com", "password123")

	// Try to create another user with same email
	formData := map[string]string{
		"name":             "Second User",
		"email":            "duplicate@example.com",
		"password":         "Password456!",
		"confirm_password": "Password456!",
	}

	status, _, _ := ts.PostForm(t, "/signup", formData)

	// Should fail with status OK (form with error)
	if status != http.StatusOK {
		t.Errorf("expected status %d for duplicate email; got %d", http.StatusOK, status)
	}

	// No welcome email should be sent
	emails := ts.Mailer.GetSentEmails()
	if len(emails) > 0 {
		t.Error("no email should be sent for duplicate signup")
	}
}
