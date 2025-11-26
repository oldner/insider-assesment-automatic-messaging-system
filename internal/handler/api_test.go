package handler_test

import (
	"bytes"
	"encoding/json"
	"insider-assessment/internal/config"
	"insider-assessment/internal/handler"
	"insider-assessment/internal/model"
	"insider-assessment/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository duplication for handler tests
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

func setupRouter() (*gin.Engine, *handler.Handler, *MockRepository) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockRepository)

	// Setup a real scheduler with mocks to avoid nil pointers,
	// though we might not assert on scheduler behavior deeply here.
	cfg := &config.Config{WorkerInterval: time.Minute}
	workerSvc := service.NewWorkerService(mockRepo, nil, cfg)
	scheduler := service.NewScheduler(workerSvc, cfg)

	h := handler.NewHandler(scheduler, mockRepo)

	r := gin.Default()
	r.POST("/start", h.StartScheduler)
	r.POST("/stop", h.StopScheduler)
	r.GET("/sent-messages", h.GetSentMessages)
	r.POST("/messages", h.AddMessage)
	r.GET("/health", h.HealthCheck)

	return r, h, mockRepo
}

func TestHandler_HealthCheck(t *testing.T) {
	r, _, _ := setupRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status": "ok"}`, w.Body.String())
}

func TestHandler_GetSentMessages(t *testing.T) {
	r, _, mockRepo := setupRouter()

	messages := []model.Message{
		{To: "+123", Content: "Test", Status: model.StatusSent},
	}
	mockRepo.On("GetAllSent").Return(messages, nil)

	req, _ := http.NewRequest("GET", "/sent-messages", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.Message
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Len(t, response, 1)
	assert.Equal(t, "+123", response[0].To)
}

func TestHandler_AddMessage(t *testing.T) {
	r, _, mockRepo := setupRouter()

	msg := model.Message{To: "+123", Content: "New Msg"}
	mockRepo.On("Create", mock.AnythingOfType("*model.Message")).Return(nil)

	body, _ := json.Marshal(msg)
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestHandler_StartStopScheduler(t *testing.T) {
	r, h, _ := setupRouter()

	// Test Start
	req, _ := http.NewRequest("POST", "/start", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Cleanup: Stop the scheduler to avoid leaking goroutines
	h.Scheduler.Stop()

	// Test Stop
	reqStop, _ := http.NewRequest("POST", "/stop", nil)
	wStop := httptest.NewRecorder()
	r.ServeHTTP(wStop, reqStop)
	assert.Equal(t, http.StatusOK, wStop.Code)
}
