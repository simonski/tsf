.PHONY: help build clean test test-go test-go-unit test-go-coverage initdb server cli orchestrator worker docker docker-up docker-down run-local

help:
	@echo "Task Management System - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build             - Build all binaries"
	@echo "  clean             - Remove build artifacts"
	@echo "  test              - Run all tests"
	@echo "  test-go           - Run all Go tests"
	@echo "  test-go-unit      - Run Go unit tests only"
	@echo "  test-go-coverage  - Run Go tests with coverage report"
	@echo "  initdb            - Build initdb command"
	@echo "  server            - Build server binary"
	@echo "  cli               - Build CLI client"
	@echo "  orchestrator      - Build orchestrator daemon"
	@echo "  worker            - Build worker daemon"
	@echo "  docker            - Build Docker image"
	@echo "  docker-up         - Start all services with Docker Compose"
	@echo "  docker-down       - Stop all Docker services"
	@echo "  run-local         - Run server and orchestrator locally"
	@echo ""

build: initdb server cli orchestrator worker

initdb:
	@echo "Building initdb..."
	@go build -o bin/task-initdb ./cmd/initdb

server:
	@echo "Building server..."
	@go build -o bin/task-server ./cmd/server

cli:
	@echo "Building CLI..."
	@go build -o bin/task ./cmd/task

orchestrator:
	@echo "Building orchestrator..."
	@go build -o bin/task-orchestrator ./cmd/orchestrator

worker:
	@echo "Building worker..."
	@go build -o bin/task-worker ./cmd/worker

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out

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
	@./bin/task-initdb -f ~/.config/task/task.db || true
	@echo "Starting server on port 8080..."
	@./bin/task-server -f ~/.config/task/task.db -port 8080
