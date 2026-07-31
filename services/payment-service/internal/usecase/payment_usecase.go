package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/repository"
)

type PaymentUseCase interface {
	CreatePayment(ctx context.Context, input domain.CreatePaymentInput) (*domain.Payment, error)
	GetPaymentByNo(ctx context.Context, paymentNo string) (*domain.Payment, error)
	GetPaymentByShortCode(ctx context.Context, shortCode string) (*domain.Payment, error)
	UpdatePaymentStatus(ctx context.Context, paymentNo string, status domain.PaymentStatus) error
}

type paymentUseCase struct {
	paymentRepo repository.PaymentRepository
}

func NewPaymentUseCase(paymentRepo repository.PaymentRepository) PaymentUseCase {
	return &paymentUseCase{paymentRepo: paymentRepo}
}

func (u *paymentUseCase) CreatePayment(ctx context.Context, input domain.CreatePaymentInput) (*domain.Payment, error) {
	if input.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if input.Currency == "" {
		input.Currency = "THB"
	}

	paymentNo := fmt.Sprintf("PAY-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8])
	if input.ShortCode == "" {
		input.ShortCode = uuid.New().String()[:6]
	}

	// Generate QR Code data (Base64 PNG)
	qrPayload := fmt.Sprintf("PAYMENT|%s|%.2f|%s", paymentNo, input.Amount, input.Currency)
	pngBytes, err := qrcode.Encode(qrPayload, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate qr code: %w", err)
	}
	qrBase64 := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(pngBytes))

	payment := &domain.Payment{
		PaymentNo:  paymentNo,
		ShortCode:  input.ShortCode,
		Amount:     input.Amount,
		Currency:   input.Currency,
		Status:     domain.StatusPending,
		QRCodeData: qrBase64,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Prepare Outbox Event
	eventPayload, _ := json.Marshal(map[string]interface{}{
		"payment_no": paymentNo,
		"short_code": input.ShortCode,
		"amount":     input.Amount,
		"currency":   input.Currency,
		"status":     payment.Status,
		"created_at": payment.CreatedAt,
	})

	outbox := &domain.Outbox{
		ID:            uuid.New().String(),
		AggregateType: "Payment",
		AggregateID:   paymentNo,
		Topic:         "payment.created",
		Payload:       string(eventPayload),
		Status:        domain.OutboxPending,
		CreatedAt:     time.Now(),
	}

	if err := u.paymentRepo.CreateWithOutbox(ctx, payment, outbox); err != nil {
		return nil, err
	}

	return payment, nil
}

func (u *paymentUseCase) GetPaymentByNo(ctx context.Context, paymentNo string) (*domain.Payment, error) {
	return u.paymentRepo.FindByPaymentNo(ctx, paymentNo)
}

func (u *paymentUseCase) GetPaymentByShortCode(ctx context.Context, shortCode string) (*domain.Payment, error) {
	return u.paymentRepo.FindByShortCode(ctx, shortCode)
}

func (u *paymentUseCase) UpdatePaymentStatus(ctx context.Context, paymentNo string, status domain.PaymentStatus) error {
	payment, err := u.paymentRepo.FindByPaymentNo(ctx, paymentNo)
	if err != nil {
		return err
	}

	eventPayload, _ := json.Marshal(map[string]interface{}{
		"payment_no": paymentNo,
		"old_status": payment.Status,
		"new_status": status,
		"updated_at": time.Now(),
	})

	outbox := &domain.Outbox{
		ID:            uuid.New().String(),
		AggregateType: "Payment",
		AggregateID:   paymentNo,
		Topic:         "payment.status_updated",
		Payload:       string(eventPayload),
		Status:        domain.OutboxPending,
		CreatedAt:     time.Now(),
	}

	return u.paymentRepo.UpdateStatus(ctx, paymentNo, status, outbox)
}
