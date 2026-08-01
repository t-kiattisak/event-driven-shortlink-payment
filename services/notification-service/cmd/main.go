package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/notification-service/internal/delivery/kafka"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/notification-service/internal/repository"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/notification-service/internal/usecase"
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func main() {
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}

	// Clean Architecture Layer Initialization
	sender := repository.NewConsoleNotificationSender()
	useCase := usecase.NewNotificationUseCase(sender)
	consumer := kafka.NewEventConsumer(kafkaBrokers, useCase)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.Start(ctx)
	log.Println("notification-service running and listening for Kafka events...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down notification-service...")
	consumer.Close()
}
