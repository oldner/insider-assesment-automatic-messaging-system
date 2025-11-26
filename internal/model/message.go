package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageStatus string

const (
	StatusPending MessageStatus = "PENDING"
	StatusSent    MessageStatus = "SENT"
	StatusFailed  MessageStatus = "FAILED"
)

type Message struct {
	ID        uuid.UUID     `gorm:"primaryKey;type:uuid;" json:"id"`
	To        string        `gorm:"not null" json:"to"`
	Content   string        `gorm:"not null" json:"content"`
	Status    MessageStatus `gorm:"default:'PENDING';index" json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	SentAt    *time.Time    `json:"sent_at,omitempty"`
}

// BeforeCreate generates a new UUID if not present
func (m *Message) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// BeforeSave is a GORM hook to validate character limit
func (m *Message) BeforeSave(tx *gorm.DB) (err error) {
	if len(m.Content) > 160 {
		return errors.New("message content exceeds 160 characters")
	}
	return nil
}
