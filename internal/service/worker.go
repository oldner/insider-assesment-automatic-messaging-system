package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"insider-assessment/internal/config"
	"insider-assessment/internal/model"
	"insider-assessment/internal/repository"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type WorkerService struct {
	Repo   repository.MessageRepository
	Redis  *redis.Client
	Config *config.Config
}

func NewWorkerService(repo repository.MessageRepository, rdb *redis.Client, cfg *config.Config) *WorkerService {
	return &WorkerService{
		Repo:   repo,
		Redis:  rdb,
		Config: cfg,
	}
}

type WebhookResponse struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

// ProcessMessages picks up messages and attempts to send them
func (s *WorkerService) ProcessMessages() {
	slog.Info("--- Ticker: Checking for pending messages ---")

	messages, err := s.Repo.GetPending(s.Config.WorkerBatchSize)
	if err != nil {
		slog.Error("error fetching messages", "error", err)
		return
	}

	if len(messages) == 0 {
		slog.Info("no pending messages found.")
		return
	}

	for _, msg := range messages {
		go s.sendMessage(msg) // send in parallel
	}
}

func (s *WorkerService) sendMessage(msg model.Message) {
	payload := map[string]string{
		"to":      msg.To,
		"content": msg.Content,
	}
	jsonVal, _ := json.Marshal(payload)

	resp, err := http.Post(s.Config.WebhookUrl, "application/json", bytes.NewBuffer(jsonVal))
	if err != nil {
		slog.Error("failed to send message", "id", msg.ID, "error", err)
		s.Repo.UpdateStatus(msg.ID, model.StatusFailed)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 202 {
		var result WebhookResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			slog.Error("failed to decode response", "id", msg.ID, "error", err)
		}

		// update DB
		s.Repo.UpdateStatus(msg.ID, model.StatusSent)
		slog.Info("message sent successfully", "id", msg.ID, "remote_id", result.MessageID)

		// cache to Redis
		if s.Redis != nil && result.MessageID != "" {
			ctx := context.Background()
			key := fmt.Sprintf("msg:%s", result.MessageID)
			val := fmt.Sprintf("sent at: %s | DB_ID: %s", time.Now().Format(time.RFC3339), msg.ID.String())

			err := s.Redis.Set(ctx, key, val, s.Config.RedisTTL).Err()
			if err != nil {
				slog.Error("redis error", "error", err)
			} else {
				slog.Info("cached msg to Redis", "remote_id", result.MessageID)
			}
		}

	} else {
		slog.Warn("webhook returned non-OK status", "status", resp.StatusCode)
		s.Repo.UpdateStatus(msg.ID, model.StatusFailed)
	}
}
