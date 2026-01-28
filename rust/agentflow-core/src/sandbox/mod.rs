//! 沙箱安全控制模块
//!
//! 提供目录白名单机制和路径穿透防护功能

use anyhow::{Context, Result, anyhow};
use std::path::{Path, PathBuf};
use std::collections::HashSet;
use thiserror::Error;
use tracing::{debug, info, warn};

/// 沙箱错误类型
#[derive(Error, Debug)]
pub enum SandboxError {
    #[error("路径不在白名单中: {0}")]
    PathNotAllowed(String),

    #[error("检测到路径穿透攻击: {0}")]
    PathTraversalDetected(String),

    #[error("检测到符号链接攻击: {0}")]
    SymlinkAttack(String),

    #[error("路径解析失败: {0}")]
    PathResolutionFailed(String),

    #[error("路径验证失败: {0}")]
    ValidationFailed(String),
}

/// 沙箱配置
///
/// 定义沙箱的安全策略和允许访问的目录
#[derive(Debug, Clone)]
pub struct SandboxConfig {
    /// 允许访问的目录白名单
    allowed_dirs: HashSet<PathBuf>,

    /// 是否启用严格模式（禁止所有白名单外的访问）
    strict_mode: bool,

    /// 是否允许符号链接跟随
    allow_symlinks: bool,

    /// 最大符号链接跟随深度
    max_symlink_depth: usize,
}

impl Default for SandboxConfig {
    fn default() -> Self {
        Self::new()
    }
}

impl SandboxConfig {
    /// 创建默认沙箱配置
    ///
    /// # 示例
    /// ```
    /// use agentflow_core::sandbox::SandboxConfig;
    ///
    /// let config = SandboxConfig::new();
    /// ```
    pub fn new() -> Self {
        Self {
            allowed_dirs: HashSet::new(),
            strict_mode: true,
            allow_symlinks: false,
            max_symlink_depth: 8,
        }
    }

    /// 创建带有指定白名单的沙箱配置
    ///
    /// # 参数
    /// * `allowed_dirs` - 允许访问的目录列表
    ///
    /// # 示例
    /// ```
    /// use agentflow_core::sandbox::SandboxConfig;
    /// use std::path::PathBuf;
    ///
    /// let allowed = vec![
    ///     PathBuf::from("/workspace"),
    ///     PathBuf::from("/tmp/task_work"),
    /// ];
    ///
    /// let config = SandboxConfig::with_allowed_dirs(allowed);
    /// ```
    pub fn with_allowed_dirs(allowed_dirs: Vec<PathBuf>) -> Self {
        let mut config = Self::new();
        for dir in allowed_dirs {
            config.add_allowed_dir(dir);
        }
        config
    }

    /// 添加允许访问的目录
    ///
    /// # 参数
    /// * `dir` - 要添加到白名单的目录路径
    ///
    /// # 注意
    /// - 路径会被规范化（解析 .. 和 .）
    /// - 路径会被转换为绝对路径
    pub fn add_allowed_dir(&mut self, dir: PathBuf) -> &mut Self {
        let normalized = Self::normalize_path(&dir);
        debug!("添加允许目录: {} -> {}", dir.display(), normalized.display());
        self.allowed_dirs.insert(normalized);
        self
    }

    /// 移除允许访问的目录
    ///
    /// # 参数
    /// * `dir` - 要从白名单移除的目录路径
    pub fn remove_allowed_dir(&mut self, dir: &Path) -> &mut Self {
        let normalized = Self::normalize_path(dir);
        debug!("移除允许目录: {}", normalized.display());
        self.allowed_dirs.remove(&normalized);
        self
    }

    /// 设置是否启用严格模式
    ///
    /// # 参数
    /// * `strict` - 是否启用严格模式
    ///
    /// # 注意
    /// 严格模式下，所有不在白名单中的访问都会被拒绝
    pub fn set_strict_mode(&mut self, strict: bool) -> &mut Self {
        debug!("设置严格模式: {}", strict);
        self.strict_mode = strict;
        self
    }

