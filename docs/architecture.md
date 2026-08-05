# Full Architecture Specifications & Technical Design

## 1. Overview
This document serves as the authoritative architectural blueprint for the **Event-Driven Shortlink & Payment Platform**. The platform is designed around **Microservices Architecture**, **Clean Architecture Principles**, and **Event-Driven Decoupling** powered by Apache Kafka.

---

## 2. Core Architectural Principles & Patterns

### 2.1 Clean Architecture (4-Layer Structure)
All microservices follow Uncle Bob's Clean Architecture:
- **Domain Layer (`internal/domain`)**: Core business entities and domain logic with zero external dependencies.
- **Use Case Layer (`internal/usecase`)**: Application workflows, business logic interfaces, and orchestration.
- **Repository Layer (`internal/repository`)**: Data mappers and drivers for PostgreSQL (GORM), Redis, and HTTP Clients.
- **Delivery Layer (`internal/delivery`)**: HTTP Fiber handlers, view templates, and Kafka consumer/producer loops.

### 2.2 Transactional Outbox Pattern
To solve the dual-write problem between PostgreSQL and Apache Kafka:
1. `payment-service` writes the payment entity and an outbox event record within the **same DB transaction**.
2. A background worker periodically polls the `outbox` table and publishes pending events to Kafka.
3. Upon successful publication, the record status is updated to `PROCESSED`.

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant PaymentSvc as payment-service (Go)
    participant DB as PostgreSQL
    participant Worker as Outbox Worker
    participant Kafka

    Client->>PaymentSvc: POST /api/v1/payments
    PaymentSvc->>DB: BEGIN TX
    PaymentSvc->>DB: INSERT INTO payments
    PaymentSvc->>DB: INSERT INTO outbox
    PaymentSvc->>DB: COMMIT TX
    PaymentSvc-->>Client: 201 Created (Fast Response)

    loop Every N ms
        Worker->>DB: SELECT * FROM outbox WHERE status = 'PENDING'
        Worker->>Kafka: Publish Event
        Worker->>DB: UPDATE outbox SET status = 'PROCESSED'
    end
```

### 2.4 Shortcode Generation, Expiration & Domain Ownership
`shortlink-service` serves as the **Single Source of Truth** for shortcode generation, 7-day expiration validation, and code management across the platform:
1. **Shortcode Creation**: `shortlink-service` exposes creation APIs (e.g. Base62 / Custom code algorithms). `payment-service` invokes this endpoint during payment creation and immediately persists the resulting `short_code` in PostgreSQL.
2. **Default Expiration Policy (7 Days TTL)**: 
   - Every shortcode carries a **default 7-day expiration policy** (604,800 seconds).
   - Code-to-Payment mappings are cached in **Redis** with a 7-day TTL (`SETEX shortlink:code 604800 payload`).
3. **Expiration Validation**: During link resolution (`GET /s/:code`), if a link is found to be expired (`ErrCodeExpired` or missing Redis key), the service responds immediately with HTTP 410 Gone (Link Expired) HTML UI.
4. **Inter-Service Communication & gRPC Roadmap**: 
   - **Current Implementation**: Services communicate internally via synchronous HTTP REST APIs.
   - **gRPC Migration Plan (Planned Phase)**: Inter-service communications between `payment-service` and `shortlink-service` are scheduled for upgrade to **gRPC Protocol (HTTP/2 + Protocol Buffers)** to maximize throughput, eliminate connection handshake overhead, and enforce strict Protobuf schema contracts across microservices.

### 2.5 gRPC Protobuf Interface Definitions & RPC Contracts
To support high-throughput internal microservice communication between `payment-service` and `shortlink-service`, the following Protobuf schemas define the gRPC contracts for the optimization phase:

#### 1. Shortlink Service Proto (`proto/shortlink.proto`)
```protobuf
syntax = "proto3";

package shortlink.v1;

option go_package = "github.com/t-kiattisak/event-driven-shortlink-payment/proto/shortlink/v1;shortlinkv1";

service ShortlinkService {
  rpc CreateShortlink (CreateShortlinkRequest) returns (CreateShortlinkResponse);
}

message CreateShortlinkRequest {
  string payment_no = 1;
  string target_url = 2;
  string custom_code = 3;
}

message CreateShortlinkResponse {
  string code = 1;
  string payment_no = 2;
  string target_url = 3;
  int64 expires_at = 4;
}
```

#### 2. Payment Service Proto (`proto/payment.proto`)
```protobuf
syntax = "proto3";

package payment.v1;

option go_package = "github.com/t-kiattisak/event-driven-shortlink-payment/proto/payment/v1;paymentv1";

service PaymentService {
  rpc FetchCheckoutHTML (FetchCheckoutHTMLRequest) returns (FetchCheckoutHTMLResponse);
}

