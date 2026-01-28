# AgentFlow 部署配置示例

本目录包含 AgentFlow 的各种部署配置示例，可直接复制使用。

## 📁 目录结构

```
deployment/examples/
├── systemd/          # systemd 服务配置
├── docker/           # Docker 容器配置
├── kubernetes/       # Kubernetes 集群配置
├── nginx/            # Nginx 反向代理配置
└── scripts/          # 自动化部署脚本
```

## 🚀 快速开始

### 方式 1: 单机 systemd 部署

适合：生产环境、单机部署、自动启动

```bash
# 1. 编译项目
cd rust
cargo build --release

# 2. 运行安装脚本
cd ../deployment/examples/scripts
sudo ./install.sh

# 3. 修改配置
sudo vi /etc/default/agentflow
# 修改 AUTH_PASSWORD 和 AUTH_API_KEY_SECRET

# 4. 启动服务
sudo systemctl start agentflow-master
sudo systemctl enable agentflow-master

# 5. 检查状态
sudo systemctl status agentflow-master
```

### 方式 2: Docker 部署

适合：容器化环境、快速部署、开发测试

```bash
# 1. 构建镜像
cd deployment/examples/docker
docker build -t agentflow-master:v1.0 .

# 2. 运行容器
docker-compose up -d

# 3. 查看日志
docker-compose logs -f agentflow-master

# 4. 停止容器
docker-compose down
```

### 方式 3: Kubernetes 集群部署

适合：云原生环境、大规模部署、高可用

```bash
# 1. 修改 Secret 配置
cd deployment/examples/kubernetes
vi secret.yaml
# 修改 AUTH_PASSWORD 和 AUTH_API_KEY_SECRET

# 2. 应用所有配置
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml

# 3. 查看状态
kubectl get pods -n agentflow
kubectl get svc -n agentflow

# 4. 查看日志
kubectl logs -f deployment/agentflow-master -n agentflow
```

### 方式 4: 单机多实例集群部署

适合：高并发场景、负载均衡、充分利用多核

```bash
# 1. 编译项目
cd rust
cargo build --release

# 2. 运行集群部署脚本
cd ../deployment/examples/scripts
sudo ./deploy-cluster.sh

# 3. 配置 Nginx 负载均衡
sudo cp ../nginx/agentflow.conf /etc/nginx/sites-available/agentflow
sudo ln -s /etc/nginx/sites-available/agentflow /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# 4. 检查所有实例
sudo systemctl status 'agentflow-master@*'
```

## 📝 配置说明

### systemd 服务配置

**文件**: `systemd/agentflow-master.service`

包含：
- 服务定义
- 资源限制（100MB 内存，50% CPU）
- 安全加固（NoNewPrivileges, ProtectSystem）
- 自动重启策略
- 日志输出到 journald

**环境变量**: `systemd/agentflow.env`

关键配置：
- `AUTH_ENABLED`: 启用认证
- `AUTH_PASSWORD`: 管理员密码
- `AUTH_API_KEY_SECRET`: API Key 签名密钥（至少 32 字符）

### Docker 配置

**Dockerfile**: `docker/Dockerfile`

特点：
- 多阶段构建，最小化镜像大小
- Alpine 基础镜像
- 非 root 用户运行
- 内置健康检查

**docker-compose.yml**: `docker/docker-compose.yml`

功能：
- 数据持久化（volumes）
- 资源限制
- 健康检查
- 日志管理（自动轮转）

### Kubernetes 配置

包含完整的 K8s 资源清单：

- `namespace.yaml`: 命名空间
- `configmap.yaml`: 配置项
- `secret.yaml`: 敏感信息（需修改）
- `deployment.yaml`: 部署配置（2 副本）
- `service.yaml`: 服务暴露
- `ingress.yaml`: Ingress 路由（需修改域名）

**注意事项**：
- 修改 `secret.yaml` 中的密码和密钥
- 修改 `ingress.yaml` 中的域名
- 配置 cert-manager 可自动签发 TLS 证书

### Nginx 配置

**文件**: `nginx/agentflow.conf`

功能：
- HTTPS/TLS 配置
- 安全头（HSTS, X-Frame-Options）
- 请求限流（10 req/s）
- WebSocket 支持
- API 代理
- 健康检查

