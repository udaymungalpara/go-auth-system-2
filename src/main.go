// src/main.go
package main

import (
	"fmt"
	"go-auth-system/src/config"
	"go-auth-system/src/middleware"
	"go-auth-system/src/routes"
	"go-auth-system/src/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	config.Load()
	dsn := config.GetDatabaseURL()

	fmt.Println("Loaded DB URL:", dsn) // Debug print

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("[error] failed to initialize database, got error: %v\n", err)
		panic("failed to connect database: " + err.Error())
	}

	// Run database migrations
	if err := utils.RunMigrations(dsn); err != nil {
		fmt.Printf("[error] failed to run migrations: %v\n", err)
		panic("failed to run migrations: " + err.Error())
	}

	// Set Gin to release mode in production
	if config.GetPort() == "8080" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add security middleware
	router.Use(middleware.SecureHeaders())

	// Add CORS middleware for development
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-CSRF-Token")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	routes.SetupRoutes(router, db)

	fmt.Printf("Server starting on port %s\n", config.GetPort())
	router.Run(":" + config.GetPort())
}
