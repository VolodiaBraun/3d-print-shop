include .env
export

.PHONY: up down restart logs ps clean status run build test migrate-up migrate-down migrate-create

# Docker Compose commands
up: ## Start all services
	docker compose up -d

down: ## Stop all services
	docker compose down

restart: ## Restart all services
	docker compose down
	docker compose up -d

logs: ## Show logs (follow mode)
	docker compose logs -f

ps: ## Show running containers
	docker compose ps

status: ## Check health of all services
	@echo "=== PostgreSQL ==="
	@docker compose exec postgres pg_isready -U shop_user -d shop_db 2>/dev/null || echo "PostgreSQL is not running"
	@echo ""
	@echo "=== Redis ==="
	@docker compose exec redis redis-cli ping 2>/dev/null || echo "Redis is not running"
	@echo ""
	@echo "=== MinIO ==="
	@docker compose exec minio mc ready local 2>/dev/null && echo "MinIO is ready" || echo "MinIO is not running"

clean: ## Stop services and remove volumes (DESTROYS DATA)
	docker compose down -v

# Backend commands
run: ## Run backend server
	cd backend && go run ./cmd/api/

build: ## Build backend binary
	cd backend && go build -o bin/api ./cmd/api/

test: ## Run backend tests
	cd backend && go test ./...

migrate-up: ## Apply all pending migrations
	migrate -path backend/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration
	migrate -path backend/migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## Create new migration (usage: make migrate-create name=create_products)
	migrate create -ext sql -dir backend/migrations -seq $(name)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
