package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/audit-service/internal/delivery/kafka"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/audit-service/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/audit-service/internal/repository"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/audit-service/internal/usecase"
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
		log.Fatalf("Failed to connect to Audit Database: %v", err)
	}

	// Auto-migrate audit_logs table
	if err := db.AutoMigrate(&domain.AuditLog{}); err != nil {
		log.Fatalf("Failed to run Audit DB migration: %v", err)
	}
	log.Println("Audit DB migration completed successfully.")

	// Clean Architecture Layer Initialization
	repo := repository.NewAuditRepository(db)
	useCase := usecase.NewAuditUseCase(repo)

	topics := []string{"payment.created", "payment.status_updated", "shortlink.created", "shortlink.clicked"}
	consumer := kafka.NewAuditConsumer(kafkaBrokers, topics, useCase)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.Start(ctx)
	log.Println("audit-service running and recording immutable compliance logs...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down audit-service...")
	consumer.Close()
}
