package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/notification-service/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/notification-service/internal/usecase"
)

type EventConsumer struct {
	createdReader *kafka.Reader
	statusReader  *kafka.Reader
	useCase       usecase.NotificationUseCase
}

func NewEventConsumer(brokers []string, useCase usecase.NotificationUseCase) *EventConsumer {
	createdReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       "payment.created",
		GroupID:     "notification-service-group",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10 * 1024,
		MaxBytes:    10 * 1024 * 1024,
	})

	statusReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       "payment.status_updated",
		GroupID:     "notification-service-group",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10 * 1024,
		MaxBytes:    10 * 1024 * 1024,
	})

	return &EventConsumer{
		createdReader: createdReader,
		statusReader:  statusReader,
		useCase:       useCase,
	}
}

func (c *EventConsumer) Start(ctx context.Context) {
	log.Println("Notification Consumer Service started...")

	go c.consumePaymentCreated(ctx)
	go c.consumePaymentStatusUpdated(ctx)
}

func (c *EventConsumer) consumePaymentCreated(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.createdReader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				continue
			}

			var event domain.PaymentCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err == nil {
				_ = c.useCase.HandlePaymentCreated(ctx, &event)
			}
			_ = c.createdReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *EventConsumer) consumePaymentStatusUpdated(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.statusReader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				continue
			}

			var event domain.PaymentStatusUpdatedEvent
			if err := json.Unmarshal(msg.Value, &event); err == nil {
				_ = c.useCase.HandlePaymentStatusUpdated(ctx, &event)
			}
			_ = c.statusReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *EventConsumer) Close() {
	_ = c.createdReader.Close()
	_ = c.statusReader.Close()
}
