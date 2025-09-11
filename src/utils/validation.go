package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ValidateEmail validates email format and normalizes it
func ValidateEmail(email string) (string, error) {
	if email == "" {
		return "", fmt.Errorf("email is required")
	}

	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Basic email regex validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return "", fmt.Errorf("invalid email format")
	}

	// Check for suspicious patterns
	if strings.Contains(email, "..") || strings.HasPrefix(email, ".") || strings.HasSuffix(email, ".") {
		return "", fmt.Errorf("invalid email format")
	}

	// Block disposable/common placeholder domains
	atIndex := strings.LastIndex(email, "@")
	if atIndex > -1 && atIndex+1 < len(email) {
		domain := email[atIndex+1:]
		blockedDomains := map[string]struct{}{
			"example.com":      {},
			"example.org":      {},
			"example.net":      {},
			"test.com":         {},
			"mailinator.com":   {},
			"10minutemail.com": {},
			"temp-mail.org":    {},
			"yopmail.com":      {},
		}
		if _, blocked := blockedDomains[strings.ToLower(domain)]; blocked {
			return "", fmt.Errorf("email domain is not allowed for registration")
		}
	}

	return email, nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Require more than 8 characters
	if len(password) < 9 {
		return fmt.Errorf("password must be at least 9 characters long")
	}

	if len(password) > 128 {

		return fmt.Errorf("password must be less than 128 characters")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check for common weak passwords
	weakPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome", "monkey",
	}

	passwordLower := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if strings.Contains(passwordLower, weak) {
			return fmt.Errorf("password is too weak")
		}
	}

	return nil
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string) string {
	// Remove null bytes and control characters
	input = strings.ReplaceAll(input, "\x00", "")
	input = strings.ReplaceAll(input, "\r", "")
	input = strings.ReplaceAll(input, "\n", "")
	input = strings.ReplaceAll(input, "\t", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// ValidateName validates and sanitizes name fields
func ValidateName(name string) (string, error) {
	if name == "" {
		return "", nil // Names are optional
	}

	name = SanitizeString(name)

	if len(name) > 50 {
		return "", fmt.Errorf("name must be less than 50 characters")
	}

	// Check for only valid characters (letters, spaces, hyphens, apostrophes)
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !nameRegex.MatchString(name) {
		return "", fmt.Errorf("name contains invalid characters")
	}

	return name, nil
}

// ValidateTokenFormat validates token format
func ValidateTokenFormat(token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}

	if len(token) < 32 {
		return fmt.Errorf("invalid token format")
	}

	// Check for valid hex characters
	tokenRegex := regexp.MustCompile(`^[a-fA-F0-9]+$`)
	if !tokenRegex.MatchString(token) {
		return fmt.Errorf("invalid token format")
	}

	return nil
}

// ValidateCSRFToken validates CSRF token format
func ValidateCSRFToken(token string) error {
	if token == "" {
		return fmt.Errorf("CSRF token is required")
	}

	if len(token) < 32 {
		return fmt.Errorf("invalid CSRF token format")
	}

	return nil
}
