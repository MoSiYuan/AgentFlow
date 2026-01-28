//! Claude CLI 集成测试程序
//!
//! 独立测试 Claude CLI 和 Skills 发现功能

use std::path::PathBuf;
use std::process::Command;

/// Skills 发现测试
fn test_skills_discovery() {
    println!("=== Skills 目录发现测试 ===\n");

    let skills_dirs = discover_skills_directories();

    println!("发现 {} 个 Skills 目录:", skills_dirs.len());
    for (i, dir) in skills_dirs.iter().enumerate() {
        println!("  {}. {:?}", i + 1, dir);
    }

    // 统计可用 skills
    let mut total_skills = 0;
    for skills_dir in &skills_dirs {
        if let Ok(entries) = std::fs::read_dir(skills_dir) {
            for entry in entries.flatten() {
                if let Ok(file_type) = entry.file_type() {
                    if file_type.is_dir() {
                        let skill_md = entry.path().join("SKILL.md");
                        if skill_md.exists() {
                            total_skills += 1;
                            println!("    ✓ {}", entry.path().display());
                        }
                    }
                }
            }
        }
    }

    println!("\n总共发现 {} 个 Skills\n", total_skills);
}

/// Claude CLI 查找测试
fn test_claude_cli() {
    println!("=== Claude CLI 查找测试 ===\n");

    if let Some(claude_path) = find_claude_cli() {
        println!("✓ 找到 Claude CLI: {:?}", claude_path);

        // 测试版本
        let output = Command::new(&claude_path)
            .arg("--version")
            .output();

        match output {
            Ok(output) => {
                let version = String::from_utf8_lossy(&output.stdout);
                println!("  版本: {}", version.trim());
            }
            Err(e) => {
                println!("  获取版本失败: {}", e);
            }
        }
    } else {
        println!("✗ 未找到 Claude CLI");
        println!("  请安装: npm install -g @anthropic-ai/claude-code");
    }

    println!();
}

/// 发现 Skills 目录
fn discover_skills_directories() -> Vec<PathBuf> {
    let mut skills_dirs = Vec::new();

    // 1. 默认目录
    if let Ok(home) = std::env::var("HOME") {
        let default_skills = PathBuf::from(home).join(".claude/skills");
        if default_skills.exists() {
            skills_dirs.push(default_skills);
        }
    }

    // 2. 项目目录
    if let Ok(cwd) = std::env::current_dir() {
        let project_skills = cwd.join(".claude/skills");
        if project_skills.exists() {
            skills_dirs.push(project_skills);
        }
    }

    // 3. 环境变量
    if let Ok(custom_paths) = std::env::var("AGENTFLOW_SKILLS_PATHS") {
        for path in custom_paths.split(':') {
            let path_buf = PathBuf::from(path.trim());
            if path_buf.exists() {
                skills_dirs.push(path_buf);
            }
        }
    }

    // 4. 全局目录
    let global_skills = PathBuf::from("/usr/local/share/claude/skills");
    if global_skills.exists() {
        skills_dirs.push(global_skills);
    }

    skills_dirs
}

/// 查找 Claude CLI
fn find_claude_cli() -> Option<PathBuf> {
    // 使用 which 命令查找
    if let Ok(output) = Command::new("which").arg("claude").output() {
        if output.status.success() {
            let path_str = String::from_utf8_lossy(&output.stdout);
            let path = path_str.trim();
            if !path.is_empty() {
                return Some(PathBuf::from(path));
            }
        }
    }

    // 检查常见位置
    let common_paths = vec![
        "/opt/homebrew/bin/claude",
        "/usr/local/bin/claude",
        "~/.npm-global/bin/claude",
    ];

    for path in common_paths {
        let expanded = shellexpand::tilde(path);
        let path_buf = PathBuf::from(expanded.as_ref());
        if path_buf.exists() {
            return Some(path_buf);
        }
    }

    None
}

fn main() {
    println!("╔════════════════════════════════════════╗");
    println!("║   AgentFlow Claude CLI 集成测试        ║");
    println!("╚════════════════════════════════════════╝\n");

    test_claude_cli();
    test_skills_discovery();

    println!("=== 测试完成 ===");
}
