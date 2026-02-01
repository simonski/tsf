.PHONY: help build clean test test-go test-go-unit test-go-coverage initdb

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
	@echo ""

build: initdb

initdb:
	@echo "Building initdb..."
	@go build -o bin/task-initdb ./cmd/initdb

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
