# Node.js SQLite 依赖更新指南

**日期**: 2026-01-23
**目的**: 更新 Node.js 版本的 SQLite 依赖以支持最新的 Node.js 版本

## 当前状态

### 现有依赖
```json
{
  "better-sqlite3": "12.6.2",
  "node-sqlite3": "^0.0.3"
}
```

### 问题分析

**Node.js v24 兼容性**:
- better-sqlite3 12.6.2: ❌ Node.js v24 无预构建二进制
- 需要本地编译，但 C++ API 不兼容

## 解决方案

### 方案 1: 更新到最新 better-sqlite3（推荐）

```bash
cd nodejs/packages/database

# 安装最新版本
pnpm install better-sqlite3@latest

# 或指定版本
pnpm install better-sqlite3@^14.0.0  # 假设最新是 14.x
```

**优势**:
- ✅ 可能包含 Node.js v24 支持
- ✅ 性能改进和 bug 修复
- ✅ 安全更新

**注意事项**:
- ⚠️ 需要验证 API 兼容性
- ⚠️ 可能需要更新类型定义

### 方案 2: 替换为 sql.js（纯 JavaScript）

```bash
cd nodejs/packages/database

# 卸载 better-sqlite3
pnpm remove better-sqlite3

# 安装 sql.js
pnpm add sql.js @types/sql.js
```

**修改代码** (`src/index.ts`):
```typescript
// 旧代码
import Database from 'better-sqlite3';
const db = new Database('data.db');

// 新代码
import initSqlJs from 'sql.js';
const SQL = await initSqlJs();
const db = new SQL.Database();
```

**优势**:
- ✅ 纯 JavaScript，无原生依赖
- ✅ 完全跨平台
- ✅ 支持 Node.js 所有版本

**劣势**:
- ❌ 性能较低（约 50%）
- ❌ API 不兼容，需要重构代码

### 方案 3: 使用 node-sqlite3（备选）

```bash
cd nodejs/packages/database

# 使用 node-sqlite3 作为主数据库
pnpm remove better-sqlite3
pnpm add node-sqlite3@latest
```

**注意事项**:
- ⚠️ API 与 better-sqlite3 不同
- ⚠️ 维护较少
- ⚠️ 性能较差

### 方案 4: 保持 Go 版本作为主实现（最推荐）

**现状**:
- ✅ Go 版本已完美工作
- ✅ 零依赖，无兼容性问题
- ✅ 性能更优

**建议**:
- Node.js 版本仅作为参考实现
- 主推 Go 版本用于生产
- Node.js 版本用于开发和调试

## 推荐行动

### 立即行动（如果必须使用 Node.js）

1. **安装 Node.js 20 LTS**（最简单）
   ```bash
   # 安装 Node.js 20
   brew install node@20

   # 切换到 Node.js 20
   node --version  # 应该显示 v20.x.x

   # 重新安装依赖
   cd nodejs
   pnpm install
   pnpm run build
   ```

2. **测试运行**
   ```bash
   node packages/master/dist/index.js
   ```

### 短期计划（1周内）

1. **验证 better-sqlite3 更新**
   - 检查是否支持 Node.js v24
   - 测试 API 兼容性
   - 运行完整测试套件

2. **评估 sql.js 迁移**
   - 性能测试
   - 代码重构工作量
   - 长期维护成本

### 长期计划（1月内）

1. **文档更新**
   - 明确推荐 Go 版本
   - 说明 Node.js 版本限制
   - 提供版本兼容性矩阵

2. **CI/CD 配置**
   - 测试多个 Node.js 版本
   - 自动化构建和测试
   - 发布兼容性警告

## 兼容性矩阵

| Node.js 版本 | better-sqlite3 12.6.2 | better-sqlite3 最新 | sql.js |
|-------------|---------------------|-------------------|--------|
| 18.x | ✅ 支持 | ✅ 可能支持 | ✅ 支持 |
| 20.x | ✅ 支持 | ✅ 可能支持 | ✅ 支持 |
| 22.x | ⚠️ 可能支持 | ✅ 可能支持 | ✅ 支持 |
| 24.x | ❌ 不支持 | ❓ 待验证 | ✅ 支持 |

## 测试清单

更新依赖后，请测试：

- [ ] Master 启动
- [ ] 任务创建
- [ ] 任务执行
- [ ] 数据库持久化
- [ ] API 响应
- [ ] 错误处理

## 代码更新示例

### 更新 package.json
```json
{
  "dependencies": {
    "better-sqlite3": "^14.0.0"
  },
  "devDependencies": {
    "@types/better-sqlite3": "^7.6.10"
  }
}
```

### 重新构建
```bash
cd nodejs
pnpm install
pnpm run build
```

### 测试
```bash
node packages/master/dist/index.js
```

## 总结

**最佳方案**: 使用 Go 版本
**备选方案**: Node.js 20 LTS + better-sqlite3
**实验方案**: sql.js（纯 JavaScript）

---

**状态**: 等待 better-sqlite3 对 Node.js v24 的支持
**推荐**: 主推 Go 版本，Node.js 版本降级到 Node.js 20
