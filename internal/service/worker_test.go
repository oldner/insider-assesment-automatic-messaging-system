package service_test

import (
	"encoding/json"
	"insider-assessment/internal/config"
	"insider-assessment/internal/model"
	"insider-assessment/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of repository.MessageRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetPending(limit int) ([]model.Message, error) {
	args := m.Called(limit)
	return args.Get(0).([]model.Message), args.Error(1)
}

func (m *MockRepository) UpdateStatus(id uuid.UUID, status model.MessageStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockRepository) GetAllSent() ([]model.Message, error) {
	args := m.Called()
	return args.Get(0).([]model.Message), args.Error(1)
}

func (m *MockRepository) Create(msg *model.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func TestWorkerService_ProcessMessages_Success(t *testing.T) {
	// 1. Setup Mock Webhook Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/webhook", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload map[string]string
		json.NewDecoder(r.Body).Decode(&payload)
		assert.Contains(t, payload, "to")
		assert.Contains(t, payload, "content")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message":   "queued",
			"messageId": "external-123",
		})
	}))
	defer server.Close()

	// 2. Setup Mock Repository
	mockRepo := new(MockRepository)
	msgID := uuid.New()
	messages := []model.Message{
		{
			ID:      msgID,
			To:      "+1234567890",
			Content: "Hello World",
			Status:  model.StatusPending,
		},
	}

	mockRepo.On("GetPending", 2).Return(messages, nil)
	mockRepo.On("UpdateStatus", msgID, model.StatusSent).Return(nil)

	// 3. Setup Service
	cfg := &config.Config{
		WebhookUrl:      server.URL + "/webhook",
		WorkerBatchSize: 2,
		RedisTTL:        time.Hour,
	}

	// Passing nil for Redis to skip cache logic
	svc := service.NewWorkerService(mockRepo, nil, cfg)

	// 4. Execute
	// ProcessMessages runs goroutines, so we might need to wait a tiny bit or use a WaitGroup if we could inject it.
	// However, ProcessMessages launches goroutines and returns. The goroutines run independently.
	// Since we can't easily wait for the internal goroutines of ProcessMessages without modifying the code,
	// we will rely on a small sleep or refactor.
	// For this test, let's try a small sleep which is flaky but simple for now,
	// or better: refactor the service to accept a WaitGroup or return a channel?
	// Given the constraints, let's assume the operations are fast enough or use a small sleep.

	svc.ProcessMessages()

	// Wait for goroutines to finish
	time.Sleep(100 * time.Millisecond)

	// 5. Verify
	mockRepo.AssertExpectations(t)
}

func TestWorkerService_ProcessMessages_Failure(t *testing.T) {
	// 1. Setup Mock Webhook Server (Returns 500)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// 2. Setup Mock Repository
	mockRepo := new(MockRepository)
	msgID := uuid.New()
	messages := []model.Message{
		{
			ID:      msgID,
			To:      "+1234567890",
			Content: "Fail Me",
			Status:  model.StatusPending,
		},
	}

	mockRepo.On("GetPending", 2).Return(messages, nil)
	mockRepo.On("UpdateStatus", msgID, model.StatusFailed).Return(nil)

	// 3. Setup Service
	cfg := &config.Config{
		WebhookUrl:      server.URL,
		WorkerBatchSize: 2,
	}

	svc := service.NewWorkerService(mockRepo, nil, cfg)

	// 4. Execute
	svc.ProcessMessages()
	time.Sleep(100 * time.Millisecond)

	// 5. Verify
	mockRepo.AssertExpectations(t)
}