    /// 设置是否允许符号链接跟随
    ///
    /// # 参数
    /// * `allow` - 是否允许符号链接
    pub fn set_allow_symlinks(&mut self, allow: bool) -> &mut Self {
        debug!("设置允许符号链接: {}", allow);
        self.allow_symlinks = allow;
        self
    }

    /// 设置最大符号链接跟随深度
    ///
    /// # 参数
    /// * `depth` - 最大深度
    pub fn set_max_symlink_depth(&mut self, depth: usize) -> &mut Self {
        debug!("设置最大符号链接深度: {}", depth);
        self.max_symlink_depth = depth;
        self
    }

    /// 验证路径是否在白名单中
    ///
    /// # 参数
    /// * `path` - 要验证的路径
    ///
    /// # 返回
    /// - Ok(()) 如果路径在白名单中
    /// - Err(SandboxError) 如果路径不在白名单中或检测到攻击
    ///
    /// # 示例
    /// ```no_run
    /// # use agentflow_core::sandbox::SandboxConfig;
    /// # use std::path::PathBuf;
    /// #
    /// # fn example() -> anyhow::Result<()> {
    /// let mut config = SandboxConfig::new();
    /// config.add_allowed_dir(PathBuf::from("/workspace"));
    ///
    /// config.validate_path(Path::new("/workspace/file.txt"))?;
    /// # Ok(())
    /// # }
    /// ```
    pub fn validate_path(&self, path: &Path) -> Result<(), SandboxError> {
        debug!("验证路径: {}", path.display());

        // 1. 规范化路径
        let normalized = Self::normalize_path(path);
        debug!("规范化后路径: {}", normalized.display());

        // 2. 检测路径穿透攻击
        if Self::contains_path_traversal(path) {
            warn!("检测到路径穿透攻击: {}", path.display());
            return Err(SandboxError::PathTraversalDetected(
                path.display().to_string(),
            ));
        }

        // 3. 检查符号链接攻击
        if !self.allow_symlinks {
            if let Err(e) = self.check_symlink_attack(&normalized, 0) {
                warn!("检测到符号链接攻击: {}", e);
                return Err(e);
            }
        }

        // 4. 检查是否在白名单中
        if self.strict_mode {
            let is_allowed = self.is_path_allowed(&normalized);
            if !is_allowed {
                warn!("路径不在白名单中: {}", normalized.display());
                return Err(SandboxError::PathNotAllowed(
                    normalized.display().to_string(),
                ));
            }
        }

        debug!("路径验证通过: {}", normalized.display());
        Ok(())
    }

    /// 验证多个路径
    ///
    /// # 参数
    /// * `paths` - 要验证的路径列表
    ///
    /// # 返回
    /// - Ok(()) 如果所有路径都通过验证
    /// - Err(SandboxError) 如果任何路径验证失败
    pub fn validate_paths(&self, paths: &[&Path]) -> Result<(), SandboxError> {
        for path in paths {
            self.validate_path(path)?;
        }
        Ok(())
    }

    /// 获取允许的目录列表
    pub fn allowed_dirs(&self) -> &HashSet<PathBuf> {
        &self.allowed_dirs
    }

    /// 检查路径是否在白名单中
    ///
    /// # 参数
    /// * `path` - 要检查的路径（必须是规范化的绝对路径）
    ///
    /// # 返回
    /// - true 如果路径在白名单中
    /// - false 如果路径不在白名单中
    fn is_path_allowed(&self, path: &Path) -> bool {
        // 检查路径或其任何父目录是否在白名单中
        for allowed_dir in &self.allowed_dirs {
            if path.starts_with(allowed_dir) {
                return true;
            }
        }
        false
    }

