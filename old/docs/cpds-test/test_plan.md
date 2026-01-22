# CPDS实战测试：10个Agent协作创作克苏鲁故事

## 测试目标

验证CPDS系统的多Worker并行协作能力：
1. **并行创作**：10个Worker同时创作不同的克苏鲁故事
2. **互相评价**：每个Worker分析其他9个故事
3. **完整流程**：Master-Worker架构、任务分发、结果汇总

## 测试场景

### 阶段1：故事创作（10个并行任务）

- **Worker 1**: 创作《深海之影》- 深海探险主题
- **Worker 2**: 创作《旧日支配者》- 古神苏醒主题
- **Worker 3**: 创作《疯狂山脉》- 南极探险主题
- **Worker 4**: 创作《印斯茅斯之影》- 变异主题
- **Worker 5**: 创作《暗夜低语》- 梦境主题
- **Worker 6**: 创作《被诅咒的城市》- 拉莱耶主题
- **Worker 7**: 创作《不可名状之恐怖》- 宇宙恐怖主题
- **Worker 8**: 创作《古籍守护者》- 禁忌知识主题
- **Worker 9**: 创作《时间的裂缝》- 时空扭曲主题
- **Worker 10**: 创作《最后的守夜人》- 疯狂与理智主题

### 阶段2：故事分析（90个分析任务）

每个Worker分析其他9个故事：
- 识别克苏鲁神话元素
- 分析恐怖氛围
- 评价创意和文笔
- 给出综合评分

## 测试步骤

### 1. 启动Master

```bash
./cpds master --mode standalone --port 8848
```

### 2. 创建故事创作任务

```bash
# 为10个Worker创建10个不同的故事创作任务
for i in {1..10}; do
  curl -X POST http://localhost:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d @tasks/story_$i.json
done
```

### 3. 启动10个Worker（并行）

```bash
# 启动10个Worker进程
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Writer-$i" \
    --oneshot \
    > logs/worker_$i.log 2>&1 &
done
```

### 4. 收集故事结果

```bash
# 查询已完成的故事
curl http://localhost:8848/api/tasks/completed | jq '.data.tasks[] | select(.task_id | startswith("STORY"))'
```

### 5. 创建分析任务

```bash
# 为每个故事创建9个分析任务（10x9=90个任务）
./scripts/create_analysis_tasks.py
```

### 6. 启动分析Worker

```bash
# 启动10个Worker进行分析
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Critic-$i" \
    --oneshot \
    > logs/critic_$i.log 2>&1 &
done
```

### 7. 汇总分析结果

```bash
# 生成最终报告
./scripts/generate_report.py
```

## 预期结果

1. **10个原创克苏鲁故事** - 每个约500-1000字
2. **90份分析报告** - 每个故事被9个不同的Agent评价
3. **综合评分** - 统计每个故事的得分
4. **最佳故事** - 得分最高的故事

## 成功标准

- ✅ 所有10个故事创作任务成功完成
- ✅ 所有90个分析任务成功完成
- ✅ 故事内容不重复，主题各异
- ✅ 分析评价合理、有深度
- ✅ Master自动关闭（所有任务完成后）
- ✅ 无Worker崩溃或超时

## 测试指标

### 性能指标
- 任务创建时间
- Worker启动时间
- 任务执行时间
- 总完成时间

### 质量指标
- 故事创意性
- 克苏鲁元素准确性
- 恐怖氛围营造
- 分析评价的深度

### 系统指标
- Master内存占用
- Worker内存占用
- HTTP请求成功率
- 错误重试次数

## 测试数据

生成的测试数据将保存在：
- `docs/cpds-test/stories/` - 10个故事
- `docs/cpds-test/analysis/` - 90份分析报告
- `docs/cpds-test/final_report.md` - 综合报告
- `docs/cpds-test/logs/` - 执行日志
