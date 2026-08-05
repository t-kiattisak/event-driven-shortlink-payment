FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY services/audit-service/go.mod services/audit-service/go.sum ./services/audit-service/
WORKDIR /app/services/audit-service
RUN go mod download

COPY services/audit-service .
RUN CGO_ENABLED=0 GOOS=linux go build -o audit-service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/audit-service/audit-service .

CMD ["./audit-service"]
