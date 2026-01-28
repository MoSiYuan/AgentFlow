这份计划书专注于将 ClawDBot 的记忆架构 移植到 AgentFlow 的 Rust Core 中。
核心思想是：Markdown 是灵魂，SQLite 是索引。所有记忆以人类可读的 Markdown 文本形式存储在文件系统中，SQLite 负责高效的向量检索和全文搜索。
记忆体系改造计划：AgentFlow Memory Core (Rust)
1. 设计目标与参考蓝本
参考对象：ClawDBot
核心特性：
文件即真理：记忆存储为 MEMORY.md 和 memory/*.md，便于人工编辑和版本控制。
双重索引：结合 向量检索（语义理解）和 BM25/FTS5（关键词匹配）。
惰性更新：通过文件监听，仅在文件变更时更新索引，无需全量重算。
目标：
使用 SQLite 作为底层存储，实现向量 (BLOB) 和全文索引 (FTS5)。
AgentFlow Master 自动监控文件变更，实时更新记忆库。
提供统一的 Search API，供 Worker (Claude) 和 Cloud Master 调用。
2. 第一阶段：数据模型与切分逻辑
Task 1.1: 定义 Schema (SQLite)
文件: rust-core/migrations/001_create_memory_tables.sql
需求细节:
创建以下表结构：
memory_chunks (存储向量化后的记忆块):
id (Integer, PK)
source_file (Text): 来源文件路径 (e.g., MEMORY.md).
chunk_index (Integer): 在文件中的块序号.
content (Text): 原始文本内容.
embedding (Blob): 向量数据 (F32数组).
metadata (JSON): 额外元数据 (tags, created_at).
updated_at (Timestamp).
memory_fts (虚拟表，用于全文搜索):
使用 SQLite FTS5 扩展: CREATE VIRTUAL TABLE memory_fts USING fts5(content, source_file);
触发器: 在 memory_chunks 插入/更新时，自动同步到 memory_fts。
Task 1.2: Markdown Chunker 模块
文件: rust-core/src/memory/chunker.rs
需求细节:
输入: Markdown 文本字符串。
逻辑:
按标题 (##) 或段落分割。
对每个段落计算 Token 数（近似按字符数/4）。
若段落过长，按 400 Token 为单位强制切分。
保留上下文：每个块携带前一个块的最后 50 token，以保证语义连贯。
输出: Vec<MemoryChunk> (包含文本和元数据)。
3. 第二阶段：嵌入与索引引擎
Task 2.1: Embedding Provider 抽象层
文件: rust-core/src/memory/embedder.rs
需求细节:
定义 Trait Embedder:
async fn embed(&self, text: &str) -> Result<Vec<f32>>;
实现:
OllamaEmbedder (推荐): 调用本地 Ollama (http://localhost:11434/api/embeddings)，使用模型 nomic-embed-text 或 mxbai-embed-large。
*理由*: 保护隐私，无需云端 API Key。
OpenAIEmbedder (可选): 调用 OpenAI API，作为回退方案。
Task 2.2: 索引构建器
文件: rust-core/src/memory/indexer.rs
需求细节:
核心流程：FileChanged -> Read -> Chunk -> Embed -> Upsert DB.
监听文件系统事件（使用 notify crate）。
触发 reindex_file(path):
读取文件内容。
调用 Chunker 切分。
并发调用 Embedder 计算向量。
将 Vec<f32> 序列化为字节 (bytemuck)。
Upsert: 删除该文件旧的 memory_chunks，插入新数据（触发器自动更新 FTS5）。
4. 第三阶段：混合检索系统
Task 3.1: 搜索算法 (Hybrid Search)
文件: rust-core/src/memory/search.rs
需求细节:
实现 search(query: &str, top_k: usize) -> Vec<MemoryChunk>.
算法步骤:
Query Embedding: 先对 Query 文本做 Embedding。
Vector Search:
在 SQLite 中加载所有 embedding Blob。
计算余弦相似度。
取前 top_k * 2 个结果 (召回)。
Keyword Filter (Rerank):
利用 memory_fts 表对结果集进行关键词匹配打分。
将关键词匹配度作为加权因子叠加到向量相似度上。
Final Sort: 返回最终的 Top K 结果。
*(注：对于小规模数据（<10万条），全量加载向量并在内存中计算余弦距离通常比使用专门的向量数据库扩展更简单、更稳定，且 Rust 足够快。)*
Task 3.2: 记忆管理 API
文件: rust-core/src/memory/mod.rs
需求细节:
提供对外的 Memory Service 接口：
write_memory(content: &str, file: &str): 写入 Markdown 并触发索引。
get_memories_by_tag(tag: &str): 查询相关记忆。
search(query: &str) -> Vec<SearchResult>: 执行混合搜索。
5. 第四阶段：系统集成与交互
Task 4.1: Git 事件集成
需求:
当 Agent 完成一个 Task 并提交 Git Commit 时，解析 Commit Message。
如果 Commit Message 包含 [Memory] 或重要标记，自动生成一条摘要写入 MEMORY.md，并触发索引更新。
Task 4.2: Worker 上下文注入
需求:
修改 Worker 调度逻辑。
在下发 Task 给 Node.js Worker 之前，Master 调用 memory.search(task_description)。
将检索到的 Top 3 记忆块拼接成 System Prompt 补充信息：## Relevant Memories
1. [2023-10-01] DX12 Crash Fix: Ensure to reset command list before execution...
2. ...
Task 4.3: Cloud Master 查询转发
需求:
Cloud Master 接收到远程查询请求。
将 Query 发送给 Local Master。
Local Master 执行本地搜索，返回结果（不包含原始 Blob 数据，只包含文本内容）。
6. 并行开发任务分配
Team A: 数据层
负责编写 SQL 脚本 (Task 1.1)。
实现 Chunker 切分逻辑 (Task 1.2)。
使用 sqlx 进行 Rust 数据库交互测试。
Team B: 引擎层
实现 Embedder Trait 和 Ollama 客户端 (Task 2.1)。
实现 Indexer 和文件监听逻辑 (Task 2.2)。
编写向量计算工具函数 (余弦相似度)。
Team C: 接口与应用层
实现混合搜索算法 (Task 3.1)。
将 Memory Service 集成到 Master 的 API 路由中 (Task 3.2, 4.2)。
处理 Git Hook 与 Memory 的联动 (Task 4.1)。
7. 验收标准
写入测试:
手动修改 MEMORY.md，添加 "DiveAdstra uses DX12 by default."。
等待 3 秒（索引更新）。
查询 SQLite memory_chunks，确认已生成 Embedding Blob。
检索测试:
查询 API: POST /api/memory/search { query: "render engine" }。
返回结果中包含 "DX12" 相关条目，且排名靠前。
性能测试:
写入 1000 条模拟记忆。
检索延迟应 < 200ms (本地 Ollama + 内存计算)。
指令提示:
Agent 请按照此计划优先实现 Task 1.1 (Schema) 和 Task 1.2 (Chunker)，这是后续所有工作的基础。确保 Rust 项目已添加依赖 sqlx, notify, tokio, serde, bytemuck。
开始执行！