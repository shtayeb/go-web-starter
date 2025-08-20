package server

import (
	"fmt"
	"go-web-starter/internal/config"
	"go-web-starter/internal/database"
	"go-web-starter/internal/jsonlog"
	"go-web-starter/internal/mailer"
	"go-web-starter/internal/queries"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestE2E(t *testing.T) {
	// Start a test server
	srv := startTestServer(t)
	if srv == nil {
		t.Skip("Could not start test server, skipping E2E tests")
		return
	}
	defer srv.Close()

	// Get server URL
	baseURL := srv.URL

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects, we want to test them explicitly
			return http.ErrUseLastResponse
		},
	}

	// Test suite
	t.Run("Health Check", func(t *testing.T) {
		testHealthEndpoint(t, client, baseURL)
	})

	t.Run("Landing Page", func(t *testing.T) {
		testLandingPage(t, client, baseURL)
	})

	t.Run("Authentication Pages", func(t *testing.T) {
		testAuthPages(t, client, baseURL)
	})

	t.Run("Protected Routes", func(t *testing.T) {
		testProtectedRoutes(t, client, baseURL)
	})

	t.Run("Static Assets", func(t *testing.T) {
		testStaticAssets(t, client, baseURL)
	})
}

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

	server := NewServer(
		cfg,
		dbService,
		queries.New(sqlDB),
		jsonlog.New(os.Stdout, jsonlog.LevelInfo),
		mailer.New(cfg.Mailer),
		NewSessionManager(sqlDB),
	)

	srv := httptest.NewServer(server.RegisterRoutes())

	return srv
}

// testHealthEndpoint tests the health check endpoint
func testHealthEndpoint(t *testing.T, client *http.Client, baseURL string) {
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// testLandingPage tests the main landing page
func testLandingPage(t *testing.T, client *http.Client, baseURL string) {
	resp, err := client.Get(baseURL + "/")
	if err != nil {
		t.Fatalf("Landing page request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected HTML content type, got: %s", contentType)
	}
}

// testAuthPages tests authentication-related pages
func testAuthPages(t *testing.T, client *http.Client, baseURL string) {
	authPages := []struct {
		path           string
		expectedStatus int
	}{
		{"/login", http.StatusOK},
		{"/signup", http.StatusOK},
		{"/forgot-password", http.StatusOK},
	}

	for _, page := range authPages {
		t.Run(fmt.Sprintf("GET %s", page.path), func(t *testing.T) {
			resp, err := client.Get(baseURL + page.path)
			if err != nil {
				t.Fatalf("Request to %s failed: %v", page.path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != page.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", page.expectedStatus, page.path, resp.StatusCode)
			}

			// Check that it's HTML
			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "text/html") {
				t.Errorf("Expected HTML content type for %s, got: %s", page.path, contentType)
			}
		})
	}
}

// testProtectedRoutes tests that protected routes require authentication
func testProtectedRoutes(t *testing.T, client *http.Client, baseURL string) {
	protectedRoutes := []string{
		"/dashboard",
		"/profile",
		"/projects",
	}

	for _, route := range protectedRoutes {
		t.Run(fmt.Sprintf("Protected %s", route), func(t *testing.T) {
			resp, err := client.Get(baseURL + route)
			if err != nil {
				t.Fatalf("Request to %s failed: %v", route, err)
			}
			defer resp.Body.Close()

			// Should redirect to login or return 401/403
			if resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
				t.Errorf("Expected redirect (302) or auth error (401/403) for protected route %s, got %d", route, resp.StatusCode)
			}

			// If it's a redirect, check the Location header
			if resp.StatusCode == http.StatusFound {
				location := resp.Header.Get("Location")
				if !strings.Contains(location, "login") {
					t.Errorf("Expected redirect to login page for %s, got: %s", route, location)
				}
			}
		})
	}
}

func testStaticAssets(t *testing.T, client *http.Client, baseURL string) {
	// Test that assets endpoint is accessible
	resp, err := client.Get(baseURL + "/assets/")
	if err != nil {
		t.Fatalf("Assets request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should either serve an index or return 404 for directory listing
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404 for /assets/, got %d", resp.StatusCode)
	}
}

// BenchmarkE2E benchmarks the server performance
func BenchmarkE2E(b *testing.B) {
	// Start test server
	srv := startTestServer(&testing.T{})
	defer srv.Close()

	baseURL := srv.URL
	client := &http.Client{Timeout: 5 * time.Second}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(baseURL + "/health")
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", resp.StatusCode)
			}
		}
	})
}
