#!/usr/bin/env python3
"""
CPDS Autonomous Worker Test Script

This script simulates CPDS workers that autonomously select topics
and create Lovecraft stories using the Claude API.
"""

import os
import sys
import time
import json
import uuid
import requests
from anthropic import Anthropic

# Configuration
MASTER_URL = "http://localhost:8848"
NUM_WORKERS = 10
API_KEY = os.environ.get("ANTHROPIC_API_KEY")

# Lovecraft story topics for inspiration
LOVECRAFT_TOPICS = [
    "深海遗迹中的古老召唤",
    "被诅咒的家族族谱",
    "南极冰层下的未知文明",
    "会自动生长的诡异书籍",
    "梦境中的呓语与现实重叠",
    "星际间的异教崇拜",
    "被遗忘的海岸小镇传说",
    "古埃及神话的黑暗面",
    "精神病人的真实见闻",
    "现代都市中的克苏鲁迹象",
    "博物馆里的诅咒文物",
    "地下墓穴的恐怖真相",
    "被时间遗忘的孤岛",
    "太空探索的惊人发现",
    "古老语言的解译后果",
]


class AutonomousWorker:
    """Autonomous worker that selects topics and executes tasks"""

    def __init__(self, worker_id, master_url, api_key):
        self.worker_id = worker_id
        self.name = f"Agent-{worker_id}"
        self.master_url = master_url
        self.client = Anthropic(api_key=api_key)
        self.selected_topic = None

    def register(self):
        """Register worker with master"""
        response = requests.post(
            f"{self.master_url}/api/workers/register",
            json={"worker_id": self.worker_id, "name": self.name},
            timeout=5
        )
        if response.status_code == 200:
            print(f"✅ {self.name} registered successfully")
            return True
        else:
            print(f"❌ {self.name} registration failed: {response.text}")
            return False

    def claim_task(self):
        """Claim a pending task"""
        response = requests.get(
            f"{self.master_url}/api/tasks/pending",
            params={"worker_id": self.worker_id},
            timeout=5
        )
        if response.status_code != 200:
            return None

        data = response.json()
        tasks = data.get("data", {}).get("tasks", [])
        if not tasks:
            return None

        # Claim first available task
        task = tasks[0]
        claim_response = requests.post(
            f"{self.master_url}/api/tasks/assign",
            json={"task_id": task["task_id"], "worker_id": self.worker_id},
            timeout=5
        )

        if claim_response.status_code == 200:
            print(f"📋 {self.name} claimed task: {task['task_id']}")
            return task
        else:
            print(f"⚠️  {self.name} failed to claim task: {claim_response.text}")
            return None

    def select_topic(self):
        """Autonomously select a unique topic"""
        # Query existing topics
        response = requests.get(f"{self.master_url}/api/topics", timeout=5)
        existing_topics = set()
        if response.status_code == 200:
            data = response.json()
            existing_topics = {t["topic"] for t in data.get("topics", [])}

        # Select a unique topic
        available_topics = [t for t in LOVECRAFT_TOPICS if t not in existing_topics]

        if not available_topics:
            # Generate custom topic if all predefined ones are taken
            self.selected_topic = f"未探索的克苏鲁领域 #{self.worker_id}"
        else:
            self.selected_topic = available_topics[0]

        # Register the topic
        register_response = requests.post(
            f"{self.master_url}/api/topics/register",
            json={
                "worker_id": self.worker_id,
                "topic": self.selected_topic,
                "story_id": f"story-{self.worker_id}"
            },
            timeout=5
        )

        if register_response.status_code == 200:
            print(f"🎯 {self.name} selected topic: {self.selected_topic}")
            return True
        elif register_response.status_code == 409:
            # Topic already taken, try again
            print(f"⚠️  Topic taken, retrying...")
            time.sleep(0.5)
            return self.select_topic()
        else:
            print(f"❌ Failed to register topic: {register_response.text}")
            return False

    def execute_task_with_claude(self, task):
        """Execute task using Claude API"""
        print(f"🤖 {self.name} is thinking about: {self.selected_topic}")

        prompt = f"""你是专业的克苏鲁神话作家。请按照以下要求创作一个故事：

【选定的题材】{self.selected_topic}

【创作要求】
- 字数：500-1000字
- 风格：洛夫克拉夫特式 Cosmic Horror
- 核心：强调未知恐惧、人类渺小、古老存在
- 语言：中文

【背景信息】
你是Agent-{self.worker_id}，这是你的原创作品。题材是自主选择的。

现在请开始创作故事："""

        try:
            message = self.client.messages.create(
                model="claude-3-5-sonnet-20241022",
                max_tokens=4096,
                messages=[{"role": "user", "content": prompt}]
            )

            story = message.content[0].text
            return story

        except Exception as e:
            error_msg = f"Error calling Claude API: {str(e)}"
            print(f"❌ {self.name} API error: {error_msg}")
            return error_msg

    def complete_task(self, task_id, output):
        """Mark task as completed"""
        response = requests.post(
            f"{self.master_url}/api/tasks/complete",
            json={
                "task_id": task_id,
                "worker_id": self.worker_id,
                "output": output,
                "status": "completed"
            },
            timeout=5
        )

        if response.status_code == 200:
            print(f"✅ {self.name} completed task {task_id}")
            return True
        else:
            print(f"❌ Failed to complete task: {response.text}")
            return False

    def run(self):
        """Run the autonomous worker workflow"""
        # Step 1: Register
        if not self.register():
            return False

        # Step 2: Claim task
        task = self.claim_task()
        if not task:
            print(f"⚠️  {self.name} found no tasks to work on")
            return False

        # Step 3: Select topic
        if not self.select_topic():
            return False

        # Step 4: Execute with Claude
        output = self.execute_task_with_claude(task)

        # Step 5: Complete task
        if not self.complete_task(task["task_id"], output):
            return False

        return True


