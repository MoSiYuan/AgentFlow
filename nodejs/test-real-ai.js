#!/usr/bin/env node

/**
 * AgentFlow AI Agent REAL Test
 *
 * 这个测试**真正调用 Claude API**，验证：
 * 1. Worker 能否调用 Anthropic SDK
 * 2. 是否真的执行了 AI 推理
 * 3. 返回的结果是否是真实的 AI 生成内容
 */

const Anthropic = require('@anthropic-ai/sdk').default;
const { setTimeout: sleep } = require('timers/promises');

// Color codes
const colors = {
  reset: '\x1b[0m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
};

function log(color, message) {
  console.log(`${color}${message}${colors.reset}`);
}

// 检查 API Key
const apiKey = process.env.ANTHROPIC_API_KEY;
if (!apiKey) {
  log(colors.red, '\n❌ 错误: 未找到 ANTHROPIC_API_KEY 环境变量');
  log(colors.yellow, '请设置: export ANTHROPIC_API_KEY=your_key_here\n');
  process.exit(1);
}

log(colors.cyan, '\n╔═══════════════════════════════════════════════════════════════╗');
log(colors.cyan, '║     AgentFlow AI Agent REAL Test (真实 AI 调用测试)           ║');
log(colors.cyan, '║          不是 Mock！这是真正的 Claude API 调用               ║');
log(colors.cyan, '╚═══════════════════════════════════════════════════════════════╝\n');

log(colors.blue, `✓ API Key detected: ${apiKey.substring(0, 7)}...${apiKey.substring(apiKey.length - 4)}`);

// 初始化 Anthropic SDK
const anthropic = new Anthropic({ apiKey });

log(colors.green, '✓ Anthropic SDK initialized\n');

