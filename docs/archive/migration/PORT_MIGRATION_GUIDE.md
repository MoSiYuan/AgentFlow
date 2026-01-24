# AgentFlow 端口变更迁移指南

## 变更概述

AgentFlow 的默认端口已从 **8848** 更改为 **6767**。

**生效版本**: AgentFlow 2.0.0
**变更日期**: 2026-01-23
**影响范围**: 所有 Master 服务、API 端点、文档和示例

## 快速迁移步骤

### 1. 更新环境变量

```bash
# 旧配置
export AGENTFLOW_MASTER_URL="http://localhost:8848"

# 新配置
export AGENTFLOW_MASTER_URL="http://localhost:6767"
```

### 2. 更新应用代码

```typescript
// 旧代码
const skill = new AgentFlowSkill({
  master_url: 'http://localhost:8848'
});

// 新代码
const skill = new AgentFlowSkill({
  master_url: 'http://localhost:6767'
});
```

### 3. 更新 Docker 配置

```yaml
# docker-compose.yml (旧)
ports:
  - "8848:8848"

# docker-compose.yml (新)
ports:
  - "6767:6767"
```

### 4. 更新 Kubernetes 配置

```yaml
# Service (旧)
spec:
  ports:
  - port: 8848

# Service (新)
spec:
  ports:
  - port: 6767
```

### 5. 更新防火墙规则

```bash
# 开放新端口
ufw allow 6767/tcp
# 或
firewall-cmd --permanent --add-port=6767/tcp
firewall-cmd --reload
```

## 受影响的组件

### ✅ 已更新的文件

#### 核心代码
- ✅ `nodejs/packages/master/src/index.ts`
- ✅ `nodejs/packages/worker/src/index.ts`
- ✅ `nodejs/packages/skill/src/index.ts`
- ✅ `nodejs/packages/cli/src/index.ts`

#### 配置文件
- ✅ `.agentflow/config.example.json`
- ✅ `golang/config.example.yaml`

#### 部署文件
- ✅ `deployment/nodejs/Dockerfile`
- ✅ `deployment/nodejs/docker-compose.yml`
- ✅ `deployment/k8s/deployment.yaml`

#### 文档
- ✅ `README.md`
- ✅ `docs/AI_INTEGRATION.md`
- ✅ `docs/SKILL.md`
- ✅ `docs/ARCHITECTURE.md`
- ✅ `examples/README.md`

#### 测试文件
- ✅ `nodejs/test-*.js`
- ✅ `examples/*.sh`

## 使用示例

### 启动 Master 服务

```bash
# 方式 1: 直接启动
cd nodejs
node packages/master/dist/index.js
# 服务运行在 http://localhost:6767

# 方式 2: Docker
docker run -d -p 6767:6767 agentflow/master:latest

# 方式 3: Docker Compose
cd deployment/nodejs
docker-compose up -d
```

### 健康检查

```bash
# 旧端口 (已弃用)
curl http://localhost:8848/health

# 新端口 (当前)
curl http://localhost:6767/health
```

### API 调用

```bash
# 获取任务列表
curl http://localhost:6767/api/v1/tasks

# 获取 Worker 状态
curl http://localhost:6767/api/v1/workers

# 创建任务
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","description":"npm test"}'
```

## Docker Compose 完整示例

```yaml
version: '3.8'

services:
  master:
    image: agentflow/master:latest
    container_name: agentflow-master
    ports:
      - "6767:6767"
    environment:
      - AGENTFLOW_MASTER_URL=http://localhost:6767
    volumes:
      - agentflow-data:/data
    restart: unless-stopped

  worker:
    image: agentflow/worker:latest
    container_name: agentflow-worker
    environment:
      - AGENTFLOW_MASTER_URL=http://master:6767
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    depends_on:
      - master
    restart: unless-stopped

volumes:
  agentflow-data:
```

## Kubernetes 完整示例

```yaml
apiVersion: v1
kind: Service
metadata:
  name: agentflow-master
  namespace: agentflow
spec:
  selector:
    app: agentflow-master
  ports:
  - port: 6767
    targetPort: 6767
  type: ClusterIP

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agentflow-master
  namespace: agentflow
spec:
  replicas: 1
  selector:
    matchLabels:
      app: agentflow-master
  template:
    metadata:
      labels:
        app: agentflow-master
    spec:
      containers:
      - name: master
        image: agentflow/master:latest
        ports:
        - containerPort: 6767
        env:
        - name: AGENTFLOW_MASTER_URL
          value: "http://agentflow-master:6767"
```

## 常见问题 (FAQ)

### Q1: 为什么要更改端口？

**A**: 端口 6767 更易于记忆（重复数字），且更不容易与其他服务冲突。

### Q2: 旧版本还能用吗？

**A**: AgentFlow 2.0.0 统一使用 6767 端口。旧版本（1.x）仍使用 8848，但不推荐使用。

### Q3: 如何处理多个环境？

**A**: 可以通过环境变量配置不同端口：

```bash
# 开发环境
export AGENTFLOW_MASTER_URL="http://localhost:6767"

# 生产环境
export AGENTFLOW_MASTER_URL="http://prod-agentflow:6767"
```

### Q4: 端口被占用怎么办？

**A**: 可以在启动时指定其他端口：

```bash
# 设置环境变量
export AGENTFLOW_PORT=8080

# 启动 Master
node nodejs/packages/master/dist/index.js
```

### Q5: 需要重启服务吗？

**A**: 是的，需要重启所有使用旧端口的 Master 和 Worker 服务。

## 回滚方案

如果需要回滚到旧端口：

```bash
# 方式 1: 环境变量
export AGENTFLOW_MASTER_URL="http://localhost:8848"

# 方式 2: 配置文件
# 修改所有 6767 回 8848

# 方式 3: 重新构建
# 使用旧版本代码或配置
```

## 验证清单

迁移后请验证：

- [ ] 环境变量已更新
- [ ] Master 服务正常启动（端口 6767）
- [ ] Worker 能连接到 Master
- [ ] 健康检查正常：`curl http://localhost:6767/health`
- [ ] API 调用正常
- [ ] Docker 部署正常
- [ ] Kubernetes 部署正常
- [ ] 防火墙规则已更新
- [ ] 文档已更新
- [ ] 团队成员已通知

## 联系支持

如有问题，请：

1. 查看完整文档：[docs/](docs/)
2. 提交 Issue：[GitHub Issues](https://github.com/MoSiYuan/AgentFlow/issues)
3. 查看部署指南：[deployment/README.md](deployment/README.md)

---

**AgentFlow 版本**: 2.0.0
**更新日期**: 2026-01-23
**变更类型**: 破坏性变更 (Breaking Change)
