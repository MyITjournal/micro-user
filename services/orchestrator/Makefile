.PHONY: help build run test clean docker-build docker-run

# Variables
APP_NAME=orchestrator
DOCKER_IMAGE=orchestrator:latest
PORT=8080

help:
	@echo "Orchestrator Service - Available Commands:"
	@echo ""
	@echo "  make build         - Build the Go binary"
	@echo "  make run           - Run the service locally"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make lint          - Run linter"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make deps          - Download dependencies"

deps:
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy

build: deps
	@echo "🏗️  Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) ./cmd/server

run: build
	@echo "🚀 Starting $(APP_NAME)..."
	@USE_MOCK_SERVICES=true LOG_FORMAT=console ./bin/$(APP_NAME)

test:
	@echo "🧪 Running tests..."
	@go test -v -race ./...

test-coverage:
	@echo "📊 Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	@echo "🔍 Running linter..."
	@golangci-lint run --timeout=5m

clean:
	@echo "🧹 Cleaning up..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean

docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build
	@echo "🐳 Running Docker container..."
	@docker run --rm -p $(PORT):$(PORT) \
		-e USE_MOCK_SERVICES=true \
		-e LOG_LEVEL=info \
		--name $(APP_NAME) \
		$(DOCKER_IMAGE)

docker-stop:
	@echo "⏹️  Stopping Docker container..."
	@docker stop $(APP_NAME) || true

# Development helpers
dev:
	@echo "🔄 Running in development mode with auto-reload..."
	@air || (echo "Install air: go install github.com/cosmtrek/air@latest" && exit 1)

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

vet:
	@echo "🔬 Vetting code..."
	@go vet ./...

# Integration test with mock data
test-integration:
	@echo "🧪 Running integration tests..."
	@./scripts/test-integration.sh