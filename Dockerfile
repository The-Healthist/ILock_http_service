# Build code
FROM golang:1.21.0-alpine AS builder

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
RUN go build -o main ./cmd/server

# Run release
FROM alpine:latest

WORKDIR /app

# 添加版本信息到最终镜像
LABEL version="2.3.0"
LABEL description="ILock HTTP Service with MQTT Call Support"
LABEL maintainer="Stone Sea"

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata curl

# 创建目录结构
RUN mkdir -p /app/cmd/server /app/logs

# Copy binary from builder
COPY --from=builder /app/main /app/cmd/server/main

# Copy Swagger docs
COPY --from=builder /app/docs /app/docs

# Create logs directory
RUN mkdir -p /app/logs

# Create a simple migration script for backward compatibility
RUN echo '#!/bin/sh\necho "Running migrations via main application"\n/app/cmd/server/main -migration=alter\nexit $?' > /app/run_migrations.sh \
    && chmod +x /app/run_migrations.sh

# Set executable permissions
RUN chmod +x /app/cmd/server/main

EXPOSE 20033

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:20033/api/ping || exit 1

# 使用重构后的入口点
ENTRYPOINT ["/app/cmd/server/main"] 