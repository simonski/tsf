.PHONY: help build clean test test-go test-go-unit test-go-coverage task docker docker-up docker-down run-local

help:
	@echo "Task Management System - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build             - Build unified task binary"
	@echo "  clean             - Remove build artifacts"
	@echo "  test              - Run all tests"
	@echo "  test-go           - Run all Go tests"
	@echo "  test-go-unit      - Run Go unit tests only"
	@echo "  test-go-coverage  - Run Go tests with coverage report"
	@echo "  task              - Build unified task binary"
	@echo "  docker            - Build Docker image"
	@echo "  docker-up         - Start all services with Docker Compose"
	@echo "  docker-down       - Stop all Docker services"
	@echo "  run-local         - Run server and orchestrator locally"
	@echo ""

build: task

task:
	@echo "Building unified task binary..."
	@go build -o task ./cmd/task-unified

clean:
	@echo "Cleaning build artifacts..."
	@rm -f task
	@rm -f coverage.out coverage.html

test: test-go

test-go:
	@echo "Running all Go tests..."
	@go test -v ./...

test-go-unit:
	@echo "Running Go unit tests..."
	@go test -v ./internal/...

test-go-coverage:
	@echo "Running Go tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

docker:
	@echo "Building Docker image..."
	@docker build -t task-management:latest .

docker-up:
	@echo "Starting Docker services..."
	@mkdir -p data
	@docker-compose up -d
	@echo "Services started. Server available at http://localhost:8080"

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

run-local: build
	@echo "Initializing database..."
	@mkdir -p ~/.config/task
	@./task initdb -f ~/.config/task/task.db || true
	@echo "Starting server on port 8080..."
	@./task server -f ~/.config/task/task.db -port 8080
