package domain

import (
	"errors"
	"time"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidAmount   = errors.New("amount must be greater than zero")
)

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "PENDING"
	StatusPaid      PaymentStatus = "PAID"
	StatusExpired   PaymentStatus = "EXPIRED"
	StatusCancelled PaymentStatus = "CANCELLED"
)

type Payment struct {
	ID         uint          `gorm:"primaryKey" json:"id"`
	PaymentNo  string        `gorm:"type:varchar(64);uniqueIndex;not null" json:"payment_no"`
	ShortCode  string        `gorm:"type:varchar(32);uniqueIndex;not null" json:"short_code"`
	Amount     float64       `gorm:"type:numeric(12,2);not null" json:"amount"`
	Currency   string        `gorm:"type:varchar(3);default:'THB'" json:"currency"`
	Status     PaymentStatus `gorm:"type:varchar(32);index;not null;default:'PENDING'" json:"status"`
	QRCodeData string        `gorm:"type:text" json:"qr_code_data"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type CreatePaymentInput struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}
