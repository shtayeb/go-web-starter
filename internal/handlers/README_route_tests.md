# Route Tests - Import Cycle Free Architecture

This directory contains comprehensive tests for the landing (`/`) and dashboard (`/dashboard`) routes, designed to avoid circular import dependencies while keeping tests in the same package as handlers.

## 🏗️ **Architecture Overview**

The tests are structured to avoid circular dependencies by using a separate testing infrastructure package:

```
internal/
├── testhelpers/         # Testing infrastructure (no circular dependencies)
│   ├── types.go         # Common types and interfaces
│   ├── database.go      # Database testing utilities
│   ├── server.go        # Test server setup (simplified to avoid cycles)
│   ├── http.go          # HTTP testing utilities
│   └── helpers.go       # Helper functions
└── handlers/            # Handler code and tests (same package)
    ├── landing.go        # Landing page handler
    ├── dashboard.go      # Dashboard page handler
    ├── landing_test.go   # Landing page tests
    ├── dashboard_test.go # Dashboard page tests
    ├── run_route_tests.sh # Test runner
    └── README.md         # This documentation
```

## 🎯 **Key Design Decisions**

### **1. Separate Testing Infrastructure**
- Created `internal/testhelpers` package for shared testing utilities
- This package doesn't import any application code, avoiding cycles
- Contains all database, server, and HTTP testing infrastructure

### **2. Simplified Test Server**
- The `TestServer` in `testhelpers` is simplified to avoid importing the `server` package
- Uses `httptest.Server` directly instead of embedding the full server
- Still provides all necessary testing functionality

### **3. Same-Package Tests**
- Tests remain in the `handlers` package as requested
- Import `testhelpers` package for testing infrastructure
- No circular dependencies between packages

## 📁 **Test Files**

### `landing_test.go`
Tests for the **public landing page** (`/`) route:

**Test Coverage:**
- ✅ **Public Access**: Works without authentication
- ✅ **Authenticated Access**: Still works with logged-in user
- ✅ **HTTP Methods**: GET (200), POST/PUT/DELETE (405)
- ✅ **Response Headers**: Content-Type, Cache-Control
- ✅ **Content Validation**: Verifies "Welcome" content

### `dashboard_test.go`
Tests for the **protected dashboard** (`/dashboard`) route:

**Test Coverage:**
- ❌ **Unauthenticated Access**: Redirects to `/login`
- ✅ **Authenticated Access**: Works with valid session
- ❌ **Invalid Session**: Redirects to `/login`
- ✅ **HTTP Methods**: All methods redirect when unauthenticated
- ✅ **Security Headers**: `Cache-Control: no-store`
- ✅ **Logout Behavior**: Redirects after logout

## 🚀 **How to Run the Tests:**

### **Option 1: Using the test runner script**
```bash
cd internal/handlers
./run_route_tests.sh
```

### **Option 2: Run specific tests**
```bash
# Landing page tests
go test -v ./internal/handlers -run TestLandingRoute

# Dashboard tests
go test -v ./internal/handlers -run TestDashboardRoute
```

### **Option 3: Run all route tests with coverage**
```bash
go test -cover ./internal/handlers -run "TestLandingRoute|TestDashboardRoute"
```

## 🔧 **Test Infrastructure Usage:**

### **Database Testing:**
```go
testhelpers.WithTestDatabase(t, func(db *sql.DB) {
    // Test with isolated PostgreSQL database
})
```

### **Server Testing:**
```go
testhelpers.WithTestServer(t, func(s *testhelpers.TestServer) {
    // Test with full HTTP server setup
})
```

### **HTTP Requests:**
```go
res := testhelpers.PerformRequest(t, s, "GET", "/", nil, nil)
testhelpers.AssertStatusCode(t, res, http.StatusOK)
```

### **Authentication:**
```go
// Register user
registerPayload := testhelpers.GenericPayload{
    "username": testhelpers.RandomUsername(t),
    "email":    testhelpers.RandomEmail(t),
    "password": testhelpers.RandomPassword(t),
}

// Login and extract session cookie
// ... authentication flow
```

## 🎯 **Test Scenarios Covered:**

### **Landing Page (`/`) - Public Route**
```go
// ✅ Works without authentication
res := testhelpers.PerformRequest(t, s, "GET", "/", nil, nil)
testhelpers.AssertStatusCode(t, res, http.StatusOK)

// ✅ Still works with authenticated user
res = testhelpers.PerformRequest(t, s, "GET", "/", headers, nil)
testhelpers.AssertStatusCode(t, res, http.StatusOK)
```

### **Dashboard (`/dashboard`) - Protected Route**
```go
// ❌ Redirects without authentication
res := testhelpers.PerformRequest(t, s, "GET", "/dashboard", nil, nil)
testhelpers.AssertStatusCode(t, res, http.StatusSeeOther) // 303 redirect to /login

// ✅ Works with authentication
res = testhelpers.PerformRequest(t, s, "GET", "/dashboard", headers, nil)
testhelpers.AssertStatusCode(t, res, http.StatusOK)
```

## 🔐 **Authentication Flow Testing:**

The tests demonstrate complete authentication workflows:

1. **User Registration** → `/signup`
2. **User Login** → `/login` (extracts session cookie)
3. **Authenticated Requests** → Uses session cookie
4. **User Logout** → `/logout`
5. **Post-Logout Access** → Verifies protection

## 🛡️ **Security Testing:**

- **CSRF Protection**: Validates CSRF tokens in responses
- **Security Headers**: Tests `X-Content-Type-Options`, `X-Frame-Options`, etc.
- **Cache Control**: Ensures protected routes have `no-store` directive
- **Session Management**: Tests invalid/expired sessions

## 📊 **Expected Test Results:**

### **Landing Page Tests:**
```
✅ TestLandingRoute - 200 OK (public access)
✅ TestLandingRouteWithAuthenticatedUser - 200 OK (authenticated access)
✅ TestLandingRouteDifferentMethods - GET:200, POST/PUT/DELETE:405
✅ TestLandingRouteHeaders - Proper content-type and security headers
```

### **Dashboard Tests:**
```
✅ TestDashboardRouteUnauthenticated - 303 redirect to /login
✅ TestDashboardRouteAuthenticated - 200 OK with valid session
✅ TestDashboardRouteWithInvalidSession - 303 redirect to /login
✅ TestDashboardRouteAfterLogout - 303 redirect to /login
✅ TestDashboardRouteHeaders - Cache-Control: no-store
```

## 🎉 **Benefits of This Architecture:**

1. **✅ No Circular Dependencies**: Tests are completely isolated
2. **✅ Same-Package Tests**: Tests remain in handlers package as requested
3. **✅ Clean Separation**: Testing concerns separated from application code
4. **✅ Reusable Infrastructure**: `testhelpers` can be used by other test packages
5. **✅ Easy to Extend**: New tests can use the same infrastructure
6. **✅ CI/CD Ready**: Works perfectly in automated testing environments

## 🔄 **Migration from Previous Structure:**

The new architecture successfully resolves the import cycle issue while maintaining:

- **Same test functionality** as the original implementation
- **Same test coverage** for both public and protected routes
- **Same testing patterns** and helper functions
- **Same ease of use** for developers

## 📝 **Adding New Route Tests:**

To add tests for new routes:

1. Create new test functions in existing test files or new test files
2. Use `testhelpers.WithTestServer()` for server setup
3. Use `testhelpers.PerformRequest()` for HTTP requests
4. Use `testhelpers.AssertStatusCode()` for response validation
5. Follow the same patterns as existing tests

This architecture provides a solid foundation for testing web routes while maintaining clean code organization and avoiding import cycles.