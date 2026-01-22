# CPDS实战测试文档总览

## 📚 文档清单

本次CPDS实战测试包含以下完整文档：

### 1. 核心指南

- **[README.md](README.md)** - 测试概述和快速开始
- **[USAGE_GUIDE.md](USAGE_GUIDE.md)** - 完整的7步使用指南（推荐）
- **[QUICKSTART.md](QUICKSTART.md)** - 快速参考

### 2. 测试数据

**10个故事任务配置** ([tasks/](tasks/)):
1. `story_01.json` - 《深海之影》- 深海探险
2. `story_02.json` - 《旧日支配者》- 古神苏醒
3. `story_03.json` - 《疯狂山脉》- 南极探险
4. `story_04.json` - 《印斯茅斯之影》- 变异
5. `story_05.json` - 《暗夜低语》- 梦境
6. `story_06.json` - 《被诅咒的城市》- 拉莱耶
7. `story_07.json` - 《不可名状之恐怖》- 宇宙恐怖
8. `story_08.json` - 《古籍守护者》- 禁忌知识
9. `story_09.json` - 《时间的裂缝》- 时空扭曲
10. `story_10.json` - 《最后的守夜人》- 疯狂与理智

### 3. 辅助脚本

- **[create_analysis_tasks.py](create_analysis_tasks.py)** - 批量创建90个分析任务
- **[generate_report.py](generate_report.py)** - 生成最终测试报告

### 4. 测试计划

- **[test_plan.md](test_plan.md)** - 详细的测试计划和成功标准

---

## 🚀 快速使用

### 方式1：跟随指南（推荐）

```bash
# 打开3个终端

# 终端1：启动Master
cd cpds-go
./cpds master --mode standalone --auto-shutdown --port 8848

# 终端2：创建任务 + 启动Workers
cd cpds-go/docs/cpds-test

# 创建10个故事任务
for i in {1..10}; do
  curl -X POST http://localhost:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d @tasks/story_$(printf '%02d' $i).json
done

# 启动10个Writer Workers
cd ../..
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Writer-$i" \
    --oneshot &
done
wait

# 创建90个分析任务
cd docs/cpds-test
python3 create_analysis_tasks.py

# 启动10个Critic Workers
cd ../..
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Critic-$i" \
    --oneshot &
done
wait

# 生成报告
cd docs/cpds-test
python3 generate_report.py
```

### 方式2：查看完整指南

```bash
cat docs/cpds-test/USAGE_GUIDE.md
```

---

## 📊 测试流程图

```
┌─────────────────────────────────────────────────────┐
│                  CPDS Master                         │
│              (localhost:8848)                       │
└───────────────┬─────────────────────────────────────┘
                │
        ┌───────┴───────┐
        │               │
    ┌───▼────┐     ┌───▼────────┐
    │Tasks   │     │Workers     │
    │10创作   │     │10并行      │
    │90分析   │     │自动执行     │
    └─────────┘     └─────────────┘
        │                   │
        └───────┬───────────┘
                │
        ┌───────▼───────────────┐
        │   Final Report        │
        │ - 10个故事            │
        │ - 90份分析            │
        │ - 统计总结            │
        └───────────────────────┘
```

---

## ✅ 测试检查清单

### 准备阶段
- [ ] CPDS二进制已编译
- [ ] Python 3已安装
- [ ] curl已安装
- [ ] jq已安装（可选）

### 执行阶段
- [ ] Master启动成功
- [ ] 10个故事任务创建成功
- [ ] 10个Writer Worker启动
- [ ] 所有故事任务完成
- [ ] 90个分析任务创建成功
- [ ] 10个Critic Worker启动
- [ ] 所有分析任务完成
- [ ] Master自动关闭

### 验证阶段
- [ ] 10个故事内容完整
- [ ] 故事主题不重复
- [ ] 90份分析报告完整
- [ ] 最终报告生成成功

---

## 📈 预期结果

### 数据统计
- **故事总数**: 10个
- **分析总数**: 90个
- **总任务数**: 100个
- **并行Worker**: 10个
- **执行时间**: 15-30分钟

### 质量标准
- 每个故事500-1000字
- 每份分析300-500字
- 故事主题各不相同
- 分析评价客观专业

---

## 🎯 测试价值

这次测试将验证：

1. ✅ **Master-Worker架构** - 分布式任务调度能力
2. ✅ **并行执行** - 10个Worker同时工作
3. ✅ **任务分发** - 自动分配任务给Workers
4. ✅ **结果收集** - 集中收集所有结果
5. ✅ **自动关闭** - 任务完成后Master自动退出
6. ✅ **数据持久化** - SQLite数据库存储
7. ✅ **REST API** - 9个API端点正常工作

---

## 📝 故事主题预览

每个故事都有独特的克苏鲁神话主题：

1. **深海探险** - 深潜器、海底遗迹、触手生物
2. **古神苏醒** - 考古挖掘、古老符文、群体疯狂
3. **南极探险** - 冰盖遗迹、非人生物、极端环境
4. **身体变异** - 海边小镇、居民异变、深潜杂交
5. **梦境侵蚀** - 诡异梦境、现实扭曲、超自然
6. **沉没古城** - 拉莱耶、巨型石构、星辰归位
7. **宇宙恐怖** - 天文观测、异星信号、宇宙尺度
8. **禁忌知识** - 古老书籍、疯狂学者、守护者诅咒
9. **时空扭曲** - 时间裂缝、循环、过去未来交织
10. **理智边缘** - 精神病院、恐怖记忆、守夜人

---

## 🔧 技术细节

### Master配置
```bash
--mode standalone      # 单机模式
--auto-shutdown        # 自动关闭
--port 8848            # 端口
```

### Worker配置
```bash
--mode standalone      # 单机模式
--master <URL>         # Master地址
--name <NAME>           # Worker名称
--oneshot               # 执行一次即退出
```

### API端点使用
- `POST /api/tasks/create` - 创建任务
- `GET /api/tasks/completed` - 获取已完成任务
- `GET /api/status` - 获取系统状态
- `GET /api/tasks/{task_id}` - 获取特定任务

---

## 🎉 完成后

测试完成后，您将获得：

1. **final_report.md** - 完整的测试报告
   - 10个克苏鲁故事全文
   - 90份专业分析评价
   - 统计数据和结论

2. **验证CPDS系统能力**:
   - 分布式任务调度 ✅
   - 并行执行能力 ✅
   - 数据持久化 ✅
   - REST API稳定性 ✅
   - 自动资源管理 ✅

---

**开始测试**：请查看 [USAGE_GUIDE.md](USAGE_GUIDE.md) 开始执行测试！
