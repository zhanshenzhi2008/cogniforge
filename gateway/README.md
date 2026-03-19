# CogniForge Gateway

Backend API service built with Go + Gin.

## Tech Stack

- Go 1.22
- Gin
- GORM
- PostgreSQL
- Redis
- JWT

## Getting Started

```bash
# Install dependencies
go mod tidy

# Run server
go run ./cmd/server
```

## Configuration

Edit `configs/config.yaml` to configure database, redis, and JWT settings.
