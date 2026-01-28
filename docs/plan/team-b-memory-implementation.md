# Team B: Memory & Git Integration Implementation Plan

**Project**: AgentFlow v0.2.1
**Team**: Team B
**Focus**: Markdown-based Memory System with Automatic Indexing and Writeback
**Date**: 2026-01-28
**Branch**: feature/0.2.1

---

## Executive Summary

This document outlines the implementation plan for adding Markdown-based memory system to AgentFlow. The system will:

1. Store memories as human-readable Markdown files in `.agentflow/memory/`
2. Automatically chunk and index Markdown files into SQLite
3. Support semantic search using embeddings via Claude CLI
4. Auto-write task execution summaries to Markdown
5. Inject relevant memories into prompts for better context

---

## Current State Analysis

### Existing Components

**MemoryCore** (`rust/agentflow-core/src/memory/mod.rs`):
- ✅ SQLite-based storage with keyword search
- ✅ Schema: `memories` table with `id`, `key`, `value`, `category`, `task_id`, `timestamp`, `expires_at`, `embedding`
- ✅ Methods: `new()`, `index()`, `search()`, `get()`, `delete()`, `cleanup_expired()`, `stats()`
- ⚠️ Embeddings are currently deprecated/unused (empty BLOB)
- ⚠️ Uses SQL LIKE for search (not semantic)

**MemoryProcessor** (`rust/agentflow-core/src/executor/memory_processor.rs`):
- ✅ Post-task execution analysis
- ✅ Calls Claude CLI to extract insights from stdout/stderr
- ✅ Stores categorized memories (Execution, Error, Result)
- ✅ Runs asynchronously in tokio::spawn
- ⚠️ Does NOT write to Markdown files
- ⚠️ Does NOT trigger reindexing

**TaskExecutor** (`rust/agentflow-core/src/executor/mod.rs`):
- ✅ Executes Claude CLI commands
- ✅ Optional memory processor integration
- ✅ Captures stdout/stderr (currently TODO)
- ⚠️ No Git diff detection
- ⚠️ No automatic writeback to Markdown

**PromptBuilder** (`rust/agentflow-core/src/executor/prompt_builder.rs`):
- ✅ Builds prompts with system instruction, memories, and task
- ✅ Accepts memory entries in `build()` method
- ⚠️ Does NOT call `MemoryCore::search()` automatically
- ⚠️ Memories must be pre-fetched and passed in

### Database Schema (Current)

```sql
CREATE TABLE memories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL,
    category TEXT NOT NULL,
    task_id TEXT,
    timestamp INTEGER NOT NULL,
    expires_at INTEGER,
    embedding BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_memories_key ON memories(key);
CREATE INDEX idx_memories_category ON memories(category);
CREATE INDEX idx_memories_task_id ON memories(task_id);
CREATE INDEX idx_memories_timestamp ON memories(timestamp DESC);
```

### Missing Components

❌ No Markdown chunker
❌ No embedder trait/implementations
❌ No file watcher for `.agentflow/memory/`
❌ No full-text search (FTS5)
❌ No `memory_chunks` table for chunked documents
❌ No automatic memory writeback
❌ No prompt memory injection

---

## Implementation Plan

### Phase 1: Data Layer Extensions

#### Task 1.1: Create Memory Chunker

**File**: `rust/agentflow-core/src/memory/chunker.rs`

**Dependencies to Add** (to `Cargo.toml`):
```toml
[dependencies]
pulldown-cmark = "0.12"  # Markdown parsing
```

**Design**:
```rust
use std::collections::HashMap;
use std::path::PathBuf;

/// A chunk of markdown text with metadata
#[derive(Debug, Clone)]
pub struct Chunk {
    pub id: String,
    pub source_file: PathBuf,
    pub index: usize,
    pub content: String,
    pub metadata: HashMap<String, String>,
}

/// Markdown chunking configuration
#[derive(Debug, Clone)]
pub struct ChunkerConfig {
    pub max_tokens: usize,        // Default: 400
    pub overlap_tokens: usize,     // Default: 50
    pub min_paragraph_tokens: usize, // Default: 50
}

impl Default for ChunkerConfig {
    fn default() -> Self {
        Self {
            max_tokens: 400,
            overlap_tokens: 50,
            min_paragraph_tokens: 50,
        }
    }
}

/// Markdown chunker
pub struct Chunker {
    config: ChunkerConfig,
}

impl Chunker {
    pub fn new(config: ChunkerConfig) -> Self {
        Self { config }
    }

    pub fn with_default_config() -> Self {
        Self::new(ChunkerConfig::default())
    }

    /// Chunk a markdown file into pieces
    pub fn chunk_markdown(&self, content: &str, source_file: PathBuf) -> Result<Vec<Chunk>> {
        // 1. Parse markdown structure (headers, paragraphs)
        // 2. Group content by headers (##, ###)
        // 3. Split long paragraphs at sentence boundaries
        // 4. Add context overlap (previous paragraph tail)
        // 5. Generate metadata (header path, tags, date)
        // Implementation in next section
    }

    /// Estimate token count (rough approximation)
    fn estimate_tokens(&self, text: &str) -> usize {
        // Chinese: ~2 chars/token, English: ~4 chars/token
        let char_count = text.chars().count();
        char_count / 3
    }
}

#[cfg(test)]
mod tests {
    // Unit tests for chunking logic
}
```

