package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"go-web-starter/internal/config"
	"go-web-starter/internal/database"
	"go-web-starter/internal/jsonlog"
	"go-web-starter/internal/queries"
	"go-web-starter/internal/server"

	"github.com/alexedwards/scs/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
)

type TestServer struct {
	Server      *httptest.Server
	Client      *http.Client
	DB          *sql.DB
	Queries     *queries.Queries
	Session     *scs.SessionManager
	Config      config.Config
	HTTPServer  *server.Server
	Mailer      *MockMailer
	PgContainer testcontainers.Container
}

// NewTestServer creates a new test server with all dependencies initialized
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	cfg := getTestConfig()

	// Create PostgreSQL
	db, container := setupTestDatabase(t)

	// Create database service wrapper
	var dbService database.Service = &testDatabaseService{db: db}

	// Create queries instance
	q := queries.New(db)

	logger := jsonlog.New(io.Discard, jsonlog.LevelInfo)

	mockMailer := NewMockMailer()

	sessionManager := setupTestSessionManager()

	s := server.NewServer(cfg, dbService, q, logger, mockMailer, sessionManager)

	ts := httptest.NewServer(s.RegisterRoutes())

	// Create HTTP client with redirect following disabled
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &TestServer{
		Server:      ts,
		Client:      client,
		DB:          db,
		Queries:     q,
		Session:     sessionManager,
		Config:      cfg,
		HTTPServer:  s,
		Mailer:      mockMailer,
		PgContainer: container,
	}
}

func (ts *TestServer) Close() {
	ts.Server.Close()
	ts.DB.Close()
	if ts.PgContainer != nil {
		ctx := context.Background()
		_ = ts.PgContainer.Terminate(ctx)
	}
}

func getTestConfig() config.Config {
	// Get the directory of the current source file
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	// Navigate from internal/tests/ to project root (go up 2 directories)
	projectRoot := filepath.Join(currentDir, "..", "..")
	envTestPath := filepath.Join(projectRoot, ".env.test")

	err := godotenv.Load(envTestPath)
	if err != nil {
		// If that fails, try current working directory as fallback
		err = godotenv.Load(".env.test")
		if err != nil {
			log.Printf("Could not load .env.test file: %v. Using default values from config.LoadConfigFromEnv()", err)
		}
	}

	return config.LoadConfigFromEnv()
}

// setupTestDatabase creates a PostgreSQL test database using testcontainers or existing database
func setupTestDatabase(t *testing.T) (*sql.DB, testcontainers.Container) {
	t.Helper()

	// Check if TEST_DATABASE_URL is set (for faster local testing)
	if testDBURL := os.Getenv("TEST_DATABASE_URL"); testDBURL != "" {
		log.Printf("Using existing test database: %s", testDBURL)

		// Connect to existing database
		db, err := sql.Open("pgx", testDBURL)
		if err != nil {
			t.Fatalf("failed to connect to existing test database: %v", err)
		}

		// Verify connection
		if err := db.Ping(); err != nil {
			t.Fatalf("failed to ping existing test database: %v", err)
		}

		// Clean the database before running tests
		cleanTestDatabase(t, db)

		// Run migrations
		if err := runMigrations(db, "postgres"); err != nil {
			t.Fatalf("failed to run migrations on existing database: %v", err)
		}

		return db, nil // No container when using existing database
	}

	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Get connection string
	connectionString, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Connect to the database
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Wait for the database to be ready
	err = db.Ping()
	if err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	// Run migrations using goose
	if err := runMigrations(db, "postgres"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db, postgresContainer
}

// cleanTestDatabase removes all data from test database tables
func cleanTestDatabase(t *testing.T, db *sql.DB) {
	t.Helper()

	// Tables to clean in reverse order of foreign key dependencies
	tables := []string{
		"tokens",
		"sessions",
		"accounts",
		"users",
		"authors",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// If table doesn't exist, that's ok (migrations will create it)
			t.Logf("Warning: could not truncate table %s: %v", table, err)
		}
	}
}

func runMigrations(db *sql.DB, dbType string) error {
	var migrationsDir string
	var dialect string

	switch dbType {
	case "sqlite", "sqlite3":
		dialect = "sqlite3"
		// Get the directory of the current source file
		_, filename, _, _ := runtime.Caller(0)
		currentDir := filepath.Dir(filename)
		// Navigate to project root and then to sqlite migrations
		projectRoot := filepath.Join(currentDir, "..", "..")
		migrationsDir = filepath.Join(projectRoot, "sql", "sqlite", "migrations")
	case "postgres", "postgresql":
		dialect = "postgres"
		// Get the directory of the current source file
		_, filename, _, _ := runtime.Caller(0)
		currentDir := filepath.Dir(filename)
		// Navigate to project root and then to postgres migrations
		projectRoot := filepath.Join(currentDir, "..", "..")
		migrationsDir = filepath.Join(projectRoot, "sql", "postgres", "migrations")
	}

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

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

	// Create user with proper PostgreSQL boolean type
	user, err := ts.Queries.CreateUser(ctx, queries.CreateUserParams{
		Name:          name,
		Email:         email,
		EmailVerified: true, // PostgreSQL handles boolean properly
		Image:         sql.NullString{},
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Create account with PostgreSQL-compatible parameters
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
