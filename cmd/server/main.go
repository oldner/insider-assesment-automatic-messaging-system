package main

import (
	"insider-assessment/internal/config"
	"insider-assessment/internal/handler"
	"insider-assessment/internal/model"
	"insider-assessment/internal/repository"
	"insider-assessment/internal/router"
	"insider-assessment/internal/service"
	"insider-assessment/pkg/database"
	"insider-assessment/pkg/logger"
	"log/slog"

	_ "insider-assessment/docs"

	"github.com/gin-gonic/gin"
)

// @title Insider Assessment API
// @version 1.0
// @description API for Automatic Message Sending System.
// @host localhost:8080
// @BasePath /
func main() {
	// initialize logger
	logger.InitLogger()

	// load config
	cfg := config.Load()

	// initialize db
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		slog.Error("database initialization failed", "error", err)
		panic(err)
	}

	// auto-migrate db
	if err := db.AutoMigrate(&model.Message{}); err != nil {
		slog.Error("database migration failed", "error", err)
	}

	// initialize redis
	rdb, err := database.NewRedisClient(cfg)
	if err != nil {
		slog.Warn("redis initialization failed. Running without cache.", "error", err)
		// we don't panic here because Redis is a "Bonus" item. app can technically run without it.
	}

	// create repos and services - dependency injection
	msgRepo := repository.NewMessageRepository(db)
	senderSvc := service.NewWorkerService(msgRepo, rdb, cfg)
	scheduler := service.NewScheduler(senderSvc, cfg)

	// start the scheduler
	scheduler.Start()

	// HTTP handler Setup
	h := handler.NewHandler(scheduler, msgRepo)

	// router setup
	r := gin.Default()
	router.InitRoutes(r, h)

	slog.Info("Server starting", "port", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		slog.Error("Server failed to start", "error", err)
	}
}
