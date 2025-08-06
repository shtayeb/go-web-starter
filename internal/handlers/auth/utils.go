package auth

import (
	"regexp"
	"unicode"
)

// ValidatePasswordStrength checks if password meets security requirements
func ValidatePasswordStrength(password string) []string {
	var errors []string

	if len(password) < 8 {
		errors = append(errors, "Password must be at least 8 characters long")
	}

	if len(password) > 128 {
		errors = append(errors, "Password must be less than 128 characters long")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors = append(errors, "Password must contain at least one uppercase letter")
	}
	if !hasLower {
		errors = append(errors, "Password must contain at least one lowercase letter")
	}
	if !hasDigit {
		errors = append(errors, "Password must contain at least one digit")
	}
	if !hasSpecial {
		errors = append(errors, "Password must contain at least one special character")
	}

	// Check for repeated characters (manual check since Go regex doesn't support backreferences)
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			errors = append(errors, "Password contains repeated characters")
			break
		}
	}

	// Check for common patterns
	commonPatterns := []string{
		`123456`,        // sequential numbers
		`qwerty`,        // keyboard patterns
		`^password\d*$`, // password followed by optional digits
	}

	for _, pattern := range commonPatterns {
		if matched, _ := regexp.MatchString(`(?i)`+pattern, password); matched {
			errors = append(errors, "Password contains common patterns and is not secure")
			break
		}
	}

	return errors
}

func IsValidRedirectPath(path string) bool {
	// Only allow relative paths that start with /
	// Reject absolute URLs, protocol-relative URLs, and paths with ..
	if path == "" {
		return false
	}

	// Must start with / but not //
	if !regexp.MustCompile(`^/[^/]`).MatchString(path) {
		return false
	}

	// Reject paths with .. to prevent directory traversal
	if regexp.MustCompile(`\.\.`).MatchString(path) {
		return false
	}

	// Reject paths with null bytes or other dangerous characters
	if regexp.MustCompile(`[\x00-\x1f\x7f]`).MatchString(path) {
		return false
	}

	return true
}

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email) && len(email) <= 255
}
