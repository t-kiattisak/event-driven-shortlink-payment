package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/analytics-consumer/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/analytics-consumer/internal/usecase"
)

type AnalyticsConsumer struct {
	clickReader   *kafka.Reader
	createdReader *kafka.Reader
	statusReader  *kafka.Reader
	useCase       usecase.AnalyticsUseCase
}

func NewAnalyticsConsumer(brokers []string, useCase usecase.AnalyticsUseCase) *AnalyticsConsumer {
	clickReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       "shortlink.clicked",
		GroupID:     "analytics-consumer-group",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10 * 1024,
		MaxBytes:    10 * 1024 * 1024,
	})

	createdReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       "payment.created",
		GroupID:     "analytics-consumer-group",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10 * 1024,
		MaxBytes:    10 * 1024 * 1024,
	})

	statusReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       "payment.status_updated",
		GroupID:     "analytics-consumer-group",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10 * 1024,
		MaxBytes:    10 * 1024 * 1024,
	})

	return &AnalyticsConsumer{
		clickReader:   clickReader,
		createdReader: createdReader,
		statusReader:  statusReader,
		useCase:       useCase,
	}
}

func (c *AnalyticsConsumer) Start(ctx context.Context) {
	log.Println("Analytics Consumer Service started listening to shortlink.clicked, payment.created, payment.status_updated...")

	go c.consumeClicks(ctx)
	go c.consumePaymentCreated(ctx)
	go c.consumePaymentStatusUpdated(ctx)
}

func (c *AnalyticsConsumer) consumeClicks(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.clickReader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				continue
			}

			var event domain.ShortlinkClickedEvent
			if err := json.Unmarshal(msg.Value, &event); err == nil {
				if err := c.useCase.ProcessClickEvent(ctx, &event); err == nil {
					log.Printf("[ANALYTICS] Recorded click for code: %s (payment: %s)", event.Code, event.PaymentNo)
				}
			}
			_ = c.clickReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *AnalyticsConsumer) consumePaymentCreated(ctx context.Context) {
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
				if err := c.useCase.ProcessPaymentCreated(ctx, &event); err == nil {
					log.Printf("[ANALYTICS] Upserted payment created: %s (amount: %.2f)", event.PaymentNo, event.Amount)
				}
			}
			_ = c.createdReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *AnalyticsConsumer) consumePaymentStatusUpdated(ctx context.Context) {
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
				if err := c.useCase.ProcessPaymentStatusUpdated(ctx, &event); err == nil {
					log.Printf("[ANALYTICS] Updated payment status: %s -> %s", event.PaymentNo, event.NewStatus)
				}
			}
			_ = c.statusReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *AnalyticsConsumer) Close() {
	_ = c.clickReader.Close()
	_ = c.createdReader.Close()
	_ = c.statusReader.Close()
}
