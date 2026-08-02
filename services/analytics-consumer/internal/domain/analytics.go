package domain

import (
	"time"
)

type ClickAnalytics struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"type:varchar(32);index" json:"code"`
	PaymentNo string    `gorm:"type:varchar(64);index" json:"payment_no"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	Referer   string    `gorm:"type:text" json:"referer"`
	ClickedAt time.Time `gorm:"index" json:"clicked_at"`
}

type PaymentAnalytics struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	PaymentNo     string    `gorm:"type:varchar(64);uniqueIndex" json:"payment_no"`
	Amount        float64   `gorm:"type:numeric(12,2)" json:"amount"`
	Currency      string    `gorm:"type:varchar(3)" json:"currency"`
	Status        string    `gorm:"type:varchar(20);index" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

type ShortlinkClickedEvent struct {
	Code      string    `json:"code"`
	PaymentNo string    `json:"payment_no"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Referer   string    `json:"referer"`
	ClickedAt time.Time `json:"clicked_at"`
}

type PaymentCreatedEvent struct {
	PaymentNo string    `json:"payment_no"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type PaymentStatusUpdatedEvent struct {
	PaymentNo string    `json:"payment_no"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	UpdatedAt time.Time `json:"updated_at"`
}
