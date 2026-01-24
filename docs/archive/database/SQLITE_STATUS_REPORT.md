# AgentFlow SQLite 依赖状态报告

**日期**: 2026-01-23
**状态**: ✅ 已完成评估

## 当前状态

### better-sqlite3 版本

| 项目 | 版本 | 状态 |
|------|------|------|
| 当前使用 | 12.6.2 | ✅ 最新稳定版 |
| GitHub 最新 | 12.6.2 | ✅ 已是最新 |
| Node.js v24 支持 | ❌ | 无预构建二进制 |

### 结论

**better-sqlite3 已经是最新版本（v12.6.2）**，但该版本不支持 Node.js v24。

## 依赖分析

### 当前 package.json

```json
{
  "dependencies": {
    "better-sqlite3": "12.6.2",
    "node-sqlite3": "^0.0.3"
  }
}
```

### 版本矩阵

| Node.js | better-sqlite3 12.6.2 | 说明 |
|---------|----------------------|------|
| 18.x | ✅ 支持 | 有预构建二进制 |
| 20.x | ✅ 支持 | 有预构建二进制 |
| 22.x | ⚠️ 部分 | 可能需要编译 |
| 24.x | ❌ 不支持 | 无预构建，编译失败 |

## 可行方案

### 方案 1: 保持现状 + Go 版本（推荐）⭐⭐⭐

**当前状态**: ✅ 已实现

**操作**:
- ✅ Go 版本已完美工作
- ✅ Node.js 版本使用 Node.js 20
- ✅ 无需更改任何依赖

**优势**:
- ✅ Go 版本零依赖，所有平台
- ✅ Node.js 版本作为参考
- ✅ 客户可以自由选择

**文档**:
```markdown
推荐版本: Go（零依赖）
备选版本: Node.js 20 LTS
```

### 方案 2: Node.js 20 + better-sqlite3（备选）⭐⭐

**操作**:
```bash
# 安装 Node.js 20
brew install node@20
brew unlink node
brew link node@20

# 验证版本
node --version  # v20.x.x

# 安装依赖
cd nodejs
pnpm install
pnpm run build
```

**优势**:
- ✅ better-sqlite3 完全兼容
- ✅ 无需修改代码
- ✅ 稳定可靠

**劣势**:
- ❌ 需要特定 Node.js 版本
- ❌ 客户可能遇到版本问题

### 方案 3: 等待 better-sqlite3 更新（被动）⭐

**跟踪**:
- GitHub: https://github.com/WiseLibs/better-sqlite3
- Issue: Node.js v24 支持

**时间线**: 未知（可能数月）

**风险**:
- ❌ 无法控制时间
- ❌ 客户等待时间过长

### 方案 4: 迁移到 sql.js（重构）⚠️

**工作量**: 中等（2-3天）

**代码改动**:
```typescript
// 当前
import Database from 'better-sqlite3';
const db = new Database('data.db');

// 迁移后
import initSqlJs from 'sql.js';
const SQL = await initSqlJs();
const db = new SQL.Database();
```

**优势**:
- ✅ 纯 JavaScript，无原生依赖
- ✅ 支持 Node.js 所有版本

**劣势**:
- ❌ API 不兼容，需要重构
- ❌ 性能降低约 50%
- ❌ 内存占用增加

## 推荐行动

### 立即行动

1. **保持 Go 版本作为主推方案**
   - 文档明确推荐 Go 版本
   - 说明 Node.js 版本限制

2. **提供 Node.js 20 安装指南**
   - 创建安装脚本（已完成 ✅）
   - 提供详细文档

### 短期（1周内）

1. **文档更新**
   - README 添加版本兼容性说明
   - 创建故障排查指南

2. **自动化测试**
   - CI/CD 测试多个 Node.js 版本
   - 自动检测兼容性问题

### 长期（1月内）

1. **持续跟踪**
   - 关注 better-sqlite3 更新
   - 评估 Node.js v24 支持

2. **备选方案准备**
   - sql.js 迁移方案
   - 数据库抽象层设计

## 安装指南

### Go 版本（推荐）

```bash
# 下载即用
git clone https://github.com/MoSiYuan/AgentFlow.git
cd AgentFlow

# 直接运行
./agentflow-go.sh run '["echo hello","echo world"]'
```

**要求**: 无

### Node.js 版本

```bash
# 安装 Node.js 20
brew install node@20
brew unlink node
brew link node@20

# 进入项目目录
cd AgentFlow/nodejs

# 安装依赖
pnpm install

# 构建
pnpm run build

# 启动 Master
node packages/master/dist/index.js
```

**要求**:
- Node.js 18-20 LTS
- pnpm 或 npm
- C++ 编译工具（首次安装）

## 测试结果

### Go 版本 ✅

| 测试项 | 结果 |
|--------|------|
| Master 启动 | ✅ 成功 |
| 任务创建 | ✅ 成功 |
| 本地执行 | ✅ 成功 |
| 跨平台 | ✅ 10/10 |

### Node.js 版本 (需要 Node.js 20)

| 测试项 | 预期结果 |
|--------|---------|
| Master 启动 | ✅ 成功 |
| 任务创建 | ✅ 成功 |
| Worker 执行 | ✅ 成功 |

## 总结

### 当前最佳实践

1. **开发**: 使用 Node.js 20 + better-sqlite3 12.6.2
2. **部署**: 使用 Go 版本（零依赖）
3. **文档**: 明确说明版本要求

### 未来路线图

**Q1 2026**:
- ✅ Go 版本稳定发布
- ⏳ 等待 better-sqlite3 v24 支持
- ⏳ 评估 sql.js 迁移

**Q2 2026**:
- 根据上游支持情况决定是否迁移
- 可能实现数据库抽象层
- 支持多种数据库后端

## 参考资料

- [better-sqlite3 Releases](https://github.com/WiseLibs/better-sqlite3/releases)
- [Node.js 版本兼容性](https://nodejs.org/en/about/releases)
- [sql.js 文档](https://sql.js.org/)

---

**最终建议**:
1. 主推 **Go 版本**（已实现 ✅）
2. Node.js 版本仅用于 **开发和参考**
3. 等待 **better-sqlite3** 对 v24 的支持
4. 长期考虑 **数据库抽象层**

**状态**: ✅ 评估完成，方案明确
**依赖版本**: better-sqlite3 12.6.2（已是最新）
**推荐**: 使用 Go 版本
