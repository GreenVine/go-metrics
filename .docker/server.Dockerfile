# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy the source code
COPY .. .

# Install required build dependencies
RUN apk add --no-cache git gcc musl-dev && \
    go mod download && \
    # Generate Protobuf files
    go generate ./... && \
    # Generate binary
    CGO_ENABLED=1 go build -ldflags="-s -w" -o /usr/local/sbin/server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /usr/local/sbin/server /usr/local/sbin/server

# Expose the gRPC port
EXPOSE 3000

# Run the binary
ENTRYPOINT ["/usr/local/sbin/server", "-host=0.0.0.0", "-dbPath=/opt/server/data/database.db"]
