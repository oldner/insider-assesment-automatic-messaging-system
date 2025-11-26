package repository

import (
	"insider-assessment/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageStatus string

const (
	StatusPending MessageStatus = "PENDING"
	StatusSent    MessageStatus = "SENT"
	StatusFailed  MessageStatus = "FAILED"
)

type MessageRepository interface {
	GetPending(limit int) ([]model.Message, error)
	UpdateStatus(id uuid.UUID, status model.MessageStatus) error
	GetAllSent() ([]model.Message, error)
	Create(msg *model.Message) error
}

type messageRepository struct {
	DB *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{DB: db}
}

func (r *messageRepository) Create(msg *model.Message) error {
	return r.DB.Create(msg).Error
}

func (r *messageRepository) UpdateStatus(id uuid.UUID, status model.MessageStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == model.StatusSent {
		updates["sent_at"] = gorm.Expr("NOW()")
	}

	return r.DB.Model(&model.Message{}).Where("id = ?", id).Updates(updates).Error
}

func (r *messageRepository) GetPending(limit int) ([]model.Message, error) {
	var messages []model.Message
	result := r.DB.Where("status = ?", model.StatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages)

	return messages, result.Error
}

func (r *messageRepository) GetAllSent() ([]model.Message, error) {
	var messages []model.Message
	result := r.DB.Where("status = ?", model.StatusSent).Find(&messages)
	return messages, result.Error
}
