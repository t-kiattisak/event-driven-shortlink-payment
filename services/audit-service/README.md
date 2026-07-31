# Audit Service (`audit-service`)

## 1. Overview & Responsibilities
`audit-service` is an **Append-Only Compliance & Traceability Consumer Service**. It listens to **ALL** system domain events across Kafka topics and persists an immutable audit log into a dedicated PostgreSQL database table (or Cold Storage S3) for compliance, auditing, and forensic analysis.

### Key Responsibilities:
- **Immutable Audit Trail**: Store raw event payloads with exact event timestamps and message headers.
- **Compliance & Security**: Maintain zero-tamper audit logs for system governance.
- **Append-Only Persistence**: Guarantee that audit records can only be inserted, never updated or deleted.

---

## 2. Clean Architecture Layer Responsibilities

```text
services/audit-service/
├── cmd/
│   └── main.go                  # Main entry point & DB migration initializer
└── internal/
    ├── domain/                  # Audit Log Entity Definition
    │   └── audit_log.go
    ├── usecase/                 # Audit Persistence UseCase
    │   └── audit_usecase.go
    ├── repository/              # PostgreSQL GORM Repository Adapter
    │   └── audit_repository.go
    └── delivery/                # Kafka Multi-Topic Consumer Group Handler
        └── kafka/
            └── audit_consumer.go
```

---

## 3. Database Schema (`PostgreSQL - audit_db`)

### Table: `audit_logs`
| Column | Type | Description |
| :--- | :--- | :--- |
| `id` | BIGSERIAL PRIMARY KEY | Auto-increment Primary Key |
| `event_id` | VARCHAR(64) NOT NULL | Unique event UUID from Kafka payload |
| `topic` | VARCHAR(128) NOT NULL | Source Kafka topic |
| `event_type` | VARCHAR(64) NOT NULL | Event action name |
| `payload` | JSONB NOT NULL | Complete, unmodified JSON message payload |
| `recorded_at` | TIMESTAMPTZ NOT NULL | Event emission timestamp |
| `created_at` | TIMESTAMPTZ NOT NULL | Local insertion timestamp |

---

## 4. Dependencies & Tech Stack
- **Language**: Go 1.22+
- **Database**: PostgreSQL 16
- **ORM**: GORM (`gorm.io/gorm`)
- **Kafka Client**: Segmentio Kafka-go
