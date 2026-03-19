# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (测试文件会被编译但不会被打包)
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -a -installsuffix cgo \
    -o main ./cmd/server

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy only the binary and config (测试文件不在这里)
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

CMD ["./main"]
