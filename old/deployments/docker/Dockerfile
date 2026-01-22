# Multi-stage build for CPDS
# Stage 1: Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev \
    sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -a -ldflags="-w -s" \
    -o cpds ./cpds

# Stage 2: Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    sqlite-libs \
    ca-certificates \
    tzdata

# Create non-root user
RUN addgroup -g 1000 cpds && \
    adduser -D -u 1000 -G cpds cpds

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/cpds /usr/local/bin/cpds

# Set ownership
RUN chown -R cpds:cpds /app

# Switch to non-root user
USER cpds

# Create data directory
RUN mkdir -p /app/.claude/cpds-manager

# Expose port
EXPOSE 8848

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8848/api/health || exit 1

# Default command
ENTRYPOINT ["/usr/local/bin/cpds"]
CMD ["master", "--host", "0.0.0.0", "--port", "8848"]
