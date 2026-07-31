# Shortlink Service (`shortlink-service`)

## 1. Overview & Responsibilities
`shortlink-service` acts as the **Core Infrastructure Service** responsible for resolving shortlink codes (`/s/:code`), performing fast lookups via **Redis Cache**, tracking user click metrics, and redirecting or dispatching requests to the appropriate Business Domain Service (`payment-service`).

### Key Responsibilities:
- **Code Generation & Single Source of Truth**: Generate unique short codes (Base62 / Hashing) via `POST /api/v1/shortlinks`.
- **Expiration Validation & 7-Day TTL**: Enforce a **default 7-day TTL (604,800 seconds)** expiration policy. Validate whether a link has expired during resolution (`GET /s/:code`) and respond with HTTP 410 Gone if expired.
- **Code Resolution**: Decode short link codes into target resources (`paymentNo` / target URL) via `GET /s/:code`.
- **Fast Caching**: Store and cache code-to-payment mappings in **Redis** with 7-day TTL.
- **Event Creation**: Produce `shortlink.created` event when a new link is registered.
- **Click Tracking (Fire-and-Forget)**: Asynchronously produce `shortlink.clicked` events to Apache Kafka whenever a code is resolved without blocking client HTTP responses.
- **Redirection / Proxy**: Seamlessly redirect users to the dynamic Checkout UI provided by `payment-service`.

---

## 2. Clean Architecture Layer Responsibilities

```text
services/shortlink-service/
├── cmd/
│   └── main.go                  # Application entry point & Dependency Injection (DI)
└── internal/
    ├── domain/                  # Core Business Entities & Domain Errors (Zero Dependencies)
    │   └── shortlink.go
    ├── usecase/                 # Business Logic & Workflow Interfaces
    │   └── shortlink_usecase.go
    ├── repository/              # Data Adapters (Redis & Payment HTTP Client)
    │   ├── redis_repository.go
    │   └── payment_client.go
    └── delivery/                # HTTP Controllers & Kafka Event Producers
        ├── http/
        │   └── shortlink_handler.go
        └── kafka/
            └── click_producer.go
```

### Layer Breakdown:
1. **`internal/domain`**: Defines `Shortlink` struct, `ClickEvent` struct, domain errors (`ErrCodeNotFound`, `ErrCodeExpired`).
2. **`internal/usecase`**: Implements code resolution logic:
   - Check Redis cache first (Cache-Aside pattern).
   - On cache miss, fetch details from `payment-service` HTTP endpoint and populate Redis cache.
   - Dispatch `shortlink.clicked` event to Kafka via interface.
3. **`internal/repository`**: Implements Redis client calls (`go-redis`) and HTTP REST client calls (`go-resty` or `net/http`) to `payment-service`.
4. **`internal/delivery`**: Fiber HTTP Handlers (`GET /s/:code`) and Sarama/Confluent Kafka Producer adapter.

---

## 3. Interfaces & Contracts

### Inbound HTTP Endpoints:
- `GET /s/:code` - Resolves code and redirects user to `/checkout/:paymentNo`.
- `GET /health` - Liveness & readiness probe.

### Outbound Kafka Topics Produced:
- `shortlink.clicked`:
  ```json
  {
    "event_id": "uuid-v4",
    "code": "pay123",
    "payment_no": "PAY-20260731-001",
    "timestamp": "2026-07-31T20:50:00Z",
    "user_agent": "Mozilla/5.0...",
    "ip_address": "203.0.113.1"
  }
  ```

---

## 4. Dependencies & Tech Stack
- **Language**: Go 1.22+
- **HTTP Framework**: Fiber (`github.com/gofiber/fiber/v2`)
- **Cache**: Redis (`github.com/redis/go-redis/v9`)
- **Kafka Client**: Segmentio Kafka-go / IBM Sarama
- **Metrics**: Prometheus Client (`github.com/prometheus/client_golang`)
