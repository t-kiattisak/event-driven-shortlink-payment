package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/notification-service/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/notification-service/internal/usecase"
)

type EventConsumer struct {
	createdReader *kafka.Reader
	statusReader  *kafka.Reader
	dlqPublisher  *DLQPublisher
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

	dlqPublisher := NewDLQPublisher(brokers, "payment.notification.dlq")

	return &EventConsumer{
		createdReader: createdReader,
		statusReader:  statusReader,
		dlqPublisher:  dlqPublisher,
		useCase:       useCase,
	}
}

func (c *EventConsumer) Start(ctx context.Context) {
	log.Println("Notification Consumer Service started with DLQ & Retry...")

	go c.consumePaymentCreated(ctx)
	go c.consumePaymentStatusUpdated(ctx)
}

func (c *EventConsumer) executeWithRetry(ctx context.Context, maxRetries int, fn func() error) error {
	var err error
	backoff := 500 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		log.Printf("[Consumer Retry] Attempt %d/%d failed: %v. Retrying in %v...", attempt, maxRetries, err, backoff)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		backoff *= 2 // Exponential backoff: 500ms, 1s, 2s...
	}
	return err
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
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[Notification Consumer] Malformed JSON in payment.created: %v. Routing to DLQ.", err)
				_ = c.dlqPublisher.PublishDLQ(ctx, "payment.created", msg.Key, msg.Value, err)
				_ = c.createdReader.CommitMessages(ctx, msg)
				continue
			}

			// Process with exponential retry (up to 3 times)
			err = c.executeWithRetry(ctx, 3, func() error {
				return c.useCase.HandlePaymentCreated(ctx, &event)
			})

			if err != nil {
				log.Printf("[Notification Consumer] Failed to process payment.created after retries: %v. Sending to DLQ.", err)
				_ = c.dlqPublisher.PublishDLQ(ctx, "payment.created", msg.Key, msg.Value, err)
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
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[Notification Consumer] Malformed JSON in payment.status_updated: %v. Routing to DLQ.", err)
				_ = c.dlqPublisher.PublishDLQ(ctx, "payment.status_updated", msg.Key, msg.Value, err)
				_ = c.statusReader.CommitMessages(ctx, msg)
				continue
			}

			// Process with exponential retry (up to 3 times)
			err = c.executeWithRetry(ctx, 3, func() error {
				return c.useCase.HandlePaymentStatusUpdated(ctx, &event)
			})

			if err != nil {
				log.Printf("[Notification Consumer] Failed to process payment.status_updated after retries: %v. Sending to DLQ.", err)
				_ = c.dlqPublisher.PublishDLQ(ctx, "payment.status_updated", msg.Key, msg.Value, err)
			}

			_ = c.statusReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *EventConsumer) Close() {
	_ = c.createdReader.Close()
	_ = c.statusReader.Close()
	if c.dlqPublisher != nil {
		_ = c.dlqPublisher.Close()
	}
}
