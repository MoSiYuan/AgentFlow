# AgentFlow Node.js 版本打包和部署指南

**目标**: 无需安装依赖，开箱即用

---

## 📦 三种部署方案

### 方案对比

| 方案 | 优点 | 缺点 | 推荐场景 |
|------|------|------|---------|
| **1. 独立可执行文件** | 无需 Node.js，单文件，启动快 | 较大 (~100MB) | 生产环境 |
| **2. 完整依赖包** | 包含所有文件，易于调试 | 需要Node.js | 开发/测试 |
| **3. Docker 镜像** | 完全隔离，易部署 | 需要 Docker | 云部署 |

---

## 🚀 方案 1: 独立可执行文件 (推荐)

### 优势

- ✅ **无需安装 Node.js**
- ✅ **无需安装依赖**
- ✅ **单文件部署**
- ✅ **跨平台支持** (macOS, Linux, Windows)

### 打包步骤

#### 1. 安装 pkg

```bash
# 全局安装
npm install -g pkg

# 或作为开发依赖
npm install --save-dev pkg
```

#### 2. 构建项目

```bash
cd nodejs

# 安装依赖
pnpm install

# 重新编译 better-sqlite3
npm rebuild better-sqlite3

# 构建项目
pnpm run build
```

#### 3. 打包

```bash
# 使用打包脚本（推荐）
chmod +x package.sh
./package.sh

# 或使用 npm script
npm run package:standalone
```

#### 4. 生成的文件

```
dist/
├── agentflow-master           # macOS/Linux 可执行文件
└── agentflow-master.exe       # Windows 可执行文件
```

### 部署

```bash
# 1. 复制可执行文件到目标服务器
scp dist/agentflow-master user@server:/usr/local/bin/

# 2. 直接运行（无需安装任何依赖）
agentflow-master

# 3. 使用自定义配置
agentflow-master --port 6767 --db /data/agentflow.db
```

### 使用示例

```bash
# 启动 Master
./agentflow-master

# 启动 Worker
./agentflow-worker

# 查看帮助
./agentflow-master --help
```

---

## 📦 方案 2: 完整依赖包

### 优势

- ✅ **包含所有依赖**
- ✅ **易于调试**
- ✅ **可以修改代码**

### 打包步骤

```bash
cd nodejs

# 运行打包脚本
./package.sh

# 生成的文件在 dist/bundle/
```

### 目录结构

```
dist/bundle/
├── master/              # Master 编译输出
├── worker/              # Worker 编译输出
├── database/            # Database package
├── shared/              # 共享类型
├── git-integration/     # Git 集成
├── sync/                # 状态同步
├── query/               # 统一查询
├── cli/                 # CLI 工具
├── node_modules/        # 所有依赖
├── start-master.sh      # Master 启动脚本
└── start-worker.sh      # Worker 启动脚本
```

### 部署

```bash
# 1. 打包
tar -czf agentflow-bundle.tar.gz -C dist/bundle .

# 2. 上传到服务器
scp agentflow-bundle.tar.gz user@server:/tmp/

# 3. 解压
ssh user@server
cd /opt
tar -xzf /tmp/agentflow-bundle.tar.gz
cd bundle

# 4. 启动
./start-master.sh
```

---

## 🐳 方案 3: Docker 镜像

### 优势

- ✅ **完全隔离**
- ✅ **易于部署和扩展**
- ✅ **支持多实例**
- ✅ **数据持久化**

### 快速开始

```bash
cd nodejs

# 构建镜像
npm run docker:build
# 或
docker build -f Dockerfile.standalone -t agentflow:latest .

# 运行容器
npm run docker:run
# 或
docker run -d \
  --name agentflow-master \
  -p 6767:6767 \
  -v ~/.claude:/root/.claude:ro \
  -v agentflow-data:/data \
  agentflow:latest
```

### 使用 Docker Compose

```bash
# 启动所有服务（Master + 2个Workers）
npm run docker:compose
# 或
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
npm run docker:compose:down
# 或
docker-compose down
```

### 部署到生产环境

```bash
# 1. 导出镜像
docker save agentflow:latest | gzip > agentflow-image.tar.gz

# 2. 传输到服务器
scp agentflow-image.tar.gz user@server:/tmp/

# 3. 导入镜像
ssh user@server
docker load < /tmp/agentflow-image.tar.gz

# 4. 运行容器
docker run -d \
  --name agentflow-master \
  --restart unless-stopped \
  -p 6767:6767 \
  -v /var/lib/agentflow:/data \
  -v /root/.claude:/root/.claude:ro \
  agentflow:latest
```

