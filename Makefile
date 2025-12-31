.PHONY: help build run dev test lint fmt clean docker-up docker-down migrate sqlc

# Default target
help:
	@echo "Vehicle Auction Go Backend"
	@echo ""
	@echo "Usage:"
	@echo "  make build        - Build the server binary"
	@echo "  make run          - Run the server"
	@echo "  make dev          - Run with hot reload (requires air)"
	@echo "  make test         - Run all tests"
	@echo "  make test-unit    - Run unit tests only"
	@echo "  make test-int     - Run integration tests only"
	@echo "  make lint         - Run linters"
	@echo "  make fmt          - Format code"
	@echo "  make sqlc         - Generate sqlc code"
	@echo "  make docker-up    - Start all services"
	@echo "  make docker-down  - Stop all services"
	@echo "  make migrate      - Run migrations on dev DB"
	@echo "  make migrate-test - Run migrations on test DB"
	@echo ""

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build variables
BINARY_NAME=server
MAIN_PATH=./cmd/server

# Database URLs
DATABASE_URL?=postgres://postgres:postgres@localhost:5432/vehicle_auc?sslmode=disable
TEST_DATABASE_URL?=postgres://postgres:postgres@localhost:5433/vehicle_auc_test?sslmode=disable

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run: build
	DATABASE_URL=$(DATABASE_URL) ./$(BINARY_NAME)

# Run with hot reload (requires: go install github.com/cosmtrek/air@latest)
dev:
	air

# Run all tests
test:
	$(GOTEST) -v ./...

# Run unit tests only
test-unit:
	$(GOTEST) -v -short ./...

# Run integration tests
test-int:
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) $(GOTEST) -v ./tests/integration/...

# Run tests with coverage
test-cover:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run linters (requires: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	golangci-lint run

# Format code
fmt:
	$(GOCMD) fmt ./...
	goimports -w .

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Generate sqlc code (requires: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest)
sqlc:
	sqlc generate

# Docker commands
docker-up:
	docker compose up -d postgres postgres-test redis jaeger

docker-down:
	docker compose down

docker-build:
	docker compose build api

docker-run:
	docker compose up api

# Database migrations
migrate:
	@echo "Running migrations on dev database..."
	psql $(DATABASE_URL) -f migrations-go/001_initial_schema.up.sql

migrate-test:
	@echo "Running migrations on test database..."
	psql $(TEST_DATABASE_URL) -f migrations-go/001_initial_schema.up.sql

migrate-down:
	@echo "Rolling back migrations on dev database..."
	psql $(DATABASE_URL) -f migrations-go/001_initial_schema.down.sql

migrate-down-test:
	@echo "Rolling back migrations on test database..."
	psql $(TEST_DATABASE_URL) -f migrations-go/001_initial_schema.down.sql

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install development tools
tools:
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Generate all
generate: sqlc
	$(GOCMD) generate ./...

# Quick start: setup everything for development
setup: docker-up
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	$(MAKE) migrate
	$(MAKE) migrate-test
	@echo ""
	@echo "Development environment ready!"
	@echo "Run 'make run' to start the server"