**Key Implementation Details**:

1. **Header-Aware Chunking**:
   - Use `pulldown-cmark` parser to identify headers
   - Track header hierarchy (h1, h2, h3...)
   - Include header path in metadata (e.g., "## 2026-01-28 > ### Changes")

2. **Paragraph Preservation**:
   - Don't split mid-sentence
   - Use regex to find sentence boundaries: `(?<=[.!?])\s+`
   - Keep list items together

3. **Context Overlap**:
   - Each chunk includes last `overlap_tokens` from previous chunk
   - Ensures semantic continuity between chunks

4. **Metadata Extraction**:
   - Extract date from headers (e.g., "## 2026-01-28")
   - Extract tags (lines starting with "- tag:")
   - Extract commit hashes (lines like "**Commit**: abc123")

---

#### Task 1.2: Database Schema Migration

**File**: `rust/agentflow-core/src/memory/migrations/002_add_chunks_table.sql`

**New Schema**:
```sql
-- Chunked markdown storage
CREATE TABLE IF NOT EXISTS memory_chunks (
    id TEXT PRIMARY KEY,              -- UUID
    source_file TEXT NOT NULL,        -- Relative path from .agentflow/memory/
    chunk_index INTEGER NOT NULL,     -- Position in file
    content TEXT NOT NULL,            -- Markdown text
    embedding BLOB,                   -- Vec<f32> serialized as bytes
    metadata TEXT,                    -- JSON: {"header": "...", "tags": [...], "date": "..."}
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_memory_chunks_source
ON memory_chunks(source_file, chunk_index);

CREATE INDEX IF NOT EXISTS idx_memory_chunks_updated
ON memory_chunks(updated_at DESC);

-- Full-text search table (SQLite FTS5)
CREATE VIRTUAL TABLE IF NOT EXISTS memory_fts USING fts5(
    content,
    source_file,
    tokenize='unicode61'
);

-- Trigger to sync chunks to FTS
CREATE TRIGGER IF NOT EXISTS memory_chunks_fts_insert
AFTER INSERT ON memory_chunks
BEGIN
    INSERT INTO memory_fts(rowid, content, source_file)
    VALUES (NEW.id, NEW.content, NEW.source_file);
END;

CREATE TRIGGER IF NOT EXISTS memory_chunks_fts_delete
AFTER DELETE ON memory_chunks
BEGIN
    DELETE FROM memory_fts WHERE rowid = OLD.id;
END;

CREATE TRIGGER IF NOT EXISTS memory_chunks_fts_update
AFTER UPDATE ON memory_chunks
BEGIN
    UPDATE memory_fts
    SET content = NEW.content, source_file = NEW.source_file
    WHERE rowid = NEW.id;
END;
```

**Migration Integration**:
- Modify `MemoryCore::init_schema()` to include new tables
- Add method to run migrations: `MemoryCore::run_migrations()`

---

### Phase 2: Embedding Engine

#### Task 2.1: Define Embedder Trait

**File**: `rust/agentflow-core/src/memory/embedder.rs`

**Dependencies to Add**:
```toml
[dependencies]
async-trait = "0.1"  # Already in workspace
```

**Design**:
```rust
use async_trait::async_trait;
use anyhow::Result;

/// Embedding provider trait
#[async_trait]
pub trait Embedder: Send + Sync {
    /// Generate embedding vector for text
    async fn embed(&self, text: &str) -> Result<Vec<f32>>;

    /// Get embedding dimension
    fn dimension(&self) -> usize;

    /// Batch embed multiple texts (optional optimization)
    async fn embed_batch(&self, texts: &[String]) -> Result<Vec<Vec<f32>>> {
        // Default implementation: sequential calls
        let mut results = Vec::with_capacity(texts.len());
        for text in texts {
            results.push(self.embed(text).await?);
        }
        Ok(results)
    }
}
```