---

## 🎯 快速部署指南

### 场景 1: 本地开发

```bash
# 使用完整依赖包
cd nodejs
./package.sh
cd dist/bundle
./start-master.sh
```

### 场景 2: 单机部署

```bash
# 使用独立可执行文件（推荐）
cd nodejs
./package.sh
cp dist/agentflow-master /usr/local/bin/
agentflow-master
```

### 场景 3: 服务器部署

```bash
# 使用 Docker（推荐）
cd nodejs
npm run docker:build
npm run docker:compose
```

### 场景 4: Kubernetes 集群

```yaml
# 使用部署到 K8s
kubectl apply -f deployment/k8s/
```

---

## 🔧 配置选项

### 环境变量

```bash
# 数据库路径
export AGENTFLOW_DB_PATH="/data/agentflow.db"

# Master URL (Worker)
export AGENTFLOW_MASTER_URL="http://localhost:6767"

# Worker 组名
export AGENTFLOW_GROUP_NAME="production"
```

### 命令行参数

```bash
# Master
agentflow-master --port 6767 --db /data/agentflow.db

# Worker
agentflow-worker --master-url http://localhost:6767
```

---

## 📊 部署检查清单

### 独立可执行文件

- [ ] pkg 已安装
- [ ] 项目已构建 (`pnpm run build`)
- [ ] better-sqlite3 已重新编译
- [ ] 可执行文件已生成
- [ ] 文件权限正确 (`chmod +x`)

### Docker 镜像

- [ ] Docker 已安装
- [ ] 镜像已构建
- [ ] 端口已映射
- [ ] 数据卷已挂载
- [ ] 健康检查已配置

---

## 🧪 验证部署

```bash
# 健康检查
curl http://localhost:6767/health

# 查看状态
curl http://localhost:6767/api/status

# 创建测试任务
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"测试任务","description":"echo Hello World"}'
```

---

## 🔄 更新部署

### 独立可执行文件

```bash
# 1. 停止服务
pkill agentflow-master

# 2. 备份旧版本
mv /usr/local/bin/agentflow-master /usr/local/bin/agentflow-master.bak

# 3. 部署新版本
cp dist/agentflow-master /usr/local/bin/

# 4. 启动服务
agentflow-master
```

### Docker

```bash
# 1. 停止容器
docker-compose down

# 2. 拉取新镜像
docker pull agentflow:latest

# 3. 启动新容器
docker-compose up -d
```

---

## 📈 性能优化

### 独立可执行文件

```bash
# 使用压缩
pkg . -C Brotli --compress GZip

# 减小体积
pkg . --public --public-packages "better-sqlite3"
```

### Docker

```bash
# 使用多阶段构建
# 使用 alpine 基础镜像
# 启用 BuildKit
DOCKER_BUILDKIT=1 docker build .
```

---

## 🐛 故障排查

### 问题 1: 独立可执行文件无法运行

**症状**: `Cannot find module 'better-sqlite3'`

**解决方案**:
```bash
# 确保 better-sqlite3 已编译
npm rebuild better-sqlite3
pnpm run build
./package.sh
```

### 问题 2: Docker 容器无法启动

**症状**: `Error: Cannot find module '../src/database'`

**解决方案**:
```bash
# 检查 COPY 指令
docker logs agentflow-master

# 重新构建
docker build --no-cache -f Dockerfile.standalone -t agentflow:latest .
```

### 问题 3: 数据库文件权限错误

**症状**: `Error: SQLITE_CANTOPEN: unable to open database file`

**解决方案**:
```bash
# 检查权限
ls -la ~/.claude/skills/agentflow/agentflow.db

# 修复权限
chmod 644 ~/.claude/skills/agentflow/agentflow.db
```

---

## 📚 相关文档

- [开发指南](../docs/NODEJS_GUIDE.md)
- [数据库配置](../docs/DATABASE_LOCATION.md)
- [API 文档](../docs/API.md)

---

## 🎉 总结

**推荐方案**:

1. **开发环境**: 使用完整依赖包 (`dist/bundle`)
2. **生产环境**: 使用独立可执行文件 (`dist/agentflow-master`)
3. **云部署**: 使用 Docker 镜像

**一键部署命令**:

```bash
# 独立可执行文件
cd nodejs && ./package.sh && sudo cp dist/agentflow-master /usr/local/bin/

# Docker
cd nodejs && npm run docker:compose
```

---

**最后更新**: 2026-01-26
**维护者**: AgentFlow Team
