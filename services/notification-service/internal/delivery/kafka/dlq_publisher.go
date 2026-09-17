package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type DLQPublisher struct {
	writer *kafka.Writer
}

func NewDLQPublisher(brokers []string, topic string) *DLQPublisher {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	return &DLQPublisher{writer: writer}
}

func (p *DLQPublisher) PublishDLQ(ctx context.Context, originalTopic string, key, value []byte, failureErr error) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "dlq-original-topic", Value: []byte(originalTopic)},
			{Key: "dlq-error", Value: []byte(failureErr.Error())},
			{Key: "dlq-failed-at", Value: []byte(time.Now().Format(time.RFC3339))},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to send message to DLQ: %w", err)
	}

	log.Printf("[DLQ] Successfully forwarded failed event to DLQ topic %s (Error: %v)", p.writer.Topic, failureErr)
	return nil
}

func (p *DLQPublisher) Close() error {
	return p.writer.Close()
}
