package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	client *redis.Client
}

type SessionData struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	LoginTime time.Time `json:"login_time"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

func NewRedisService(redisURL string) (*RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "cache:6379", // Use service name in Docker
		Password: "",
		DB:       0,
	})

	// Test connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisService{client: client}, nil
}

// Token Management
func (rs *RedisService) BlacklistToken(token string, expiration time.Duration) error {
	ctx := context.Background()
	return rs.client.Set(ctx, "blacklist:"+token, "true", expiration).Err()
}

func (rs *RedisService) IsTokenBlacklisted(token string) (bool, error) {
	ctx := context.Background()
	result, err := rs.client.Get(ctx, "blacklist:"+token).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return result == "true", nil
}

// Session Management
func (rs *RedisService) StoreSession(sessionID string, sessionData SessionData, expiration time.Duration) error {
	ctx := context.Background()
	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	return rs.client.Set(ctx, "session:"+sessionID, data, expiration).Err()
}

func (rs *RedisService) GetSession(sessionID string) (*SessionData, error) {
	ctx := context.Background()
	result, err := rs.client.Get(ctx, "session:"+sessionID).Result()
	if err == redis.Nil {
		return nil, nil // Session not found
	}
	if err != nil {
		return nil, err
	}

	var sessionData SessionData
	if err := json.Unmarshal([]byte(result), &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &sessionData, nil
}

func (rs *RedisService) DeleteSession(sessionID string) error {
	ctx := context.Background()
	return rs.client.Del(ctx, "session:"+sessionID).Err()
}

// User Session Management
func (rs *RedisService) StoreUserSession(userID uint, sessionID string, expiration time.Duration) error {
	ctx := context.Background()
	return rs.client.Set(ctx, fmt.Sprintf("user_session:%d", userID), sessionID, expiration).Err()
}

func (rs *RedisService) GetUserSession(userID uint) (string, error) {
	ctx := context.Background()
	result, err := rs.client.Get(ctx, fmt.Sprintf("user_session:%d", userID)).Result()
	if err == redis.Nil {
		return "", nil // No active session
	}
	return result, err
}

func (rs *RedisService) DeleteUserSession(userID uint) error {
	ctx := context.Background()
	return rs.client.Del(ctx, fmt.Sprintf("user_session:%d", userID)).Err()
}

// Rate Limiting
func (rs *RedisService) IncrementRateLimit(key string, expiration time.Duration) (int64, error) {
	ctx := context.Background()

	// Use pipeline for atomic operations
	pipe := rs.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

func (rs *RedisService) GetRateLimit(key string) (int64, error) {
	ctx := context.Background()
	result, err := rs.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return result, err
}

// Cache Management
func (rs *RedisService) SetCache(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	return rs.client.Set(ctx, "cache:"+key, data, expiration).Err()
}

func (rs *RedisService) GetCache(key string, dest interface{}) error {
	ctx := context.Background()
	result, err := rs.client.Get(ctx, "cache:"+key).Result()
	if err == redis.Nil {
		return nil // Cache miss
	}
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(result), dest)
}

func (rs *RedisService) DeleteCache(key string) error {
	ctx := context.Background()
	return rs.client.Del(ctx, "cache:"+key).Err()
}

// Cleanup expired sessions and tokens
func (rs *RedisService) CleanupExpired() error {
	ctx := context.Background()

	// This is a simple cleanup - in production you might want to use Redis TTL
	// or implement a more sophisticated cleanup mechanism

	// Get all session keys
	sessionKeys, err := rs.client.Keys(ctx, "session:*").Result()
	if err != nil {
		return err
	}

	// Check TTL for each session and delete if expired
	for _, key := range sessionKeys {
		ttl, err := rs.client.TTL(ctx, key).Result()
		if err != nil {
			continue
		}
		if ttl == -1 { // No expiration set
			rs.client.Del(ctx, key)
		}
	}

	return nil
}

// Close connection
func (rs *RedisService) Close() error {
	return rs.client.Close()
}