---

#### Task 2.2: Claude CLI Embedder Implementation

**Design**:
```rust
use std::process::Stdio;
use tokio::process::Command;
use tracing::{debug, warn};

/// Claude CLI embedding provider
///
/// Uses Claude CLI to generate embeddings by requesting
/// the model to output a fixed-dimensional vector representation
pub struct ClaudeCliEmbedder {
    /// Claude CLI command path (default: "claude")
    command: String,
    /// Embedding dimension (fixed for consistency)
    dimension: usize,
    /// Timeout in seconds
    timeout: u64,
}

impl ClaudeCliEmbedder {
    pub fn new() -> Self {
        Self {
            command: "claude".to_string(),
            dimension: 1536, // Common embedding dimension
            timeout: 30,
        }
    }

    pub fn with_dimension(mut self, dimension: usize) -> Self {
        self.dimension = dimension;
        self
    }

    pub fn with_timeout(mut self, timeout: u64) -> Self {
        self.timeout = timeout;
        self
    }

    /// Build prompt for embedding generation
    fn build_embedding_prompt(&self, text: &str) -> String {
        format!(
            r#"Generate a {}-dimensional embedding vector for the following text.
Return ONLY a JSON array of floats, nothing else.

Text: {}"#,
            self.dimension,
            text
        )
    }

    /// Parse Claude response to extract vector
    fn parse_embedding_response(&self, response: &str) -> Result<Vec<f32>> {
        // Extract JSON array from response
        let json_str = self.extract_json_array(response)?;

        // Parse as JSON array
        let vec: Vec<f32> = serde_json::from_str(json_str)
            .with_context(|| format!("Failed to parse embedding: {}", json_str))?;

        // Validate dimension
        if vec.len() != self.dimension {
            warn!(
                "Embedding dimension mismatch: expected {}, got {}",
                self.dimension,
                vec.len()
            );
        }

        Ok(vec)
    }

    fn extract_json_array<'a>(&self, text: &'a str) -> Result<&'a str> {
        let start = text.find('[').context("No JSON array found")?;
        let end = text.rfind(']').context("No JSON array end found")?;
        Ok(&text[start..=end])
    }
}

#[async_trait]
impl Embedder for ClaudeCliEmbedder {
    async fn embed(&self, text: &str) -> Result<Vec<f32>> {
        debug!("Generating embedding for text ({} chars)", text.len());

        let prompt = self.build_embedding_prompt(text);

        // Execute Claude CLI
        let output = tokio::time::timeout(
            std::time::Duration::from_secs(self.timeout),
            Command::new(&self.command)
                .arg(&prompt)
                .env_clear()
                .envs(std::env::vars())
                .stdout(Stdio::piped())
                .stderr(Stdio::piped())
                .output(),
        )
        .await
        .context("Embedding generation timed out")?
        .context("Failed to execute Claude CLI")?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            return Err(anyhow::anyhow!("Claude CLI failed: {}", stderr));
        }

        let response = String::from_utf8_lossy(&output.stdout);
        let embedding = self.parse_embedding_response(&response)?;

        debug!("Generated embedding: {} dimensions", embedding.len());
        Ok(embedding)
    }

    fn dimension(&self) -> usize {
        self.dimension
    }
}
```

**Limitations & Notes**:
- Claude CLI doesn't natively support embeddings
- This is a workaround: asking Claude to generate vectors manually
- Quality may vary compared to dedicated embedding models
- Future: Add `GLM4Embedder` for proper embeddings via API

---

#### Task 2.3: Placeholder GLM4 Embedder

**Design**:
```rust
/// GLM-4 embedding provider (future implementation)
///
/// Will use Zhipu AI's embedding API:
/// https://open.bigmodel.cn/dev/api#embedding
pub struct GLM4Embedder {
    api_key: String,
    base_url: String,
    model: String,
    timeout: u64,
}

impl GLM4Embedder {
    pub fn new(api_key: String) -> Self {
        Self {
            api_key,
            base_url: "https://open.bigmodel.cn/api/paas/v4/embeddings".to_string(),
            model: "embedding-2".to_string(),
            timeout: 30,
        }
    }
}

#[async_trait]
impl Embedder for GLM4Embedder {
    async fn embed(&self, _text: &str) -> Result<Vec<f32>> {
        // TODO: Implement HTTP call to GLM-4 API
        Err(anyhow::anyhow!("GLM4Embedder not yet implemented"))
    }

    fn dimension(&self) -> usize {
        1024 // GLM-4 embedding dimension
    }
}
```

---

### Phase 3: File Watcher & Indexer

#### Task 3.1: Memory Indexer

