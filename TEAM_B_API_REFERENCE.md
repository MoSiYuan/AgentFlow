# Team B API 快速参考

## MemoryCore - 记忆核心系统

### 初始化
```rust
use agentflow_core::memory::MemoryCore;

// 创建记忆核心
let memory = MemoryCore::new("/path/to/memory.db").await?;
```

### 核心操作
```rust
// 索引记忆
memory.index(&entry).await?;

// 语义检索
let results = memory.search("查询文本", 10).await?;

// 获取单个记忆
let entry = memory.get("key").await?;

// 删除记忆
memory.delete("key").await?;

// 清理过期记忆
let count = memory.cleanup_expired().await?;

// 获取统计信息
let stats = memory.stats().await?;
```

---

## SandboxConfig - 沙箱安全控制

### 初始化
```rust
use agentflow_core::sandbox::SandboxConfig;
use std::path::PathBuf;

// 默认配置
let config = SandboxConfig::new();

// 使用白名单创建
let config = SandboxConfig::with_allowed_dirs(vec![
    PathBuf::from("/workspace"),
    PathBuf::from("/tmp/task_work"),
]);
```

### 配置管理
```rust
// 添加允许的目录
config.add_allowed_dir(PathBuf::from("/new/path"));

// 设置严格模式
config.set_strict_mode(true);

// 设置符号链接策略
config.set_allow_symlinks(false);
config.set_max_symlink_depth(8);
```

### 路径验证
```rust
// 验证单个路径
config.validate_path(Path::new("/workspace/file.txt"))?;

// 验证多个路径
config.validate_paths(&[path1, path2, path3])?;

// 创建安全子路径
let safe_path = config.create_safe_path(
    Path::new("/workspace"),
    Path::new("subdir/file.txt")
)?;

// 检查文件名安全性
let is_safe = SandboxConfig::is_safe_filename("file.txt");
```

---

## PromptBuilder - Prompt 构建器

### 初始化
```rust
use agentflow_core::executor::PromptBuilder;

// 默认配置
let builder = PromptBuilder::new();

// 自定义配置
let config = PromptBuilderConfig::new(
    "自定义系统指令".to_string(),
    100_000 // 最大 tokens
);
let builder = PromptBuilder::with_config(config);
```

### 构建 Prompt
```rust
use agentflow_core::types::{MemoryEntry, MemoryCategory};
use serde_json::json;

// 简单构建
let task = "实现用户登录功能";
let memories = vec![/* ... */];
let prompt = builder.build(task, &memories);

// 带元数据构建
let metadata = json!({
    "environment": "production",
    "version": "1.0.0"
});
let prompt = builder.build_with_metadata(task, &memories, &metadata);

// 便捷函数
let prompt = build_prompt(task, &memories);
```

### Token 管理
```rust
// 估算 token 数量
let tokens = builder.estimate_tokens(&prompt);

// 验证长度
builder.validate_length(&prompt)?;

// 获取/设置配置
let config = builder.config();
builder.set_config(new_config);
```

---

## 配置选项

### PromptBuilderConfig
```rust
pub struct PromptBuilderConfig {
    pub system_instruction: String,    // 系统指令
    pub max_tokens: usize,             // 最大 tokens（默认 200K）
    pub include_memory: bool,          // 包含记忆（默认 true）
    pub max_memory_entries: usize,     // 最大记忆数（默认 10）
    pub include_metadata: bool,        // 包含元数据（默认 true）
}
```

### SandboxConfig
```rust
pub struct SandboxConfig {
    allowed_dirs: HashSet<PathBuf>,    // 白名单目录
    strict_mode: bool,                 // 严格模式（默认 true）
    allow_symlinks: bool,              // 允许符号链接（默认 false）
    max_symlink_depth: usize,          // 最大递归深度（默认 8）
}
```

---

## 错误类型

### SandboxError
```rust
pub enum SandboxError {
    PathNotAllowed(String),              // 路径不在白名单
    PathTraversalDetected(String),       // 路径穿透攻击
    SymlinkAttack(String),               // 符号链接攻击
    PathResolutionFailed(String),        // 路径解析失败
    ValidationFailed(String),            // 验证失败
}
```

---

## 完整示例

