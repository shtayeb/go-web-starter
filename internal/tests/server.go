package tests

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestServer struct {
	Server      *httptest.Server
	Client      *http.Client
	DB          *sql.DB
	DBService   database.Service
	Queries     *queries.Queries
	Session     *scs.SessionManager
	Config      config.Config
	HTTPServer  *server.Server
	Mailer      *MockMailer
	PgContainer testcontainers.Container // nil for SQLite
}

// NewTestServer creates a new test server with all dependencies initialized
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Reset database singleton before each test suite
	database.Reset()

	cfg := getTestConfig()

	cfg.AppEnv = "test"

	db, container := setupTestDatabase(t)

	// Create database service with existing connection
	dbService := database.NewWithDB(db)

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
		DBService:   dbService,
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

// setupTestDatabase creates a test database (PostgreSQL or SQLite) using testcontainers or existing database
func setupTestDatabase(t *testing.T) (*sql.DB, testcontainers.Container) {
	t.Helper()

	return setupPostgresTestDatabase(t)
}

// setupPostgresTestDatabase creates a PostgreSQL test database using testcontainers or existing database
func setupPostgresTestDatabase(t *testing.T) (*sql.DB, testcontainers.Container) {
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

		// Configure connection pool for tests
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		// Clean the database before running tests
		cleanTestDatabase(t, db)

		// Run migrations
		if err := runMigrations(db); err != nil {
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

	// Configure connection pool for tests
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Wait for the database to be ready
	err = db.Ping()
	if err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	// Run migrations using goose
	if err := runMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db, postgresContainer
}

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

	// Detect database type by trying a SQLite-specific query
	var isSQLite bool
	var result string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&result)
	isSQLite = err == nil

	for _, table := range tables {
		var query string
		if isSQLite {
			// SQLite uses DELETE FROM (no TRUNCATE support)
			query = fmt.Sprintf("DELETE FROM %s", table)
		} else {
			// PostgreSQL uses TRUNCATE with CASCADE
			query = fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		}

		_, err := db.Exec(query)
		if err != nil {
			// If table doesn't exist, that's ok (migrations will create it)
			t.Logf("Warning: could not clean table %s: %v", table, err)
		}
	}
}

func runMigrations(db *sql.DB) error {
	dialect := "postgres"
	// Get the directory of the current source file
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)
	// Navigate to project root and then to postgres migrations
	projectRoot := filepath.Join(currentDir, "..", "..")
	migrationsDir := filepath.Join(projectRoot, "sql", "postgres", "migrations")

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
