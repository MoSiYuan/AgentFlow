# AgentFlow Node.js 版本 - 快速开始

**5分钟快速部署，无需担心依赖**

---

## 🎯 三种快速部署方式

### 方式 1: 独立可执行文件（最简单）⭐

```bash
# 1. 打包（开发者一次性操作）
cd nodejs
chmod +x package.sh
./package.sh

# 2. 复制到目标位置
cp dist/agentflow-master /usr/local/bin/

# 3. 直接运行（无需安装任何依赖）
agentflow-master
```

**优势**:
- ✅ 无需 Node.js
- ✅ 无需 npm/pnpm
- ✅ 单文件，约 100MB
- ✅ 开箱即用

---

### 方式 2: 完整依赖包

```bash
# 1. 打包
cd nodejs
./package.sh

# 2. 解压使用
cd dist/bundle
./start-master.sh
```

**优势**:
- ✅ 包含所有依赖
- ✅ 可以查看和修改代码
- ✅ 适合开发调试

---

### 方式 3: Docker（一键部署）🐳

```bash
cd nodejs

# 一键启动
npm run docker:compose

# 或使用 docker-compose
docker-compose up -d
```

**优势**:
- ✅ 完全隔离
- ✅ 自动重启
- ✅ 数据持久化
- ✅ 支持多实例

---

## 📝 详细步骤

### 开发者：打包发布

```bash
# 1. 准备环境
cd nodejs
pnpm install
npm rebuild better-sqlite3

# 2. 构建项目
pnpm run build

# 3. 打包
chmod +x package.sh
./package.sh

# 生成的文件:
# - dist/agentflow-master (独立可执行文件)
# - dist/bundle/ (完整依赖包)
```

### 用户：部署使用

#### 使用独立可执行文件

```bash
# 1. 下载或复制文件
scp agentflow-master user@server:/usr/local/bin/

# 2. 添加执行权限
chmod +x /usr/local/bin/agentflow-master

# 3. 运行
agentflow-master

# 4. 验证
curl http://localhost:6767/health
```

#### 使用 Docker

```bash
# 1. 构建镜像（开发者）
docker build -f Dockerfile.standalone -t agentflow:latest .

# 2. 导出镜像
docker save agentflow:latest | gzip > agentflow.tar.gz

# 3. 传输到服务器
scp agentflow.tar.gz user@server:/tmp/

# 4. 导入并运行（服务器）
docker load < /tmp/agentflow.tar.gz
docker run -d --name agentflow -p 6767:6767 agentflow:latest
```

---

## 🚀 快速验证

```bash
# 健康检查
curl http://localhost:6767/health

# 查看状态
curl http://localhost:6767/api/status

# 创建测试任务
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"测试任务","description":"echo Hello World"}'

# 查看任务
curl http://localhost:6767/api/v1/tasks/1
```

---

## 📦 文件说明

### 独立可执行文件

```
dist/
├── agentflow-master          # Master 服务器
├── agentflow-master.exe      # Windows 版本
└── agentflow-worker          # Worker 进程
```

### 完整依赖包

```
dist/bundle/
├── master/                   # Master 编译输出
├── worker/                   # Worker 编译输出
├── node_modules/             # 所有依赖（已包含）
├── start-master.sh           # 启动脚本
└── start-worker.sh           # 启动脚本
```

---

## ⚙️ 配置

### 环境变量

```bash
# 数据库路径
export AGENTFLOW_DB_PATH="/data/agentflow.db"

# Master URL (Worker)
export AGENTFLOW_MASTER_URL="http://localhost:6767"
```

### 命令行参数

```bash
# Master
agentflow-master --port 6767 --db /data/agentflow.db

# Worker
agentflow-worker --master-url http://localhost:6767
```

---

## 🎓 常见问题

### Q1: 独立可执行文件需要 Node.js 吗？

**A**: 不需要！独立可执行文件已包含 Node.js 运行时和所有依赖。

### Q2: 文件为什么这么大？

**A**: 因为包含了 Node.js 运行时和所有依赖包（约 100MB）。这是正常的，确保用户无需安装任何依赖。

### Q3: 如何更新？

**A**:
```bash
# 独立可执行文件
pkill agentflow-master
cp new-agentflow-master /usr/local/bin/agentflow-master
agentflow-master

# Docker
docker-compose down
docker-compose pull
docker-compose up -d
```

### Q4: 数据库在哪里？

**A**: 默认位置: `~/.claude/skills/agentflow/agentflow.db`

可通过 `--db` 参数或环境变量自定义。

---

## 📚 更多信息

- [完整打包指南](./PACKAGING_GUIDE.md)
- [开发指南](../docs/NODEJS_GUIDE.md)
- [数据库配置](../docs/DATABASE_LOCATION.md)

---

## 🎉 开始使用

```bash
# 最快的方式（独立可执行文件）
cd nodejs
./package.sh
sudo cp dist/agentflow-master /usr/local/bin/
agentflow-master

# 或使用 Docker（最省心）
cd nodejs
npm run docker:compose
```

---

**最后更新**: 2026-01-26
**版本**: 1.0.0
