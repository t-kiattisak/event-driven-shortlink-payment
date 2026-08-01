package repository

import (
	"context"
	"log"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/notification-service/internal/domain"
)

type consoleNotificationSender struct{}

func NewConsoleNotificationSender() NotificationSender {
	return &consoleNotificationSender{}
}

func (s *consoleNotificationSender) Send(ctx context.Context, n *domain.Notification) error {
	log.Printf("[NOTIFICATION DISPATCHED] Channel: %s | Recipient: %s | Invoice: %s | Message: %s",
		n.Channel, n.Recipient, n.PaymentNo, n.Message)
	return nil
}
