# AgentFlow Dashboard

AgentFlow v0.5.0 分布式管理大屏前端项目。

## 技术栈

- **框架**: React 19 + TypeScript
- **构建工具**: Vite
- **样式**: Tailwind CSS
- **状态管理**: Redux Toolkit
- **UI 组件**: Ant Design
- **图表**: @ant-design/charts
- **拓扑图**: ReactFlow
- **通信**: Socket.IO Client

## 项目结构

```
dashboard/
├── src/
│   ├── components/          # React 组件
│   │   ├── DashboardLayout.tsx   # 主布局
│   │   ├── NodeTopology.tsx      # 节点拓扑图
│   │   ├── TaskList.tsx          # 任务列表
│   │   ├── LogStream.tsx         # 日志流
│   │   └── ChatPanel.tsx         # Claude 对话面板
│   ├── hooks/              # 自定义 Hooks
│   │   └── useWebSocket.ts       # WebSocket Hook
│   ├── types/              # TypeScript 类型定义
│   │   └── messages.ts            # 消息类型
│   ├── App.tsx             # 应用入口
│   ├── main.tsx            # React 挂载点
│   └── index.css           # 全局样式
├── public/                 # 静态资源
├── package.json            # 项目依赖
├── vite.config.ts          # Vite 配置
├── tailwind.config.js      # Tailwind CSS 配置
└── tsconfig.json           # TypeScript 配置
```

## 快速开始

### 安装依赖

```bash
cd /Users/jiangxiaolong/work/project/AgentFlow/dashboard
npm install
```

### 开发模式

```bash
npm run dev
```

访问 http://localhost:5173

### 生产构建

```bash
npm run build
```

构建产物在 `dist/` 目录。

### 预览生产构建

```bash
npm run preview
```

## 功能特性

### 已实现

1. **React 项目基础架构**
   - Vite + React + TypeScript
   - Tailwind CSS 样式系统
   - 路径别名 (@/)

2. **WebSocket Hook**
   - 自动重连机制
   - 消息分发系统
   - 类型安全的消息处理

3. **监控大屏布局**
   - 节点拓扑图（ReactFlow）
   - 任务列表（Ant Design Table）
   - 实时日志流
   - 响应式布局

4. **Claude 对话组件**
   - 上下文切换选择器
   - 消息输入和发送
   - 对话历史记录

## WebSocket 消息协议

### 客户端消息

```typescript
// 认证
{ type: 'AUTH'; token: string }

// 聊天
{ type: 'CHAT'; text: string; targetContext: string }

// 订阅主题
{ type: 'SUBSCRIBE'; topics: string[] }
```

### 服务端消息

```typescript
// 节点状态更新
{ type: 'NODE_STATUS'; node: NodeStatus }

// 任务更新
{ type: 'TASK_UPDATE'; task: TaskInfo }

// Claude 响应
{ type: 'CHAT_RESPONSE'; text: string }

// 告警
{ type: 'ALERT'; level: 'error' | 'warning'; message: string }
```

## 使用示例

### 使用 WebSocket Hook

```typescript
import { useWebSocket } from '@/hooks/useWebSocket';

function MyComponent() {
  const { isConnected, on, emit } = useWebSocket({
    url: 'ws://localhost:3000'
  });

  useEffect(() => {
    const cleanup = on('NODE_STATUS', (data) => {
      console.log('Node status:', data);
    });

    return cleanup;
  }, [on]);

  return <div>{isConnected ? '已连接' : '未连接'}</div>;
}
```

### 自定义组件

所有组件使用函数式组件 + Hooks，遵循 React 最佳实践。

## 开发规范

- 使用 TypeScript 严格模式
- 组件使用函数式组件 + Hooks
- 样式优先使用 Tailwind CSS
- UI 组件优先使用 Ant Design
- 类型定义统一放在 `src/types/` 目录

## 构建优化

当前打包大小警告：

```
(!) Some chunks are larger than 500 kB after minification.
```

建议后续优化：
- 使用动态 import() 进行代码分割
- 配置 build.rollupOptions.output.manualChunks

## 依赖安全

```bash
npm audit fix
```

## 许可证

MIT
