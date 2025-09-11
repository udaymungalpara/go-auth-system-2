package routes

import (
	"go-auth-system/src/handlers"
	"go-auth-system/src/middleware"
	"go-auth-system/src/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	userHandler := handlers.NewUserHandler(db)
	rateLimiter := middleware.NewRateLimiter()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Welcome endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "welcome to authorization system"})
	})

	// CSRF token endpoint
	router.GET("/csrf-token", func(c *gin.Context) {
		token := utils.GenerateCSRFToken()
		// Set the CSRF token as a cookie
		c.SetCookie("csrf_token", token, 3600, "/", "", false, true)
		c.JSON(200, gin.H{"csrf_token": token})
	})

	// Public routes with rate limiting
	publicGroup := router.Group("/")
	{
		// General rate limiting for all public endpoints
		publicGroup.Use(rateLimiter.RateLimitByIP(100, 15*60)) // 100 requests per 15 minutes per IP

		// Auth routes
		authGroup := publicGroup.Group("/auth")
		{
			// Routes that don't need CSRF protection (GET requests)
			authGroup.GET("/verify", authHandler.VerifyEmail)

			// Routes that need CSRF protection
			csrfGroup := authGroup.Group("/")
			csrfGroup.Use(middleware.CSRFProtection())
			{
				csrfGroup.POST("/register", authHandler.Register)
				csrfGroup.POST("/login",
					rateLimiter.LoginRateLimit(5, 15*60), // 5 login attempts per 15 minutes
					authHandler.Login)
				csrfGroup.POST("/refresh", authHandler.RefreshToken)
				csrfGroup.POST("/password/forgot",
					rateLimiter.PasswordResetRateLimit(3, 60*60), // 3 password reset attempts per hour
					authHandler.ForgotPassword)
				csrfGroup.POST("/password/reset", authHandler.ResetPassword)
			}
		}
	}

	// Protected routes
	protectedGroup := router.Group("/")
	protectedGroup.Use(middleware.AuthMiddleware())
	{
		// Authenticated "me" endpoint
		protectedGroup.GET("/auth/me", authHandler.Me)

		// Logout endpoint (requires authentication)
		protectedGroup.POST("/auth/logout", authHandler.Logout)

		// User routes
		userGroup := protectedGroup.Group("/user")
		{
			userGroup.GET("/profile/:id", userHandler.GetUser)
			userGroup.PUT("/update/:id", userHandler.UpdateUser)
		}

	}
}
