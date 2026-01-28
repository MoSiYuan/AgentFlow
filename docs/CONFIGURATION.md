# AgentFlow Configuration Guide

Complete reference for configuring AgentFlow v0.2.1 (Pure Rust).

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration Methods](#configuration-methods)
- [Configuration File Format](#configuration-file-format)
- [Environment Variables](#environment-variables)
- [Operating Modes](#operating-modes)
- [Complete Configuration Reference](#complete-configuration-reference)
- [Security Configuration](#security-configuration)
- [Advanced Configuration](#advanced-configuration)
- [Examples](#examples)

---

## Quick Start

### Minimal Configuration

Create `~/.agentflow/config.toml`:

```toml
[server]
port = 6767

[database]
url = "sqlite://agentflow.db"
```

Then start AgentFlow:
```bash
agentflow server
```

### Using Environment Variables

```bash
export AGENTFLOW_SERVER_PORT=6767
export AGENTFLOW_DATABASE_URL="sqlite://agentflow.db"
agentflow server
```

---

## Configuration Methods

AgentFlow supports multiple configuration methods, applied in the following priority order (highest to lowest):

1. **Command-line arguments** (highest priority)
2. **Environment variables**
3. **Configuration file** (`~/.agentflow/config.toml`)
4. **Default values** (lowest priority)

### Example: Priority in Action

```bash
# Default: port = 6767
# Config file: port = 8080
# Environment: AGENTFLOW_SERVER_PORT=9000
# Command line: agentflow --port 3000

# Result: port 3000 (command line wins)
```

---

## Configuration File Format

### Default Location

The configuration file is located at:
- **Linux/macOS**: `~/.agentflow/config.toml`
- **Windows**: `%USERPROFILE%\.agentflow\config.toml`

### Custom Location

You can specify a custom configuration file:

```bash
agentflow server --config /path/to/custom-config.toml
```

---

## Environment Variables

All configuration options can be set via environment variables. Use the `AGENTFLOW_` prefix:

| Environment Variable | Config Section | Description |
|---------------------|---------------|-------------|
| `AGENTFLOW_SERVER_ADDR` | `[server]` | Server bind address |
| `AGENTFLOW_SERVER_PORT` | `[server]` | Server port |
| `AGENTFLOW_DATABASE_URL` | `[database]` | Database connection URL |
| `AGENTFLOW_LOG_LEVEL` | `[server]` | Log level (trace, debug, info, warn, error) |
| `AGENTFLOW_MAX_CONCURRENT_TASKS` | `[executor]` | Maximum concurrent tasks |
| `AGENTFLOW_TASK_TIMEOUT` | `[executor]` | Task timeout in seconds |
| `AGENTFLOW_MEMORY_BACKEND` | `[memory]` | Memory backend type |
| `AGENTFLOW_SANDBOX_ENABLED` | `[sandbox]` | Enable sandbox security |

---

## Operating Modes

AgentFlow supports three operating modes:

### 1. Local Mode (Default)

Executes tasks locally using Claude CLI.

```bash
agentflow server local
# or
agentflow server
```

**Characteristics:**
- Tasks execute on the same machine
- Uses local Claude CLI installation
- Lowest latency
- No network requirements

**Use Cases:**
- Development and testing
- Local automation
- Single-machine deployments

### 2. Cloud Mode

Receives tasks via webhooks from external platforms (e.g., Zhipu AI).

```bash
agentflow server cloud
```

**Characteristics:**
- Exposes webhook endpoints
- Integrates with AI platforms
- Supports remote task execution
- Requires public URL

**Use Cases:**
- AI platform integration
- Multi-user systems
- Distributed task execution

### 3. Planner-Only Mode

Only plans tasks without execution (useful for preview/validation).

```bash
agentflow server planner-only
```

**Characteristics:**
- Creates task plans
- Validates workflows
- No actual execution
- Dry-run mode

**Use Cases:**
- Task planning validation
- Workflow testing
- Preview mode

---

## Complete Configuration Reference

### Server Configuration

```toml
[server]
# Server bind address (default: "0.0.0.0")
# Use "127.0.0.1" for local-only access
host = "0.0.0.0"

# Server port (default: 6767)
port = 6767

# Log level (default: "info")
# Options: trace, debug, info, warn, error
log_level = "info"

# Worker heartbeat timeout in seconds (default: 60)
# How long to wait before considering a worker "lost"
worker_heartbeat_timeout = 60
```

### Database Configuration

```toml
[database]
# Database connection URL (default: "sqlite://agentflow.db")
# Supported formats:
# - SQLite: sqlite://path/to/database.db
# - PostgreSQL: postgresql://user:password@localhost/dbname
# - MySQL: mysql://user:password@localhost/dbname
url = "sqlite://agentflow.db"

# Maximum connections in pool (default: 10)
max_connections = 10

# Connection timeout in seconds (default: 30)
connection_timeout = 30
```

### Executor Configuration

```toml
[executor]
# Maximum concurrent tasks (default: 10)
# Number of tasks that can run simultaneously
max_concurrent_tasks = 10

# Task timeout in seconds (default: 300)
# Maximum time a task can run before being killed
task_timeout = 300

# Maximum retry attempts (default: 3)
max_retries = 3

# Retry delay in seconds (default: 5)
retry_delay = 5

# Executor backend (default: "claude-cli")
# Options: claude-cli, docker, kubernetes
backend = "claude-cli"
```

### Memory Configuration

```toml
[memory]
# Memory backend (default: "memory")
# Options:
# - memory: In-memory storage (fast, not persistent)
# - sqlite: SQLite database (persistent)
# - redis: Redis server (distributed)
backend = "memory"

# Database URL for SQLite backend (optional)
database_url = "sqlite://agentflow_memory.db"

# Redis connection URL for Redis backend (optional)
redis_url = "redis://localhost:6379"

# Default TTL (time-to-live) in seconds (default: 3600)
# How long to keep memory entries
default_ttl = 3600

# Maximum entries in memory (default: 10000)
# LRU eviction when limit is reached
max_entries = 10000

# Enable persistence (default: false)
# Save memory to disk periodically
enable_persistence = false

# Persistence interval in seconds (default: 300)
# How often to save to disk
persistence_interval = 300
```

### Sandbox Configuration

```toml
[sandbox]
# Enable sandbox security (default: true)
# WARNING: Disabling sandbox may expose security risks
enabled = true

# Default workspace directory (default: "/tmp/agentflow/workspace")
# Where task execution happens
default_workspace = "/tmp/agentflow/workspace"

# Allow network access (default: false)
# Enable if tasks need to make network requests
allow_network = false

# Maximum memory limit (optional)
# Format: "1G", "512M", "2048K"
max_memory = "1G"

# Maximum CPU cores (default: 2)
max_cpu = 2

# Process timeout in seconds (default: 300)
timeout = 300

# Path whitelist (optional)
# Only allow access to these paths
whitelist = [
    "/home/user/projects",
    "/tmp/agentflow"
]

# Path blacklist (optional)
# Never allow access to these paths
blacklist = [
    "/etc",
    "/root",
    "/var/log"
]
```

### Webhook Configuration

```toml
[webhook]
# Enable webhook support (default: false)
enabled = false

# Webhook endpoint path (default: "/api/v1/webhook")
path = "/api/v1/webhook"

# Webhook secret for signature verification (optional)
secret = "your-webhook-secret-key"

# Signature algorithm (default: "sha256")
# Options: sha256, sha512
algorithm = "sha256"

# Header name containing signature (default: "X-Webhook-Signature")
signature_header = "X-Webhook-Signature"

# Enable IP whitelist (default: false)
ip_whitelist_enabled = false

# Allowed IP addresses (optional)
ip_whitelist = ["203.119.0.0/16"]

# Rate limiting: requests per minute (default: 100)
rate_limit = 100

# Rate limit burst size (default: 20)
rate_limit_burst = 20

# Require authentication (default: false)
require_auth = false

# Authentication header name (default: "X-API-Key")
auth_header = "X-API-Key"

# Expected API key (optional)
auth_key = "your-expected-api-key"
```

### Zhipu AI Integration

```toml
[zhipu]
# Enable Zhipu AI integration (default: false)
enabled = false

# Zhipu AI API key (required if enabled)
api_key = "your-zhipu-api-key"

# Model to use (default: "glm-4")
# Options: glm-4, glm-3-turbo, glm-4-flash
model = "glm-4"

# API endpoint (default: "https://open.bigmodel.cn/api/paas/v4")
endpoint = "https://open.bigmodel.cn/api/paas/v4"

# Maximum tokens per request (default: 8192)
max_tokens = 8192

# Temperature for response generation (default: 0.7)
# Range: 0.0 - 1.0 (higher = more creative)
temperature = 0.7

# Request timeout in seconds (default: 300)
timeout = 300

# Enable retry on failure (default: true)
retry_enabled = true

# Maximum retry attempts (default: 3)
max_retries = 3

# Retry delay in seconds (default: 1)
retry_delay = 1

# Callback URL for task updates (optional)
callback_url = "https://your-domain.com/api/v1/callback"
```

### CORS Configuration

```toml
[cors]
# Enable CORS (default: true)
enabled = true

# Allowed origins (default: ["*"])
# Use specific origins for better security:
# allowed_origins = ["https://example.com", "https://app.example.com"]
allowed_origins = ["*"]

# Allowed methods (default: ["GET", "POST", "PUT", "DELETE"])
allowed_methods = ["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"]

# Allowed headers (default: ["*"])
allowed_headers = ["*"]

# Allow credentials (default: false)
allow_credentials = false

# Max age for preflight requests (default: 3600)
max_age = 3600
```

---

## Security Configuration

### 1. API Key Authentication

```toml
[auth]
# Enable API key authentication (default: false)
enabled = true

# API keys (at least one required)
# Generate with: openssl rand -hex 32
api_keys = [
    "sk-1234567890abcdef1234567890abcdef",
    "sk-abcdef1234567890abcdef1234567890"
]

# Header name for API key (default: "X-API-Key")
header = "X-API-Key"
```

### 2. TLS/SSL Configuration

```toml
[tls]
# Enable TLS (default: false)
enabled = true

# Certificate file path (required if enabled)
cert_file = "/path/to/cert.pem"

# Private key file path (required if enabled)
key_file = "/path/to/key.pem"

# CA certificate file path (optional)
ca_file = "/path/to/ca.pem"

# Minimum TLS version (default: "1.2")
# Options: 1.0, 1.1, 1.2, 1.3
min_version = "1.2"

# Client authentication (default: false)
client_auth_enabled = false
```

### 3. Rate Limiting

```toml
[rate_limit]
# Enable rate limiting (default: false)
enabled = true

# Requests per minute (default: 60)
requests_per_minute = 60

# Requests per hour (default: 1000)
requests_per_hour = 1000

# Burst size (default: 10)
burst = 10

# Storage backend (default: "memory")
# Options: memory, redis
storage = "memory"

# Redis URL for distributed rate limiting (optional)
redis_url = "redis://localhost:6379"
```

---

## Advanced Configuration

### 1. Logging

```toml
[logging]
# Log level (default: "info")
level = "info"

# Log format (default: "pretty")
# Options: pretty, json, compact
format = "pretty"

# Log file path (optional)
# If not specified, logs to stdout/stderr
file = "/var/log/agentflow/agentflow.log"

# Enable log rotation (default: false)
rotation_enabled = true

# Maximum log file size in MB (default: 100)
max_size_mb = 100

# Maximum number of log files to keep (default: 10)
max_files = 10

# Maximum age of log files in days (default: 30)
max_age_days = 30
```

### 2. Metrics and Monitoring

```toml
[metrics]
# Enable metrics collection (default: false)
enabled = true

# Metrics endpoint (default: "/metrics")
endpoint = "/metrics"

# Metrics format (default: "prometheus")
# Options: prometheus, statsd
format = "prometheus"

# StatsD server (if using statsd format)
statsd_host = "localhost:8125"
```

### 3. Task Queue Configuration

```toml
[queue]
# Queue backend (default: "memory")
# Options: memory, redis, postgresql
backend = "memory"

# Redis connection URL (if using redis)
redis_url = "redis://localhost:6379"

# Queue name (default: "agentflow")
name = "agentflow"

# Maximum queue size (default: 10000)
max_size = 10000

# Task priority levels (default: ["low", "normal", "high", "urgent"])
priorities = ["low", "normal", "high", "urgent"]
```

---

## Examples

### Example 1: Development Configuration

```toml
# ~/.agentflow/config.toml (development)
[server]
host = "127.0.0.1"
port = 6767
log_level = "debug"

[database]
url = "sqlite://dev_agentflow.db"

[executor]
max_concurrent_tasks = 5
task_timeout = 600

[memory]
backend = "memory"
default_ttl = 7200

[sandbox]
enabled = true
allow_network = true
default_workspace = "./workspace"
```

### Example 2: Production Configuration

```toml
# ~/.agentflow/config.toml (production)
[server]
host = "0.0.0.0"
port = 6767
log_level = "info"

[database]
url = "postgresql://agentflow:password@db.example.com/agentflow"
max_connections = 20
connection_timeout = 30

[executor]
max_concurrent_tasks = 50
task_timeout = 300
max_retries = 5

[memory]
backend = "redis"
redis_url = "redis://redis.example.com:6379"
default_ttl = 3600
enable_persistence = true

[sandbox]
enabled = true
allow_network = false
default_workspace = "/var/lib/agentflow/workspace"
whitelist = ["/var/lib/agentflow", "/data/projects"]

[webhook]
enabled = true
secret = "${WEBHOOK_SECRET}"
rate_limit = 1000

[zhipu]
enabled = true
api_key = "${ZHIPU_API_KEY}"
model = "glm-4"
max_tokens = 8192

[auth]
enabled = true
api_keys = ["${AGENTFLOW_API_KEY}"]

[tls]
enabled = true
cert_file = "/etc/ssl/certs/agentflow.crt"
key_file = "/etc/ssl/private/agentflow.key"
```

### Example 3: Cloud Mode with Webhooks

```toml
# ~/.agentflow/config.toml (cloud mode)
[server]
host = "0.0.0.0"
port = 6767

[database]
url = "sqlite://agentflow_cloud.db"

[executor]
max_concurrent_tasks = 20
task_timeout = 300

[memory]
backend = "sqlite"
database_url = "sqlite://agentflow_memory.db"
default_ttl = 86400

[webhook]
enabled = true
secret = "your-webhook-secret"
ip_whitelist_enabled = true
ip_whitelist = ["203.119.0.0/16"]
rate_limit = 500

[zhipu]
enabled = true
api_key = "your-zhipu-api-key"
model = "glm-4"
callback_url = "https://your-domain.com/callback"
```

### Example 4: Minimal Configuration

```toml
# ~/.agentflow/config.toml (minimal)
[server]
port = 6767

[database]
url = "sqlite://agentflow.db"
```

---

## Configuration Validation

AgentFlow validates configuration on startup. Common errors:

### Invalid Port

```
Error: Invalid AGENTFLOW_SERVER_PORT: must be between 1 and 65535
```

### Invalid Database URL

```
Error: Invalid database URL: unknown scheme
```

### Missing Required Fields

```
Error: Missing required field: zhipu.api_key (when zhipu.enabled = true)
```

### File Not Found

```
Error: TLS certificate file not found: /path/to/cert.pem
```

---

## Best Practices

1. **Use Environment Variables for Secrets**
   ```toml
   # Don't do this:
   api_key = "sk-1234567890"

   # Do this:
   api_key = "${ZHIPU_API_KEY}"
   ```

2. **Separate Configurations per Environment**
   - Development: `config.dev.toml`
   - Staging: `config.staging.toml`
   - Production: `config.prod.toml`

3. **Use Version Control for Configuration Templates**
   ```bash
   config.toml.template    # Commit to git
   config.toml            # Add to .gitignore
   ```

4. **Document Custom Configurations**
   ```toml
   # Custom configuration for XYZ project
   # Contact: user@example.com
   [custom]
   project_id = "xyz-123"
   ```

5. **Regularly Rotate Secrets**
   - API keys
   - Webhook secrets
   - Database passwords

6. **Monitor Configuration Changes**
   - Use configuration checksums
   - Log configuration changes
   - Track configuration versions

---

## Troubleshooting

### Issue: Configuration Not Loading

**Check:**
```bash
# Verify config file exists
ls -la ~/.agentflow/config.toml

# Check syntax
cat ~/.agentflow/config.toml
```

**Solution:** Ensure TOML syntax is valid (use [TOML Lint](https://toml-lint.com/))

### Issue: Environment Variables Ignored

**Check:**
```bash
# Verify variable is set
echo $AGENTFLOW_SERVER_PORT

# Check for typos
env | grep AGENTFLOW
```

**Solution:** Ensure variable name matches pattern `AGENTFLOW_*`

### Issue: Default Values Used

**Check:**
```bash
# Run with debug logging
agentflow server --log-level debug
```

**Solution:** Check configuration file syntax and variable names

---

## Additional Resources

- **[README.md](README.md)**: Main documentation
- **[ZHIPU_INTEGRATION.md](ZHIPU_INTEGRATION.md)**: Zhipu AI integration guide
- **[RUST_V3_QUICKSTART.md](RUST_V3_QUICKSTART.md)**: Quick start guide
- **[API Documentation](rust/agentflow-master/API.md)**: REST API reference

---

## Support

For configuration issues:
- Check logs: `agentflow server --log-level debug`
- Validate config: `agentflow config validate`
- GitHub Issues: [https://github.com/MoSiYuan/AgentFlow/issues](https://github.com/MoSiYuan/AgentFlow/issues)
