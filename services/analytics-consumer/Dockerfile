FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY services/analytics-consumer/go.mod services/analytics-consumer/go.sum ./services/analytics-consumer/
WORKDIR /app/services/analytics-consumer
RUN go mod download

COPY services/analytics-consumer .
RUN CGO_ENABLED=0 GOOS=linux go build -o analytics-consumer ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/analytics-consumer/analytics-consumer .

CMD ["./analytics-consumer"]
