FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY proto ./proto
COPY services/payment-service/go.mod services/payment-service/go.sum ./services/payment-service/
WORKDIR /app/services/payment-service
RUN go mod download

COPY services/payment-service .
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/payment-service/payment-service .
COPY --from=builder /app/services/payment-service/internal/delivery/views ./internal/delivery/views

CMD ["./payment-service"]
