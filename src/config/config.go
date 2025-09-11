package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	port         string
	databaseURL  string
	redisURL     string
	jwtSecret    string
	emailService string
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
)

func Load() {
	// Load .env file if exists
	_ = godotenv.Load(".env")

	port = os.Getenv("PORT")
	if port == "" {
		port = "8080" // default port
	}

	databaseURL = os.Getenv("DATABASE_URL")
	redisURL = os.Getenv("REDIS_URL")
	jwtSecret = os.Getenv("JWT_SECRET")

	emailService = os.Getenv("EMAIL_SERVICE")
	smtpHost = os.Getenv("SMTP_HOST")
	smtpUsername = os.Getenv("SMTP_USERNAME")
	smtpPassword = os.Getenv("SMTP_PASSWORD")

	// Set default SMTP values if not provided
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com" // Def
		// ault to Gmail SMTP
	}

	// Parse SMTP port
	if portStr := os.Getenv("SMTP_PORT"); portStr != "" {
		if parsedPort, err := strconv.Atoi(portStr); err == nil {
			smtpPort = parsedPort
		} else {
			smtpPort = 587 // default SMTP port
		}
	} else {
		smtpPort = 587 // default SMTP port
	}

	// Set default JWT secret if not provided
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-change-in-production"
	}
}

func GetPort() string {
	return port
}

func GetDatabaseURL() string {
	return databaseURL
}

func GetRedisURL() string {
	return redisURL
}

func GetJWTSecret() string {
	return jwtSecret
}

func GetEmailService() string {
	return emailService
}

func GetSMTPHost() string {
	return smtpHost
}

func GetSMTPPort() int {
	return smtpPort
}

func GetSMTPUsername() string {
	return smtpUsername
}

func GetSMTPPassword() string {
	return smtpPassword
}
