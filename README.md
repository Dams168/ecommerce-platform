# Ecommerce Platform — Microservices with Go

Platform e-commerce mini untuk belajar microservice, gRPC, Kafka, Redis, dan WebSocket menggunakan Go.

## Arsitektur

```
Client → API Gateway (REST/WS)
              ↓ gRPC
    ┌─────────┼─────────┐
    User   Order    Notification
    Service Service  Service
              ↓ Kafka
         Payment Worker
```

## Services & Ports

| Service              | Port  | Protocol |
|----------------------|-------|----------|
| API Gateway          | 8080  | HTTP/WS  |
| User Service         | 50051 | gRPC     |
| Order Service        | 50052 | gRPC     |
| Notification Service | 50053 | gRPC     |

## Infrastructure

| Service     | Port | Keterangan          |
|-------------|------|---------------------|
| PostgreSQL   | 5432 | Database per service|
| Redis        | 6379 | Cache + Pub/Sub     |
| Kafka        | 9092 | Message broker      |
| Kafka UI     | 8090 | Debug topics        |

## Kafka Topics

| Topic           | Publisher     | Consumer             |
|-----------------|---------------|----------------------|
| order.created   | Order Service | Payment Worker       |
| payment.done    | Payment Worker| Notification Service |
| payment.failed  | Payment Worker| Order Service        |

## Quick Start

```bash
# 1. Install tools
# Lihat docs/setup.md

# 2. Jalankan infrastruktur
make up

# 3. Generate proto
make proto

# 4. Buat semua database
make db-create

# 5. Jalankan service (masing-masing terminal)
make dev-gateway
make dev-user
make dev-order
```

## Struktur Folder

```
ecommerce-platform/
├── api-gateway/          # Entry point, HTTP + WebSocket
├── services/
│   ├── user-service/     # Auth, profile
│   ├── order-service/    # CRUD order
│   └── notification-service/
├── workers/
│   └── payment-worker/   # Async payment processing
├── pkg/                  # Shared libraries
│   ├── kafka/
│   ├── redis/
│   ├── jwt/
│   └── logger/
├── proto/                # gRPC contracts
├── infra/                # Docker Compose
└── Makefile
```

## Development

```bash
make help     # lihat semua command
make up       # start infra
make down     # stop infra
make proto    # generate dari .proto
make test     # run tests
make lint     # run linter
make tidy     # go mod tidy semua
```
