package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"go-auth-system/src/config"
	"go-auth-system/src/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func AuthMiddleware() gin.HandlerFunc {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "cache:6379",
		Password: "",
		DB:       0,
	})

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Check if token is blacklisted
		blacklisted, err := rdb.Get(context.Background(), "blacklist:"+tokenString).Result()
		if err == nil && blacklisted == "true" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenString, utils.AccessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userIDString", strconv.FormatUint(uint64(claims.UserID), 10))
		c.Next()
	}
}

func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			c.Next()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims, err := utils.ValidateToken(tokenString, utils.AccessToken)
		if err != nil {
			c.Next()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userIDString", strconv.FormatUint(uint64(claims.UserID), 10))
		c.Next()
	}
}

// CSRFProtection middleware for protecting against CSRF attacks
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for GET, HEAD, OPTIONS requests
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Check for CSRF token in header
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token missing"})
			c.Abort()
			return
		}

		// In a real implementation, you would validate the CSRF token
		// For now, we'll just check if it exists
		if len(csrfToken) < 32 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecureHeaders middleware to add security headers
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict Transport Security (HSTS) - only in production with HTTPS
		if config.GetPort() != "8080" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// CSRFMiddleware for validating CSRF tokens
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		csrfTokenFromHeader := c.GetHeader("X-CSRF-Token")
		csrfTokenFromCookie, err := c.Cookie("csrf_token")
		if err != nil || csrfTokenFromHeader == "" || csrfTokenFromHeader != csrfTokenFromCookie {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token invalid or missing"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func UserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, gin.H{"userID": userID})
		c.Next()
	}
}