**File**: `rust/agentflow-core/src/memory/indexer.rs`

**Dependencies to Add**:
```toml
[dependencies]
notify = "7.0"  # File system watcher
```

**Design**:
```rust
use crate::memory::chunker::{Chunk, Chunker};
use crate::memory::embedder::Embedder;
use anyhow::{Context, Result};
use notify::{RecursiveMode, Watcher, recommended_watcher, Event, EventKind};
use sqlx::SqlitePool;
use std::path::{Path, PathBuf};
use std::sync::Arc;
use tokio::sync::Mutex;
use tracing::{debug, info, warn};

/// Memory indexer
///
/// Watches `.agentflow/memory/` directory and automatically
/// reindexes changed markdown files
pub struct Indexer {
    /// Database connection pool
    pool: SqlitePool,
    /// Markdown chunker
    chunker: Chunker,
    /// Embedding provider
    embedder: Arc<dyn Embedder>,
    /// Memory directory path
    memory_dir: PathBuf,
    /// Notify watcher
    watcher: Option<Box<dyn Watcher>>,
}

impl Indexer {
    /// Create new indexer
    pub fn new(
        pool: SqlitePool,
        chunker: Chunker,
        embedder: Arc<dyn Embedder>,
        memory_dir: PathBuf,
    ) -> Self {
        Self {
            pool,
            chunker,
            embedder,
            memory_dir,
            watcher: None,
        }
    }

    /// Start watching for file changes
    pub fn start_watching(mut self) -> Result<Self> {
        info!("Starting file watcher: {:?}", self.memory_dir);

        // Create watcher
        let (tx, rx) = std::sync::mpsc::channel();
        let mut watcher = recommended_watcher(move |res: Result<Event, _>| {
            if let Ok(event) = res {
                let _ = tx.send(event);
            }
        })?;

        // Watch memory directory recursively
        watcher.watch(&self.memory_dir, RecursiveMode::Recursive)?;

        self.watcher = Some(Box::new(watcher));

        // Spawn background task to handle events
        let pool = self.pool.clone();
        let chunker = self.chunker.clone();
        let embedder = self.embedder.clone();
        let memory_dir = self.memory_dir.clone();

        tokio::spawn(async move {
            for event in rx {
                if let Err(e) = Self::handle_event(
                    &pool,
                    &chunker,
                    &embedder,
                    &memory_dir,
                    event,
                ).await {
                    warn!("Error handling file event: {}", e);
                }
            }
        });

        Ok(self)
    }

    /// Handle a file system event
    async fn handle_event(
        pool: &SqlitePool,
        chunker: &Chunker,
        embedder: &Arc<dyn Embedder>,
        memory_dir: &Path,
        event: Event,
    ) -> Result<()> {
        // Only process .md files
        let path = event.paths.first()?;
        if path.extension().and_then(|s| s.to_str()) != Some("md") {
            return Ok(());
        }

        // Debounce: wait 1 second before processing
        tokio::time::sleep(std::time::Duration::from_secs(1)).await;

        match event.kind {
            EventKind::Create(_) | EventKind::Modify(_) => {
                Self::reindex_file(pool, chunker, embedder, memory_dir, path).await?;
            }
            EventKind::Remove(_) => {
                Self::remove_file_chunks(pool, path).await?;
            }
            _ => {}
        }

        Ok(())
    }

    /// Reindex a markdown file
    pub async fn reindex_file(
        pool: &SqlitePool,
        chunker: &Chunker,
        embedder: &Arc<dyn Embedder>,
        memory_dir: &Path,
        file_path: &Path,
    ) -> Result<()> {
        info!("Reindexing file: {:?}", file_path);

        // Read file content
        let content = tokio::fs::read_to_string(file_path).await
            .context("Failed to read file")?;

        // Chunk the content
        let chunks = chunker.chunk_markdown(&content, file_path.to_path_buf())?;

        // Get relative path from memory_dir
        let relative_path = file_path.strip_prefix(memory_dir)
            .unwrap_or(file_path)
            .to_string_lossy()
            .to_string();

        // Clear old chunks for this file
        sqlx::query("DELETE FROM memory_chunks WHERE source_file = ?1")
            .bind(&relative_path)
            .execute(pool)
            .await?;

        // Embed and insert new chunks
        for chunk in chunks {
            // Generate embedding
            let embedding = embedder.embed(&chunk.content).await
                .context("Failed to generate embedding")?;

            // Serialize embedding to bytes
            let embedding_bytes = Self::serialize_embedding(&embedding);

            // Serialize metadata to JSON
            let metadata_json = serde_json::to_string(&chunk.metadata)?;

            // Insert into database
            sqlx::query(
                r#"
                INSERT INTO memory_chunks (id, source_file, chunk_index, content, embedding, metadata)
                VALUES (?1, ?2, ?3, ?4, ?5, ?6)
                "#
            )
            .bind(&chunk.id)
            .bind(&relative_path)
            .bind(chunk.index as i64)
            .bind(&chunk.content)
            .bind(embedding_bytes)
            .bind(metadata_json)
            .execute(pool)
            .await?;
        }

        info!("Reindexed {} chunks from {:?}", chunks.len(), file_path);
        Ok(())
    }

    /// Remove all chunks for a file
    async fn remove_file_chunks(pool: &SqlitePool, file_path: &Path) -> Result<()> {
        let file_str = file_path.to_string_lossy();
        sqlx::query("DELETE FROM memory_chunks WHERE source_file = ?1")
            .bind(&*file_str)
            .execute(pool)
            .await?;
        Ok(())
    }

    /// Serialize embedding vector to bytes
    fn serialize_embedding(vec: &[f32]) -> Vec<u8> {
        bytemuck::cast_slice(vec).to_vec()
    }

    /// Deserialize bytes to embedding vector
    fn deserialize_embedding(bytes: &[u8]) -> Vec<f32> {
        bytemuck::cast_slice(bytemuck::try_cast_slice(bytes).unwrap()).to_vec()
    }
}
```

