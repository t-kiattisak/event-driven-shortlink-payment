package domain

import (
	"time"
)

type OutboxStatus string

const (
	OutboxPending   OutboxStatus = "PENDING"
	OutboxProcessed OutboxStatus = "PROCESSED"
	OutboxFailed    OutboxStatus = "FAILED"
)

type Outbox struct {
	ID            string       `gorm:"type:uuid;primaryKey" json:"id"`
	AggregateType string       `gorm:"type:varchar(64);not null" json:"aggregate_type"`
	AggregateID   string       `gorm:"type:varchar(64);not null" json:"aggregate_id"`
	Topic         string       `gorm:"type:varchar(128);not null" json:"topic"`
	Payload       string       `gorm:"type:jsonb;not null" json:"payload"`
	Status        OutboxStatus `gorm:"type:varchar(32);default:'PENDING'" json:"status"`
	RetryCount    int          `gorm:"default:0" json:"retry_count"`
	LastError     string       `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
}
