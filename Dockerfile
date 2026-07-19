# =====================================================================
# 阶段一：构建
# =====================================================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# 预复制 go mod 文件（利用 Docker 缓存层）
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o cogniforge \
    ./cmd/server

# =====================================================================
# 阶段二：运行
# =====================================================================
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget

COPY --from=builder /app/cogniforge .

RUN mkdir -p /app/uploads

EXPOSE 8080

ENV GIN_MODE=release

ENTRYPOINT ["./cogniforge"]
CMD ["serve"]
