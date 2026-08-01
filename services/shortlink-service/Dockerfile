FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY proto ./proto
COPY services/shortlink-service/go.mod services/shortlink-service/go.sum ./services/shortlink-service/
WORKDIR /app/services/shortlink-service
RUN go mod download

COPY services/shortlink-service .
RUN CGO_ENABLED=0 GOOS=linux go build -o shortlink-service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/shortlink-service/shortlink-service .

CMD ["./shortlink-service"]