def main():
    """Main test function"""
    if not API_KEY:
        print("❌ Error: ANTHROPIC_API_KEY environment variable not set")
        print("   Please run: export ANTHROPIC_API_KEY='your-key-here'")
        sys.exit(1)

    print("=" * 60)
    print("🚀 Starting CPDS Autonomous Worker Test")
    print("=" * 60)
    print(f"📊 Workers: {NUM_WORKERS}")
    print(f"🌐 Master: {MASTER_URL}")
    print("=" * 60)

    # Create workers
    workers = []
    for i in range(1, NUM_WORKERS + 1):
        worker_id = str(uuid.uuid4())
        worker = AutonomousWorker(worker_id, MASTER_URL, API_KEY)
        workers.append(worker)

    # Start workers with slight delays to avoid race conditions
    print("\n🎬 Starting workers...\n")
    for i, worker in enumerate(workers):
        worker.run()
        if i < len(workers) - 1:
            time.sleep(0.5)  # Small delay between starts

    # Wait for all tasks to complete
    print("\n⏳ Waiting for tasks to complete...")
    time.sleep(5)

    # Display results
    print("\n" + "=" * 60)
    print("📊 TEST RESULTS")
    print("=" * 60)

    # Check selected topics
    response = requests.get(f"{MASTER_URL}/api/topics", timeout=5)
    if response.status_code == 200:
        data = response.json()
        topics = data.get("topics", [])
        print(f"\n🎯 Selected Topics ({len(topics)}):")
        for i, topic in enumerate(topics, 1):
            print(f"  {i}. {topic['topic']} (by {topic.get('worker_id', 'unknown')[:8]}...)")

    # Check completed tasks
    response = requests.get(f"{MASTER_URL}/api/tasks/completed", timeout=5)
    if response.status_code == 200:
        data = response.json()
        tasks = data.get("data", {}).get("tasks", [])
        print(f"\n✅ Completed Tasks ({len(tasks)}):")
        for task in tasks:
            output_preview = task.get('output', '')[:100]
            print(f"  - {task['task_id']}: {output_preview}...")

    print("\n" + "=" * 60)
    print("✨ Test Complete!")
    print("=" * 60)


if __name__ == "__main__":
    main()
