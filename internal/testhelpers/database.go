package testhelpers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

// DatabaseDumpConfig represents configuration for creating a test database from a SQL dump
type DatabaseDumpConfig struct {
	DumpFile          string // absolute path to .sql dump file
	ApplyMigrations   bool   // whether to apply migrations after dump
	ApplyTestFixtures bool   // whether to apply fixtures after dump
}

// IntegreSQLManager manages isolated test databases using Docker
type IntegreSQLManager struct {
	pool     *dockertest.Pool
	resource *dockertest.Resource
	dbURL    string
	template string
	hash     string
}

// NewIntegreSQLManager creates a new IntegreSQL manager
func NewIntegreSQLManager(t TestingT, template string) *IntegreSQLManager {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	err = pool.Client.Ping()
	require.NoError(t, err)

	// Start PostgreSQL container
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=testpass",
			"POSTGRES_USER=testuser",
			"POSTGRES_DB=testdb",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(t, err)

	// Get database URL
	dbURL := fmt.Sprintf("postgres://testuser:testpass@localhost:%s/testdb?sslmode=disable", resource.GetPort("5432/tcp"))

	// Wait for database to be ready
	pool.MaxWait = 60 * time.Second
	err = pool.Retry(func() error {
		db, err := sql.Open("pgx", dbURL)
		if err != nil {
			return err
		}
		defer db.Close()
		return db.Ping()
	})
	require.NoError(t, err)

	manager := &IntegreSQLManager{
		pool:     pool,
		resource: resource,
		dbURL:    dbURL,
		template: template,
	}

	// Create template database if specified
	if template != "" {
		manager.createTemplateDatabase(t)
	}

	return manager
}

// createTemplateDatabase creates a template database for faster test setup
func (m *IntegreSQLManager) createTemplateDatabase(t TestingT) {
	db, err := sql.Open("pgx", m.dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Drop template database if it exists
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", m.template))
	require.NoError(t, err)

	// Create template database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", m.template))
	require.NoError(t, err)

	// Connect to template database and set it up
	templateURL := fmt.Sprintf("postgres://testuser:testpass@localhost:%s/%s?sslmode=disable", m.resource.GetPort("5432/tcp"), m.template)
	templateDB, err := sql.Open("pgx", templateURL)
	require.NoError(t, err)
	defer templateDB.Close()

	// Apply migrations if needed
	if m.template != "empty" {
		ApplyMigrations(t, templateDB)
	}

	// Mark as template
	_, err = templateDB.Exec("UPDATE pg_database SET datistemplate = true WHERE datname = $1", m.template)
	require.NoError(t, err)
}

// CreateTestDatabase creates a new isolated test database
func (m *IntegreSQLManager) CreateTestDatabase(t TestingT) *sql.DB {
	db, err := sql.Open("pgx", m.dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Generate unique database name
	dbName := fmt.Sprintf("test_%d_%d", time.Now().Unix(), t.(*testing.T).Name())

	// Create database from template if available
	if m.template != "" && m.template != "empty" {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s", dbName, m.template))
		require.NoError(t, err)
	} else {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		require.NoError(t, err)
	}

	// Connect to the new database
	testDBURL := fmt.Sprintf("postgres://testuser:testpass@localhost:%s/%s?sslmode=disable", m.resource.GetPort("5432/tcp"), dbName)
	testDB, err := sql.Open("pgx", testDBURL)
	require.NoError(t, err)

	// Ensure database is cleaned up after test
	if tt, ok := t.(*testing.T); ok {
		tt.Cleanup(func() {
			testDB.Close()
			_, err := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
			if err != nil {
				tt.Logf("Failed to drop test database %s: %v", dbName, err)
			}
		})
	}

	return testDB
}

// Close cleans up the IntegreSQL manager
func (m *IntegreSQLManager) Close() error {
	if m.pool != nil && m.resource != nil {
		return m.pool.Purge(m.resource)
	}
	return nil
}

// Global manager instance for simple tests
var globalManager *IntegreSQLManager

// WithTestDatabase creates a test database and passes it to the closure
func WithTestDatabase(t TestingT, closure func(db *sql.DB)) {
	if globalManager == nil {
		globalManager = NewIntegreSQLManager(t, "test_template")
	}

	db := globalManager.CreateTestDatabase(t)
	closure(db)
}

// WithTestDatabaseEmpty creates an empty test database
func WithTestDatabaseEmpty(t TestingT, closure func(db *sql.DB)) {
	manager := NewIntegreSQLManager(t, "empty")
	defer manager.Close()

	db := manager.CreateTestDatabase(t)
	closure(db)
}

// WithTestDatabaseFromDump creates a test database from a SQL dump
func WithTestDatabaseFromDump(t TestingT, config DatabaseDumpConfig, closure func(db *sql.DB)) {
	manager := NewIntegreSQLManager(t, "empty")
	defer manager.Close()

	db := manager.CreateTestDatabase(t)

	// Load dump file
	dumpData, err := os.ReadFile(config.DumpFile)
	require.NoError(t, err)

	// Execute dump
	_, err = db.Exec(string(dumpData))
	require.NoError(t, err)

	// Apply migrations if requested
	if config.ApplyMigrations {
		ApplyMigrations(t, db)
	}

	// Apply test fixtures if requested
	if config.ApplyTestFixtures {
		ApplyTestFixtures(context.Background(), t, db)
	}

	closure(db)
}

// ApplyMigrations applies database migrations to the test database
func ApplyMigrations(t TestingT, db *sql.DB) {
	// Find migration files
	migrationDir := "sql/migrations"
	files, err := filepath.Glob(filepath.Join(migrationDir, "*.sql"))
	require.NoError(t, err)

	// Sort files by name (assuming they follow naming convention like 001_, 002_, etc.)
	// For simplicity, we'll just execute them in the order they're found
	for _, file := range files {
		migrationSQL, err := os.ReadFile(file)
		require.NoError(t, err)

		_, err = db.Exec(string(migrationSQL))
		require.NoError(t, err)
	}
}

// ApplyTestFixtures applies test fixtures to the database
func ApplyTestFixtures(ctx context.Context, t TestingT, db *sql.DB) {
	// This would typically load fixture data from files or generate test data
	// For now, we'll leave this as a placeholder
	t.Log("Applying test fixtures...")
}

// GetDatabaseHash generates a hash for database template caching
func GetDatabaseHash(files ...string) string {
	h := sha256.New()
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			f, err := os.Open(file)
			if err != nil {
				continue
			}
			defer f.Close()
			io.Copy(h, f)
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