    /// 检测路径穿透攻击
    ///
    /// # 参数
    /// * `path` - 要检查的路径
    ///
    /// # 返回
    /// - true 如果检测到路径穿透
    /// - false 如果路径安全
    fn contains_path_traversal(path: &Path) -> bool {
        let path_str = path.to_string_lossy();

        // 检查可疑模式
        // 1. 包含 "../" 或 "..\" (父目录遍历)
        if path_str.contains("../") || path_str.contains("..\\") {
            return true;
        }

        // 2. 路径组件以 ".." 开头
        if path.components().any(|c| {
            c.as_os_str().to_string_lossy().starts_with("..")
        }) {
            return true;
        }

        false
    }

    /// 检查符号链接攻击
    ///
    /// # 参数
    /// * `path` - 要检查的路径
    /// * `depth` - 当前递归深度
    ///
    /// # 返回
    /// - Ok(()) 如果没有符号链接攻击
    /// - Err(SandboxError) 如果检测到符号链接攻击
    fn check_symlink_attack(&self, path: &Path, depth: usize) -> Result<(), SandboxError> {
        // 防止无限递归
        if depth > self.max_symlink_depth {
            return Err(SandboxError::SymlinkAttack(
                "符号链接深度超过限制".to_string(),
            ));
        }

        // 检查路径是否存在
        if !path.exists() {
            // 路径不存在，无法检查符号链接，假设安全
            return Ok(());
        }

        // 检查是否是符号链接
        if path.is_symlink() {
            // 解析符号链接目标
            let target = std::fs::read_link(path)
                .map_err(|e| SandboxError::PathResolutionFailed(format!(
                    "无法读取符号链接: {} - {}",
                    path.display(),
                    e
                )))?;

            debug!("发现符号链接: {} -> {}", path.display(), target.display());

            // 检查目标是否在白名单中
            let normalized_target = Self::normalize_path(&target);
            if !self.is_path_allowed(&normalized_target) {
                return Err(SandboxError::SymlinkAttack(format!(
                    "符号链接指向白名单外的路径: {} -> {}",
                    path.display(),
                    target.display()
                )));
            }

            // 递归检查符号链接目标
            return self.check_symlink_attack(&normalized_target, depth + 1);
        }

        // 如果是目录，递归检查父目录
        if let Some(parent) = path.parent() {
            self.check_symlink_attack(parent, depth)?;
        }

        Ok(())
    }

    /// 规范化路径
    ///
    /// 解析路径中的所有相对组件（. 和 ..）和符号链接
    ///
    /// # 参数
    /// * `path` - 要规范化的路径
    ///
    /// # 返回
    /// 规范化后的绝对路径
    fn normalize_path(path: &Path) -> PathBuf {
        // 转换为绝对路径
        let abs_path = if path.is_absolute() {
            path.to_path_buf()
        } else {
            std::env::current_dir()
                .unwrap_or_else(|_| PathBuf::from("/"))
                .join(path)
        };

        // 规范化路径（移除 . 和 ..）
        let normalized = abs_path
            .canonicalize()
            .unwrap_or_else(|_| abs_path.to_path_buf());

        normalized
    }

    /// 创建安全的子路径
    ///
    /// 在允许的目录下创建安全的子路径
    ///
    /// # 参数
    /// * `base_dir` - 基础目录（必须在白名单中）
    /// * `sub_path` - 子路径
    ///
    /// # 返回
    /// - Ok(PathBuf) 安全的完整路径
    /// - Err(SandboxError) 如果路径不安全
    pub fn create_safe_path(&self, base_dir: &Path, sub_path: &Path) -> Result<PathBuf, SandboxError> {
        // 验证基础目录
        self.validate_path(base_dir)?;

        // 构建完整路径
        let full_path = base_dir.join(sub_path);

        // 规范化路径
        let normalized = Self::normalize_path(&full_path);

        // 确保结果路径仍然在基础目录下
        let normalized_base = Self::normalize_path(base_dir);
        if !normalized.starts_with(&normalized_base) {
            return Err(SandboxError::PathTraversalDetected(format!(
                "子路径逃逸了基础目录: {} (base: {})",
                normalized.display(),
                normalized_base.display()
            )));
        }

        // 验证最终路径
        self.validate_path(&normalized)?;

        Ok(normalized)
    }

