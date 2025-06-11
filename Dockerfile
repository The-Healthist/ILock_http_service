# Build code
FROM golang:1.23.0-alpine AS builder

# 设置Go环境变量
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 添加版本信息
LABEL version="2.3.0"
LABEL description="ILock HTTP Service with MQTT Call Support"
LABEL maintainer="Stone Sea"

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application with new structure
RUN go build -o ilock_http_service ./cmd/server

# Run release
FROM alpine:latest

WORKDIR /app

# 添加版本信息到最终镜像
LABEL version="2.3.0"
LABEL description="ILock HTTP Service with MQTT Call Support"
LABEL maintainer="Stone Sea"

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata curl

# Copy binary from builder
COPY --from=builder /app/ilock_http_service .

# Copy configuration structure
COPY --from=builder /app/internal /app/internal
COPY --from=builder /app/pkg /app/pkg

# Create logs directory
RUN mkdir -p /app/logs

# Create a simple migration script
RUN echo '#!/bin/sh\necho "No migrations needed or migrations handled by application"\nexit 0' > /app/run_migrations.sh \
    && chmod +x /app/run_migrations.sh

# Set executable permissions
RUN chmod +x /app/ilock_http_service

EXPOSE 20033

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:20033/api/ping || exit 1

ENTRYPOINT ["./ilock_http_service"] 