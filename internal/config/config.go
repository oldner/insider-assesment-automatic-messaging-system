package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBPort          string
	RedisAddr       string
	WebhookUrl      string
	ServerPort      string
	WorkerBatchSize int
	WorkerInterval  time.Duration
	RedisTTL        time.Duration
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found")
	}

	return &Config{
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "postgres"),
		DBPort:          getEnv("DB_PORT", "5432"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		WebhookUrl:      getEnv("WEBHOOK_URL", ""),
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		WorkerBatchSize: getEnvInt("WORKER_BATCH_SIZE", 2),
		WorkerInterval:  getEnvDuration("WORKER_INTERVAL", 2*time.Minute),
		RedisTTL:        getEnvDuration("REDIS_TTL", 24*time.Hour),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
