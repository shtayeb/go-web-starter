package security

import (
	"testing"
)

func TestIsValidRedirectPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
		name     string
	}{
		{"/dashboard", true, "valid relative path"},
		{"/user/profile", true, "valid nested path"},
		{"//evil.com", false, "protocol-relative URL"},
		{"http://evil.com", false, "absolute URL"},
		{"https://evil.com", false, "absolute HTTPS URL"},
		{"/path/../../../etc/passwd", false, "directory traversal"},
		{"/path/with/..hidden", false, "path with .. component"},
		{"", false, "empty path"},
		{"/", false, "root path only"},
		{"/valid/path", true, "valid multi-segment path"},
		{"/path\x00null", false, "path with null byte"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidRedirectPath(tt.path)
			if result != tt.expected {
				t.Errorf("IsValidRedirectPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		password    string
		expectValid bool
		name        string
	}{
		{"Password123!", true, "strong password"},
		{"weak", false, "too short"},
		{"password123", false, "no uppercase or special chars"},
		{"PASSWORD123!", false, "no lowercase"},
		{"Password!", false, "no digits"},
		{"Password123", false, "no special characters"},
		{"Aa1!", false, "too short but has all character types"},
		{"Password1111!", false, "repeated characters"},
		{"Passwordaaaa!", false, "repeated letters"},
		{"Password123456!", false, "sequential numbers"},
		{"PasswordQwerty1!", false, "keyboard pattern"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidatePasswordStrength(tt.password)
			isValid := len(errors) == 0
			if isValid != tt.expectValid {
				t.Errorf("ValidatePasswordStrength(%q) valid=%v, want %v. Errors: %v",
					tt.password, isValid, tt.expectValid, errors)
			}
		})
	}
}
