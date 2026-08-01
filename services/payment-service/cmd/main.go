package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/delivery/http"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/repository"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/usecase"
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
	// 1. Environment & DB Setup
	dbHost := getEnv("POSTGRES_HOST", "localhost")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "postgres")
	dbPass := getEnv("POSTGRES_PASSWORD", "postgrespassword")
	dbName := getEnv("POSTGRES_DB", "payment_db")
	shortlinkSvcURL := getEnv("SHORTLINK_SERVICE_URL", "http://localhost:8082")
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// 2. Auto Migration
	if err := db.AutoMigrate(&domain.Payment{}, &domain.Outbox{}); err != nil {
		log.Fatalf("Failed to run GORM auto-migration: %v", err)
	}
	log.Println("Database migration completed successfully.")

	// 3. Layer Initialization (Clean Architecture Dependency Injection)
	paymentRepo := repository.NewPaymentRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)
	shortlinkClient := repository.NewShortlinkClient(shortlinkSvcURL)
	paymentUseCase := usecase.NewPaymentUseCase(paymentRepo, shortlinkClient)

	// 4. Start Outbox Worker Background Goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outboxWorker := usecase.NewOutboxWorker(outboxRepo, kafkaBrokers)
	go outboxWorker.Start(ctx, 3*time.Second)
	log.Println("Outbox worker started polling every 3 seconds...")

	// 5. Fiber Web Server Setup with Views Engine
	engine := html.New("./internal/delivery/views", ".html")
	engine.Reload(true)
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Use(logger.New())

	http.NewPaymentHandler(app, paymentUseCase)

	// 6. Graceful Shutdown Setup
	port := getEnv("PORT", "8081")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Fiber server error: %v", err)
		}
	}()
	log.Printf("payment-service running on port %s", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down payment-service...")
	cancel()
	_ = app.Shutdown()
}
