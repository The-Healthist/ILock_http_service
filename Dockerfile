# Build code
FROM golang:1.21.0-alpine AS builder

# 设置Go环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 添加版本信息
LABEL version="2.3.0" \
      description="ILock HTTP Service with MQTT Call Support" \
      maintainer="Stone Sea"

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application with optimizations
RUN go build -ldflags="-s -w" -o main ./cmd/server

# Run release
FROM alpine:latest

WORKDIR /app

# 添加版本信息到最终镜像
LABEL version="2.3.0" \
      description="ILock HTTP Service with MQTT Call Support" \
      maintainer="Stone Sea"

# 安装必要的运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/cache/apk/*

# 创建目录结构
RUN mkdir -p /app/cmd/server /app/logs /app/docs

# Copy binary and docs from builder
COPY --from=builder /app/main /app/cmd/server/main
COPY --from=builder /app/docs /app/docs

# Set executable permissions
RUN chmod +x /app/cmd/server/main

EXPOSE 20033

# 更全面的健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:20033/api/health || exit 1

# 使用重构后的入口点
ENTRYPOINT ["/app/cmd/server/main"] 