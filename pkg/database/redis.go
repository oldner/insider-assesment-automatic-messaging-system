package database

import (
	"context"
	"insider-assessment/internal/config"
	"log/slog"

	"github.com/go-redis/redis/v8"
)

// NewRedisClient initializes the Redis connection and pings it to ensure it's alive
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		return nil, err
	}

	slog.Info("Successfully connected to Redis")
	return rdb, nil
}