**限流配置**（需在 http 块中添加）：

```nginx
http {
    limit_req_zone $binary_remote_addr zone=agentflow:10m rate=10r/s;
}
```

## 🔧 高级配置

### 资源限制调整

**systemd**:
编辑 `agentflow-master.service` 中的：
- `MemoryMax`: 内存限制
- `CPUQuota`: CPU 限制

**Docker**:
编辑 `docker-compose.yml` 中的 `deploy.resources` 部分

**Kubernetes**:
编辑 `deployment.yaml` 中的 `resources` 部分

### 数据持久化

**systemd**: 数据存储在 `/var/lib/agentflow`

**Docker**: 使用 named volume `agentflow-data`

**Kubernetes**:
- 默认使用 `emptyDir`（Pod 删除后数据丢失）
- 生产环境建议使用 PVC：

```yaml
volumeClaimTemplates:
- metadata:
    name: data
  spec:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 10Gi
```

### 高可用配置

**单机多实例**：
使用 `scripts/deploy-cluster.sh` 部署多个实例

**Kubernetes 多副本**：
编辑 `deployment.yaml` 中的 `replicas` 数量

**多机集群**：
参考 [集群部署指南](../../docs/CLUSTERING.md)

## 🔍 监控和日志

### systemd 日志

```bash
# 实时查看
sudo journalctl -u agentflow-master -f

# 查看最近 100 行
sudo journalctl -u agentflow-master -n 100

# 查看启动错误
sudo journalctl -u agentflow-master -b
```

### Docker 日志

```bash
# 查看容器日志
docker logs -f agentflow-master

# 查看最近日志
docker logs --tail 100 agentflow-master
```

### Kubernetes 日志

```bash
# 查看 Pod 日志
kubectl logs -f deployment/agentflow-master -n agentflow

# 查看所有 Pod 状态
kubectl get pods -n agentflow -w
```

### 日志文件

```bash
# 查看 AgentFlow 日志
sudo tail -f /var/log/agentflow/agentflow.log
```

## 🔒 安全建议

1. **修改默认密码**
   - systemd: 编辑 `/etc/default/agentflow`
   - Docker: 编辑 `docker-compose.yml` 环境变量
   - Kubernetes: 编辑 `secret.yaml`

2. **使用 HTTPS**
   - 配置 Nginx 反向代理
   - 申请 Let's Encrypt 证书

3. **限制访问**
   - 配置防火墙（只开放必要端口）
   - 使用 Nginx 限流
   - 配置 IP 白名单

4. **定期备份**
   - 备份 `/var/lib/agentflow` 数据目录
   - 备份配置文件
   - 备份日志

5. **更新升级**
   - 定期更新到最新版本
   - 关注安全公告

## 📖 更多文档

- [生产部署指南](../../docs/DEPLOYMENT.md) - 详细部署说明
- [集群部署指南](../../docs/CLUSTERING.md) - 集群架构和方案
- [技能集成指南](../../docs/SKILL_INTEGRATION.md) - 系统集成说明
- [系统架构](../../docs/ARCHITECTURE.md) - 深入理解架构

## 🆘 故障排查

### 服务启动失败

```bash
# 检查服务状态
sudo systemctl status agentflow-master

# 查看详细日志
sudo journalctl -u agentflow-master -n 50 --no-pager

# 检查配置文件
sudo /opt/agentflow/bin/agentflow-master --help
```

### 端口占用

```bash
# 检查端口占用
sudo netstat -tulpn | grep 6767

# 修改端口
sudo vi /etc/default/agentflow
# 修改 AGENTFLOW_SERVER_PORT
```

### 权限问题

```bash
# 检查文件权限
ls -la /opt/agentflow/bin/
ls -la /var/lib/agentflow

# 修复权限
sudo chown -R agentflow:agentflow /opt/agentflow
sudo chown -R agentflow:agentflow /var/lib/agentflow
```

## 📞 获取帮助

遇到问题？
- 提交 Issue: https://github.com/MoSiYuan/AgentFlow/issues
- 查看文档: [项目 README](../../README.md)

---

**部署配置版本**: 1.0.0
**最后更新**: 2026-01-28
