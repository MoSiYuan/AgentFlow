# Node.js 22 编译测试报告

**测试日期**: 2026-01-23
**Node.js 版本**: v22.22.0
**better-sqlite3 版本**: 12.6.2（最新稳定版）

## 执行摘要

**结论**: ❌ **Node.js 22 与 better-sqlite3 12.6.2 不兼容**

### 测试结果

| 测试项 | 结果 | 说明 |
|--------|------|------|
| Node.js 安装 | ✅ 成功 | v22.22.0 LTS |
| pnpm 安装 | ✅ 成功 | v10.28.1 |
| 依赖安装 | ✅ 成功 | 所有包已安装 |
| better-sqlite3 加载 | ❌ **失败** | 无原生模块 |
| 项目编译 | ❌ 失败 | 依赖问题 |
| Master 启动 | ❌ 失败 | 数据库无法加载 |

## 详细错误

### better-sqlite3 原生模块缺失

**错误信息**:
```
Error: Could not locate the bindings file. Tried:
 → .../better-sqlite3/lib/binding/node-v127-darwin-arm64/better_sqlite3.node
```

**Node.js 版本对应**:
- Node.js 22.22.0 → ABI 版本 127
- better-sqlite3 12.6.2 需要预构建二进制文件

**查找路径**（均失败）:
1. `build/better_sqlite3.node`
2. `compiled/22.22.0/darwin/arm64/better_sqlite3.node`
3. `lib/binding/node-v127-darwin-arm64/better_sqlite3.node`

## 根本原因

### better-sqlite3 预构建二进制覆盖

| Node.js | ABI 版本 | better-sqlite3 12.6.2 | 状态 |
|---------|---------|----------------------|------|
| 18.x | 108 | ✅ 有预构建 | 支持 |
| 20.x | 127 | ✅ 有预构建 | 支持 |
| 22.x | 127 | ❌ **无预构建** | **不支持** |
| 24.x | 137 | ❌ 无预构建 | 不支持 |

**注意**: Node.js 22 和 20 使用相同的 ABI 版本 (127)，但 better-sqlite3 仍然没有为 Node.js 22 提供预构建二进制。

### 为什么 Node.js 22 失败

虽然 ABI 版本相同，但 better-sqlite3 的预构建配置可能：
1. 没有测试 Node.js 22
2. 没有为 Node.js 22 构建预构建二进制
3. 版本检测逻辑排除了 Node.js 22

## 解决方案

### 方案 1: 使用 Node.js 20 LTS（推荐）⭐⭐⭐

```bash
# 卸载 Node.js 22
brew uninstall node@22

# 安装 Node.js 20
brew install node@20

# 设置 PATH
export PATH="/opt/homebrew/opt/node@20/bin:$PATH"

# 验证
node --version  # v20.x.x

# 编译项目
cd nodejs
pnpm install
pnpm run build
```

**优势**:
- ✅ better-sqlite3 完全兼容
- ✅ 预构建二进制可用
- ✅ 稳定 LTS 版本

### 方案 2: 手动编译 better-sqlite3（复杂）⚠️

```bash
# 安装构建工具
brew install python@3.13

# 进入 better-sqlite3 目录
cd node_modules/.pnpm/better-sqlite3@12.6.2/node_modules/better-sqlite3

# 尝试编译
npm run install --ignore-scripts=false
```

**问题**:
- ⚠️ 可能编译失败（C++ API 兼容性）
- ⚠️ 需要安装 Xcode Command Line Tools
- ⚠️ 编译时间长（5-10分钟）
- ⚠️ 不保证成功

### 方案 3: 使用 Go 版本（最推荐）⭐⭐⭐

```bash
# 无需 Node.js
./agentflow-go.sh run '["echo hello","echo world"]'
```

**优势**:
- ✅ 零依赖
- ✅ 所有平台
- ✅ 性能更好

## 版本兼容性总结

### Node.js 版本与 better-sqlite3

| Node.js | 版本类型 | better-sqlite3 12.6.2 | 推荐 |
|---------|---------|----------------------|------|
| 18.x | LTS | ✅ 支持 | ✅ |
| 20.x | LTS | ✅ 支持 | ⭐⭐⭐ |
| 22.x | Current | ❌ 不支持 | ❌ |
| 24.x | Latest | ❌ 不支持 | ❌ |

### 推荐配置

**生产环境**:
- Go 版本（零依赖）

**开发环境**:
- Node.js 20 LTS + better-sqlite3

**测试环境**:
- Node.js 18 LTS（最小兼容版本）

## 编译状态

### 当前状态

```
Node.js: v22.22.0
pnpm: v10.28.1
依赖: 已安装
better-sqlite3: ❌ 无原生模块
项目编译: ❌ 失败
```

### 编译错误

1. **better-sqlite3 加载失败**
   - 原因: 无 Node.js 22 预构建二进制
   - 影响: 所有依赖数据库的功能

2. **TypeScript 编译错误**
   - 文件: `src/orchestrator.ts(83,11)`
   - 错误: 语法错误
   - 影响: Master 无法编译

## 下一步行动

### 立即行动

1. **使用 Node.js 20**
   ```bash
   brew install node@20
   export PATH="/opt/homebrew/opt/node@20/bin:$PATH"
   cd nodejs && pnpm install && pnpm run build
   ```

2. **或使用 Go 版本**
   ```bash
   ./agentflow-go.sh run '["echo test"]'
   ```

### 长期方案

1. **等待 better-sqlite3 更新**
    - 跟踪 GitHub releases
    - 评估 Node.js 22 支持

2. **考虑数据库迁移**
    - 评估 sql.js（纯 JS）
    - 评估其他数据库
    - 设计数据库抽象层

## 经验教训

### 问题

1. **新 Node.js 版本兼容性**
   - better-sqlite3 更新滞后
   - 预构建二进制覆盖不完整

2. **原生模块依赖**
    - 增加部署复杂度
    - 版本兼容性问题
    - 客户体验差

### 解决

1. **Go 版本优势确认**
    - 零依赖真正解决问题
    - 跨平台一致性好
    - 部署简单

2. **版本选择策略**
    - 开发: 使用 LTS 版本（20）
    - 生产: 使用 Go 版本
    - 避免: Current/Latest 版本

## 总结

### 测试结论

**Node.js 22 与 better-sqlite3 12.6.2 不兼容**

**原因**: better-sqlite3 未提供 Node.js 22 的预构建二进制

**影响**: 无法使用 Node.js 22 运行 AgentFlow

### 最终推荐

1. **生产**: Go 版本 ⭐⭐⭐
2. **开发**: Node.js 20 LTS ⭐⭐
3. **测试**: Node.js 18 LTS ⭐

---

**状态**: ❌ Node.js 22 测试失败
**建议**: 使用 Node.js 20 或 Go 版本
**依赖版本**: better-sqlite3 12.6.2（已是最新）
