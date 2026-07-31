# Analytics Consumer Service (`analytics-consumer`)

## 1. Overview & Responsibilities
`analytics-consumer` is an **Asynchronous Data Ingestion Consumer Service**. It ingests high-volume click stream events (`shortlink.clicked`) and payment lifecycle events (`payment.created`, `payment.status_updated`) from Kafka, transforms them, and performs bulk indexing directly into **Elasticsearch**.

### Key Responsibilities:
- **Event Aggregation**: Consume click events and payment state changes from multiple Kafka topics.
- **Bulk Indexing**: Batch events and write them efficiently into Elasticsearch indices (e.g., `analytics-clicks-2026.07`, `analytics-payments-2026.07`).
- **Dashboard Enablement**: Provide rich aggregated data source for **Kibana** dashboards (Click-through rates, Payment Conversion Rates, Latency metrics).

---

## 2. Clean Architecture Layer Responsibilities

```text
services/analytics-consumer/
├── cmd/
│   └── main.go                  # Main consumer runner & Elasticsearch client setup
└── internal/
    ├── domain/                  # Analytics Metrics models & Index schemas
    │   └── analytics.go
    ├── usecase/                 # Event Transformation & Batching UseCase
    │   └── analytics_usecase.go
    ├── repository/              # Elasticsearch Bulk Indexer Repository
    │   └── elasticsearch_repository.go
    └── delivery/                # Kafka Consumer Group Handler
        └── kafka/
            └── event_consumer.go
```

---

## 3. Subscribed Kafka Topics & Indices

| Kafka Topic | Elasticsearch Index Pattern | Purpose |
| :--- | :--- | :--- |
| `shortlink.clicked` | `analytics-clicks-YYYY.MM` | Track link click volumes, GEO IP, Referral sources |
| `payment.created` | `analytics-payments-YYYY.MM` | Track payment creation rates & revenue intent |
| `payment.status_updated` | `analytics-payments-YYYY.MM` | Track payment conversion, completion & expiration rates |

---

## 4. Dependencies & Tech Stack
- **Language**: Go 1.22+
- **Database / Indexer**: Elasticsearch 8.x Client (`github.com/elastic/go-elasticsearch/v8`)
- **Kafka Client**: Segmentio Kafka-go
