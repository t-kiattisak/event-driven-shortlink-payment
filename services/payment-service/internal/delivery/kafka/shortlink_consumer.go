package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/usecase"
)

type ShortlinkCreatedConsumer struct {
	reader         *kafka.Reader
	paymentUseCase usecase.PaymentUseCase
}

func NewShortlinkCreatedConsumer(brokers []string, paymentUseCase usecase.PaymentUseCase) *ShortlinkCreatedConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "shortlink.created",
		GroupID:  "payment-service-group",
		MinBytes: 10 * 1024,
		MaxBytes: 10 * 1024 * 1024,
	})

	return &ShortlinkCreatedConsumer{
		reader:         reader,
		paymentUseCase: paymentUseCase,
	}
}

func (c *ShortlinkCreatedConsumer) Start(ctx context.Context) {
	log.Println("[ShortlinkCreatedConsumer] Started listening for shortlink.created events...")

	for {
		select {
		case <-ctx.Done():
			c.Close()
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[ShortlinkCreatedConsumer] Error fetching message: %v", err)
				continue
			}

			var payload struct {
				Code      string `json:"code"`
				PaymentNo string `json:"payment_no"`
			}

			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				log.Printf("[ShortlinkCreatedConsumer] Invalid payload format: %v", err)
				_ = c.reader.CommitMessages(ctx, msg)
				continue
			}

			if payload.PaymentNo != "" && payload.Code != "" {
				if err := c.paymentUseCase.AssociateShortCode(ctx, payload.PaymentNo, payload.Code); err != nil {
					log.Printf("[ShortlinkCreatedConsumer] Failed to associate shortcode %s with payment %s: %v", payload.Code, payload.PaymentNo, err)
				} else {
					log.Printf("[ShortlinkCreatedConsumer] Successfully associated shortcode %s with payment %s", payload.Code, payload.PaymentNo)
				}
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("[ShortlinkCreatedConsumer] Failed to commit message offset: %v", err)
			}
		}
	}
}

func (c *ShortlinkCreatedConsumer) Close() {
	_ = c.reader.Close()
}
