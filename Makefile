.PHONY: proto up down restart logs test lint help

# ─── Warna untuk output ───────────────────────
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RESET  := \033[0m

help: ## Tampilkan semua command
	@echo ""
	@echo "  Ecommerce Platform — available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""

proto: ## Generate kode Go dari semua .proto file
	@echo "$(YELLOW)→ Generating proto files...$(RESET)"
	@mkdir -p proto/gen
	protoc \
		--go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		-I proto \
		proto/user/v1/user.proto \
		proto/order/v1/order.proto \
		proto/notification/v1/notification.proto
	@echo "$(GREEN)✓ Proto generated$(RESET)"

up: ## Jalankan semua infrastruktur (Kafka, Redis, PostgreSQL)
	@echo "$(YELLOW)→ Starting infrastructure...$(RESET)"
	docker compose -f infra/docker-compose.yml up -d
	@echo "$(GREEN)✓ Infrastructure running$(RESET)"
	@echo "  Kafka UI  → http://localhost:8090"
	@echo "  PostgreSQL → localhost:5432"
	@echo "  Redis      → localhost:6379"

down: ## Hentikan semua infrastruktur
	docker compose -f infra/docker-compose.yml down

restart: down up ## Restart infrastruktur

logs: ## Lihat logs semua container
	docker compose -f infra/docker-compose.yml logs -f

logs-%: ## Lihat log service tertentu. Contoh: make logs-kafka
	docker compose -f infra/docker-compose.yml logs -f $*

ps: ## Status semua container
	docker compose -f infra/docker-compose.yml ps

test: ## Jalankan semua test
	go test ./...

lint: ## Jalankan linter
	golangci-lint run ./...

tidy: ## go mod tidy semua module
	@for dir in api-gateway services/user-service services/order-service \
		services/notification-service workers/payment-worker \
		pkg/kafka pkg/redis pkg/jwt pkg/logger proto; do \
		echo "$(YELLOW)→ tidy $$dir$(RESET)"; \
		(cd $$dir && go mod tidy); \
	done

clean: ## Hapus semua binary build
	find . -name 'tmp' -type d -exec rm -rf {} + 2>/dev/null || true
	find . -name '*.test' -delete 2>/dev/null || true

# ─── Database ─────────────────────────────────
db-create: ## Buat semua database di PostgreSQL
	docker compose -f infra/docker-compose.yml exec postgres \
		psql -U postgres -f /docker-entrypoint-initdb.d/init.sql

migrate-up: ## Jalankan semua migration
	@echo "$(YELLOW)→ Running migrations...$(RESET)"
	# akan ditambahkan setelah service dibuat

# ─── Dev (jalankan service dengan hot reload) ─
dev-gateway: ## Jalankan API Gateway dengan hot reload
	cd api-gateway && air

dev-user: ## Jalankan User Service dengan hot reload
	cd services/user-service && air

dev-order: ## Jalankan Order Service dengan hot reload
	cd services/order-service && air
