package handler

import (
	"insider-assessment/internal/model"
	"insider-assessment/internal/repository"
	"insider-assessment/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Scheduler *service.Scheduler
	Repo      repository.MessageRepository
}

func NewHandler(scheduler *service.Scheduler, repo repository.MessageRepository) *Handler {
	return &Handler{Scheduler: scheduler, Repo: repo}
}

// StartScheduler godoc
// @Summary Start the automatic message sender
// @Description Resumes the background ticker that sends messages every 2 minutes.
// @Tags Control
// @Produce json
// @Success 200 {object} map[string]string
// @Router /start [post]
func (h *Handler) StartScheduler(c *gin.Context) {
	h.Scheduler.Start()
	c.JSON(http.StatusOK, gin.H{"message": "Automatic message sending started"})
}

// StopScheduler godoc
// @Summary Stop the automatic message sender
// @Description Pauses the background ticker.
// @Tags Control
// @Produce json
// @Success 200 {object} map[string]string
// @Router /stop [post]
func (h *Handler) StopScheduler(c *gin.Context) {
	h.Scheduler.Stop()
	c.JSON(http.StatusOK, gin.H{"message": "Automatic message sending stopped"})
}

// GetSentMessages godoc
// @Summary Get list of sent messages
// @Tags Messages
// @Produce json
// @Success 200 {array} model.Message
// @Router /sent-messages [get]
func (h *Handler) GetSentMessages(c *gin.Context) {
	msgs, err := h.Repo.GetAllSent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

type CreateMessageRequest struct {
	To      string `json:"to" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// AddMessage godoc
// @Summary Add a new message (Test Helper)
// @Tags Messages
// @Accept json
// @Produce json
// @Param message body CreateMessageRequest true "Message Content"
// @Success 201 {object} model.Message
// @Router /messages [post]
func (h *Handler) AddMessage(c *gin.Context) {
	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := model.Message{
		To:      req.To,
		Content: req.Content,
		Status:  model.StatusPending,
	}

	if err := h.Repo.Create(&msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, msg)
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Returns 200 OK if the server is running
// @Tags System
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetAllCachedMessages godoc
// @Summary Get all cached messages
// @Description Retrieves all messages currently stored in Redis cache.
// @Tags Messages
// @Produce json
// @Success 200 {object} map[string]string
// @Router /messages/cache [get]
func (h *Handler) GetAllCachedMessages(c *gin.Context) {
	if h.Scheduler == nil || h.Scheduler.Sender == nil || h.Scheduler.Sender.Redis == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis not available"})
		return
	}

	keys, err := h.Scheduler.Sender.Redis.Keys(c.Request.Context(), "msg:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan cache"})
		return
	}

	results := make(map[string]string)
	for _, key := range keys {
		val, _ := h.Scheduler.Sender.Redis.Get(c.Request.Context(), key).Result()
		results[key] = val
	}

	c.JSON(http.StatusOK, results)
}