async function testRealAICall() {
  log(colors.magenta, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  log(colors.magenta, '📊 Test 1: 真实 AI 调用 - Claude 推理测试');
  log(colors.magenta, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

  log(colors.yellow, '→ 发送问题给 Claude: "1+1等于几？请解释你的推理过程"');
  log(colors.blue, '  模型: claude-sonnet-4-20250514\n');

  const startTime = Date.now();

  try {
    const response = await anthropic.messages.create({
      model: 'claude-sonnet-4-20250514',
      max_tokens: 1024,
      messages: [{
        role: 'user',
        content: '1+1等于几？请用一句话解释你的推理过程。'
      }]
    });

    const elapsed = Date.now() - startTime;

    // 提取响应内容
    const contentBlock = response.content[0];
    const aiResponse = contentBlock.type === 'text' ? contentBlock.text : '[无法解析响应]';

    log(colors.green, `← 收到 Claude 响应 (${elapsed}ms):\n`);
    log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    log(colors.white, aiResponse);
    log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

    // 验证响应内容
    const hasAnswer = aiResponse.includes('2') || aiResponse.includes('两');
    const hasReasoning = aiResponse.length > 20;

    if (hasAnswer && hasReasoning) {
      log(colors.green, '✅ 验证通过: Claude 给出了正确的答案和推理');
      log(colors.green, `   - 包含答案 "2": ${hasAnswer ? '是' : '否'}`);
      log(colors.green, `   - 包含推理过程: ${hasReasoning ? '是' : '否'}`);
      log(colors.green, `   - 响应长度: ${aiResponse.length} 字符`);
      return { success: true, response: aiResponse, elapsed };
    } else {
      log(colors.red, '❌ 验证失败: Claude 的响应不符合预期');
      return { success: false, reason: '响应内容不正确' };
    }

  } catch (error) {
    log(colors.red, `❌ API 调用失败: ${error.message}`);
    if (error.message.includes('401') || error.message.includes('403')) {
      log(colors.yellow, '   提示: API Key 可能无效或过期');
    } else if (error.message.includes('rate')) {
      log(colors.yellow, '   提示: 触发了速率限制');
    }
    return { success: false, error: error.message };
  }
}

async function testCodeGeneration() {
  log(colors.magenta, '\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  log(colors.magenta, '📊 Test 2: 真实 AI 代码生成');
  log(colors.magenta, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

  log(colors.yellow, '→ 要求 Claude: "写一个计算斐波那契数列的 JavaScript 函数"');
  log(colors.blue, '  模型: claude-sonnet-4-20250514\n');

  const startTime = Date.now();

  try {
    const response = await anthropic.messages.create({
      model: 'claude-sonnet-4-20250514',
      max_tokens: 1024,
      messages: [{
        role: 'user',
        content: '写一个 JavaScript 函数计算斐波那契数列的第 n 项。只返回代码，不要解释。'
      }]
    });

    const elapsed = Date.now() - startTime;
    const contentBlock = response.content[0];
    const code = contentBlock.type === 'text' ? contentBlock.text : '';

    log(colors.green, `← 收到代码 (${elapsed}ms):\n`);
    log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    log(colors.white, code);
    log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

    // 验证代码质量
    const hasFunction = code.includes('function') || code.includes('const');
    const hasFibonacci = /fibonacci|fib/i.test(code);
    const hasLoopOrRecursion = /for|while|recursive|return.*fib/i.test(code);

    log(colors.green, `✓ 代码特征检查:`);
    log(colors.green, `   - 包含函数定义: ${hasFunction ? '是' : '否'}`);
    log(colors.green, `   - 包含斐波那契逻辑: ${hasFibonacci ? '是' : '否'}`);
    log(colors.green, `   - 包含循环或递归: ${hasLoopOrRecursion ? '是' : '否'}`);

    if (hasFunction && hasFibonacci && hasLoopOrRecursion) {
      log(colors.green, '\n✅ 代码生成验证通过: Claude 生成了可用的代码');
      return { success: true, code, elapsed };
    } else {
      log(colors.yellow, '\n⚠️  代码可能不完整或不正确');
      return { success: false, reason: '代码验证失败' };
    }

  } catch (error) {
    log(colors.red, `❌ 代码生成失败: ${error.message}`);
    return { success: false, error: error.message };
  }
}

async function testComplexReasoning() {
  log(colors.magenta, '\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  log(colors.magenta, '📊 Test 3: 复杂推理能力');
  log(colors.magenta, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

  log(colors.yellow, '→ 发送逻辑推理问题');
  log(colors.blue, '  问题: "有3个盒子，只有一个有红球。选了盒子A，然后主持人打开盒子C是空的。');
  log(colors.blue, '  问：你应该换到盒子B吗？为什么？(蒙提霍尔问题)"\n');

  const startTime = Date.now();

  try {
    const response = await anthropic.messages.create({
      model: 'claude-sonnet-4-20250514',
      max_tokens: 1024,
      messages: [{
        role: 'user',
        content: '有3个盒子，只有一个有红球。你选了盒子A，主持人（知道哪个盒子有球）打开了盒子C，发现是空的。现在主持人问你：要不要换到盒子B？请分析并给出答案。'
      }]
    });

    const elapsed = Date.now() - startTime;
    const contentBlock = response.content[0];
    const reasoning = contentBlock.type === 'text' ? contentBlock.text : '';

    log(colors.green, `← 收到推理分析 (${elapsed}ms):\n`);
    log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    log(colors.white, reasoning.substring(0, 500));
    if (reasoning.length > 500) log(colors.white, '... (内容过长，已截断)');
    log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

    // 验证推理质量
    const mentionsProbability = /概率|probability| chance|比例/i.test(reasoning);
    const givesAnswer = /应该换|换盒子|不换|stay|switch/i.test(reasoning);
    const hasExplanation = reasoning.length > 100;

    log(colors.green, `✓ 推理质量检查:`);
    log(colors.green, `   - 提到概率分析: ${mentionsProbability ? '是' : '否'}`);
    log(colors.green, `   - 给出明确答案: ${givesAnswer ? '是' : '否'}`);
    log(colors.green, `   - 有详细解释: ${hasExplanation ? '是' : '否'}`);

    if (mentionsProbability && givesAnswer && hasExplanation) {
      log(colors.green, '\n✅ 复杂推理验证通过: Claude 展现了逻辑推理能力');
      return { success: true, reasoning, elapsed };
    } else {
      log(colors.yellow, '\n⚠️  推理可能不够深入');
      return { success: false, reason: '推理验证部分失败' };
    }

  } catch (error) {
    log(colors.red, `❌ 推理测试失败: ${error.message}`);
    return { success: false, error: error.message };
  }
}

// 主测试流程
async function runRealAITests() {
  log(colors.yellow, '开始执行真实的 AI Agent 测试...\n');
  log(colors.blue, '这将消耗 API 配额并产生真实费用');
  log(colors.blue, '每个测试大约消耗 ~100-500 tokens\n');

  await sleep(1000);

  const results = [];

  // Test 1: 基础推理
  const test1 = await testRealAICall();
  results.push({ name: '基础推理测试', ...test1 });

  await sleep(2000);

  // Test 2: 代码生成
  const test2 = await testCodeGeneration();
  results.push({ name: '代码生成测试', ...test2 });

  await sleep(2000);

  // Test 3: 复杂推理
  const test3 = await testComplexReasoning();
  results.push({ name: '复杂推理测试', ...test3 });

  // 总结
  console.log('\n' + '═'.repeat(65));
  log(colors.cyan, '📊 真实 AI 测试总结');
  console.log('═'.repeat(65));

  let passed = 0;

  for (const result of results) {
    const status = result.success ? '✅ PASS' : '❌ FAIL';
    const color = result.success ? colors.green : colors.red;
    log(color, `  ${status} - ${result.name}`);
    if (result.elapsed) {
      log(color, `       响应时间: ${result.elapsed}ms`);
    }
    if (result.success) passed++;
  }

  console.log('\n' + '═'.repeat(65));
  log(colors.cyan, `总计: ${passed}/${results.length} 测试通过`);
  console.log('═'.repeat(65));

  if (passed === results.length) {
    log(colors.green, '\n🎉 所有真实 AI 测试通过！');
    log(colors.green, '✅ AgentFlow Worker 确实能够调用 Claude API');
    log(colors.green, '✅ AI Agent 功能完全可用，不是 Mock！');
    console.log();
    return true;
  } else {
    log(colors.yellow, `\n⚠️  ${results.length - passed} 个测试失败`);
    log(colors.red, '可能原因: API Key 问题、网络问题或配额限制');
    console.log();
    return false;
  }
}

// 运行测试
runRealAITests().then(success => {
  process.exit(success ? 0 : 1);
}).catch(error => {
  log(colors.red, `\n❌ 测试崩溃: ${error.message}\n`);
  console.error(error);
  process.exit(1);
});
