package tests

import (
	"go-web-starter/internal/config"
	"go-web-starter/internal/database"
	"go-web-starter/internal/jsonlog"
	"go-web-starter/internal/mailer"
	"go-web-starter/internal/queries"
	"go-web-starter/internal/server"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// startTestServer starts a test HTTP server for testing
func startTestServer(t *testing.T) *httptest.Server {
	if err := godotenv.Load("../../.env.test"); err != nil {
		t.Logf("Warning: Could not load .env.test file: %v", err)
	}

	cfg := config.LoadConfigFromEnv()

	// Try to connect to database, if it fails, skip the test
	dbService := database.New(cfg.Database)
	if dbService == nil {
		return nil
	}

	health := dbService.Health()
	if health["status"] != "up" {
		t.Logf("Database is not available: %s", health["error"])
		return nil
	}

	sqlDB := dbService.GetDB()

	server := server.NewServer(
		cfg,
		dbService,
		queries.New(sqlDB),
		jsonlog.New(os.Stdout, jsonlog.LevelInfo),
		mailer.New(cfg.Mailer),
		server.NewSessionManager(sqlDB),
	)

	srv := httptest.NewServer(server.RegisterRoutes())

	return srv
}
