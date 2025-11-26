package database

import (
	"fmt"
	"insider-assessment/internal/config"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB handles the connection to PostgreSQL
func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	var db *gorm.DB
	var err error

	// retry logic for Docker startup
	for i := 1; i <= 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			break
		}
		slog.Warn("Failed to connect to database. Retrying...", "attempt", i)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("could not connect to database after 5 attempts: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10) // maximum 10 connections in pool

	sqlDB.SetMaxOpenConns(100) // default max 100 open connections to the db

	sqlDB.SetConnMaxLifetime(time.Hour) // 1 hour after a connection may be reused

	slog.Info("Successfully connected to PostgreSQL with Connection Pooling")
	return db, nil
}
