package testhelpers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"

	"go-web-starter/internal/config"
	"go-web-starter/internal/database"
	"go-web-starter/internal/jsonlog"
	"go-web-starter/internal/mailer"
	"go-web-starter/internal/queries"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

// TestServer wraps the server with test-specific functionality
type TestServer struct {
	httpServer *httptest.Server
	db         *sql.DB
}

// NewTestServer creates a new test server with isolated database
func NewTestServer(t TestingT, dbConfig config.Database) *TestServer {
	// Create database service
	dbService := database.New(dbConfig)
	sqlDB := dbService.GetDB()

	// Create session manager
	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(sqlDB)
	sessionManager.Lifetime = 12 * 60 * 60 * 1000000000 // 12 hours in nanoseconds
	sessionManager.Cookie.Secure = false                // Disable secure cookies for tests

	// Create test config
	testConfig := config.Config{
		AppName:  "Test App",
		AppEnv:   "test",
		Debug:    true,
		Port:     0, // Let the system assign a port
		Database: dbConfig,
		Mailer: config.SMTP{
			Host:     "localhost",
			Port:     587,
			Username: "test",
			Password: "test",
			Sender:   "test@example.com",
		},
	}

	// Setup social logins for tests
	goth.UseProviders(
		google.New(
			"test-client-id",
			"test-client-secret",
			"http://localhost:8080/auth/google/callback",
		),
	)

	// Create a simple test handler that mimics the server behavior
	// This avoids importing the server package which causes cycles
	testHandler := createTestHandler(testConfig, dbService, queries.New(sqlDB), jsonlog.New(nil, jsonlog.LevelError), mailer.New(testConfig.Mailer), sessionManager)

	// Wrap in httptest.Server for easy testing
	testHTTPServer := httptest.NewServer(testHandler)

	return &TestServer{
		httpServer: testHTTPServer,
		db:         sqlDB,
	}
}

// createTestHandler creates a test HTTP handler that mimics server functionality
// This is a simplified version to avoid import cycles
func createTestHandler(cfg config.Config, dbService database.Service, queries *queries.Queries, logger *jsonlog.Logger, mailer mailer.Mailer, sessionManager *scs.SessionManager) http.Handler {
	// For now, return a simple handler that just returns 404
	// In a real implementation, you would need to replicate the routing logic
	// without importing the handlers package
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple test handler - just return 404 for now
		// This would need to be expanded to handle the actual routes
		http.NotFound(w, r)
	})
}

// URL returns the test server URL
func (ts *TestServer) URL() string {
	return ts.httpServer.URL
}

// Close shuts down the test server
func (ts *TestServer) Close() {
	ts.httpServer.Close()
}

// GetDB returns the underlying database connection
func (ts *TestServer) GetDB() *sql.DB {
	return ts.db
}

// WithTestServer creates a test server and passes it to the closure
func WithTestServer(t TestingT, closure func(s *TestServer)) {
	// Create test database
	WithTestDatabase(t, func(db *sql.DB) {
		// Create database config for test
		dbConfig := config.Database{
			DBUrl:    "", // Will use the test database
			Database: "testdb",
			Password: "testpass",
			Username: "testuser",
			Port:     "5432",
			Host:     "localhost",
			Schema:   "public",
		}

		// Create test server
		ts := NewTestServer(t, dbConfig)
		defer ts.Close()

		closure(ts)
	})
}

// WithTestServerConfigurable creates a test server with custom config
func WithTestServerConfigurable(t TestingT, cfg config.Config, closure func(s *TestServer)) {
	// Create test database
	WithTestDatabase(t, func(db *sql.DB) {
		// Override database config to use test database
		cfg.Database = config.Database{
			DBUrl:    "", // Will use the test database
			Database: "testdb",
			Password: "testpass",
			Username: "testuser",
			Port:     "5432",
			Host:     "localhost",
			Schema:   "public",
		}

		// Create test server
		ts := NewTestServer(t, cfg.Database)
		defer ts.Close()

		closure(ts)
	})
}

// WithTestServerFromDump creates a test server with database from SQL dump
func WithTestServerFromDump(t TestingT, dumpConfig DatabaseDumpConfig, closure func(s *TestServer)) {
	WithTestDatabaseFromDump(t, dumpConfig, func(db *sql.DB) {
		// Create database config for test
		dbConfig := config.Database{
			DBUrl:    "", // Will use the test database
			Database: "testdb",
			Password: "testpass",
			Username: "testuser",
			Port:     "5432",
			Host:     "localhost",
			Schema:   "public",
		}

		// Create test server
		ts := NewTestServer(t, dbConfig)
		defer ts.Close()

		closure(ts)
	})
}