**Dependencies to Add** (to workspace Cargo.toml):
```toml
bytemuck = "1.20"  # For casting f32 slice to bytes
```

---

### Phase 4: Auto Writeback System

#### Task 4.1: Modify TaskExecutor for Git Integration

**File**: `rust/agentflow-core/src/executor/mod.rs`

**Changes to `execute()` method**:

```rust
pub async fn execute(&self, prompt: &str) -> Result<ExecutionResult> {
    // ... existing execution code ...

    let result = ExecutionResult {
        success,
        stdout,
        stderr,
        exit_code,
    };

    // NEW: Detect Git changes after task execution
    if let Ok(git_diff) = self.detect_git_changes().await {
        if !git_diff.is_empty() {
            debug!("Detected Git changes, will write to memory");

            // Write memory entry
            if let Err(e) = self.write_task_memory(&task_description, &git_diff, &result).await {
                warn!("Failed to write task memory: {}", e);
            }
        }
    }

    // NEW: Check for <save_memory> tags in output
    if self.contains_save_memory_tag(&result.stdout) || self.contains_save_memory_tag(&result.stderr) {
        debug!("Detected <save_memory> tag, will write to memory");
        // Extract and save memory
    }

    // ... existing memory processor call ...
}
```

**New Methods**:

```rust
impl TaskExecutor {
    /// Detect Git changes after task execution
    async fn detect_git_changes(&self) -> Result<String> {
        let output = Command::new("git")
            .args(["diff", "HEAD"])
            .current_dir(&self.workspace_path)
            .output()
            .await?;

        if output.status.success() {
            Ok(String::from_utf8_lossy(&output.stdout).to_string())
        } else {
            Ok(String::new())
        }
    }

    /// Get current Git commit hash
    async fn get_git_commit(&self) -> Result<String> {
        let output = Command::new("git")
            .args(["rev-parse", "HEAD"])
            .current_dir(&self.workspace_path)
            .output()
            .await?;

        if output.status.success() {
            Ok(String::from_utf8_lossy(&output.stdout).trim().to_string())
        } else {
            Ok("unknown".to_string())
        }
    }

    /// Write task execution summary to memory markdown file
    async fn write_task_memory(
        &self,
        task: &str,
        git_diff: &str,
        result: &ExecutionResult,
    ) -> Result<()> {
        let date = Utc::now().format("%Y-%m-%d").to_string();
        let memory_file = self.workspace_path
            .join(".agentflow/memory")
            .join(format!("{}.md", date));

        // Ensure directory exists
        tokio::fs::create_dir_all(memory_file.parent().unwrap()).await?;

        // Get commit hash
        let commit = self.get_git_commit().await?;

        // Parse Git diff to extract changed files
        let changes = self.parse_git_diff(git_diff);

        // Build markdown content
        let content = format!(
            r#"## {} Task Summary

**Commit**: {}

**Task**: {}

### Changes
{}

### Result
- **Success**: {}
- **Exit Code**: {:?}

### Errors Encountered
{}

---
"#,
            date,
            commit,
            task,
            changes,
            result.success,
            result.exit_code,
            if result.stderr.is_empty() {
                "None".to_string()
            } else {
                format!("```\n{}\n```", result.stderr)
            }
        );

        // Append to file
        tokio::fs::write(&memory_file, content).await
            .context("Failed to write memory file")?;

        info!("Memory written to: {:?}", memory_file);

        // NEW: Trigger indexer reindex
        // This will be handled by the file watcher automatically

        Ok(())
    }

    /// Parse Git diff to extract changed files summary
    fn parse_git_diff(&self, diff: &str) -> String {
        // Extract file names from diff
        let mut files = Vec::new();
        for line in diff.lines() {
            if line.starts_with("+++ b/") || line.starts_with("--- a/") {
                let file = line.split('/').last().unwrap_or("");
                if !file.is_empty() && !file.contains("/dev/null") {
                    files.push(format!("- {}", file));
                }
            }
        }
        if files.is_empty() {
            "No files changed".to_string()
        } else {
            files.join("\n")
        }
    }

    /// Check if text contains <save_memory> tag
    fn contains_save_memory_tag(&self, text: &str) -> bool {
        text.contains("<save_memory>") || text.contains("<save memory>")
    }
}
```

