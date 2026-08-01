package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/domain"
)

type eventProducer struct {
	createdWriter *kafka.Writer
	clickedWriter *kafka.Writer
}

func NewEventProducer(brokers []string) EventProducer {
	return &eventProducer{
		createdWriter: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  "shortlink.created",
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
		clickedWriter: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  "shortlink.clicked",
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *eventProducer) PublishShortlinkCreated(ctx context.Context, shortlink *domain.Shortlink) error {
	payload, err := json.Marshal(shortlink)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(shortlink.Code),
		Value: payload,
	}
	return p.createdWriter.WriteMessages(ctx, msg)
}

func (p *eventProducer) PublishShortlinkClicked(ctx context.Context, event *domain.ClickEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.Code),
		Value: payload,
	}

	// Fire-and-forget in background goroutine to prevent blocking HTTP handler
	go func() {
		if err := p.clickedWriter.WriteMessages(context.Background(), msg); err != nil {
			log.Printf("[EventProducer] Failed to produce shortlink.clicked event: %v", err)
		}
	}()

	return nil
}

func (p *eventProducer) Close() error {
	_ = p.createdWriter.Close()
	_ = p.clickedWriter.Close()
	return nil
}
