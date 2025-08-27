# Test Infrastructure

This package provides test helpers and utilities for running integration tests with a real database (PostgreSQL or SQLite).

## Overview

The test infrastructure supports database testing for both PostgreSQL and SQLite based on the `BLUEPRINT_DB_TYPE` environment variable (falling back to `DATABASE_TYPE` for backward compatibility).

### Database Types

- **PostgreSQL**: Uses testcontainers or existing database
- **SQLite**: Uses temporary file-based database

### Modes of Operation

1. **Testcontainers Mode** (PostgreSQL only): Automatically spins up a PostgreSQL container for each test suite
2. **Existing Database Mode** (PostgreSQL only): Uses an existing PostgreSQL database for faster test execution
3. **SQLite Mode**: Uses a temporary SQLite database file for each test

## Test Database Setup

The test database type is determined by the `BLUEPRINT_DB_TYPE` environment variable (or `DATABASE_TYPE` if not set).

### PostgreSQL Testing

#### Using Testcontainers (Default for PostgreSQL)

When `BLUEPRINT_DB_TYPE=postgres` (or `DATABASE_TYPE=postgres`), tests will automatically create a PostgreSQL container using testcontainers. This ensures complete isolation between test runs but takes longer (~20 seconds per test suite).

```bash
# Set database type to PostgreSQL
export BLUEPRINT_DB_TYPE=postgres

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

#### Using Existing Database (Fast Mode for PostgreSQL)

For faster local development with PostgreSQL:

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

### SQLite Testing

When `BLUEPRINT_DB_TYPE=sqlite` (or `DATABASE_TYPE=sqlite`), tests will use a temporary SQLite database file that is created and destroyed for each test.

```bash
# Set database type to SQLite
export BLUEPRINT_DB_TYPE=sqlite

# Run tests with SQLite
go test ./...
```

**Pros:**
- Very fast test execution
- No external dependencies
- Automatic cleanup
- Works anywhere

**Cons:**
- Less realistic than PostgreSQL testing
- Some PostgreSQL-specific features may not be tested



## How It Works

### Database Initialization

1. **Database Type Detection**: The test reads `DATABASE_TYPE` from environment configuration
2. **Database Setup**:
   - **PostgreSQL**: Container creation (testcontainers) or connection to existing database
   - **SQLite**: Temporary database file creation
3. **Connection**: The test helper establishes a connection to the database
4. **Migration**: Goose migrations are automatically applied based on database type:
   - PostgreSQL: Uses `sql/postgres/migrations/`
   - SQLite: Uses `sql/sqlite/migrations/`
5. **Cleanup**: Tables are cleaned before each test run (TRUNCATE for PostgreSQL, DELETE for SQLite)

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
- `BLUEPRINT_DB_TYPE=postgres` (or `sqlite`; legacy: `DATABASE_TYPE`)

## Troubleshooting

### Docker Not Running
If you see errors about Docker not being available when using PostgreSQL, make sure Docker Desktop or Docker Engine is running.

### SQLite Driver Issues
If you encounter SQLite driver issues, ensure the `github.com/mattn/go-sqlite3` package is properly installed and the CGO is enabled.

## Best Practices

1. **Use testcontainers for CI/CD** (PostgreSQL) to ensure consistent, isolated test environments
2. **Use SQLite for local development** for fastest feedback loops
3. **Use existing PostgreSQL database** for comprehensive testing with realistic data
4. **Clean up test data** properly in test teardown functions
5. **Test with both database types** to ensure compatibility