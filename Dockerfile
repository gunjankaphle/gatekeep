# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN make build

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates postgresql-client

# Create non-root user
RUN addgroup -g 1000 gatekeep && \
    adduser -D -u 1000 -G gatekeep gatekeep

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/bin/gatekeep /usr/local/bin/gatekeep
COPY --from=builder /build/bin/gatekeep-server /usr/local/bin/gatekeep-server

# Change ownership
RUN chown -R gatekeep:gatekeep /app

# Switch to non-root user
USER gatekeep

# Expose API port
EXPOSE 8080

# Default command (can be overridden)
CMD ["gatekeep-server"]