---

#### Task 4.2: Memory Directory Structure

**Create Directory**:
```
.agentflow/
└── memory/
    ├── 2026-01-28.md
    ├── 2026-01-27.md
    └── ...
```

**Example Memory File** (`2026-01-28.md`):
```markdown
## 2026-01-28 Task Summary

**Commit**: a1b2c3d4

**Task**: Implement user authentication with JWT

### Changes
- src/auth/jwt.rs
- src/api/routes.rs
- src/models/user.rs

### Result
- **Success**: true
- **Exit Code**: Some(0)

### Errors Encountered
None

---

## 2026-01-28 Task Summary

**Commit**: e5f6g7h8

**Task**: Fix database connection timeout

### Changes
- src/db/connection.rs

### Result
- **Success**: true
- **Exit Code**: Some(0)

### Errors Encountered
```
Error: Connection timeout after 30s
Solution: Increased timeout to 60s in config.toml
```

---
```

---

### Phase 5: Prompt Integration

#### Task 5.1: Enhance PromptBuilder

**File**: `rust/agentflow-core/src/executor/prompt_builder.rs`

**Changes**:

```rust
impl PromptBuilder {
    /// NEW: Build with automatic memory search
    ///
    /// This version calls MemoryCore::search() to find relevant memories
    pub async fn build_with_memory_search(
        &self,
        task: &str,
        memory_core: &crate::memory::MemoryCore,
    ) -> String {
        // Search for relevant memories (top 3)
        let memories = match memory_core.search(task, 3).await {
            Ok(m) => m,
            Err(e) => {
                warn!("Memory search failed: {}", e);
                Vec::new()
            }
        };

        // Use existing build method
        self.build(task, &memories)
    }

    /// Modify build_memory_section to include dates
    fn build_memory_section(&self, memories: &[MemoryEntry]) -> String {
        // ... existing code ...

        for (index, memory) in limited_memories.iter().enumerate() {
            let timestamp = DateTime::<Utc>::from_timestamp(memory.timestamp, 0)
                .unwrap()
                .format("%Y-%m-%d")
                .to_string();

            memory_text.push_str(&format!(
                "### {} [{}] ({})\n",
                category_label, index + 1, timestamp
            ));

            // ... rest of formatting ...
        }

        memory_text
    }
}
```

**Usage in TaskExecutor**:

```rust
// In execute() method:
let builder = PromptBuilder::new();
let prompt = if let Some(ref memory_core) = self.memory_core {
    builder.build_with_memory_search(&prompt, memory_core).await
} else {
    builder.build(&prompt, &[])
};
```

---

### Phase 6: Enhanced Search

#### Task 6.1: Hybrid Search Implementation

**File**: `rust/agentflow-core/src/memory/mod.rs`

**Add to MemoryCore**:

