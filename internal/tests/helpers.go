package tests

import (
	"go-web-starter/internal/queries"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// ExtractCSRFToken extracts the CSRF token from an HTML response
// It looks for a hidden input field with name="csrf_token" or similar patterns
func ExtractCSRFToken(t *testing.T, html string) string {
	// Log first 500 chars of HTML for debugging
	if t != nil && testing.Verbose() {
		htmlPreview := html
		if len(htmlPreview) > 500 {
			htmlPreview = htmlPreview[:500] + "..."
		}
		t.Logf("ExtractCSRFToken: Searching in HTML (first 500 chars): %s", htmlPreview)
	}

	// Try multiple patterns that nosurf might use
	patterns := []string{
		`<input[^>]*name="csrf_token"[^>]*value="([^"]+)"`,
		`<input[^>]*value="([^"]+)"[^>]*name="csrf_token"`,
		`name="csrf_token"[^>]*value="([^"]+)"`,
		`csrf_token["']?\s*:\s*["']([^"']+)["']`,
	}

	for i, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			if t != nil && testing.Verbose() {
				t.Logf("ExtractCSRFToken: Found token with pattern %d: %s", i, matches[1])
			}
			return matches[1]
		}
	}

	if t != nil && testing.Verbose() {
		t.Logf("ExtractCSRFToken: No CSRF token found in HTML")
	}
	return ""
}