    /// 检查路径是否为安全的文件路径（不是目录穿越）
    ///
    /// # 参数
    /// * `filename` - 文件名
    ///
    /// # 返回
    /// - true 如果文件名安全
    /// - false 如果文件名包含路径分隔符或特殊字符
    pub fn is_safe_filename(filename: &str) -> bool {
        // 检查路径分隔符
        if filename.contains('/') || filename.contains('\\') {
            return false;
        }

        // 检查特殊字符和字符串
        let forbidden_chars = ['~', '$', '\0'];

        for forbidden_char in &forbidden_chars {
            if filename.contains(*forbidden_char) {
                return false;
            }
        }

        // 检查特殊的危险字符串（必须完全匹配或作为路径组件）
        if filename == ".." || filename == "." {
            return false;
        }

        true
    }

    /// 获取沙箱配置摘要
    ///
    /// # 返回
    /// 返回配置摘要信息
    pub fn summary(&self) -> SandboxSummary {
        SandboxSummary {
            allowed_dirs_count: self.allowed_dirs.len(),
            strict_mode: self.strict_mode,
            allow_symlinks: self.allow_symlinks,
            max_symlink_depth: self.max_symlink_depth,
        }
    }
}

/// 沙箱配置摘要
#[derive(Debug, Clone)]
pub struct SandboxSummary {
    /// 允许的目录数量
    pub allowed_dirs_count: usize,

    /// 是否启用严格模式
    pub strict_mode: bool,

    /// 是否允许符号链接
    pub allow_symlinks: bool,

    /// 最大符号链接深度
    pub max_symlink_depth: usize,
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    #[test]
    fn test_sandbox_config() {
        let mut config = SandboxConfig::new();
        config.add_allowed_dir(PathBuf::from("/workspace"));

        // 测试有效路径
        assert!(config.validate_path(Path::new("/workspace/file.txt")).is_ok());
        assert!(config.validate_path(Path::new("/workspace/subdir/file.txt")).is_ok());

        // 测试路径穿透检测
        assert!(matches!(
            config.validate_path(Path::new("/workspace/../../../etc/passwd")),
            Err(SandboxError::PathTraversalDetected(_))
        ));

        // 测试白名单外路径
        assert!(matches!(
            config.validate_path(Path::new("/etc/passwd")),
            Err(SandboxError::PathNotAllowed(_))
        ));
    }

    #[test]
    fn test_is_safe_filename() {
        assert!(SandboxConfig::is_safe_filename("file.txt"));
        assert!(SandboxConfig::is_safe_filename("document.pdf"));

        assert!(!SandboxConfig::is_safe_filename("../file.txt"));
        assert!(!SandboxConfig::is_safe_filename("/etc/passwd"));
        assert!(!SandboxConfig::is_safe_filename("file?.txt"));
    }

    #[test]
    fn test_create_safe_path() {
        let config = SandboxConfig::with_allowed_dirs(vec![PathBuf::from("/workspace")]);

        // 测试安全子路径
        let result = config.create_safe_path(Path::new("/workspace"), Path::new("subdir/file.txt"));
        assert!(result.is_ok());

        // 测试路径穿透
        let result = config.create_safe_path(Path::new("/workspace"), Path::new("../etc/passwd"));
        assert!(matches!(result, Err(SandboxError::PathTraversalDetected(_))));
    }

    #[test]
    fn test_sandbox_summary() {
        let mut config = SandboxConfig::new();
        config.add_allowed_dir(PathBuf::from("/workspace"));
        config.add_allowed_dir(PathBuf::from("/tmp"));

        let summary = config.summary();
        assert_eq!(summary.allowed_dirs_count, 2);
        assert!(summary.strict_mode);
    }
}
