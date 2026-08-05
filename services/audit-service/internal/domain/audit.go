package domain

import (
	"time"
)

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EventID   string    `gorm:"type:varchar(64);index" json:"event_id"`
	EventType string    `gorm:"type:varchar(64);index" json:"event_type"`
	Actor     string    `gorm:"type:varchar(64)" json:"actor"`
	Payload   string    `gorm:"type:jsonb" json:"payload"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
