#!/usr/bin/env python3
"""
CPDS实战测试 - 批量创建分析任务
为每个克苏鲁故事创建9个分析任务（10x9=90个任务）
"""

import requests
import json
import sys

MASTER_URL = "http://localhost:8848"

def main():
    print("=" * 60)
    print("CPDS实战测试 - 创建分析任务")
    print("=" * 60)
    print()

    # 检查Master是否运行
    try:
        resp = requests.get(f"{MASTER_URL}/api/health", timeout=5)
        if resp.status_code != 200:
            print("❌ Master health check failed")
            sys.exit(1)
    except Exception as e:
        print(f"❌ Cannot connect to Master: {e}")
        print("   Please start Master first:")
        print("   ./cpds master --mode standalone --auto-shutdown")
        sys.exit(1)

    print("✅ Master is running")
    print()

    # 获取已完成的故事
    print("📚 Fetching completed stories...")
    resp = requests.get(f"{MASTER_URL}/api/tasks/completed")
    all_tasks = resp.json()['data']['tasks']
    stories = [t for t in all_tasks if t['task_id'].startswith('STORY')]

    if len(stories) == 0:
        print("❌ No stories found!")
        print("   Please complete story creation first")
        sys.exit(1)

    print(f"✅ Found {len(stories)} stories")
    print()

    # 创建分析任务
    print("📝 Creating analysis tasks...")
    print()

    count = 0
    failed = 0

    for story in stories:
        story_id = story['task_id']
        story_num = story_id.split('-')[1]
        story_title = story['title']
        story_content = story.get('output', '')

        if not story_content:
            print(f"⚠️  Skipping {story_id} - no content")
            continue

        print(f"Story {story_num}: {story_title}")

        # 为每个故事创建9个分析任务（排除作者自己）
        for critic_num in range(1, 11):
            # 跳过作者分析自己的故事
            if int(story_num) == critic_num:
                continue

            analysis_id = f"ANALYSIS-{story_num}-{critic_num:02d}"
            critic_name = f"Critic-{critic_num:02d}"

            # 构建分析任务
            task = {
                "task_id": analysis_id,
                "title": f"{critic_name}分析故事{story_num}",
                "description": f"""作为专业的克苏鲁神话文学评论家，请对以下故事进行深入分析：

【故事标题】
{story_title}

【故事内容】
{story_content}

【分析要求】
1. 字数：300-500字
2. 分析维度：
   - 克苏鲁元素识别（识别3-5个具体的克苏鲁神话元素）
   - 恐怖氛围营造（评价恐怖氛围的营造效果）
   - 创意性和独特性（评估故事的创意程度）
   - 文笔和叙述技巧（评价文笔和叙述手法）
   - 符合洛夫克拉夫特风格（评估是否符合经典洛夫克拉夫特风格）

3. 给出具体评分（每项1-10分）：
   - 克苏鲁元素准确性：__ / 10
   - 恐怖氛围：__ / 10
   - 创意性：__ / 10
   - 文笔：__ / 10
   - 总分：__ / 40

4. 总结评价（2-3句话总结整体印象）

5. 改进建议（给出1-2条具体建议）

请确保分析客观、专业、有深度。""",
                "priority": "medium",
                "tags": ["analysis", "critique", "lovecraft"],
                "deployment_mode": "standalone"
            }

            # 创建任务
            try:
                resp = requests.post(
                    f"{MASTER_URL}/api/tasks/create",
                    json=task,
                    timeout=10
                )

                if resp.status_code == 200:
                    result = resp.json()
                    if result.get('success'):
                        print(f"  ✅ {analysis_id}")
                        count += 1
                    else:
                        print(f"  ❌ {analysis_id}: {result.get('error', 'Unknown error')}")
                        failed += 1
                else:
                    print(f"  ❌ {analysis_id}: HTTP {resp.status_code}")
                    failed += 1
            except Exception as e:
                print(f"  ❌ {analysis_id}: {e}")
                failed += 1

        print()

    # 汇总
    print("=" * 60)
    print("Task Creation Summary")
    print("=" * 60)
    print(f"✅ Successfully created: {count} analysis tasks")
    if failed > 0:
        print(f"❌ Failed: {failed} tasks")
    print()

    if count > 0:
        print("✅ Analysis tasks are ready!")
        print("   Next step: Start Critic Workers")
        print()
        print("Example:")
        print("  for i in {1..10}; do")
        print("    ./cpds worker --mode standalone --master http://localhost:8848 --name Critic-$i --oneshot &")
        print("  done")
    else:
        print("⚠️  No tasks were created")
        print("   Check the error messages above")

if __name__ == "__main__":
    main()
