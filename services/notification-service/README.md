# Notification Service (`notification-service`)

## 1. Overview & Responsibilities
`notification-service` is an **Asynchronous Consumer Service** that reacts to domain events emitted over Apache Kafka. It listens for payment lifecycle changes and dispatches alerts/notifications (Email, SMS, or LINE Notify) to customers or administrators.

### Key Responsibilities:
- **Event-Driven Notification**: Consume `payment.status_updated` and `payment.created` Kafka topics.
- **Stateless Execution**: Operate as a lightweight, stateless consumer group capable of horizontal scaling.
- **Idempotency Safeguard**: Ensure notifications are delivered at most once per status transition, preventing duplicate SMS/Emails caused by Kafka message re-deliveries.

---

## 2. Clean Architecture Layer Responsibilities

```text
services/notification-service/
├── cmd/
│   └── main.go                  # Consumer initialization & signal handling
└── internal/
    ├── domain/                  # Notification templates & event models
    │   └── notification.go
    ├── usecase/                 # Notification logic & Dispatcher interfaces
    │   └── notification_usecase.go
    ├── repository/              # Notification Adapters (Email/SMS/LINE APIs)
    │   ├── email_adapter.go
    │   └── line_adapter.go
    └── delivery/                # Kafka Consumer Group Handler
        └── kafka/
            └── payment_consumer.go
```

---

## 3. Subscribed Kafka Topics
- **`payment.status_updated`**: Trigger notification when payment changes to `PAID` or `EXPIRED`.
- **Consumer Group ID**: `notification-service-group`

---

## 4. Dependencies & Tech Stack
- **Language**: Go 1.22+
- **Kafka Client**: Segmentio Kafka-go / IBM Sarama
- **Notification Clients**: Go SMTP / SendGrid / LINE Messaging API SDK
