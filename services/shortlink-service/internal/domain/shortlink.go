package domain

import (
	"errors"
	"time"
)

var (
	ErrCodeNotFound = errors.New("shortlink code not found")
	ErrCodeExpired  = errors.New("shortlink code has expired")
	ErrInvalidURL   = errors.New("target url or payment_no is required")
)

const (
	DefaultTTL = 7 * 24 * time.Hour // 7 Days TTL
)

type Shortlink struct {
	Code      string    `json:"code"`
	PaymentNo string    `json:"payment_no"`
	TargetURL string    `json:"target_url"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateShortlinkInput struct {
	PaymentNo string `json:"payment_no"`
	TargetURL string `json:"target_url"`
	CustomCode string `json:"custom_code,omitempty"`
}

type ClickEvent struct {
	EventID   string    `json:"event_id"`
	Code      string    `json:"code"`
	PaymentNo string    `json:"payment_no"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	Referer   string    `json:"referer"`
	Timestamp time.Time `json:"timestamp"`
}