```rust
impl MemoryCore {
    /// NEW: Hybrid semantic + keyword search
    ///
    /// Combines vector similarity with FTS5 keyword matching
    pub async fn search_hybrid(
        &self,
        query: &str,
        embedder: &dyn Embedder,
        limit: usize,
    ) -> Result<Vec<MemoryEntry>> {
        // 1. Generate query embedding
        let query_embedding = embedder.embed(query).await?;

        // 2. Load all chunks with embeddings
        let rows = sqlx::query(
            r#"
            SELECT id, source_file, chunk_index, content, embedding, metadata
            FROM memory_chunks
            "#
        )
        .fetch_all(&self.pool)
        .await?;

        // 3. Calculate cosine similarities
        let mut scored_results: Vec<(f64, MemoryEntry)> = Vec::new();

        for row in rows {
            let embedding_bytes: Vec<u8> = row.get("embedding");
            let content: String = row.get("content");
            let metadata_json: String = row.get("metadata");

            // Deserialize embedding
            let chunk_embedding: Vec<f32> = Indexer::deserialize_embedding(&embedding_bytes);

            // Calculate cosine similarity
            let similarity = cosine_similarity(&query_embedding, &chunk_embedding);

            // Create memory entry
            let metadata: HashMap<String, String> = serde_json::from_str(&metadata_json)?;
            let entry = MemoryEntry {
                key: format!("{}:{}:{}", row.get::<String, _>("source_file"),
                             row.get::<i64, _>("chunk_index"),
                             Uuid::new_v4()),
                value: serde_json::json!(content),
                expires_at: None,
                category: MemoryCategory::Context,
                task_id: None,
                timestamp: Utc::now().timestamp(),
            };

            scored_results.push((similarity, entry));
        }

        // 4. Sort by similarity
        scored_results.sort_by(|a, b| b.0.partial_cmp(&a.0).unwrap());

        // 5. Take top results
        let results: Vec<MemoryEntry> = scored_results
            .into_iter()
            .take(limit)
            .map(|(_, entry)| entry)
            .collect();

        Ok(results)
    }
}

/// Calculate cosine similarity between two vectors
fn cosine_similarity(a: &[f32], b: &[f32]) -> f64 {
    if a.len() != b.len() {
        return 0.0;
    }

    let dot_product: f32 = a.iter().zip(b.iter()).map(|(x, y)| x * y).sum();
    let norm_a: f32 = a.iter().map(|x| x * x).sum::<f32>().sqrt();
    let norm_b: f32 = b.iter().map(|x| x * x).sum::<f32>().sqrt();

    if norm_a == 0.0 || norm_b == 0.0 {
        0.0
    } else {
        (dot_product / (norm_a * norm_b)) as f64
    }
}
```

---

## Integration Points

### 1. Initialize Indexer on Server Startup

**File**: `rust/agentflow-master/src/main.rs`

```rust
async fn main() -> Result<()> {
    // ... existing initialization ...

    // Create memory directory
    let memory_dir = std::path::PathBuf::from(".agentflow/memory");
    tokio::fs::create_dir_all(&memory_dir).await?;

    // Create indexer components
    let chunker = Chunker::with_default_config();
    let embedder = Arc::new(ClaudeCliEmbedder::new());

    // Create and start indexer
    let indexer = Indexer::new(
        db.clone(),
        chunker,
        embedder,
        memory_dir,
    ).start_watching()?;

    info!("📝 Memory indexer started");

    // Initial index of existing files
    indexer.index_all_existing().await?;

    // ... rest of main ...
}
```

### 2. Update lib.rs Exports

**File**: `rust/agentflow-core/src/lib.rs`

```rust
pub mod types;
pub mod database;
pub mod memory;
pub mod sandbox;
pub mod executor;

// NEW: Export memory submodules
pub use memory::{chunker, embedder, indexer};

pub use types::*;
pub use database::*;
pub use memory::*;
pub use sandbox::*;
pub use executor::*;
```

---

## Testing Strategy

### Unit Tests

1. **Chunker Tests**:
   - Test header-based chunking
   - Test paragraph preservation
   - Test context overlap
   - Test metadata extraction

2. **Embedder Tests**:
   - Test Claude CLI integration
   - Test dimension validation
   - Test error handling

3. **Indexer Tests**:
   - Test file change detection
   - Test database upsert
   - Test embedding serialization

### Integration Tests

1. **End-to-End Flow**:
   ```
   1. Create markdown file in .agentflow/memory/
   2. Wait 2 seconds for indexer to process
   3. Query SQLite for chunks
   4. Verify embeddings exist
   5. Search via MemoryCore::search_hybrid()
   ```

2. **Git Integration**:
   ```
   1. Execute task that modifies files
   2. Check memory file created
   3. Verify content includes Git diff
   ```

3. **Prompt Injection**:
   ```
   1. Add memories to database
   2. Build prompt with search
   3. Verify memories appear in prompt
   ```

---

## Performance Considerations

### Embedding Generation

- **Batching**: Process multiple chunks concurrently
- **Caching**: Cache embeddings for unchanged chunks
- **Timeout**: Fail fast if Claude CLI hangs

### Database

- **Indexes**: Ensure all foreign keys have indexes
- **FTS5**: Use for keyword pre-filtering before vector search
- **Connection Pooling**: Reuse connections via SqlitePool

