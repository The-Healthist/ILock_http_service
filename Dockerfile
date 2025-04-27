# Build code
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -o ilock_service

# Run release
FROM alpine:latest

WORKDIR /app

# Copy binary and config
COPY --from=builder /app/ilock_service .

# Create logs directory
RUN mkdir -p /app/logs

# Set executable permissions
RUN chmod +x /app/ilock_service

EXPOSE 20033

ENTRYPOINT ["./ilock_service"] 