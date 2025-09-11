package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"
)

type SecurityEvent struct {
	EventType string    `json:"event_type"`
	UserID    *uint     `json:"user_id,omitempty"`
	Email     string    `json:"email,omitempty"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Details   string    `json:"details,omitempty"`
	RiskLevel string    `json:"risk_level"` // low, medium, high, critical
}

type SecurityLogger struct{}

func NewSecurityLogger() *SecurityLogger {
	return &SecurityLogger{}
}

func (sl *SecurityLogger) LogEvent(event SecurityEvent) {
	// In production, you would send this to a proper logging system
	// For now, we'll use structured JSON logging
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal security event: %v", err)
		return
	}

	log.Printf("SECURITY_EVENT: %s", string(eventJSON))
}

func (sl *SecurityLogger) LogLoginAttempt(email, ipAddress, userAgent string, success bool, userID *uint) {
	riskLevel := "low"
	if !success {
		riskLevel = "medium"
	}

	sl.LogEvent(SecurityEvent{
		EventType: "login_attempt",
		UserID:    userID,
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   success,
		RiskLevel: riskLevel,
	})
}

func (sl *SecurityLogger) LogRegistration(email, ipAddress, userAgent string, success bool, userID *uint) {
	riskLevel := "low"
	if !success {
		riskLevel = "medium"
	}

	sl.LogEvent(SecurityEvent{
		EventType: "registration",
		UserID:    userID,
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   success,
		RiskLevel: riskLevel,
	})
}

func (sl *SecurityLogger) LogPasswordReset(email, ipAddress, userAgent string, success bool) {
	riskLevel := "medium"
	if !success {
		riskLevel = "high"
	}

	sl.LogEvent(SecurityEvent{
		EventType: "password_reset",
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   success,
		RiskLevel: riskLevel,
	})
}

func (sl *SecurityLogger) LogTokenRefresh(userID uint, ipAddress, userAgent string, success bool) {
	riskLevel := "low"
	if !success {
		riskLevel = "medium"
	}

	sl.LogEvent(SecurityEvent{
		EventType: "token_refresh",
		UserID:    &userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   success,
		RiskLevel: riskLevel,
	})
}

func (sl *SecurityLogger) LogLogout(userID uint, ipAddress, userAgent string) {
	sl.LogEvent(SecurityEvent{
		EventType: "logout",
		UserID:    &userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   true,
		RiskLevel: "low",
	})
}

func (sl *SecurityLogger) LogSuspiciousActivity(eventType, ipAddress, userAgent, details string) {
	sl.LogEvent(SecurityEvent{
		EventType: eventType,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   false,
		Details:   details,
		RiskLevel: "high",
	})
}

func (sl *SecurityLogger) LogAccountLockout(email, ipAddress, userAgent string) {
	sl.LogEvent(SecurityEvent{
		EventType: "account_lockout",
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   false,
		RiskLevel: "high",
		Details:   "Account locked due to multiple failed login attempts",
	})
}

func GenerateCSRFToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		// fallback: use a timestamp or panic, but crypto/rand should not fail in normal cases
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
