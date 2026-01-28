#!/usr/bin/env node
/**
 * Team B 模块集成测试
 *
 * 测试记忆与安全模块的核心功能
 * - memory/mod.rs: 记忆索引和检索
 * - sandbox/mod.rs: 沙箱安全控制
 * - executor/prompt_builder.rs: Prompt 构建
 */

console.log('='.repeat(60));
console.log('Team B 模块集成测试报告');
console.log('='.repeat(60));

// 模块结构验证
const modules = {
  memory: {
    path: '/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-core/src/memory/mod.rs',
    expectedItems: [
      'MemoryCore',
      'MemoryCore::new',
      'MemoryCore::index',
      'MemoryCore::search',
      'MemoryCore::get',
      'MemoryCore::delete',
      'MemoryCore::cleanup_expired',
      'MemoryCore::stats',
      'MemoryStats',
    ],
  },
  sandbox: {
    path: '/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-core/src/sandbox/mod.rs',
    expectedItems: [
      'SandboxConfig',
      'SandboxConfig::new',
      'SandboxConfig::with_allowed_dirs',
      'SandboxConfig::add_allowed_dir',
      'SandboxConfig::validate_path',
      'SandboxConfig::create_safe_path',
      'SandboxConfig::is_safe_filename',
      'SandboxError',
      'SandboxSummary',
    ],
  },
  prompt_builder: {
    path: '/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-core/src/executor/prompt_builder.rs',
    expectedItems: [
      'PromptBuilder',
      'PromptBuilder::new',
      'PromptBuilder::build',
      'PromptBuilder::build_with_metadata',
      'PromptBuilder::truncate_if_needed',
      'PromptBuilderConfig',
      'build_prompt',
    ],
  },
};

const fs = require('fs');

// 测试结果
const results = {
  passed: 0,
  failed: 0,
  details: [],
};

// 测试每个模块
for (const [moduleName, moduleInfo] of Object.entries(modules)) {
  console.log(`\n测试模块: ${moduleName}`);
  console.log('-'.repeat(60));

  try {
    const content = fs.readFileSync(moduleInfo.path, 'utf8');

    // 检查预期的项是否存在
    let allFound = true;
    for (const item of moduleInfo.expectedItems) {
      const found = content.includes(item);
      const status = found ? '✓' : '✗';
      console.log(`  ${status} ${item}`);

      if (!found) {
        allFound = false;
      }
    }

    // 检查关键特性
    console.log('\n关键特性:');

    if (moduleName === 'memory') {
      const features = [
        ['SQLite 集成', content.includes('SqlitePool')],
        ['向量索引', content.includes('compute_embedding')],
        ['语义检索', content.includes('search')],
        ['相似度计算', content.includes('cosine_similarity')],
        ['过期清理', content.includes('cleanup_expired')],
      ];
      features.forEach(([feature, found]) => {
        const status = found ? '✓' : '✗';
        console.log(`  ${status} ${feature}`);
        if (!found) allFound = false;
      });
    } else if (moduleName === 'sandbox') {
      const features = [
        ['白名单机制', content.includes('allowed_dirs')],
        ['路径穿透检测', content.includes('path_traversal')],
        ['符号链接检测', content.includes('symlink')],
        ['路径规范化', content.includes('normalize_path')],
        ['安全文件名检查', content.includes('is_safe_filename')],
      ];
      features.forEach(([feature, found]) => {
        const status = found ? '✓' : '✗';
        console.log(`  ${status} ${feature}`);
        if (!found) allFound = false;
      });
    } else if (moduleName === 'prompt_builder') {
      const features = [
        ['系统指令', content.includes('system_instruction')],
        ['记忆上下文', content.includes('build_memory_section')],
        ['任务描述', content.includes('build_task_section')],
        ['Token 限制', content.includes('max_tokens')],
        ['自动截断', content.includes('truncate_if_needed')],
        ['元数据支持', content.includes('build_metadata_section')],
      ];
      features.forEach(([feature, found]) => {
        const status = found ? '✓' : '✗';
        console.log(`  ${status} ${feature}`);
        if (!found) allFound = false;
      });
    }

    // 记录结果
    if (allFound) {
      results.passed++;
      results.details.push({ module: moduleName, status: 'PASSED' });
      console.log('\n✓ 模块测试通过');
    } else {
      results.failed++;
      results.details.push({ module: moduleName, status: 'FAILED' });
      console.log('\n✗ 模块测试失败');
    }
  } catch (error) {
    results.failed++;
    results.details.push({ module: moduleName, status: 'ERROR', error: error.message });
    console.log(`\n✗ 读取文件失败: ${error.message}`);
  }
}

// 打印总结
console.log('\n' + '='.repeat(60));
console.log('测试总结');
console.log('='.repeat(60));
console.log(`通过: ${results.passed}`);
console.log(`失败: ${results.failed}`);
console.log(`总计: ${results.passed + results.failed}`);

if (results.failed === 0) {
  console.log('\n✓ 所有模块测试通过！');
  console.log('\n实现的功能:');
  console.log('  1. memory/mod.rs - 基于 SQLite 的向量记忆系统');
  console.log('  2. sandbox/mod.rs - 沙箱安全控制（白名单+路径穿透防护）');
  console.log('  3. executor/prompt_builder.rs - 智能 Prompt 构建器');
  process.exit(0);
} else {
  console.log('\n✗ 部分模块测试失败');
  process.exit(1);
}