// GetPageWithCSRF fetches a page and extracts the CSRF token
func (ts *TestServer) GetPageWithCSRF(t *testing.T, client *http.Client, urlPath string) (string, error) {
	t.Helper()

	if client == nil {
		client = ts.Client
	}

	req, err := http.NewRequest(http.MethodGet, ts.Server.URL+urlPath, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if testing.Verbose() {
		t.Logf("GetPageWithCSRF: Response status: %d", resp.StatusCode)
		t.Logf("GetPageWithCSRF: Cookies received from server:")
		for _, cookie := range resp.Cookies() {
			t.Logf("  - %s = %s (Path: %s, HttpOnly: %v)", cookie.Name, cookie.Value, cookie.Path, cookie.HttpOnly)
		}
		// Also check what's in the jar
		if parsedURL, err := url.Parse(ts.Server.URL); err == nil {
			jarCookies := client.Jar.Cookies(parsedURL)
			t.Logf("GetPageWithCSRF: Cookies in jar after request:")
			for _, cookie := range jarCookies {
				t.Logf("  - %s = %s", cookie.Name, cookie.Value)
			}
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	token := ExtractCSRFToken(t, string(body))
	// Note: Do NOT use the cookie value directly as the CSRF token!
	// NoSurf uses a double-submit pattern where the cookie contains the base token
	// and the form field contains a different token that's validated against the cookie.
	// We must extract the token from the HTML form, not from the cookie.

	if testing.Verbose() {
		t.Logf("GetPageWithCSRF: Final token value: %s", token)
	}
	return token, nil
}

// LoginUserWithCSRF logs in a user with proper CSRF token handling
func (ts *TestServer) LoginUserWithCSRF(t *testing.T, email, password string) *http.Client {
	t.Helper()

	// Create a client with cookie jar
	client := ts.NewClientWithCookies(t)

	// First, get the login page to obtain CSRF token and session
	csrfToken, err := ts.GetPageWithCSRF(t, client, "/login")
	if err != nil {
		t.Fatalf("failed to get CSRF token: %v", err)
	}

	// Log cookies after GET request
	if testing.Verbose() {
		if parsedURL, err := url.Parse(ts.Server.URL); err == nil {
			cookies := client.Jar.Cookies(parsedURL)
			t.Logf("LoginUserWithCSRF: Cookies after GET /login:")
			hasSession := false
			hasCSRF := false
			for _, cookie := range cookies {
				t.Logf("  - %s = %s (HttpOnly: %v, Secure: %v)", cookie.Name, cookie.Value, cookie.HttpOnly, cookie.Secure)
				if cookie.Name == "session" {
					hasSession = true
				}
				if cookie.Name == "csrf_token" {
					hasCSRF = true
				}
			}
			if !hasSession {
				t.Logf("  WARNING: No session cookie found!")
			}
			if !hasCSRF {
				t.Logf("  WARNING: No CSRF cookie found!")
			}
		}
	}

	// Prepare login form data with CSRF token
	formData := url.Values{
		"email":    {email},
		"password": {password},
	}

	// Add CSRF token if we found one
	if csrfToken != "" {
		formData.Set("csrf_token", csrfToken)
		if testing.Verbose() {
			t.Logf("LoginUserWithCSRF: Using CSRF token: %s", csrfToken)
		}
	} else {
		if testing.Verbose() {
			t.Logf("LoginUserWithCSRF: WARNING - No CSRF token found!")
		}
	}

	if testing.Verbose() {
		t.Logf("LoginUserWithCSRF: Sending form data: %v", formData)
	}

	// Make login request
	req, err := http.NewRequest(
		http.MethodPost,
		ts.Server.URL+"/login",
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Log request details before sending
	if testing.Verbose() {
		t.Logf("LoginUserWithCSRF: Request URL: %s", req.URL)
		t.Logf("LoginUserWithCSRF: Request headers: %v", req.Header)
		// Log cookies that will be sent
		if parsedURL, err := url.Parse(ts.Server.URL); err == nil {
			cookies := client.Jar.Cookies(parsedURL)
			t.Logf("LoginUserWithCSRF: Cookies being sent with POST:")
			hasSession := false
			hasCSRF := false
			for _, cookie := range cookies {
				t.Logf("  - %s = %s", cookie.Name, cookie.Value)
				if cookie.Name == "session" {
					hasSession = true
				}
				if cookie.Name == "csrf_token" {
					hasCSRF = true
				}
			}
			if !hasSession {
				t.Logf("  WARNING: No session cookie being sent with POST!")
			}
			if !hasCSRF {
				t.Logf("  WARNING: No CSRF cookie being sent with POST!")
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if testing.Verbose() {
		t.Logf("LoginUserWithCSRF: Response status: %d", resp.StatusCode)
		t.Logf("LoginUserWithCSRF: Response headers: %v", resp.Header)
	}

	// Check if login was successful
	if resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Log more details about the failure
		if testing.Verbose() {
			t.Logf("LoginUserWithCSRF: Login failed!")
			t.Logf("LoginUserWithCSRF: Response body: %s", string(body))
			// Check if it's a CSRF error specifically
			if resp.StatusCode == http.StatusBadRequest && string(body) == "Bad Request" {
				t.Logf("LoginUserWithCSRF: This appears to be a CSRF validation failure")
			}
		}
		t.Fatalf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	return client
}

// PostFormWithCSRF makes a POST request with CSRF token handling
func (ts *TestServer) PostFormWithCSRF(t *testing.T, client *http.Client, urlPath string, formData map[string]string) (int, http.Header, string) {
	t.Helper()

	if client == nil {
		client = ts.Client
	}

	// Get CSRF token for the form page
	// Many forms are on GET pages, so we try to get the token from there
	csrfToken, _ := ts.GetPageWithCSRF(t, client, urlPath)

	// Prepare form data
	data := make(url.Values)
	for key, value := range formData {
		data.Set(key, value)
	}

	// Add CSRF token if we have one
	if csrfToken != "" {
		data.Set("csrf_token", csrfToken)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		ts.Server.URL+urlPath,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return resp.StatusCode, resp.Header, string(body)
}

// NewClientWithCookies creates a new HTTP client with cookie jar for maintaining sessions
func (ts *TestServer) NewClientWithCookies(t *testing.T) *http.Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// CreateAndLoginUser is a convenience method that creates a user and logs them in
func (ts *TestServer) CreateAndLoginUser(t *testing.T, name, email, password string) (*http.Client, *queries.User) {
	t.Helper()

	user := ts.CreateTestUser(t, name, email, password)
	client := ts.LoginUserWithCSRF(t, email, password)

	return client, user
}

// AssertRedirect checks if a response is a redirect to the expected location
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

// AssertContains checks if the response body contains the expected text
func AssertContains(t *testing.T, body, expected string) {
	t.Helper()

	if !strings.Contains(body, expected) {
		t.Errorf("expected body to contain %q", expected)
	}
}

// AssertNotContains checks if the response body does not contain the text
func AssertNotContains(t *testing.T, body, shouldNotContain string) {
	t.Helper()

	if strings.Contains(body, shouldNotContain) {
		t.Errorf("expected body to NOT contain %q", shouldNotContain)
	}
}

// AssertStatus checks if the status code matches the expected value
func AssertStatus(t *testing.T, actual, expected int) {
	t.Helper()

	if actual != expected {
		t.Errorf("expected status %d; got %d", expected, actual)
	}
}

// ExtractSessionID extracts the session ID from cookies
func ExtractSessionID(cookies []*http.Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			return cookie.Value
		}
	}
	return ""
}

// WaitForEmail waits for an email to be sent and returns it
// This is useful when testing async email sending
func (ts *TestServer) WaitForEmail(t *testing.T, recipient string, maxWait int) *SentEmail {
	t.Helper()

	for i := 0; i < maxWait; i++ {
		emails := ts.Mailer.GetSentEmails()
		for _, email := range emails {
			if email.Recipient == recipient {
				return &email
			}
		}
		// Small delay before checking again
		if i < maxWait-1 {
			// In real implementation, you might want to use time.Sleep
			// For now, we'll just check immediately
		}
	}

	return nil
}

// ClearEmails clears all sent emails from the mock mailer
func (ts *TestServer) ClearEmails() {
	ts.Mailer.Clear()
}

// GetLastEmail returns the last email that was sent
func (ts *TestServer) GetLastEmail() *SentEmail {
	return ts.Mailer.LastEmail()
}

// AssertEmailSent checks if an email was sent to the recipient
func (ts *TestServer) AssertEmailSent(t *testing.T, recipient string, templateFile string) {
	t.Helper()

	emails := ts.Mailer.GetSentEmails()
	found := false

	for _, email := range emails {
		if email.Recipient == recipient && email.TemplateFile == templateFile {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected email with template %s to be sent to %s", templateFile, recipient)
	}
}

// AssertNoEmailsSent checks that no emails were sent
func (ts *TestServer) AssertNoEmailsSent(t *testing.T) {
	t.Helper()

	count := ts.Mailer.EmailCount()
	if count > 0 {
		t.Errorf("expected no emails to be sent, but %d were sent", count)
	}
}

// MakeAuthenticatedRequest makes a request with an authenticated client
func (ts *TestServer) MakeAuthenticatedRequest(t *testing.T, method, urlPath string, body io.Reader, email, password string) (int, http.Header, string) {
	t.Helper()

	// Create and login user if needed
	client := ts.LoginUserWithCSRF(t, email, password)

	req, err := http.NewRequest(method, ts.Server.URL+urlPath, body)
	if err != nil {
		t.Fatal(err)
	}

	if body != nil && method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return resp.StatusCode, resp.Header, string(respBody)
}

// TestFormValidation is a helper struct for testing form validation
type TestFormValidation struct {
	Name           string
	FormData       map[string]string
	ExpectedStatus int
	ExpectedErrors []string
	ShouldSucceed  bool
}

// RunFormValidationTests runs a series of form validation tests
func (ts *TestServer) RunFormValidationTests(t *testing.T, urlPath string, tests []TestFormValidation) {
	t.Helper()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			status, _, body := ts.PostForm(t, urlPath, tc.FormData)

			AssertStatus(t, status, tc.ExpectedStatus)

			if tc.ShouldSucceed {
				AssertNotContains(t, body, "error")
			} else {
				for _, expectedError := range tc.ExpectedErrors {
					AssertContains(t, body, expectedError)
				}
			}
		})
	}
}

// SetupTestData creates common test data
func (ts *TestServer) SetupTestData(t *testing.T) {
	t.Helper()

	// Create some test users
	ts.CreateTestUser(t, "Admin User", "admin@example.com", "AdminPass123!")
	ts.CreateTestUser(t, "Regular User", "user@example.com", "UserPass123!")
	ts.CreateTestUser(t, "Test User", "test@example.com", "TestPass123!")
}

// CleanupTestData cleans up test data
func (ts *TestServer) CleanupTestData(t *testing.T) {
	t.Helper()

	// In SQLite with in-memory database, this happens automatically
	// when the connection is closed, but you might want to
	// explicitly clean tables for other database types
}
