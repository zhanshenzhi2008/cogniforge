# Makefile for gateway service

.PHONY: build run test test-unit test-integration docker-build docker-run clean lint fmt

BINARY_NAME=gateway
DOCKER_IMAGE=orjrs/gateway
VERSION=$(shell git describe --tags --always --dirty)
PORT=8080

# Build the binary (测试文件会被自动排除)
build:
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) ./cmd/server

# Build with debug info
build-dev:
	go build -o bin/$(BINARY_NAME) ./cmd/server

# Run the service
run:
	go run ./cmd/server

# Run all tests
test:
	go test -v -race ./...

# Run unit tests only (同包测试文件)
test-unit:
	go test -v -short ./...

# Run integration tests only (test/ 目录)
test-integration:
	go test -v ./test/...

# Run tests with coverage
test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Docker build (测试文件不会被打包)
docker-build:
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# Docker run
docker-run:
	docker run -p $(PORT):$(PORT) --env-file .env $(DOCKER_IMAGE):latest

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Lint
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy

# Help
help:
	@echo "Gateway Service Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build          - Build the binary (production)"
	@echo "  build-dev      - Build with debug info"
	@echo "  run            - Run the service"
	@echo "  test           - Run all tests"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-cover     - Run tests with coverage report"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  clean          - Clean build artifacts"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  tidy           - Tidy dependencies"