### File Watching

- **Debouncing**: Wait 1 second after file change before processing
- **Batching**: Process multiple changes in single transaction
- **Error Recovery**: Log errors but continue watching

---

## Dependencies Summary

### New Dependencies (workspace Cargo.toml)

```toml
[workspace.dependencies]
# ... existing dependencies ...

# File watching
notify = "7.0"

# Markdown parsing
pulldown-cmark = "0.12"

# Byte manipulation
bytemuck = "1.20"

# Already present
async-trait = "0.1"
```

---

## Migration Path

### Step 1: Add Dependencies (Week 1, Day 1)
- Update `rust/Cargo.toml`
- Run `cargo check` to verify

### Step 2: Implement Chunker (Week 1, Day 2-3)
- Create `chunker.rs`
- Write unit tests
- Verify chunking quality manually

### Step 3: Database Migration (Week 1, Day 4)
- Create migration SQL
- Update `MemoryCore::init_schema()`
- Test migration on fresh database

### Step 4: Implement Embedder (Week 1, Day 5)
- Create `embedder.rs`
- Implement `ClaudeCliEmbedder`
- Test with real Claude CLI

### Step 5: Implement Indexer (Week 2, Day 1-2)
- Create `indexer.rs`
- Integrate file watcher
- Test with manual file creation

### Step 6: Auto Writeback (Week 2, Day 3-4)
- Modify `TaskExecutor`
- Test Git diff detection
- Verify memory file creation

### Step 7: Prompt Integration (Week 2, Day 5)
- Modify `PromptBuilder`
- Test memory injection
- Measure prompt quality

### Step 8: Integration Testing (Week 3)
- End-to-end tests
- Performance benchmarks
- Documentation

---

## Open Questions

1. **Claude CLI Embeddings**:
   - Q: Will Claude CLI reliably generate consistent embeddings?
   - A: Unknown - this is experimental. May need fallback to dedicated embedding service.

2. **Embedding Dimension**:
   - Q: What dimension should we target?
   - A: 1536 is standard (OpenAI), but we can make it configurable.

3. **Large File Performance**:
   - Q: How to handle very large markdown files (>10MB)?
   - A: Implement streaming chunking and batch embedding.

4. **Memory File Rotation**:
   - Q: When to create new memory files?
   - A: Daily rotation (one file per date) is simple and effective.

5. **Conflict Resolution**:
   - Q: What if manual edits conflict with auto-writeback?
   - A: Append mode avoids conflicts. Manual edits at top, auto at bottom.

---

## Success Criteria

### Functional

- ✅ Markdown files in `.agentflow/memory/` are automatically indexed
- ✅ Task executions with Git changes write to memory
- ✅ Prompts include relevant memory context
- ✅ Search returns semantically similar results

### Performance

- ✅ File indexing within 5 seconds of file change
- ✅ Search latency < 500ms for <1000 chunks
- ✅ Memory overhead < 50MB for embeddings

### Reliability

- ✅ No crashes on malformed markdown
- ✅ Graceful degradation if Claude CLI fails
- ✅ Database transactions always succeed or rollback

---

## Next Steps

1. **Review this plan** with Team B members
2. **Assign tasks** to team members
3. **Set up feature branch**: `feature/memory-indexing`
4. **Create tracking issues** in GitHub
5. **Start implementation** with Phase 1

---

## Appendix: Code Snippets

### A. Creating Test Memory Files

```bash
# Create memory directory
mkdir -p .agentflow/memory

# Create sample memory file
cat > .agentflow/memory/2026-01-28.md << 'EOF'
## 2026-01-28 Development Log

### Bug Fix: Database Timeout
- **Issue**: Connection timing out after 30s
- **Solution**: Increased timeout to 60s in config.toml
- **Commit**: abc123

### Feature: User Authentication
- Added JWT-based authentication
- Files: src/auth/jwt.rs, src/api/routes.rs
- **Commit**: def456
EOF
```

### B. Testing Indexer Manually

```bash
# Wait for indexer to process
sleep 3

# Query SQLite
sqlite3 .agentflow/agentflow.db << 'EOF'
SELECT source_file, chunk_index, substr(content, 1, 50) as preview
FROM memory_chunks
LIMIT 10;
EOF
```

### C. Testing Search

```bash
# Start server
cargo run --bin agentflow-master

# In another terminal, test search API
curl -X POST http://localhost:6767/api/v1/memory/search \
  -H "Content-Type: application/json" \
  -d '{"query": "database timeout", "limit": 3}'
```

---

**Document Version**: 1.0
**Last Updated**: 2026-01-28
**Status**: Ready for Review
