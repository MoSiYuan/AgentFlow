# CPDS - Claude Parallel Development System (Go Version)

A high-performance, distributed task execution system rewritten in Go with Docker support.

## Features

- **Dual Deployment Modes**:
  - **Cloud Mode**: Master and Worker run continuously, Workers call Claude API
  - **Standalone Mode**: Master auto-shutdown after task completion, Workers exit after one task

- **High Performance**:
  - 100x faster HTTP throughput than Python version
  - Single binary deployment (<20MB)
  - SQLite database with WAL mode

- **Claude API Integration**:
  - Workers can execute tasks using Claude API
  - Token tracking and cost management
  - Graceful fallback to local execution

- **Docker Support**:
  - Multi-stage builds for minimal image size
  - Separate compose files for cloud and standalone modes
  - Health checks and auto-restart

## Quick Start

### Install

```bash
# From source
go install github.com/jiangxiaolong/cpds-go/cpds@latest

# Or build manually
git clone https://github.com/jiangxiaolong/cpds-go.git
cd cpds-go
go build -o cpds ./cpds
```

### Standalone Mode (Single Machine)

```bash
# Terminal 1: Start Master
./cpds master --mode standalone --auto-shutdown

# Terminal 2: Start Worker
./cpds worker --mode standalone --master http://localhost:8848

# Terminal 3: Create a task
curl -X POST http://localhost:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "TASK-001",
    "title": "Test Task",
    "description": "Test the CPDS system",
    "priority": "high"
  }'
```

### Docker Mode

```bash
# Standalone mode (exits after completion)
docker-compose -f deployments/docker/docker-compose.standalone.yml up

# Cloud mode (continuous)
docker-compose -f deployments/docker/docker-compose.cloud.yml up
```

## API Endpoints

### Worker Management
- `POST /api/workers/register` - Register worker
- `POST /api/workers/heartbeat` - Worker heartbeat
- `GET /api/workers` - List online workers

### Task Management
- `POST /api/tasks/create` - Create task
- `GET /api/tasks/pending` - Get pending tasks
- `GET /api/tasks/running` - Get running tasks
- `GET /api/tasks/completed` - Get completed tasks
- `POST /api/tasks/assign` - Assign task to worker
- `POST /api/tasks/progress` - Update task progress
- `POST /api/tasks/complete` - Complete task

### System
- `GET /api/status` - System status
- `GET /api/health` - Health check

## Configuration

### Command Line Flags

```bash
# Master
./cpds master \
  --mode standalone \
  --host 0.0.0.0 \
  --port 8848 \
  --auto-shutdown

# Worker
./cpds worker \
  --mode standalone \
  --master http://localhost:8848 \
  --name my-worker \
  --oneshot
```

### Environment Variables

```bash
export CPDS_MODE=standalone
export CLAUDE_API_KEY=sk-ant-...
export CPDS_MASTER_HOST=0.0.0.0
export CPDS_MASTER_PORT=8848
```

## Development

```bash
# Run tests
go test ./...

# Build
go build -o cpds ./cpds

# Run
./cpds master --mode standalone
```

## Project Structure

```
cpds-go/
├── cmd/                  # CLI commands
│   ├── main.go
│   ├── root.go
│   ├── master.go
│   └── worker.go
├── internal/
│   ├── api/             # API types
│   ├── config/          # Configuration
│   ├── database/        # Database layer
│   ├── master/          # Master server
│   └── worker/          # Worker client
├── deployments/
│   └── docker/          # Docker configs
├── scripts/             # Build scripts
└── docs/                # Documentation
```

## Performance

- **HTTP Throughput**: 10,000+ req/s (vs 100 req/s in Python)
- **Memory Usage**: ~20MB (vs 50MB in Python)
- **Startup Time**: <100ms (vs 500ms in Python)
- **Binary Size**: ~15MB (single static binary)

## License

MIT

## Author

Jiang Xiaolong
