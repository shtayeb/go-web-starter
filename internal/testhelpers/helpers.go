package testhelpers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/stretchr/testify/require"
)

// RandomString generates a random string of the specified length
func RandomString(t TestingT, length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result strings.Builder
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		require.NoError(t, err)
		result.WriteByte(charset[num.Int64()])
	}
	return result.String()
}

// RandomEmail generates a random email address for testing
func RandomEmail(t TestingT) string {
	username := RandomString(t, 10)
	domain := RandomString(t, 5)
	return fmt.Sprintf("%s@%s.com", username, domain)
}

// RandomPassword generates a random password for testing
func RandomPassword(t TestingT) string {
	return RandomString(t, 12) + "A1!"
}

// RandomUsername generates a random username for testing
func RandomUsername(t TestingT) string {
	return "user_" + RandomString(t, 8)
}

// RandomID generates a random ID string
func RandomID(t TestingT) string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	require.NoError(t, err)
	return hex.EncodeToString(bytes)
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t TestingT, condition func() bool, timeout time.Duration, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for condition: %s", message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// AssertTimeWithinRange asserts that a time is within a range of another time
func AssertTimeWithinRange(t TestingT, actual, expected time.Time, delta time.Duration) {
	diff := actual.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	if diff > delta {
		t.Errorf("Time %v is not within %v of expected time %v", actual, delta, expected)
	}
}

// AssertContainsIgnoreCase asserts that a string contains a substring (case-insensitive)
func AssertContainsIgnoreCase(t TestingT, s, substr string) {
	if !strings.Contains(strings.ToLower(s), strings.ToLower(substr)) {
		t.Errorf("String '%s' does not contain '%s' (case-insensitive)", s, substr)
	}
}

// AssertStringLength asserts that a string has a specific length
func AssertStringLength(t TestingT, s string, expectedLength int) {
	if len(s) != expectedLength {
		t.Errorf("String '%s' has length %d, expected %d", s, len(s), expectedLength)
	}
}

// AssertStringMinLength asserts that a string has a minimum length
func AssertStringMinLength(t TestingT, s string, minLength int) {
	if len(s) < minLength {
		t.Errorf("String '%s' has length %d, expected at least %d", s, len(s), minLength)
	}
}

// AssertValidEmail asserts that a string is a valid email format
func AssertValidEmail(t TestingT, email string) {
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		t.Errorf("String '%s' is not a valid email format", email)
	}
}

// AssertValidURL asserts that a string is a valid URL format
func AssertValidURL(t TestingT, url string) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		t.Errorf("String '%s' is not a valid URL format", url)
	}
}
