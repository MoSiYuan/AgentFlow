# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o agentflow ./cmd/agentflow

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache sqlite-libs ca-certificates curl

# Copy binary from builder
COPY --from=builder /app/agentflow /app/agentflow

# Create data directory
RUN mkdir -p /data

# Set entrypoint
ENTRYPOINT ["/app/agentflow"]

# Default command shows help
CMD ["--help"]
