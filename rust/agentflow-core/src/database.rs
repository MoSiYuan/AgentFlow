//! 数据库模块
//!
//! 提供数据库连接和操作的抽象层

use anyhow::Result;
use sqlx::{Pool, Sqlite};
use std::path::Path;

/// 数据库连接池
pub type DbPool = Pool<Sqlite>;

/// 创建数据库连接池
///
/// # 参数
///
/// * `database_path` - SQLite 数据库文件路径
///
/// # 返回
///
/// 返回数据库连接池
pub async fn create_pool(database_path: &Path) -> Result<DbPool> {
    let connection_string = format!("sqlite:{}", database_path.display());

    let pool = sqlx::SqlitePool::connect(&connection_string).await?;

    Ok(pool)
}

/// 运行数据库迁移
///
/// # 参数
///
/// * `pool` - 数据库连接池
///
/// # 注意
///
/// 当前版本不使用 migrations，此函数保留以备将来使用
pub async fn run_migrations(_pool: &DbPool) -> Result<()> {
    // TODO: 实现数据库迁移逻辑
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_create_pool() {
        let pool = create_pool(Path::new(":memory:")).await;
        assert!(pool.is_ok());
    }
}
