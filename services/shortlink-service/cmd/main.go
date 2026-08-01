package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/redis/go-redis/v9"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/delivery/http"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/delivery/kafka"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/repository"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/usecase"
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func main() {
	// 1. Environment Configurations
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	paymentSvcURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:8081")
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}

	// 2. Redis Client Initialization
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort,
	})

	// 3. Layer Initialization (Clean Architecture DI)
	redisRepo := repository.NewRedisRepository(redisClient)
	paymentClient := repository.NewPaymentClient(paymentSvcURL)
	eventProducer := kafka.NewEventProducer(kafkaBrokers)
	defer eventProducer.Close()

	shortlinkUseCase := usecase.NewShortlinkUseCase(redisRepo, paymentClient, eventProducer)

	// 4. Fiber Web Server Initialization
	app := fiber.New()
	app.Use(logger.New())

	http.NewShortlinkHandler(app, shortlinkUseCase)

	// 5. Graceful Shutdown Setup
	port := getEnv("PORT", "8082")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Fiber server error: %v", err)
		}
	}()
	log.Printf("shortlink-service running on port %s", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down shortlink-service...")
	_ = app.Shutdown()
}
