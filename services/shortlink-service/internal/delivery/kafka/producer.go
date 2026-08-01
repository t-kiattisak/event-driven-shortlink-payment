package kafka

import (
	"context"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/domain"
)

type EventProducer interface {
	PublishShortlinkCreated(ctx context.Context, shortlink *domain.Shortlink) error
	PublishShortlinkClicked(ctx context.Context, event *domain.ClickEvent) error
	Close() error
}
