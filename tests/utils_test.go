package tests

import (
	"testing"

	"go-auth-system/src/utils"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
		hasError bool
	}{
		{
			name:     "Valid email",
			email:    "test@example.com",
			expected: "test@example.com",
			hasError: false,
		},
		{
			name:     "Valid email with uppercase",
			email:    "TEST@EXAMPLE.COM",
			expected: "test@example.com",
			hasError: false,
		},
		{
			name:     "Valid email with spaces",
			email:    "  test@example.com  ",
			expected: "test@example.com",
			hasError: false,
		},
		{
			name:     "Invalid email - no @",
			email:    "testexample.com",
			expected: "",
			hasError: true,
		},
		{
			name:     "Invalid email - no domain",
			email:    "test@",
			expected: "",
			hasError: true,
		},
		{
			name:     "Invalid email - no local part",
			email:    "@example.com",
			expected: "",
			hasError: true,
		},
		{
			name:     "Invalid email - double dots",
			email:    "test..test@example.com",
			expected: "",
			hasError: true,
		},
		{
			name:     "Empty email",
			email:    "",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.ValidateEmail(tt.email)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		hasError bool
	}{
		{
			name:     "Valid password",
			password: "MySecurePass123!",
			hasError: false,
		},
		{
			name:     "Password too short",
			password: "Short1!",
			hasError: true,
		},
		{
			name:     "Password too long",
			password: "ThisPasswordIsWayTooLongAndExceedsTheMaximumAllowedLengthOfOneHundredAndTwentyEightCharactersAndShouldFailValidation123!",
			hasError: true,
		},
		{
			name:     "Password without uppercase",
			password: "lowercase123!",
			hasError: true,
		},
		{
			name:     "Password without lowercase",
			password: "UPPERCASE123!",
			hasError: true,
		},
		{
			name:     "Password without number",
			password: "NoNumbers!",
			hasError: true,
		},
		{
			name:     "Password without special character",
			password: "NoSpecial123",
			hasError: true,
		},
		{
			name:     "Weak password - contains 'password'",
			password: "Password123!",
			hasError: true,
		},
		{
			name:     "Weak password - contains '123456'",
			password: "MyPassword123456!",
			hasError: true,
		},
		{
			name:     "Empty password",
			password: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidatePassword(tt.password)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Valid name",
			input:    "John",
			expected: "John",
			hasError: false,
		},
		{
			name:     "Valid name with spaces",
			input:    "John Doe",
			expected: "John Doe",
			hasError: false,
		},
		{
			name:     "Valid name with hyphen",
			input:    "Mary-Jane",
			expected: "Mary-Jane",
			hasError: false,
		},
		{
			name:     "Valid name with apostrophe",
			input:    "O'Connor",
			expected: "O'Connor",
			hasError: false,
		},
		{
			name:     "Empty name (optional)",
			input:    "",
			expected: "",
			hasError: false,
		},
		{
			name:     "Name with numbers",
			input:    "John123",
			expected: "",
			hasError: true,
		},
		{
			name:     "Name with special characters",
			input:    "John@Doe",
			expected: "",
			hasError: true,
		},
		{
			name:     "Name too long",
			input:    "ThisNameIsWayTooLongAndExceedsTheMaximumAllowedLengthOfFiftyCharacters",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.ValidateName(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidateTokenFormat(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		hasError bool
	}{
		{
			name:     "Valid token",
			token:    "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456",
			hasError: false,
		},
		{
			name:     "Token too short",
			token:    "short",
			hasError: true,
		},
		{
			name:     "Empty token",
			token:    "",
			hasError: true,
		},
		{
			name:     "Token with invalid characters",
			token:    "invalid-token-with-dashes-and-special-chars!",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidateTokenFormat(tt.token)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "String with null bytes",
			input:    "Hello\x00World",
			expected: "HelloWorld",
		},
		{
			name:     "String with carriage returns",
			input:    "Hello\rWorld",
			expected: "HelloWorld",
		},
		{
			name:     "String with newlines",
			input:    "Hello\nWorld",
			expected: "HelloWorld",
		},
		{
			name:     "String with tabs",
			input:    "Hello\tWorld",
			expected: "HelloWorld",
		},
		{
			name:     "String with leading/trailing spaces",
			input:    "  Hello World  ",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "TestPassword123!"

	hash, err := utils.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Test that the same password produces different hashes (due to salt)
	hash2, err := utils.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEqual(t, hash, hash2)
}

func TestCheckPasswordHash(t *testing.T) {
	password := "TestPassword123!"
	wrongPassword := "WrongPassword123!"

	hash, err := utils.HashPassword(password)
	assert.NoError(t, err)

	// Test correct password
	assert.True(t, utils.CheckPasswordHash(password, hash))

	// Test wrong password
	assert.False(t, utils.CheckPasswordHash(wrongPassword, hash))

	// Test empty password
	assert.False(t, utils.CheckPasswordHash("", hash))

	// Test empty hash
	assert.False(t, utils.CheckPasswordHash(password, ""))
}
