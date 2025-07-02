# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy the source code
COPY .. .

RUN apk add --no-cache git && \
    go mod download && \
    # Generate Protobuf files
    go generate ./... && \
    # Generate binary
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /usr/local/sbin/metrics-server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /usr/local/sbin/metrics-server /usr/local/sbin/metrics-server

# Expose the gRPC port
EXPOSE 3000

# Run the binary
ENTRYPOINT ["/usr/local/sbin/metrics-server", "--host=0.0.0.0"]
