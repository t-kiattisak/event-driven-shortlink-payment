package repository

import (
	"context"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/notification-service/internal/domain"
)

type NotificationSender interface {
	Send(ctx context.Context, notification *domain.Notification) error
}
