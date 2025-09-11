package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-auth-system/src/models"
	"go-auth-system/src/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB             *gorm.DB
	RedisClient    *redis.Client
	SecurityLogger *utils.SecurityLogger
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "cache:6379",
		Password: "",
		DB:       0,
	})
	return &AuthHandler{
		DB:             db,
		RedisClient:    rdb,
		SecurityLogger: utils.NewSecurityLogger(),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validate inputs
	normalizedEmail, err := utils.ValidateEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := utils.ValidatePassword(input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	firstName, err := utils.ValidateName(input.FirstName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	lastName, err := utils.ValidateName(input.LastName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check for existing user
	var existing models.User
	if err := h.DB.Where("email = ?", normalizedEmail).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Generate custom user_id (e.g., user1, user2, ...)
	var count int64
	h.DB.Model(&models.User{}).Count(&count)

	// Hash password
	passwordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password error"})
		return
	}

	user := models.User{
		Email:            normalizedEmail,
		PasswordHash:     passwordHash,
		FirstName:        firstName,
		LastName:         lastName,
		IsEmailVerified:  false,
		FailedLoginCount: 0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := h.DB.Create(&user).Error; err != nil {
		fmt.Println("DB error:", err) // Debugging
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	// Generate email verification token for testing
	verificationToken, err := utils.GenerateEmailVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate verification token"})
		return
	}

	// Store email verification token
	verificationTokenRecord := models.EmailVerificationToken{
		UserID:    user.ID,
		Token:     verificationToken,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours expiry
	}

	if err := h.DB.Create(&verificationTokenRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create verification token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":            "User registered successfully",
		"user_id":            user.ID,
		"verification_token": verificationToken, // For testing purposes
		"verification_url":   fmt.Sprintf("/auth/verify?token=%s", verificationToken),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	// Validate and normalize email
	normalizedEmail, err := utils.ValidateEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", normalizedEmail).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if account is locked
	if user.IsAccountLocked() {
		h.SecurityLogger.LogAccountLockout(normalizedEmail, c.ClientIP(), c.GetHeader("User-Agent"))
		c.JSON(http.StatusLocked, gin.H{"error": "Account is temporarily locked due to too many failed login attempts"})
		return
	}

	if !user.CheckPassword(input.Password) {
		user.IncrementFailedLogin()
		h.DB.Save(&user)
		h.SecurityLogger.LogLoginAttempt(normalizedEmail, c.ClientIP(), c.GetHeader("User-Agent"), false, &user.ID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Reset failed login count on successful login
	user.ResetFailedLoginCount()
	now := time.Now()
	user.LastLoginAt = &now
	h.DB.Save(&user)

	// Log successful login
	h.SecurityLogger.LogLoginAttempt(normalizedEmail, c.ClientIP(), c.GetHeader("User-Agent"), true, &user.ID)

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate refresh token"})
		return
	}

	// Store refresh token in database
	refreshTokenRecord := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.DB.Create(&refreshTokenRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not store refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    100000, // 100 seconds
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validate refresh token
	claims, err := utils.ValidateToken(input.RefreshToken, utils.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Check if refresh token is blacklisted
	blacklisted, err := h.RedisClient.Get(context.Background(), "blacklist:"+input.RefreshToken).Result()
	if err == nil && blacklisted == "true" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has been revoked"})
		return
	}

	// Check if refresh token exists in database
	var refreshTokenRecord models.RefreshToken
	if err := h.DB.Where("token = ? AND expires_at > ?", input.RefreshToken, time.Now()).First(&refreshTokenRecord).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Blacklist the current access token (if provided) so it stops working immediately
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		currentAccessToken := strings.TrimPrefix(authHeader, "Bearer ")
		// Default TTL for blacklist; will be overridden by remaining token lifetime if available
		ttl := 15 * time.Minute
		if accessClaims, err := utils.ValidateToken(currentAccessToken, utils.AccessToken); err == nil && accessClaims.ExpiresAt != nil {
			remaining := time.Until(accessClaims.ExpiresAt.Time)
			if remaining > 0 {
				ttl = remaining
			}
		}
		_ = h.RedisClient.Set(context.Background(), "blacklist:"+currentAccessToken, "true", ttl).Err()
	}

	// Generate new tokens (token rotation)
	newAccessToken, err := utils.GenerateAccessToken(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate access token"})
		return
	}

	newRefreshToken, err := utils.GenerateRefreshToken(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate refresh token"})
		return
	}

	// Blacklist the old refresh token
	h.RedisClient.Set(context.Background(), "blacklist:"+input.RefreshToken, "true", 7*24*time.Hour)

	// Remove old refresh token from database
	h.DB.Where("token = ?", input.RefreshToken).Delete(&models.RefreshToken{})

	// Store new refresh token
	newRefreshTokenRecord := models.RefreshToken{
		UserID:    claims.UserID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.DB.Create(&newRefreshTokenRecord).Error; err != nil {
		h.SecurityLogger.LogTokenRefresh(claims.UserID, c.ClientIP(), c.GetHeader("User-Agent"), false)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not store new refresh token"})
		return
	}

	// Log successful token refresh
	h.SecurityLogger.LogTokenRefresh(claims.UserID, c.ClientIP(), c.GetHeader("User-Agent"), true)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    100000, // 100 seconds
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user ID from authenticated context
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	// Optional: Accept refresh token in request body for additional cleanup
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}

	// Try to bind JSON, but don't require it
	c.ShouldBindJSON(&input)

	// If refresh token is provided, remove it from database and blacklist it
	if input.RefreshToken != "" {
		// Validate refresh token first
		_, err := utils.ValidateToken(input.RefreshToken, utils.RefreshToken)
		if err == nil {
			// Remove refresh token from database
			h.DB.Where("token = ?", input.RefreshToken).Delete(&models.RefreshToken{})

			// Blacklist the refresh token in Redis
			h.RedisClient.Set(context.Background(), "blacklist:"+input.RefreshToken, "true", 7*24*time.Hour)
		}
	}

	// Blacklist the current access token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		currentAccessToken := strings.TrimPrefix(authHeader, "Bearer ")
		// Calculate remaining token lifetime for proper TTL
		ttl := 15 * time.Minute // Default TTL
		if accessClaims, err := utils.ValidateToken(currentAccessToken, utils.AccessToken); err == nil && accessClaims.ExpiresAt != nil {
			remaining := time.Until(accessClaims.ExpiresAt.Time)
			if remaining > 0 {
				ttl = remaining
			}
		}
		// Blacklist the access token
		h.RedisClient.Set(context.Background(), "blacklist:"+currentAccessToken, "true", ttl)
	}

	// Clear any cached user sessions
	h.RedisClient.Del(context.Background(), fmt.Sprintf("user_session:%d", userID))

	// Log logout
	h.SecurityLogger.LogLogout(userID, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	var verificationToken models.EmailVerificationToken
	if err := h.DB.Where("token = ? AND expires_at > ? AND used = ?", token, time.Now(), false).First(&verificationToken).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Mark token as used
	verificationToken.Used = true
	h.DB.Save(&verificationToken)

	// Update user email verification status
	now := time.Now()
	if err := h.DB.Model(&models.User{}).Where("id = ?", verificationToken.UserID).Updates(map[string]interface{}{
		"is_email_verified": true,
		"email_verified_at": &now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	// Validate and normalize email
	normalizedEmail, err := utils.ValidateEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", normalizedEmail).First(&user).Error; err != nil {
		// Don't reveal if user exists or not
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a password reset link has been sent"})
		return
	}

	// Generate password reset token
	resetToken, err := utils.GeneratePasswordResetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate reset token"})
		return
	}

	// Store password reset token
	resetTokenRecord := models.PasswordResetToken{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour expiry
	}

	if err := h.DB.Create(&resetTokenRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create reset token"})
		return
	}

	// Send password reset email
	mailService := utils.NewMailService()
	if err := mailService.SendPasswordResetEmail(user.Email, resetToken); err != nil {
		// Log error but don't reveal if user exists
		fmt.Printf("Failed to send password reset email: %v\n", err)
		h.SecurityLogger.LogPasswordReset(normalizedEmail, c.ClientIP(), c.GetHeader("User-Agent"), false)

		// Return the token for testing purposes when email fails
		c.JSON(http.StatusOK, gin.H{
			"message":     "If the email exists, a password reset link has been sent",
			"reset_token": resetToken, // For testing when email fails
			"note":        "Email sending failed - using token for testing",
			"error":       err.Error(),
		})
		return
	}

	// Log successful password reset request
	h.SecurityLogger.LogPasswordReset(normalizedEmail, c.ClientIP(), c.GetHeader("User-Agent"), true)

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	// Validate token
	if err := utils.ValidateTokenFormat(input.Token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate new password
	if err := utils.ValidatePassword(input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var resetToken models.PasswordResetToken
	if err := h.DB.Where("token = ? AND expires_at > ? AND used = ?", input.Token, time.Now(), false).First(&resetToken).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Mark token as used
	resetToken.Used = true
	h.DB.Save(&resetToken)

	// Update user password
	var user models.User
	if err := h.DB.First(&user, resetToken.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := user.SetPassword(input.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not reset password"})
		return
	}

	// Reset failed login count
	user.ResetFailedLoginCount()

	if err := h.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// Me returns the authenticated user's profile based on the userID from context
func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	})
}
