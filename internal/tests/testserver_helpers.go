package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"go-web-starter/internal/queries"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func (ts *TestServer) CleanDatabase(t *testing.T) {
	t.Helper()
	cleanTestDatabase(t, ts.DB)
}

func (ts *TestServer) WithTransaction(t *testing.T, fn func(*testing.T, *sql.Tx)) {
	t.Helper()

	tx, err := ts.DB.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	fn(t, tx)
}

func (ts *TestServer) ExecuteRequest(req *http.Request) *http.Response {
	resp, err := ts.Client.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}

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
