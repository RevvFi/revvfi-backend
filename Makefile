.PHONY: help build run test clean docker-build docker-up docker-down migrate lint fmt

BINARY_API=revvfi-api
BINARY_INDEXER=revvfi-indexer
BINARY_SCHEDULER=revvfi-scheduler

help:
	@echo "RevvFi Backend - Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  make build          - Build all binaries"
	@echo "  make run-api        - Run API server locally"
	@echo "  make run-indexer    - Run indexer locally"
	@echo "  make run-scheduler  - Run scheduler locally"
	@echo "  make test           - Run all tests"
	@echo "  make test-unit      - Run unit tests only"
	@echo "  make test-integration - Run integration tests"
	@echo "  make coverage       - Generate test coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint           - Run linter"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Run go vet"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up     - Run migrations"
	@echo "  make migrate-down   - Rollback migrations"
	@echo "  make migrate-status - Show migration status"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build   - Build Docker images"
	@echo "  make docker-up      - Start Docker containers"
	@echo "  make docker-down    - Stop Docker containers"
	@echo "  make docker-logs    - View container logs"
	@echo ""
	@echo "Contracts:"
	@echo "  make generate-bindings - Generate contract bindings"
	@echo ""
	@echo "Clean:"
	@echo "  make clean          - Remove binaries and artifacts"

# Build targets
build: build-api build-indexer build-scheduler

build-api:
	@echo "Building API server..."
	go build -o bin/$(BINARY_API) ./cmd/api

build-indexer:
	@echo "Building indexer..."
	go build -o bin/$(BINARY_INDEXER) ./cmd/indexer

build-scheduler:
	@echo "Building scheduler..."
	go build -o bin/$(BINARY_SCHEDULER) ./cmd/scheduler

# Run targets (development)
run-api: build-api
	@echo "Starting API server..."
	./bin/$(BINARY_API)

run-indexer: build-indexer
	@echo "Starting indexer..."
	./bin/$(BINARY_INDEXER)

run-scheduler: build-scheduler
	@echo "Starting scheduler..."
	./bin/$(BINARY_SCHEDULER)

# Test targets
test:
	@echo "Running all tests..."
	go test -v -race ./...

test-unit:
	@echo "Running unit tests..."
	go test -v -race ./test/unit/...

test-integration:
	@echo "Running integration tests..."
	go test -v -race -tags=integration ./test/integration/...

coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality
lint:
	@echo "Running linter..."
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

# Database
migrate-up:
	@echo "Running migrations..."
	@if [ -f scripts/migrate.sh ]; then \
		bash scripts/migrate.sh up; \
	else \
		echo "Migration script not found"; \
	fi

migrate-down:
	@echo "Rolling back migrations..."
	@if [ -f scripts/migrate.sh ]; then \
		bash scripts/migrate.sh down; \
	else \
		echo "Migration script not found"; \
	fi

migrate-status:
	@echo "Checking migration status..."
	@if [ -f scripts/migrate.sh ]; then \
		bash scripts/migrate.sh status; \
	else \
		echo "Migration script not found"; \
	fi

# Docker
docker-build:
	@echo "Building Docker images..."
	docker-compose -f docker/docker-compose.yml build

docker-up:
	@echo "Starting Docker containers..."
	docker-compose -f docker/docker-compose.yml up -d
	@echo "Services started!"
	@docker-compose -f docker/docker-compose.yml ps

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose -f docker/docker-compose.yml down

docker-logs:
	@docker-compose -f docker/docker-compose.yml logs -f

docker-shell-db:
	@docker-compose -f docker/docker-compose.yml exec postgres psql -U postgres -d revvfi_db

# Contract bindings
generate-bindings:
	@echo "Generating contract bindings..."
	@if [ -f scripts/generate_bindings.sh ]; then \
		bash scripts/generate_bindings.sh; \
	else \
		echo "Generate bindings script not found"; \
	fi

# Dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Clean
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache -testcache

install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
