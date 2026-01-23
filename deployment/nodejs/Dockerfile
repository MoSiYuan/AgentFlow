# AgentFlow Master - Node.js Dockerfile
# Multi-stage build for optimal image size

FROM node:20-alpine AS builder

WORKDIR /app

# Install build dependencies for better-sqlite3
RUN apk add --no-cache python3 make g++ sqlite-dev

# Copy monorepo files
COPY nodejs/package*.json ./
COPY nodejs/pnpm-workspace.yaml ./
COPY nodejs/pnpm-lock.yaml ./

# Install pnpm
RUN npm install -g pnpm

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy source code
COPY nodejs/packages ./packages

# Build all packages
RUN pnpm run build

# Production stage
FROM node:20-alpine

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache sqlite-libs ca-certificates curl

# Copy built packages and node_modules
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/packages ./packages

# Create data directory
RUN mkdir -p /data

# Set working directory
WORKDIR /app/packages/master

# Expose port
EXPOSE 8848

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD curl -f http://localhost:8848/health || exit 1

# Default command: start Master
CMD ["node", "dist/index.js"]
