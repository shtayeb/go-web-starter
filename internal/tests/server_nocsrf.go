package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"go-web-starter/internal/config"
	"go-web-starter/internal/database"
	"go-web-starter/internal/handlers"
	"go-web-starter/internal/handlers/auth"
	"go-web-starter/internal/jsonlog"
	"go-web-starter/internal/queries"
	"go-web-starter/internal/server"
	"go-web-starter/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
)

// TestServerNoCSRF is a test server with CSRF protection disabled
type TestServerNoCSRF struct {
	*TestServer
}

// NewTestServerNoCSRF creates a test server with CSRF protection disabled
// This is useful for testing handlers without dealing with CSRF tokens
func NewTestServerNoCSRF(t *testing.T) *TestServerNoCSRF {
	t.Helper()

	// Create test configuration
	cfg := getTestConfig()

	// Create in-memory SQLite database for testing
	db := setupTestDatabase(t)

	// Create database service wrapper
	var dbService database.Service = &testDatabaseService{db: db}

	// Create queries instance
	q := queries.New(db)

	// Create test logger (discards output during tests)
	logger := jsonlog.New(io.Discard, jsonlog.LevelInfo)

	// Create test mailer
	mockMailer := NewMockMailer()

	// Create session manager with in-memory store
	sessionManager := setupTestSessionManager()

	// Create the server instance (we'll override the routes)
	s := server.NewServer(cfg, dbService, q, logger, mockMailer, sessionManager)

	// Create custom routes WITHOUT CSRF protection
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)

	// Session middleware (needed for authentication)
	r.Use(sessionManager.LoadAndSave)

	// Authentication middleware (custom version without CSRF)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simple authentication check without CSRF
			id := sessionManager.GetInt32(r.Context(), "authenticatedUserID")
			if id != 0 {
				user, err := q.GetUserById(r.Context(), id)
				if err != nil {
					// User lookup failed - clear the session
					sessionManager.Remove(r.Context(), "authenticatedUserID")
				} else {
					ctx := context.WithValue(r.Context(), config.IsAuthenticatedContextKey, true)
					ctx = context.WithValue(ctx, config.UserContextKey, user)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	})

	// Register all the same routes but without CSRF middleware
	// We'll use the handlers from the server instance
	handlers := handlers.NewHandlers(*q, dbService, logger, mockMailer, sessionManager, cfg)
	authService := service.NewAuthService(q, dbService)
	authHandlers := auth.NewAuthHandler(handlers, authService)

	// Public routes
	r.Get("/", handlers.LandingViewHandler)
	r.Get("/health", handlers.HealthHandler)

	// Auth routes (no CSRF check)
	r.Get("/login", authHandlers.LoginViewHandler)
	r.Post("/login", authHandlers.LoginPostHandler)
	r.Get("/signup", authHandlers.SignUpViewHandler)
	r.Post("/signup", authHandlers.SignUpPostHandler)
	r.Post("/logout", authHandlers.LogoutPostHandler)

	// Protected routes
	r.Group(func(r chi.Router) {
		// Add requireAuth middleware
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				isAuth, _ := r.Context().Value(config.IsAuthenticatedContextKey).(bool)
				if !isAuth && r.Method == http.MethodGet {
					http.Redirect(w, r, "/login?next="+r.URL.Path, http.StatusSeeOther)
					return
				}
				next.ServeHTTP(w, r)
			})
		})

		r.Get("/dashboard", handlers.DashboardViewHandler)
		r.Get("/projects", handlers.ProjectViewHandler)
		r.Get("/profile", authHandlers.ProfileViewHandler)
		r.Post("/profile/update", authHandlers.UpdateUserNameAndImageHandler)
		r.Post("/profile/update-password", authHandlers.UpdateAccountPasswordHandler)
		r.Post("/profile/delete-account", authHandlers.DeleteAccountHandler)
		r.Post("/hello", handlers.HelloWebHandler)
	})

	// Create test HTTP server with our custom routes
	ts := httptest.NewServer(r)

	// Create HTTP client
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	baseTestServer := &TestServer{
		Server:     ts,
		Client:     client,
		DB:         db,
		Queries:    q,
		Session:    sessionManager,
		Config:     cfg,
		HTTPServer: s,
		Mailer:     mockMailer,
	}

	return &TestServerNoCSRF{
		TestServer: baseTestServer,
	}
}

// LoginUserSimple logs in a user without CSRF token handling
func (ts *TestServerNoCSRF) LoginUserSimple(t *testing.T, email, password string) *http.Client {
	t.Helper()

	// Create a client with cookie jar
	client := ts.NewClientWithCookies(t)

	// Make login request directly without CSRF
	formData := url.Values{
		"email":    {email},
		"password": {password},
	}

	req, err := http.NewRequest(
		http.MethodPost,
		ts.Server.URL+"/login",
		strings.NewReader(formData.Encode()),
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

	// Check if login was successful
	if resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	return client
}

// PostFormSimple makes a POST request without CSRF token
func (ts *TestServerNoCSRF) PostFormSimple(t *testing.T, client *http.Client, urlPath string, formData map[string]string) (int, http.Header, string) {
	t.Helper()

	if client == nil {
		client = ts.Client
	}

	// Prepare form data
	data := make(url.Values)
	for key, value := range formData {
		data.Set(key, value)
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
