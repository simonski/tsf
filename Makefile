.PHONY: help build clean test test-go test-go-unit test-go-coverage test-e2e test-e2e-setup task docker docker-up docker-down run-local

help:
	@echo "sf - Software Factory - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build             - Build binary"
	@echo "  clean             - Remove build artifacts"
	@echo "  test              - Run all tests (Go + E2E)"
	@echo "  test-go           - Run all Go tests"
	@echo "  test-go-unit      - Run Go unit tests only"
	@echo "  test-go-coverage  - Run Go tests with coverage report"
	@echo "  test-e2e-setup    - Install Playwright dependencies"
	@echo "  test-e2e          - Run Playwright E2E tests"
	@echo "  test-e2e-ui       - Run Playwright tests in UI mode"
	@echo "  docker            - Build Docker image"
	@echo "  docker-up         - Start all services with Docker Compose"
	@echo "  docker-down       - Stop all Docker services"
	@echo "  run-local         - Run server and orchestrator locally"
	@echo ""

build: 
	@echo "Building sf binary..."
	@go build -o sf .

clean:
	@echo "Cleaning build artifacts..."
	@rm -f sf
	@rm -f coverage.out coverage.html
	@rm -f sf.test.db

test: test-go test-e2e
	@echo "All tests completed successfully!"

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

test-e2e-setup:
	@echo "Installing Playwright dependencies..."
	@cd tests/e2e && npm install
	@cd tests/e2e && npx playwright install

test-e2e: build
	@echo "Running Playwright E2E tests..."
	@./sf initdb -f sf.test.db --force -password admin || true
	@cd tests/e2e && npm test || (pkill -f 'sf server' || true; exit 1)
	@pkill -f 'sf server' || true

test-e2e-ui: build
	@echo "Running Playwright E2E tests in UI mode..."
	@./sf initdb -f sf.test.db --force -password admin || true
	@cd tests/e2e && npm run test:ui

docker:
	@echo "Building Docker image..."
	@docker build -t sf:latest .

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
	@mkdir -p ~/.config/sf
	@./sf initdb -f ~/.config/sf/sf.db || true
	@echo "Starting server on port 8080..."
	@./sf server -f ~/.config/sf/sf.db -port 8080