message FetchCheckoutHTMLRequest {
  string payment_no = 1;
}

### 2.6 Analytics Consumer & Data Pipeline
`analytics-consumer` functions as an asynchronous worker that ingests multi-topic events from Apache Kafka and aggregates analytics into PostgreSQL (`audit_db`):
1. **Multi-Topic Ingestion**:
   - **`shortlink.clicked`**: Captures shortlink click telemetry (IP address, User-Agent, referer, shortcode, payment reference).
   - **`payment.created`**: Records new payment invoice creations for revenue analytics.
   - **`payment.status_updated`**: Tracks payment lifecycle state transitions (`PENDING` $\rightarrow$ `PAID` / `EXPIRED`).
2. **Database Persistence**: Persists structured metrics into `click_analytics` and `payment_analytics` tables using GORM `Upsert` and conflict-free clauses.

### 2.7 Full Observability Stack Architecture (Planned Roadmap)
To achieve production-grade visibility across all microservices, the platform incorporates a full Observability Stack:
1. **Metrics Collection (Prometheus)**:
   - Scrapes `/metrics` endpoints across `payment-service` (`:8081`), `shortlink-service` (`:8082`), and Kafka brokers every 5 seconds.
2. **Centralized Visual Dashboards (Grafana)**:
   - Pre-configured Grafana instance (`:3000`) visualizing request rates, P99 latency, error rates, and conversion metrics.
3. **Distributed Waterfall Tracing (Jaeger & OpenTelemetry)**:
   - Propagates W3C trace context headers across HTTP/gRPC boundaries to visualize end-to-end request flows in Jaeger UI (`:16686`).

### 2.8 Audit Service & Immutable Compliance Logs
`audit-service` functions as a dedicated compliance engine for financial auditing:
1. **Multi-Topic Compliance Monitoring**: Subscribes to all domain topics (`payment.created`, `payment.status_updated`, `shortlink.created`, `shortlink.clicked`).
2. **Immutable Audit Trail**: Writes raw event payloads into the `audit_logs` table in PostgreSQL (`audit_db`) with timestamping and event ID tracking for regulatory compliance.

```mermaid
sequenceDiagram
    autonumber
    actor User as User Browser
    participant Gateway as API Gateway (Nginx)
    participant Shortlink as shortlink-service (Go)
    participant Redis
    participant Payment as payment-service (Go)
    participant Kafka

    User->>Gateway: GET /s/:code (e.g., /s/pay99)
    Gateway->>Shortlink: Forward GET /s/:code
    activate Shortlink
    
    Shortlink->>Redis: GET shortlink:pay99
    
    alt Case 1: Link Expired (TTL > 7 Days or Key Missing)
        Shortlink-->>User: 410 Gone / HTML (Notification: Link Expired)
        
    else Case 2: Link Valid & Active (TTL <= 7 Days)
        Shortlink->>Shortlink: Fire-and-forget (Produce shortlink.clicked -> Kafka)
        
        Note over Shortlink,Payment: Internal Inter-Service Call via gRPC / HTTP RPC (Browser URL remains /s/:code)
        Shortlink->>Payment: FetchCheckoutHTML RPC / HTTP GET
        Payment-->>Shortlink: HTML UI + QR Code (text/html)
        Shortlink-->>User: 200 OK (Content-Type: text/html - Checkout UI with QR Code)
    end
    deactivate Shortlink
```

---

## 3. Data Models & Schemas

### PostgreSQL Database (`payment_db`)

#### Table: `payments`
```sql
CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    payment_no VARCHAR(64) UNIQUE NOT NULL,
    short_code VARCHAR(32) UNIQUE NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    qr_code_data TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_short_code ON payments(short_code);
```

#### Table: `outbox`
```sql
CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id VARCHAR(64) NOT NULL,
    topic VARCHAR(128) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(32) DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_outbox_status ON outbox(status);
```

### Redis Cache (`shortlink-service`)
- **Key**: `shortlink:{code}`
- **Value**: `{"payment_no": "PAY-12345", "target_url": "/checkout/PAY-12345", "expires_at": "2026-08-07T20:00:00Z"}`
- **TTL**: 604800 seconds (Default 7 Days / 168 Hours)

---

## 4. Observability Architecture (Three Pillars)

1. **Logging**: Container stdout/stderr $\rightarrow$ Fluent Bit Shipper $\rightarrow$ Elasticsearch $\rightarrow$ Kibana
2. **Metrics**: `/metrics` Endpoints $\rightarrow$ Prometheus $\rightarrow$ Grafana Dashboards
3. **Tracing**: OpenTelemetry SDK $\rightarrow$ Jaeger (Distributed Waterfall Request Tracing)
