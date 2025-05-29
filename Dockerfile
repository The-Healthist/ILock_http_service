# Build code
FROM golang:1.23.0-alpine AS builder

# 设置Go环境变量
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 添加版本信息
LABEL version="1.1.0"
LABEL description="ILock HTTP Service with MQTT Call Support"
LABEL maintainer="Stone Sea"

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN go build -o ilock_http_service

# Run release
FROM alpine:latest

WORKDIR /app

# 添加版本信息到最终镜像
LABEL version="1.1.0"
LABEL description="ILock HTTP Service with MQTT Call Support"
LABEL maintainer="Stone Sea"

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# Copy binary and config
COPY --from=builder /app/ilock_http_service .

# Create logs directory
RUN mkdir -p /app/logs

# Set executable permissions
RUN chmod +x /app/ilock_http_service

EXPOSE 20033

ENTRYPOINT ["./ilock_http_service"] 