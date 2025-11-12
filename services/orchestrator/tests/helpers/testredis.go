package helpers

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/config"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/database"
)

// TestRedisConfig returns a test Redis configuration
func TestRedisConfig() config.RedisConfig {
	host := os.Getenv("TEST_REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("TEST_REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	password := os.Getenv("TEST_REDIS_PASSWORD")
	if password == "" {
		password = ""
	}

	// Use a separate test database (DB 1) to avoid conflicts
	db := 1
	if testDB := os.Getenv("TEST_REDIS_DB"); testDB != "" {
		fmt.Sscanf(testDB, "%d", &db)
	}

	return config.RedisConfig{
		Host:           host,
		Port:           port,
		Password:       password,
		DB:             db,
		IdempotencyTTL: 3600, // 1 hour default TTL
	}
}

// SetupTestRedis creates a test Redis connection
func SetupTestRedis(t *testing.T) (*database.RedisClient, func()) {
	cfg := TestRedisConfig()

	// Connect to Redis
	redisClient, err := database.NewRedis(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test Redis: %v", err)
	}

	// Cleanup function - flush the test database
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Flush the test database
		if err := redisClient.FlushDB(ctx).Err(); err != nil {
			t.Logf("Warning: Failed to flush Redis test database: %v", err)
		}

		redisClient.Close()
	}

	return redisClient, cleanup
}

// FlushTestRedis flushes all keys from the test Redis database
func FlushTestRedis(client *database.RedisClient, t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Logf("Warning: Failed to flush Redis: %v", err)
	}
}

// GetTestRedisKey generates a test-specific Redis key
func GetTestRedisKey(prefix, testName, key string) string {
	return fmt.Sprintf("test:%s:%s:%s", prefix, testName, key)
}
