package usecase

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/repository"
)

type OutboxWorker struct {
	outboxRepo   repository.OutboxRepository
	kafkaBrokers []string
	writers      map[string]*kafka.Writer
}

func NewOutboxWorker(outboxRepo repository.OutboxRepository, kafkaBrokers []string) *OutboxWorker {
	return &OutboxWorker{
		outboxRepo:   outboxRepo,
		kafkaBrokers: kafkaBrokers,
		writers:      make(map[string]*kafka.Writer),
	}
}

func (w *OutboxWorker) getWriter(topic string) *kafka.Writer {
	if writer, exists := w.writers[topic]; exists {
		return writer
	}
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(w.kafkaBrokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	w.writers[topic] = writer
	return writer
}

func (w *OutboxWorker) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.Close()
			return
		case <-ticker.C:
			w.processPendingEvents(ctx)
		}
	}
}

func (w *OutboxWorker) processPendingEvents(ctx context.Context) {
	records, err := w.outboxRepo.FetchPending(ctx, 20)
	if err != nil || len(records) == 0 {
		return
	}

	for _, record := range records {
		writer := w.getWriter(record.Topic)
		msg := kafka.Message{
			Key:   []byte(record.AggregateID),
			Value: []byte(record.Payload),
			Time:  record.CreatedAt,
		}

		if err := writer.WriteMessages(ctx, msg); err != nil {
			log.Printf("[OutboxWorker] Failed to publish message %s to topic %s: %v", record.ID, record.Topic, err)
			_ = w.outboxRepo.MarkFailed(ctx, record.ID)
			continue
		}

		if err := w.outboxRepo.MarkProcessed(ctx, record.ID); err != nil {
			log.Printf("[OutboxWorker] Failed to mark outbox record %s as PROCESSED: %v", record.ID, err)
		} else {
			log.Printf("[OutboxWorker] Successfully published event %s to topic %s", record.ID, record.Topic)
		}
	}
}

func (w *OutboxWorker) Close() {
	for _, writer := range w.writers {
		_ = writer.Close()
	}
}
