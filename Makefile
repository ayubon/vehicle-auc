.PHONY: help build run dev test lint fmt clean docker-up docker-down migrate sqlc test-e2e test-e2e-ui test-all seed seed-clear seed-sql

# Default target
help:
	@echo "Vehicle Auction Go Backend"
	@echo ""
	@echo "Usage:"
	@echo "  make build        - Build the server binary"
	@echo "  make run          - Run the server"
	@echo "  make dev          - Run with hot reload (requires air)"
	@echo "  make test         - Run all Go tests"
	@echo "  make test-unit    - Run unit tests only"
	@echo "  make test-int     - Run integration tests only"
	@echo "  make test-e2e     - Run E2E tests (Playwright)"
	@echo "  make test-e2e-ui  - Run E2E tests with Playwright UI"
	@echo "  make test-all     - Run all tests (Go + E2E)"
	@echo "  make lint         - Run linters"
	@echo "  make fmt          - Format code"
	@echo "  make sqlc         - Generate sqlc code"
	@echo "  make docker-up    - Start all services"
	@echo "  make docker-down  - Stop all services"
	@echo "  make migrate      - Run migrations on dev DB"
	@echo "  make migrate-test - Run migrations on test DB"
	@echo "  make seed         - Seed database via API (requires running server)"
	@echo "  make seed-clear   - Clear seed data via API"
	@echo "  make seed-sql     - Run seed SQL directly"
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

# Seed data
seed:
	@echo "Seeding database via API..."
	curl -X POST http://localhost:8080/debug/seed

seed-clear:
	@echo "Clearing seed data via API..."
	curl -X DELETE http://localhost:8080/debug/seed

seed-sql:
	@echo "Running seed SQL script..."
	psql $(DATABASE_URL) -f migrations-go/002_seed_data.sql

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

# E2E Tests (Playwright)
# Requires: backend running on :8080, frontend will auto-start

test-e2e:
	@echo "Running E2E tests..."
	@echo "Note: Backend must be running on localhost:8080"
	cd frontend && npm run test:e2e

test-e2e-ui:
	@echo "Running E2E tests with UI..."
	@echo "Note: Backend must be running on localhost:8080"
	cd frontend && npm run test:e2e:ui

test-e2e-debug:
	@echo "Running E2E tests in debug mode..."
	cd frontend && npm run test:e2e:debug

# Run all tests (Go backend + E2E)
test-all:
	@echo "Running Go tests..."
	$(GOTEST) -v ./...
	@echo ""
	@echo "Running E2E tests..."
	cd frontend && npm run test:e2e

# Frontend development
frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

frontend-install:
	cd frontend && npm install

