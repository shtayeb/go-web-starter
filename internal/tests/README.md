# Test Infrastructure

This package provides test helpers and utilities for running integration tests with a real PostgreSQL database.

## Overview

The test infrastructure supports two modes of operation:
1. **Testcontainers Mode** (default): Automatically spins up a PostgreSQL container for each test suite
2. **Existing Database Mode**: Uses an existing PostgreSQL database for faster test execution

## Test Database Setup

### Using Testcontainers (Default)

By default, tests will automatically create a PostgreSQL container using testcontainers. This ensures complete isolation between test runs but takes longer (~20 seconds per test suite).

```bash
# Run tests with testcontainers
go test ./...

# Or use the make target
make test-container
```

**Pros:**
- Complete isolation between test runs
- No setup required
- Works in CI/CD environments
- Automatic cleanup

**Cons:**
- Slower test execution
- Requires Docker to be running

### Using Existing Database (Fast Mode)

For faster local development, you can use an existing PostgreSQL database:

```bash
# Start a local test database
make test-db-start

# Run tests using the existing database
make test-fast

# Stop the test database when done
make test-db-stop
```

You can also manually specify a database URL:

```bash
TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable" go test ./...
```



## How It Works

### Database Initialization

1. **Container Creation**: When using testcontainers, a PostgreSQL 16 Alpine container is created
2. **Connection**: The test helper establishes a connection to the database
3. **Migration**: Goose migrations from `sql/migrations/` are automatically applied
4. **Cleanup**: For existing databases, tables are truncated before each test run

### Test Helpers

The package provides several test helpers:

- `TestServer`: A complete test server with database, session management, and mailer mocks
- Helper methods for creating test users, logging in, and making authenticated requests

## Available Make Commands

| Command | Description |
|---------|-------------|
| `make test` | Run all tests with default settings |
| `make test-container` | Run tests using testcontainers |
| `make test-fast` | Run tests using existing database (fast mode) |
| `make test-db-start` | Start a PostgreSQL container for testing |
| `make test-db-stop` | Stop and remove the test database container |
| `make test-one TEST_NAME=TestName` | Run a specific test |

## Example Test

```go
func TestExample(t *testing.T) {
    // Create a test server with all dependencies
    ts := tests.NewTestServer(t)
    defer ts.Close() // Cleanup database and containers

    // Create a test user
    user := ts.CreateTestUser(t, "John Doe", "john@example.com", "password123")

    // Login the user
    client := ts.LoginUser(t, "john@example.com", "password123")

    // Make authenticated requests
    status, _, body := ts.GetWithClient(t, client, "/dashboard")
    
    // Assert expectations
    if status != http.StatusOK {
        t.Errorf("expected status 200, got %d", status)
    }
}
```

## Configuration

The test setup uses configuration from `.env.test` if available, falling back to default values. The configuration is loaded using:

Key configuration values for testing:
- `APP_ENV=test`

## Troubleshooting

### Docker Not Running
If you see errors about Docker not being available, make sure Docker Desktop or Docker Engine is running.

## Best Practices

1. **Use testcontainers for CI/CD** to ensure consistent, isolated test environments
2. **Use existing database for local development** for faster feedback loops
3. **Clean up test data** properly in test teardown functions