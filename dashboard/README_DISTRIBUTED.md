# AgentFlow Dashboard

AgentFlow 分布式执行系统的 Web 管理界面

## ✨ 功能特性

- 🏛️ **集群拓扑** - 实时可视化 Master 集群、Worker 节点和连接关系
- ⚙️ **Worker 监控** - 实时监控 Worker 状态、资源使用和负载情况
- 📊 **工作流管理** - 创建、查看和管理 DAG 工作流
- 🔒 **分布式锁** - 管理和监控分布式锁状态
- 📋 **任务列表** - 查看和管理任务执行状态
- 📜 **日志流** - 实时查看系统日志

## 🚀 快速开始

### 1. 安装依赖

```bash
cd dashboard
npm install
```

### 2. 配置环境变量

编辑 `.env` 文件：

```bash
VITE_API_URL=http://localhost:6767
VITE_WS_URL=ws://localhost:8849
```

### 3. 启动开发服务器

```bash
npm run dev
```

Dashboard 将在 `http://localhost:5173` 启动。

### 4. 构建生产版本

```bash
npm run build
npm run preview
```

## 📦 依赖项

主要依赖：
- **React 19** - UI 框架
- **TypeScript** - 类型安全
- **Ant Design 5** - UI 组件库
- **ReactFlow** - 节点拓扑图
- **Vite** - 构建工具

## 🎨 组件结构

```
src/
├── components/
│   ├── DashboardLayout.tsx    # 主布局（标签页导航）
│   ├── ClusterTopology.tsx     # 集群拓扑图
│   ├── WorkerMonitor.tsx       # Worker 监控
│   ├── WorkflowManager.tsx     # 工作流管理
│   ├── LockManager.tsx         # 分布式锁管理
│   ├── TaskList.tsx            # 任务列表
│   └── LogStream.tsx           # 日志流
├── services/
│   └── api.ts                  # API 服务层
└── types/
    ├── distributed.ts          # 分布式系统类型
    └── messages.ts             # 消息类型
```

## 🔌 API 集成

Dashboard 通过 REST API 与 AgentFlow Master 通信：

```typescript
// 集群状态
GET /api/v1/cluster/status
GET /api/v1/cluster/leader

// Worker 管理
GET /api/v1/workers
PATCH /api/v1/workers/:id

// 工作流
GET /api/v1/workflows
POST /api/v1/workflows
GET /api/v1/workflows/:name/graph

// 任务
GET /api/v1/tasks
POST /api/v1/tasks
PATCH /api/v1/tasks/:id/priority

// 分布式锁
GET /api/v1/locks
POST /api/v1/locks/acquire
POST /api/v1/locks/release
```

## 🎯 使用场景

### 场景 1: 监控集群健康

1. 打开"集群拓扑"标签页
2. 查看 Leader 状态和节点连接
3. 实时观察节点健康状态

### 场景 2: 管理 Workers

1. 打开"Worker 监控"标签页
2. 查看所有 Worker 的状态和资源使用
3. 调整 Worker 状态（Active/Busy/Draining）

### 场景 3: 创建工作流

1. 打开"工作流管理"标签页
2. 点击"创建工作流"
3. 输入工作流名称和任务列表（JSON 格式）
4. 查看执行顺序和关键路径

### 场景 4: 管理分布式锁

1. 打开"分布式锁"标签页
2. 查看所有活跃的锁
3. 获取新锁或释放现有锁

## 🔧 开发

### 添加新组件

1. 在 `src/components/` 创建新组件
2. 在 `DashboardLayout.tsx` 中添加标签页
3. 在 `src/services/api.ts` 添加 API 调用

### 添加新类型

在 `src/types/distributed.ts` 中添加类型定义。

### 环境变量

所有环境变量定义在 `.env` 文件中，通过 `import.meta.env.VITE_*` 访问。

## 🐛 故障排查

### 问题 1: 无法连接到 API

**症状**: Dashboard 显示错误，无法加载数据

**解决方案**:
1. 确认 AgentFlow Master 正在运行：`curl http://localhost:6767/health`
2. 检查 `.env` 中的 `VITE_API_URL` 配置
3. 查看浏览器控制台的错误信息

### 问题 2: 节点拓扑图不显示

**症状**: 集群拓扑页面空白

**解决方案**:
1. 检查 Master 集群是否已启动
2. 确认至少有 3 个 Master 节点在运行
3. 查看浏览器控制台的 API 响应

### 问题 3: 实时更新不工作

**症状**: 数据不自动刷新

**解决方案**:
1. 检查 WebSocket 连接状态
2. 确认 `.env` 中的 `VITE_WS_URL` 配置正确
3. 查看 Network 面板的 WebSocket 连接

## 📊 性能优化

- 使用 React.memo 减少不必要的重渲染
- API 调用使用防抖和节流
- 虚拟化长列表（如果需要）
- 图表数据缓存

## 🚢 部署

### Docker 部署

```dockerfile
FROM node:20-alpine as builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
```

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agentflow-dashboard
spec:
  replicas: 2
  selector:
    matchLabels:
      app: agentflow-dashboard
  template:
    metadata:
      labels:
        app: agentflow-dashboard
    spec:
      containers:
      - name: dashboard
        image: agentflow-dashboard:latest
        ports:
        - containerPort: 80
        env:
        - name: VITE_API_URL
          value: "http://agentflow-master:6767"
```

## 📝 License

MIT

---

**Made with ❤️ by the AgentFlow Team**
