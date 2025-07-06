# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy the source code
COPY .. .

# Install required build dependencies
RUN apk add --no-cache git && \
    go mod download && \
    # Generate Protobuf files
    go generate ./... && \
    # Generate binary
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /usr/local/sbin/stresser ./cmd/stresser

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /usr/local/sbin/stresser /usr/local/sbin/stresser

# Run the binary
ENTRYPOINT ["/usr/local/sbin/stresser", "-server=server:3000"]
