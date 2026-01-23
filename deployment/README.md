# AgentFlow Deployment Guide

> **⚠️ Port Change Notice**: AgentFlow v2.0.0 uses port **6767** instead of 8848.
> All deployment configurations have been updated. See [Migration Guide](../PORT_MIGRATION_GUIDE.md) for details.

This directory contains deployment configurations for AgentFlow.

## Directory Structure

```
deployment/
├── nodejs/              # Node.js deployment (Recommended)
│   ├── Dockerfile
│   ├── docker-compose.yml
│   ├── .dockerignore
│   └── nginx.conf.example
├── k8s/                 # Kubernetes deployment
│   ├── deployment.yaml
│   └── secrets.yaml.example
└── obsolete/            # Old Go deployment configs
    ├── Dockerfile
    ├── docker-compose.yml
    └── .dockerignore
```

## Quick Start

### Docker Compose (Recommended)

```bash
# From project root
cd deployment/nodejs

# Set environment variables
export ANTHROPIC_API_KEY="sk-ant-..."

# Start AgentFlow
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Accessing Services

- **Master API**: http://localhost:6767
- **Health Check**: http://localhost:6767/health
- **Admin UI**: http://localhost:8080 (if enabled)

## Docker Deployment

### Build Image

```bash
# Build Master image
docker build -f deployment/nodejs/Dockerfile -t agentflow/master:latest .

# Build Worker image
docker build -f deployment/nodejs/Dockerfile -t agentflow/worker:latest .
```

### Run Containers

```bash
# Master only
docker run -d \
  --name agentflow-master \
  -p 6767:6767 \
  -v agentflow-data:/data \
  agentflow/master:latest

# Worker
docker run -d \
  --name agentflow-worker \
  -e AGENTFLOW_MASTER_URL=http://agentflow-master:6767 \
  -e ANTHROPIC_API_KEY=sk-ant-... \
  agentflow/worker:latest
```

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (minikube, kind, or cloud provider)
- kubectl configured
- Anthropic API key

### Deploy

```bash
cd deployment/k8s

# Create namespace and secrets
kubectl apply -f secrets.yaml.example

# Deploy AgentFlow
kubectl apply -f deployment.yaml

# Check status
kubectl get pods -n agentflow
kubectl get services -n agentflow

# View logs
kubectl logs -f deployment/agentflow-master -n agentflow

# Port forward to access locally
kubectl port-forward -n agentflow svc/agentflow-master 6767:6767
```

### Scale Workers

```bash
# Scale to 5 workers
kubectl scale deployment/agentflow-worker --replicas=5 -n agentflow

# Check status
kubectl get pods -n agentflow
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NODE_ENV` | Environment | `production` |
| `AGENTFLOW_MASTER_URL` | Master server URL | `http://localhost:6767` |
| `AGENTFLOW_GROUP` | Worker group name | `default` |
| `ANTHROPIC_API_KEY` | Anthropic API key | Required for AI features |

## Production Considerations

### Security

1. **API Keys**: Use Kubernetes secrets or Docker secrets
2. **Network**: Use private networks or VPN
3. **TLS**: Enable HTTPS in production
4. **Authentication**: Add authentication layer

### Persistence

```yaml
volumes:
  agentflow-data:
    driver: local
    driver_opts:
      type: none
      device: /path/to/data
      o: bind
```

### Monitoring

```bash
# Health check endpoint
curl http://localhost:6767/health

# View task stats
curl http://localhost:6767/api/v1/tasks/stats

# Worker status
curl http://localhost:6767/api/v1/workers
```

### Scaling

#### Docker Compose

```yaml
services:
  worker:
    # ...
    deploy:
      replicas: 5
```

#### Kubernetes

```bash
kubectl autoscale deployment/agentflow-worker \
  --min=2 --max=10 \
  --cpu-percent=80 \
  -n agentflow
```

## Troubleshooting

### Master not starting

```bash
# Check logs
docker-compose logs master

# Common issues:
# - Port 6767 already in use
# - Database permissions
# - Missing environment variables
```

### Workers not connecting

```bash
# Check worker logs
docker-compose logs worker

# Verify Master URL
echo $AGENTFLOW_MASTER_URL

# Test connectivity
curl http://master:6767/health
```

### Database issues

```bash
# Check database file
ls -la data/

# Reset database (WARNING: deletes all data)
rm data/agentflow.db
docker-compose restart
```

## Advanced Configuration

### Custom Worker Configuration

```yaml
services:
  worker-custom:
    environment:
      - AGENTFLOW_GROUP=custom
      - WORKER_TIMEOUT=300000
      - MAX_RETRIES=5
      - CHECKPOINT_INTERVAL=60000
```

### Resource Limits

```yaml
services:
  master:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

### Multi-Master Setup

For high availability, deploy multiple masters with a load balancer:

```yaml
services:
  lb:
    image: nginx:alpine
    volumes:
      - ./nginx-lb.conf:/etc/nginx/nginx.conf
    ports:
      - "6767:6767"
    depends_on:
      - master-1
      - master-2
```

## Backup and Restore

### Backup Database

```bash
# Copy database from container
docker cp agentflow-master:/data/agentflow.db ./backup.db

# Or using volume
docker run --rm -v agentflow-data:/data -v $(pwd):/backup \
  alpine tar czf /backup/agentflow-backup.tar.gz /data
```

### Restore Database

```bash
# Copy database to container
docker cp ./backup.db agentflow-master:/data/agentflow.db

# Or restore from archive
docker run --rm -v agentflow-data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/agentflow-backup.tar.gz -C /
```

## Migration from Go Version

If migrating from the old Go deployment:

1. Backup existing database
2. Stop Go containers
3. Deploy Node.js version
4. Restore database (if compatible)
5. Update API endpoints
6. Verify functionality

## Support

- **Documentation**: [../docs/](../docs/)
- **Issues**: [GitHub Issues](https://github.com/MoSiYuan/AgentFlow/issues)
- **Examples**: [../examples/](../examples/)

---

**AgentFlow Version**: 2.0.0
**Last Updated**: 2026-01-23
