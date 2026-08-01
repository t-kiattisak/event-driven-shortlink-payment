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
`shortlink-service` เป็น **Single Source of Truth** สำหรับการสร้าง, ตรวจสอบสถานะวันหมดอายุ, และจัดการ Short Code ทั้งหมดในระบบ:
1. **Shortcode Creation**: `shortlink-service` ให้บริการ API สำหรับสร้าง Shortcode (เช่น Base62 / Custom code) โดย `payment-service` จะเรียกสร้างในจังหวะสร้าง Payment และบันทึก `short_code` ลง PostgreSQL ทันที
2. **Default Expiration Policy (7 Days TTL)**: 
   - ทุก Shortcode จะมีอายุการใช้งาน **Default 7 วัน** (604,800 วินาที)
   - บันทึก Mapping ลง **Redis Cache** ด้วย TTL `7 Days` (`SETEX shortlink:code 604800 payload`)
3. **Expiration Validation**: เมื่อมีผู้ใช้งานเปิดลิงก์ หากพบว่าลิงก์หมดอายุแล้ว (`ErrCodeExpired` หรือ Key ใน Redis หายไป) ระบบจะตอบกลับด้วยหน้า UI / HTTP 410 Gone (Link Expired) ทันที
4. **Inter-Service Communication**: ในปัจจุบันบริการคุยกันผ่าน HTTP REST API และพร้อมสำหรับยกระดับเป็น **gRPC (HTTP/2 + Protobuf)** สำหรับ High Throughput ในอนาคต

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
        
        Note over Shortlink,Payment: Content Forwarding (Browser URL remains /s/:code)
        Shortlink->>Payment: GET /checkout/:paymentNo (Internal HTTP Fetch)
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
