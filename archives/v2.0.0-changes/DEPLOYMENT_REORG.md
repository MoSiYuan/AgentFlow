# Deployment Directory Reorganization

## Summary

Docker and deployment files have been reorganized into a dedicated `deployment/` directory for better structure and maintainability.

## Changes Made

### Moved from Root to deployment/

**Previous Structure:**
```
AgentFlow/
├── Dockerfile              # Old Go version
├── docker-compose.yml      # Old Go version
└── .dockerignore
```

**New Structure:**
```
deployment/
├── nodejs/                    # Node.js deployment (Current)
│   ├── Dockerfile             # Multi-stage Node.js build
│   ├── docker-compose.yml     # Docker Compose config
│   ├── .dockerignore          # Node.js ignore rules
│   └── nginx.conf.example     # Reverse proxy config
├── k8s/                       # Kubernetes deployment
│   ├── deployment.yaml        # K8s manifests
│   └── secrets.yaml.example   # Secret templates
├── obsolete/                  # Old Go deployment
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── .dockerignore
└── README.md                  # Deployment guide
```

## New Deployment Files

### 1. Node.js Dockerfile

**Features:**
- Multi-stage build (builder + runtime)
- Optimized image size
- SQLite support included
- Health checks configured
- Production-ready

**Usage:**
```bash
docker build -f deployment/nodejs/Dockerfile -t agentflow/master:latest .
```

### 2. Docker Compose

**Services:**
- **Master**: HTTP API on port 6767
- **Worker 1**: General task execution
- **Worker 2**: Scalable worker
- **Admin**: Optional nginx admin UI

**Usage:**
```bash
cd deployment/nodejs
export ANTHROPIC_API_KEY="sk-ant-..."
docker-compose up -d
```

### 3. Kubernetes Deployment

**Resources:**
- Namespace: `agentflow`
- Master deployment (1 replica)
- Worker deployment (3 replicas, auto-scalable)
- Service: ClusterIP on port 6767
- PVC: 1Gi storage
- Ingress: For external access

**Usage:**
```bash
cd deployment/k8s
kubectl apply -f deployment.yaml
kubectl get pods -n agentflow
```

## Benefits

### 1. **Organized Structure** 📁
- All deployment files in one place
- Clear separation of concerns (Node.js vs K8s)
- Easy to find and maintain

### 2. **Version Control** 🎯
- Old configs preserved in `obsolete/`
- Clear migration path
- Historical reference

### 3. **Flexibility** 🔄
- Multiple deployment options
- Easy to add new environments (staging, prod)
- Environment-specific configs

### 4. **Documentation** 📚
- Comprehensive deployment guide
- Examples for common scenarios
- Troubleshooting section

## Quick Start Commands

### Docker (Local Development)
```bash
cd deployment/nodejs
docker-compose up -d
```

### Docker (Production)
```bash
docker build -f deployment/nodejs/Dockerfile -t agentflow/master:latest .
docker run -d -p 6767:6767 -v agentflow-data:/data agentflow/master:latest
```

### Kubernetes
```bash
cd deployment/k8s
kubectl apply -f deployment.yaml
kubectl port-forward -n agentflow svc/agentflow-master 6767:6767
```

## File Sizes

| File | Size | Description |
|------|------|-------------|
| `deployment/nodejs/Dockerfile` | ~1KB | Multi-stage build |
| `deployment/nodejs/docker-compose.yml` | ~2KB | 4 services |
| `deployment/nodejs/.dockerignore` | ~1KB | Build optimization |
| `deployment/k8s/deployment.yaml` | ~3KB | Complete K8s stack |
| `deployment/k8s/secrets.yaml.example` | ~0.5KB | Secret template |
| `deployment/README.md` | ~8KB | Full guide |

## Migration Guide

### From Old Docker Setup

**Before:**
```bash
docker build -t agentflow .
docker-compose up -d
```

**After:**
```bash
cd deployment/nodejs
docker-compose up -d
```

### Path Changes

| Old Path | New Path |
|----------|----------|
| `./Dockerfile` | `deployment/nodejs/Dockerfile` |
| `./docker-compose.yml` | `deployment/nodejs/docker-compose.yml` |
| `./.dockerignore` | `deployment/nodejs/.dockerignore` |

## Next Steps

1. **Update CI/CD** pipelines to use new paths
2. **Add monitoring** (Prometheus, Grafana)
3. **Add logging** (ELK stack)
4. **Create staging environment** config
5. **Add production** optimization

## Related Files

- Updated `README.md` to reference deployment/
- Old deployment scripts archived in `archives/old-scripts/`

## Verification

```bash
# Test Docker build
docker build -f deployment/nodejs/Dockerfile -t test .

# Test docker-compose (dry-run)
cd deployment/nodejs
docker-compose config

# Test K8s manifests (dry-run)
cd deployment/k8s
kubectl apply --dry-run=client -f deployment.yaml
```

---

**Date**: 2026-01-23
**AgentFlow Version**: 2.0.0
