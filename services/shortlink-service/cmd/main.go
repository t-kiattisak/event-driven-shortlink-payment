package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	shortlinkv1 "github.com/t-kiattisak/event-driven-shortlink-payment/proto/shortlink/v1"
	sgrpc "github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/delivery/grpc"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/delivery/http"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/delivery/kafka"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/repository"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/usecase"
	"google.golang.org/grpc"
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
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}

	// 2. Redis Client Initialization
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort,
	})

	// 4. Layer Initialization (Clean Architecture)
	redisRepo := repository.NewRedisRepository(redisClient)
	
	paymentGRPCAddress := getEnv("PAYMENT_SERVICE_GRPC_URL", "payment-service:50051")
	paymentGRPCClient, err := repository.NewPaymentGRPCClient(paymentGRPCAddress)
	if err != nil {
		log.Fatalf("Failed to initialize Payment gRPC Client: %v", err)
	}
	defer paymentGRPCClient.Close()

	paymentClient := repository.NewPaymentClient(paymentGRPCClient)
	kafkaProducer := kafka.NewEventProducer(kafkaBrokers)
	shortlinkUseCase := usecase.NewShortlinkUseCase(redisRepo, paymentClient, kafkaProducer)

	// Start gRPC Server in Background Goroutine
	grpcPort := getEnv("GRPC_PORT", "50052")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err == nil {
		grpcServer := grpc.NewServer()
		shortlinkv1.RegisterShortlinkServiceServer(grpcServer, sgrpc.NewShortlinkGRPCServer(shortlinkUseCase))
		go func() {
			log.Printf("shortlink-service gRPC server listening on port %s", grpcPort)
			if err := grpcServer.Serve(lis); err != nil {
				log.Printf("gRPC server error: %v", err)
			}
		}()
	}

	// 5. Fiber Web Server Initialization with Prometheus Metrics
	app := fiber.New()
	app.Use(logger.New())

	// Prometheus Metrics Endpoint
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

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
