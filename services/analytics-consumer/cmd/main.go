package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/analytics-consumer/internal/delivery/kafka"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/analytics-consumer/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/analytics-consumer/internal/repository"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/analytics-consumer/internal/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func main() {
	dbHost := getEnv("POSTGRES_HOST", "localhost")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "postgres")
	dbPass := getEnv("POSTGRES_PASSWORD", "postgrespassword")
	dbName := getEnv("POSTGRES_DB", "audit_db")
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Analytics Database: %v", err)
	}

	// Auto-migrate analytics tables
	if err := db.AutoMigrate(&domain.ClickAnalytics{}, &domain.PaymentAnalytics{}); err != nil {
		log.Fatalf("Failed to run Analytics DB migration: %v", err)
	}
	log.Println("Analytics DB migration completed successfully.")

	// Clean Architecture Layer Initialization
	repo := repository.NewAnalyticsRepository(db)
	useCase := usecase.NewAnalyticsUseCase(repo)
	consumer := kafka.NewAnalyticsConsumer(kafkaBrokers, useCase)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.Start(ctx)
	log.Println("analytics-consumer running and processing events...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down analytics-consumer...")
	consumer.Close()
}
