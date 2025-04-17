# Build code
FROM golang:1.23.0-alpine AS build-stage

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app
COPY . /app

# Download dependencies and build
RUN go mod tidy
RUN go build -o main .

# Run release
FROM alpine:3.14 AS release-stage

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary and config
COPY --from=build-stage /app/main /app/
COPY --from=build-stage /app/config /app/config

# Create logs directory
RUN mkdir -p /app/logs && \
  chmod 755 /app/logs

# Set executable permissions
RUN chmod +x /app/main

EXPOSE 8080

ENTRYPOINT ["/app/main"] 