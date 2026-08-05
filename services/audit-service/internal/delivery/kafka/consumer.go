package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/audit-service/internal/usecase"
)

type AuditConsumer struct {
	readers []*kafka.Reader
	useCase usecase.AuditUseCase
}

func NewAuditConsumer(brokers []string, topics []string, useCase usecase.AuditUseCase) *AuditConsumer {
	var readers []*kafka.Reader
	for _, topic := range topics {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			Topic:       topic,
			GroupID:     "audit-service-group",
			StartOffset: kafka.FirstOffset,
			MinBytes:    10 * 1024,
			MaxBytes:    10 * 1024 * 1024,
		})
		readers = append(readers, r)
	}

	return &AuditConsumer{
		readers: readers,
		useCase: useCase,
	}
}

func (c *AuditConsumer) Start(ctx context.Context) {
	log.Println("Audit Consumer Service started listening to all compliance topics...")
	for _, reader := range c.readers {
		go c.consumeTopic(ctx, reader)
	}
}

func (c *AuditConsumer) consumeTopic(ctx context.Context, reader *kafka.Reader) {
	topic := reader.Config().Topic
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				continue
			}

			key := string(msg.Key)
			if key == "" {
				key = "auto-assigned"
			}

			if err := c.useCase.RecordEvent(ctx, topic, key, string(msg.Value)); err == nil {
				log.Printf("[AUDIT LOGGED] Topic: %s | Key: %s", topic, key)
			}
			_ = reader.CommitMessages(ctx, msg)
		}
	}
}

func (c *AuditConsumer) Close() {
	for _, reader := range c.readers {
		_ = reader.Close()
	}
}
