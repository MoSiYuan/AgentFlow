# 🎯 CPDS实战测试 - 立即开始！

## 📋 测试内容

**10个Agent协作**：
- 10个Writer → 创作10个不同主题的克苏鲁故事
- 90个Critic → 每个故事被9个Agent分析评价
- 1个系统 → 自动汇总生成报告

## ⚡ 3分钟快速开始

### 第1步：启动Master（终端1）

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go
./cpds master --mode standalone --auto-shutdown --port 8848
```

### 第2步：创建任务 + 启动Workers（终端2）

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go

# 创建10个故事任务
for i in {1..10}; do
  curl -X POST http://localhost:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d @docs/cpds-test/tasks/story_$(printf '%02d' $i).json
done

# 启动10个Writer Worker（并行创作）
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Writer-$i" \
    --oneshot &
done
wait
echo "✅ 10个故事创作完成！"
```

### 第3步：创建分析任务 + 启动Critic Workers（终端2）

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go/docs/cpds-test

# 创建90个分析任务
python3 create_analysis_tasks.py

# 启动10个Critic Worker（并行分析）
cd ../..
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Critic-$i" \
    --oneshot &
done
wait
echo "✅ 90份分析完成！"
```

### 第4步：生成报告（终端2）

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go/docs/cpds-test
python3 generate_report.py

# 查看报告
less final_report.md
```

## 📊 预期结果

```
✅ 10个克苏鲁故事（500-1000字/个）
   - 《深海之影》- 深海探险
   - 《旧日支配者》- 古神苏醒
   - 《疯狂山脉》- 南极探险
   - ... (共10个主题)

✅ 90份专业分析报告（300-500字/份）
   - 每个故事被9个Critic分析
   - 包含评分、评价、建议

✅ 完整测试报告
   - 所有故事全文
   - 所有分析全文
   - 统计数据和结论
```

## 📚 完整文档

详细指南请查看：
- **[USAGE_GUIDE.md](USAGE_GUIDE.md)** - 7步完整指南（推荐）
- **[README.md](README.md)** - 测试概述
- **[test_plan.md](test_plan.md)** - 详细计划

## 🎉 开始测试

现在就可以开始！打开3个终端，按照上面的步骤执行。

测试完成后，您将看到：
- ✅ CPDS分布式系统正常工作
- ✅ 10个Agent并行协作
- ✅ 100个任务成功执行
- ✅ 完整的克苏鲁故事集和分析报告

**立即开始**：打开终端1，执行第1步！
