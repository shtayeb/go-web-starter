package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"go-web-starter/internal/config"
	"go-web-starter/internal/database"
	"go-web-starter/internal/jsonlog"
	"go-web-starter/internal/mailer"
	"go-web-starter/internal/queries"
	"go-web-starter/internal/server"

	"github.com/alexedwards/scs/v2"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// TestServer wraps all the components needed for testing
type TestServer struct {
	Server     *httptest.Server
	Client     *http.Client
	DB         *sql.DB
	Queries    *queries.Queries
	Session    *scs.SessionManager
	Config     config.Config
	HTTPServer *server.Server
	Mailer     *MockMailer
}

// NewTestServer creates a new test server with all dependencies initialized
func NewTestServer(t *testing.T) *TestServer {
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

	// Create test mailer with test configuration
	testMailer := mailer.New(cfg.Mailer)
	mockMailer := NewMockMailer()

	// Create session manager with in-memory store
	sessionManager := setupTestSessionManager()

	// Create the server instance
	s := server.NewServer(cfg, dbService, q, logger, testMailer, sessionManager)

	// Create test HTTP server
	ts := httptest.NewServer(s.RegisterRoutes())

	// Create HTTP client with redirect following disabled
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &TestServer{
		Server:     ts,
		Client:     client,
		DB:         db,
		Queries:    q,
		Session:    sessionManager,
		Config:     cfg,
		HTTPServer: s,
		Mailer:     mockMailer,
	}
}

// Close cleans up the test server and its resources
func (ts *TestServer) Close() {
	ts.Server.Close()
	ts.DB.Close()
}

// getTestConfig returns a configuration suitable for testing
func getTestConfig() config.Config {
	return config.Config{
		AppName: "Test App",
		AppEnv:  "test",
		AppURL:  "http://localhost:8080",
		Debug:   false,
		Port:    8080,
		Database: config.Database{
			DBUrl:    ":memory:",
			Database: "test",
			Password: "",
			Username: "",
			Port:     "",
			Host:     "",
			Schema:   "",
		},
		Mailer: config.SMTP{
			Host:     "localhost",
			Port:     1025,
			Username: "test",
			Password: "test",
			Sender:   "test@example.com",
		},
		SocialLogins: config.SocialLogins{
			GoogleClientID:     "test-client-id",
			GoogleClientSecret: "test-client-secret",
		},
	}
}

// setupTestDatabase creates an in-memory SQLite database with test schema
func setupTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create test schema matching your PostgreSQL schema
	// Note: SQLite syntax is slightly different from PostgreSQL
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		email_verified INTEGER DEFAULT 0,
		image TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		account_id TEXT NOT NULL,
		provider_id TEXT,
		password TEXT,
		access_token TEXT,
		refresh_token TEXT,
		id_token TEXT,
		access_token_expires_at DATETIME,
		refresh_token_expires_at DATETIME,
		scope TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, provider_id)
	);

	CREATE TABLE IF NOT EXISTS tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		hash BLOB NOT NULL,
		scope TEXT NOT NULL,
		expiry DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		expiry REAL NOT NULL
	);

	CREATE INDEX idx_sessions_expiry ON sessions(expiry);
	CREATE INDEX idx_tokens_user_id ON tokens(user_id);
	CREATE INDEX idx_accounts_user_id ON accounts(user_id);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create test schema: %v", err)
	}

	return db
}

// setupTestSessionManager creates a session manager with in-memory store
func setupTestSessionManager() *scs.SessionManager {
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = false // Not using HTTPS in tests
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	return sessionManager
}

// testDatabaseService implements the database.Service interface for testing
type testDatabaseService struct {
	db *sql.DB
}

func (s *testDatabaseService) Health() map[string]string {
	return map[string]string{
		"status":  "up",
		"message": "test database",
	}
}

func (s *testDatabaseService) Close(cfg config.Database) error {
	return s.db.Close()
}

func (s *testDatabaseService) GetDB() *sql.DB {
	return s.db
}

func (s *testDatabaseService) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// Helper methods for making requests

// Get makes a GET request to the test server
func (ts *TestServer) Get(t *testing.T, urlPath string) (int, http.Header, string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, ts.Server.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := ts.Client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	return resp.StatusCode, resp.Header, string(body)
}

// PostForm makes a POST request with form data to the test server
func (ts *TestServer) PostForm(t *testing.T, urlPath string, form map[string]string) (int, http.Header, string) {
	t.Helper()

	formData := make(url.Values)
	for key, value := range form {
		formData.Set(key, value)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		ts.Server.URL+urlPath,
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := ts.Client.Do(req)
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

// PostJSON makes a POST request with JSON data to the test server
func (ts *TestServer) PostJSON(t *testing.T, urlPath string, jsonData interface{}) (int, http.Header, string) {
	t.Helper()

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, ts.Server.URL+urlPath, bytes.NewReader(jsonBytes))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.Client.Do(req)
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

// CreateTestUser creates a test user in the database
func (ts *TestServer) CreateTestUser(t *testing.T, name, email, password string) *queries.User {
	t.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	ctx := context.Background()

	// Create user
	user, err := ts.Queries.CreateUser(ctx, queries.CreateUserParams{
		Name:          name,
		Email:         email,
		EmailVerified: true,
		Image:         sql.NullString{},
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Create account
	_, err = ts.Queries.CreateAccount(ctx, queries.CreateAccountParams{
		UserID:    user.ID,
		AccountID: name,
		Password:  sql.NullString{String: string(hashedPassword), Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create test account: %v", err)
	}

	return &user
}

// LoginUser logs in a user and returns a client with session cookies
func (ts *TestServer) LoginUser(t *testing.T, email, password string) *http.Client {
	t.Helper()

	return ts.LoginUserWithCSRF(t, email, password)
}

// GetWithClient makes a GET request with a specific client (for authenticated requests)
func (ts *TestServer) GetWithClient(t *testing.T, client *http.Client, urlPath string) (int, http.Header, string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, ts.Server.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}

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

// PostFormWithClient makes a POST request with a specific client (for authenticated requests)
func (ts *TestServer) PostFormWithClient(t *testing.T, client *http.Client, urlPath string, form map[string]string) (int, http.Header, string) {
	t.Helper()

	formData := make(url.Values)
	for key, value := range form {
		formData.Set(key, value)
	}

	req, err := http.NewRequest(http.MethodPost, ts.Server.URL+urlPath, strings.NewReader(formData.Encode()))
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
