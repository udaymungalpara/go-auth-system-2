package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-auth-system/src/handlers"
	"go-auth-system/src/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AuthTestSuite struct {
	suite.Suite
	db       *gorm.DB
	handler  *handlers.AuthHandler
	router   *gin.Engine
	testUser models.User
}

func (suite *AuthTestSuite) SetupSuite() {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(suite.T(), err)

	// Auto-migrate for testing
	err = db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.PasswordResetToken{},
		&models.EmailVerificationToken{},
	)
	assert.NoError(suite.T(), err)

	suite.db = db
	suite.handler = handlers.NewAuthHandler(db)

	// Setup test router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Setup routes
	suite.setupTestRoutes()

	// Create test user
	suite.createTestUser()
}

func (suite *AuthTestSuite) setupTestRoutes() {
	suite.router.POST("/auth/register", suite.handler.Register)
	suite.router.POST("/auth/login", suite.handler.Login)
	suite.router.POST("/auth/refresh", suite.handler.RefreshToken)
	suite.router.POST("/auth/logout", suite.handler.Logout)
	suite.router.GET("/auth/verify", suite.handler.VerifyEmail)
	suite.router.POST("/auth/password/forgot", suite.handler.ForgotPassword)
	suite.router.POST("/auth/password/reset", suite.handler.ResetPassword)
}

func (suite *AuthTestSuite) createTestUser() {
	user := models.User{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	err := user.SetPassword("TestPassword123!")
	assert.NoError(suite.T(), err)

	err = suite.db.Create(&user).Error
	assert.NoError(suite.T(), err)

	suite.testUser = user
}

func (suite *AuthTestSuite) TestRegister() {
	// Test successful registration
	registerData := map[string]string{
		"email":      "newuser@example.com",
		"password":   "NewPassword123!",
		"first_name": "New",
		"last_name":  "User",
	}

	jsonData, _ := json.Marshal(registerData)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	// Test duplicate email
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	// Test invalid email
	invalidData := map[string]string{
		"email":      "invalid-email",
		"password":   "NewPassword123!",
		"first_name": "New",
		"last_name":  "User",
	}

	jsonData, _ = json.Marshal(invalidData)
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// Test weak password
	weakPasswordData := map[string]string{
		"email":      "weak@example.com",
		"password":   "123",
		"first_name": "Weak",
		"last_name":  "User",
	}

	jsonData, _ = json.Marshal(weakPasswordData)
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *AuthTestSuite) TestLogin() {
	// Test successful login
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "TestPassword123!",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "access_token")
	assert.Contains(suite.T(), response, "refresh_token")

	// Test invalid credentials
	invalidLoginData := map[string]string{
		"email":    "test@example.com",
		"password": "WrongPassword",
	}

	jsonData, _ = json.Marshal(invalidLoginData)
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	// Test non-existent user
	nonExistentData := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "SomePassword123!",
	}

	jsonData, _ = json.Marshal(nonExistentData)
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthTestSuite) TestRefreshToken() {
	// First login to get tokens
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "TestPassword123!",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var loginResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(suite.T(), err)

	refreshToken := loginResponse["refresh_token"].(string)

	// Test refresh token
	refreshData := map[string]string{
		"refresh_token": refreshToken,
	}

	jsonData, _ = json.Marshal(refreshData)
	req, _ = http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var refreshResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &refreshResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), refreshResponse, "access_token")
	assert.Contains(suite.T(), refreshResponse, "refresh_token")

	// Test invalid refresh token
	invalidRefreshData := map[string]string{
		"refresh_token": "invalid-token",
	}

	jsonData, _ = json.Marshal(invalidRefreshData)
	req, _ = http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthTestSuite) TestLogout() {
	// First login to get tokens
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "TestPassword123!",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(suite.T(), err)

	refreshToken := loginResponse["refresh_token"].(string)

	// Test logout
	logoutData := map[string]string{
		"refresh_token": refreshToken,
	}

	jsonData, _ = json.Marshal(logoutData)
	req, _ = http.NewRequest("POST", "/auth/logout", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test logout with invalid token
	invalidLogoutData := map[string]string{
		"refresh_token": "invalid-token",
	}

	jsonData, _ = json.Marshal(invalidLogoutData)
	req, _ = http.NewRequest("POST", "/auth/logout", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthTestSuite) TestForgotPassword() {
	// Test forgot password with existing user
	forgotData := map[string]string{
		"email": "test@example.com",
	}

	jsonData, _ := json.Marshal(forgotData)
	req, _ := http.NewRequest("POST", "/auth/password/forgot", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test forgot password with non-existing user (should still return OK)
	nonExistentData := map[string]string{
		"email": "nonexistent@example.com",
	}

	jsonData, _ = json.Marshal(nonExistentData)
	req, _ = http.NewRequest("POST", "/auth/password/forgot", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test invalid email format
	invalidData := map[string]string{
		"email": "invalid-email",
	}

	jsonData, _ = json.Marshal(invalidData)
	req, _ = http.NewRequest("POST", "/auth/password/forgot", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
