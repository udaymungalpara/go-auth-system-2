package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	redisClient *redis.Client
}

func NewRateLimiter() *RateLimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "cache:6379",
		Password: "",
		DB:       0,
	})
	return &RateLimiter{redisClient: rdb}
}

func (rl *RateLimiter) RateLimitByIP(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:ip:%s", clientIP)

		count, err := rl.redisClient.Incr(context.Background(), key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit error"})
			c.Abort()
			return
		}

		if count == 1 {
			rl.redisClient.Expire(context.Background(), key, window)
		}

		if count > int64(maxRequests) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) RateLimitByUser(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		key := fmt.Sprintf("rate_limit:user:%d", userID)

		count, err := rl.redisClient.Incr(context.Background(), key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit error"})
			c.Abort()
			return
		}

		if count == 1 {
			rl.redisClient.Expire(context.Background(), key, window)
		}

		if count > int64(maxRequests) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) LoginRateLimit(maxAttempts int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("login_attempts:%s", clientIP)

		count, err := rl.redisClient.Incr(context.Background(), key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit error"})
			c.Abort()
			return
		}

		if count == 1 {
			rl.redisClient.Expire(context.Background(), key, window)
		}

		if count > int64(maxAttempts) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many login attempts",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) PasswordResetRateLimit(maxAttempts int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("password_reset_attempts:%s", clientIP)

		count, err := rl.redisClient.Incr(context.Background(), key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit error"})
			c.Abort()
			return
		}

		if count == 1 {
			rl.redisClient.Expire(context.Background(), key, window)
		}

		if count > int64(maxAttempts) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many password reset attempts",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