### 记忆系统使用流程
```rust
use agentflow_core::memory::MemoryCore;
use agentflow_core::types::{MemoryEntry, MemoryCategory};
use serde_json::json;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // 1. 初始化
    let memory = MemoryCore::new("/tmp/memory.db").await?;

    // 2. 创建记忆
    let entry = MemoryEntry {
        key: "task_123".to_string(),
        value: json!("任务完成"),
        expires_at: None,
        category: MemoryCategory::Result,
        task_id: Some("task_123".to_string()),
        timestamp: chrono::Utc::now().timestamp(),
    };

    // 3. 索引
    memory.index(&entry).await?;

    // 4. 检索
    let results = memory.search("任务", 10).await?;

    // 5. 清理
    memory.cleanup_expired().await?;

    Ok(())
}
```

### 沙箱系统使用流程
```rust
use agentflow_core::sandbox::SandboxConfig;
use std::path::{Path, PathBuf};

fn main() -> anyhow::Result<()> {
    // 1. 配置沙箱
    let mut config = SandboxConfig::new();
    config.add_allowed_dir(PathBuf::from("/workspace"));

    // 2. 验证路径
    let path = Path::new("/workspace/file.txt");
    config.validate_path(path)?;

    // 3. 创建安全路径
    let safe_path = config.create_safe_path(
        Path::new("/workspace"),
        Path::new("subdir/file.txt")
    )?;

    // 4. 检查文件名
    let filename = "safe_file.txt";
    if SandboxConfig::is_safe_filename(filename) {
        println!("文件名安全");
    }

    Ok(())
}
```

### Prompt 构建器使用流程
```rust
use agentflow_core::executor::PromptBuilder;
use agentflow_core::types::{MemoryEntry, MemoryCategory};
use serde_json::json;

fn main() {
    // 1. 创建构建器
    let builder = PromptBuilder::new();

    // 2. 准备数据
    let task = "实现用户认证";
    let memories = vec![
        MemoryEntry {
            key: "auth_context".to_string(),
            value: json!("使用 JWT"),
            expires_at: None,
            category: MemoryCategory::Context,
            task_id: None,
            timestamp: 0,
        }
    ];

    // 3. 构建 Prompt
    let prompt = builder.build(task, &memories);

    // 4. 检查长度
    let tokens = builder.estimate_tokens(&prompt);
    println!("Token 数量: {}", tokens);

    // 5. 使用 Prompt
    println!("{}", prompt);
}
```

---

## 类型引用

### MemoryEntry
```rust
pub struct MemoryEntry {
    pub key: String,                          // 唯一键
    pub value: serde_json::Value,             // JSON 值
    pub expires_at: Option<i64>,              // 过期时间戳
    pub category: MemoryCategory,             // 分类
    pub task_id: Option<String>,              // 关联任务 ID
    pub timestamp: i64,                       // 创建时间
}
```

### MemoryCategory
```rust
pub enum MemoryCategory {
    Execution,      // 执行记录
    Context,        // 上下文
    Result,         // 结果
    Error,          // 错误
    Checkpoint,     // 检查点
}
```

---

## 性能考虑

### MemoryCore
- **索引**: O(1)
- **搜索**: O(n) - 可优化为 O(log n)
- **建议**: 定期调用 `cleanup_expired()`

### SandboxConfig
- **验证**: O(p) - p 为路径深度
- **建议**: 缓存验证结果

### PromptBuilder
- **构建**: O(n) - n 为记忆条目数
- **建议**: 限制 `max_memory_entries`

---

## 安全建议

### MemoryCore
- ✅ 定期清理过期记忆
- ✅ 验证输入数据
- ✅ 使用适当的过期时间

### SandboxConfig
- ✅ 启用严格模式
- ✅ 禁用符号链接（除非必要）
- ✅ 限制白名单目录
- ✅ 使用 `create_safe_path()` 而非手动拼接

### PromptBuilder
- ✅ 设置合理的 `max_tokens`
- ✅ 限制 `max_memory_entries`
- ✅ 验证 Prompt 长度

---

## 文件位置

```
/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-core/src/
├── memory/
│   └── mod.rs                  # 记忆核心系统
├── sandbox/
│   └── mod.rs                  # 沙箱安全控制
└── executor/
    └── prompt_builder.rs       # Prompt 构建器
```

---

## 导入路径

```rust
// 记忆系统
use agentflow_core::memory::MemoryCore;

// 沙箱系统
use agentflow_core::sandbox::SandboxConfig;

// Prompt 构建器
use agentflow_core::executor::PromptBuilder;

// 类型
use agentflow_core::types::{MemoryEntry, MemoryCategory};
```

---

## 更多信息

详细文档请参阅：
- `TEAM_B_IMPLEMENTATION_REPORT.md` - 完整实现报告
- `TEAM_B_CHECKLIST.md` - 功能清单
- 源码中的 Rust 文档注释
