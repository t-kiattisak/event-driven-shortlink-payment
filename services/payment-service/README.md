# Payment Service (`payment-service`)

## 1. Overview & Responsibilities
`payment-service` is the **Business Domain Service** and **Source of Truth** for payments. It manages payment life cycles, processes payment creation, renders server-side HTML checkout pages with dynamic QR Codes, and ensures event delivery via the **Transactional Outbox Pattern**.

### Key Responsibilities:
- **Payment Domain Logic**: Create payments, update statuses (PENDING, PAID, EXPIRED), and manage pricing/amounts.
- **Server-Side UI Rendering**: Render HTML Checkout UI with QR Code payload directly to the client browser (Fiber HTML Engine).
- **Primary Persistence**: Store payment entities in **PostgreSQL** via **GORM**.
- **Transactional Outbox Pattern**: Persist events to an `outbox` table within the same DB transaction during payment creation/updates to guarantee zero message loss to Kafka.

---

## 2. Clean Architecture Layer Responsibilities

```text
services/payment-service/
├── cmd/
│   └── main.go                  # Entry point, DI, GORM DB migration & Fiber bootstrapper
└── internal/
    ├── domain/                  # Pure Payment Models & Domain Interfaces (Zero External Imports)
    │   ├── payment.go
    │   └── outbox.go
    ├── usecase/                 # Business Workflows & Outbox Relay Worker
    │   ├── payment_usecase.go
    │   └── outbox_worker.go
    ├── repository/              # PostgreSQL (GORM) Adapter Implementations
    │   ├── payment_repository.go
    │   └── outbox_repository.go
    └── delivery/                # HTTP Controllers, HTML Templates, & Kafka Outbox Publisher
        ├── http/
        │   └── payment_handler.go
        ├── views/               # Fiber HTML Views & Templates
        │   └── checkout.html
        └── kafka/
            └── payment_producer.go
```

### Layer Breakdown:
1. **`internal/domain`**: Defines `Payment` struct, `Outbox` struct, `PaymentStatus` constants (`PENDING`, `PAID`, `EXPIRED`), and domain error types.
2. **`internal/usecase`**:
   - `PaymentUseCase`: Handles payment creation logic, QR Code generation (PromptPay / Base64 SVG), and outbox message creation within atomic DB transactions.
   - `OutboxWorker`: Background goroutine worker polling pending records from `outbox` table and publishing them to Kafka.
3. **`internal/repository`**: Implements GORM database queries (`Create`, `FindByPaymentNo`, `UpdateStatus`, `FetchPendingOutbox`, `MarkOutboxProcessed`).
4. **`internal/delivery`**:
   - `http/payment_handler.go`: Fiber handlers for REST APIs (`POST /api/v1/payments`) and HTML rendering (`GET /checkout/:paymentNo`).
   - `views/checkout.html`: Responsive HTML checkout page template with embedded QR Code image.

---

## 3. Database Schema (`PostgreSQL`)

### Table: `payments`
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | BIGSERIAL | PRIMARY KEY | Internal ID |
| `payment_no` | VARCHAR(64) | UNIQUE, NOT NULL | Unique payment number (e.g. PAY-20260731-001) |
| `short_code` | VARCHAR(32) | INDEX | Associated shortlink code |
| `amount` | NUMERIC(12,2)| NOT NULL | Payment amount |
| `currency` | VARCHAR(3) | DEFAULT 'THB' | Currency code |
| `status` | VARCHAR(32) | INDEX, NOT NULL | PENDING, PAID, EXPIRED, CANCELLED |
| `qr_code_data` | TEXT | NULL | Base64 QR Code string / Payload |
| `created_at` | TIMESTAMPTZ | NOT NULL | Creation timestamp |
| `updated_at` | TIMESTAMPTZ | NOT NULL | Update timestamp |

### Table: `outbox`
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | UUID | PRIMARY KEY | Unique Outbox record ID |
| `aggregate_type`| VARCHAR(64) | NOT NULL | Entity type (e.g. `Payment`) |
| `aggregate_id` | VARCHAR(64) | NOT NULL | Entity ID / PaymentNo |
| `topic` | VARCHAR(128) | NOT NULL | Target Kafka Topic |
| `payload` | JSONB | NOT NULL | Full event payload JSON |
| `status` | VARCHAR(32) | DEFAULT 'PENDING' | PENDING, PROCESSED, FAILED |
| `created_at` | TIMESTAMPTZ | NOT NULL | Timestamp |

---

## 4. Interfaces & Contracts

### Inbound Endpoints:
- `POST /api/v1/payments` - Creates a new payment & shortlink code.
- `GET /checkout/:paymentNo` - Renders HTML Checkout Page with QR Code.
- `GET /api/v1/payments/code/:code` - Internal API called by `shortlink-service` on cache miss.

### Outbound Kafka Topics Produced (via Outbox Worker):
- `payment.created`: Broadcast when a new payment is created.
- `payment.status_updated`: Broadcast when status transitions (e.g. `PENDING` -> `PAID`).

---

## 5. Dependencies & Tech Stack
- **Language**: Go 1.22+
- **HTTP Framework**: Fiber (`github.com/gofiber/fiber/v2`)
- **Template Engine**: Fiber HTML Engine (`github.com/gofiber/template/html/v2`)
- **ORM**: GORM (`gorm.io/gorm` & `gorm.io/driver/postgres`)
- **QR Code Generator**: `github.com/skip2/go-qrcode`
- **Database**: PostgreSQL 16
