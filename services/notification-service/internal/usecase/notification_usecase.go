package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/notification-service/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/notification-service/internal/repository"
)

type NotificationUseCase interface {
	HandlePaymentCreated(ctx context.Context, event *domain.PaymentCreatedEvent) error
	HandlePaymentStatusUpdated(ctx context.Context, event *domain.PaymentStatusUpdatedEvent) error
}

type notificationUseCase struct {
	sender         repository.NotificationSender
	processedKeys  map[string]bool
	processedMutex sync.Mutex
}

func NewNotificationUseCase(sender repository.NotificationSender) NotificationUseCase {
	return &notificationUseCase{
		sender:        sender,
		processedKeys: make(map[string]bool),
	}
}

// IsDuplicate Checks idempotency key (paymentNo + status) to prevent duplicate notifications
func (u *notificationUseCase) isDuplicate(idempotencyKey string) bool {
	u.processedMutex.Lock()
	defer u.processedMutex.Unlock()

	if u.processedKeys[idempotencyKey] {
		return true
	}
	u.processedKeys[idempotencyKey] = true
	return false
}

func (u *notificationUseCase) HandlePaymentCreated(ctx context.Context, event *domain.PaymentCreatedEvent) error {
	idempotencyKey := fmt.Sprintf("created:%s", event.PaymentNo)
	if u.isDuplicate(idempotencyKey) {
		return nil
	}

	notification := &domain.Notification{
		ID:        idempotencyKey,
		PaymentNo: event.PaymentNo,
		Recipient: "customer@example.com",
		Channel:   domain.ChannelEmail,
		Message:   fmt.Sprintf("Invoice %s created for amount %.2f %s. Please complete payment.", event.PaymentNo, event.Amount, event.Currency),
		SentAt:    time.Now(),
	}

	return u.sender.Send(ctx, notification)
}

func (u *notificationUseCase) HandlePaymentStatusUpdated(ctx context.Context, event *domain.PaymentStatusUpdatedEvent) error {
	idempotencyKey := fmt.Sprintf("status:%s:%s", event.PaymentNo, event.NewStatus)
	if u.isDuplicate(idempotencyKey) {
		return nil
	}

	var message string
	var channel domain.NotificationChannel = domain.ChannelEmail

	switch event.NewStatus {
	case "PAID":
		message = fmt.Sprintf("SUCCESS: Payment for invoice %s has been PAID successfully!", event.PaymentNo)
		channel = domain.ChannelSMS
	case "EXPIRED":
		message = fmt.Sprintf("EXPIRED: Payment link for invoice %s has expired.", event.PaymentNo)
		channel = domain.ChannelEmail
	default:
		message = fmt.Sprintf("Payment status for invoice %s updated to %s", event.PaymentNo, event.NewStatus)
	}

	notification := &domain.Notification{
		ID:        idempotencyKey,
		PaymentNo: event.PaymentNo,
		Recipient: "customer@example.com",
		Channel:   channel,
		Message:   message,
		SentAt:    time.Now(),
	}

	return u.sender.Send(ctx, notification)
}
