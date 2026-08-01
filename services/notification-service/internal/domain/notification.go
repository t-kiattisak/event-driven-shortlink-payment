package domain

import (
	"time"
)

type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "EMAIL"
	ChannelSMS   NotificationChannel = "SMS"
	ChannelLINE  NotificationChannel = "LINE"
)

type Notification struct {
	ID        string              `json:"id"`
	PaymentNo string              `json:"payment_no"`
	Recipient string              `json:"recipient"`
	Channel   NotificationChannel `json:"channel"`
	Message   string              `json:"message"`
	SentAt    time.Time           `json:"sent_at"`
}

type PaymentStatusUpdatedEvent struct {
	PaymentNo string    `json:"payment_no"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PaymentCreatedEvent struct {
	PaymentNo string    `json:"payment_no"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
