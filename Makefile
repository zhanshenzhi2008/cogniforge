.PHONY: build run test clean dev lint fmt

# Binary name
BINARY_NAME=server

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build the server
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/server/main.go

# Run the server (development mode)
run: build
	POSTGRES_PORT=5433 ./$(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v -short ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -short -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Development mode with hot reload (requires air)
dev:
	cd cmd/server && $(GOCMD) run main.go

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Format code
fmt:
	$(GOFMT) ./...

# Lint code
lint:
	golangci-lint run ./...

# Build for Linux (Docker)
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux ./cmd/server/main.go

# Docker build
docker-build:
	docker build -t cogniforge:latest -f docker/Dockerfile .

# Run database migrations (placeholder)
migrate:
	@echo "Migrations handled via GORM AutoMigrate on startup"

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the server binary"
	@echo "  run            - Build and run the server"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  dev            - Run in development mode"
	@echo "  clean          - Clean build artifacts"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  build-linux    - Build for Linux"
	@echo "  docker-build   - Build Docker image"
