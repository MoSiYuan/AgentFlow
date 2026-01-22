#!/usr/bin/env python3
"""
CPDS AI SDK - Python SDK for AI Agents

让 Claude Code 和其他 AI Agent 可以轻松使用 CPDS
"""

import os
import subprocess
import time
import json
from typing import Optional, List, Dict
from dataclasses import dataclass
from enum import Enum


class TaskStatus(Enum):
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"


class TaskType(Enum):
    AI = "task"          # AI 智能任务
    SHELL = "shell"      # Shell 命令
    SCRIPT = "script"    # 脚本文件
    FILE = "file"        # 文件操作


@dataclass
class Task:
    id: str
    title: str
    description: str
    group: str
    status: str
    result: Optional[str] = None
    error: Optional[str] = None

    def __repr__(self):
        return f"Task(id={self.id}, title={self.title}, status={self.status})"


class CPDSClient:
    """CPDS 客户端 - 面向 AI Agent"""

    def __init__(
        self,
        db_path: str = "agentflow.db",
        master_url: str = None,
        worker_group: str = "local"
    ):
        self.db_path = db_path
        self.master_url = master_url or os.getenv("MASTER_URL", "http://localhost:8848")
        self.worker_group = worker_group or os.getenv("WORKER_GROUP", "local")

    def create_task(
        self,
        title: str,
        description: str,
        group: str = None,
        wait: bool = False,
        timeout: int = 300
    ) -> Task:
        """
        创建任务

        Args:
            title: 任务标题
            description: 任务描述（支持 task:, shell:, script:, file: 前缀）
            group: Worker 组
            wait: 是否等待完成
            timeout: 等待超时（秒）

        Returns:
            Task 对象
        """
        group = group or self.worker_group

        cmd = ["agentflow", "add", title, "--desc", description, "--group", group]

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            raise RuntimeError(f"创建任务失败: {result.stderr}")

        # 解析任务 ID
        task_id = self._parse_task_id(result.stdout)

        task = Task(
            id=task_id,
            title=title,
            description=description,
            group=group,
            status="pending"
        )

        if wait:
            task = self.wait_for_task(task_id, timeout)

        return task

    def wait_for_task(
        self,
        task_id: str,
        timeout: int = 300,
        check_interval: int = 2
    ) -> Task:
        """
        等待任务完成

        Args:
            task_id: 任务 ID
            timeout: 超时时间（秒）
            check_interval: 检查间隔（秒）

        Returns:
            完成的 Task 对象
        """
        start_time = time.time()

        while time.time() - start_time < timeout:
            task = self.get_task(task_id)

            if task.status in [TaskStatus.COMPLETED.value, TaskStatus.FAILED.value]:
                return task

            time.sleep(check_interval)

        raise TimeoutError(f"任务 {task_id} 在 {timeout} 秒内未完成")

    def get_task(self, task_id: str) -> Task:
        """
        获取任务详情

        Args:
            task_id: 任务 ID

        Returns:
            Task 对象
        """
        # 使用 list 命令获取任务
        result = subprocess.run(
            ["agentflow", "list"],
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            raise RuntimeError(f"获取任务失败: {result.stderr}")

        # 解析输出（简化版）
        lines = result.stdout.split('\n')
        for line in lines:
            if f"[{task_id}]" in line:
                return self._parse_task_line(line, task_id)

        raise ValueError(f"未找到任务 {task_id}")

    def list_tasks(
        self,
        status: str = None,
        group: str = None
    ) -> List[Task]:
        """
        列出任务

        Args:
            status: 过滤状态
            group: 过滤组

        Returns:
            Task 列表
        """
        cmd = ["agentflow", "list"]

        if status:
            cmd.extend(["--status", status])
        if group:
            cmd.extend(["--group", group])

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            raise RuntimeError(f"列出任务失败: {result.stderr}")

        return self._parse_task_list(result.stdout)

    def create_ai_task(
        self,
        feature: str,
        wait: bool = True,
        timeout: int = 600
    ) -> Task:
        """
        创建 AI 任务（自动分解）

        Args:
            feature: 功能描述
            wait: 是否等待完成
            timeout: 超时时间

        Returns:
            Task 对象
        """
        return self.create_task(
            title=f"实现{feature}",
            description=f"task:implement:{feature}",
            wait=wait,
            timeout=timeout
        )

    def create_shell_task(
        self,
        title: str,
        command: str,
        group: str = None,
        wait: bool = False
    ) -> Task:
        """
        创建 Shell 任务

        Args:
            title: 任务标题
            command: Shell 命令
            group: Worker 组
            wait: 是否等待完成

        Returns:
            Task 对象
        """
        return self.create_task(
            title=title,
            description=f"shell:{command}",
            group=group,
            wait=wait
        )

    def create_test_task(
        self,
        wait: bool = True
    ) -> Task:
        """
        创建测试任务

        Args:
            wait: 是否等待完成

        Returns:
            Task 对象
        """
        return self.create_task(
            title="运行测试",
            description="task:test",
            wait=wait
        )

    def create_build_task(
        self,
        wait: bool = True
    ) -> Task:
        """
        创建构建任务

        Args:
            wait: 是否等待完成

        Returns:
            Task 对象
        """
        return self.create_shell_task(
            title="构建项目",
            command="go build -v ./...",
            wait=wait
        )

    def _parse_task_id(self, output: str) -> str:
        """解析任务 ID"""
        for line in output.split('\n'):
            if "ID:" in line:
                return line.split("ID:")[1].strip()
        raise ValueError("无法解析任务 ID")

    def _parse_task_line(self, line: str, task_id: str) -> Task:
        """解析任务行"""
        # 简化版解析
        parts = line.split(']')
        if len(parts) < 2:
            raise ValueError(f"无法解析任务行: {line}")

        status = "pending"
        if "⏳" in line:
            status = "pending"
        elif "▶️" in line:
            status = "running"
        elif "✅" in line:
            status = "completed"
        elif "❌" in line:
            status = "failed"

        # 提取标题
        title = parts[2].strip() if len(parts) > 2 else "Unknown"

        return Task(
            id=task_id,
            title=title,
            description="",
            group=self.worker_group,
            status=status
        )

    def _parse_task_list(self, output: str) -> List[Task]:
        """解析任务列表"""
        tasks = []
        for line in output.split('\n'):
            if '[' in line and ']' in line:
                # 提取任务 ID
                start = line.find('[') + 1
                end = line.find(']')
                task_id = line[start:end]

                try:
                    task = self._parse_task_line(line, task_id)
                    tasks.append(task)
                except ValueError:
                    continue
        return tasks


# === 便捷函数 ===

def quick_task(title: str, description: str, wait: bool = True) -> Task:
    """
    快速创建并等待任务

    Args:
        title: 任务标题
        description: 任务描述
        wait: 是否等待完成

    Returns:
        Task 对象
    """
    client = CPDSClient()
    return client.create_task(title, description, wait=wait)


def develop(feature: str, wait: bool = True) -> Task:
    """
    开发功能（AI 自动分解）

    Args:
        feature: 功能描述
        wait: 是否等待完成

    Returns:
        Task 对象
    """
    client = CPDSClient()
    return client.create_ai_task(feature, wait=wait)


# === 使用示例 ===

if __name__ == "__main__":
    # 示例 1：创建简单任务
    print("示例 1：创建简单任务")
    task = quick_task(
        title="系统信息",
        description="shell:uname -a",
        wait=True
    )
    print(f"任务完成: {task}")

    # 示例 2：开发功能
    print("\n示例 2：开发功能")
    task = develop("用户登录功能", wait=True)
    print(f"开发完成: {task}")

    # 示例 3：批量创建任务
    print("\n示例 3：批量创建任务")
    client = CPDSClient()

    tasks = [
        client.create_shell_task("格式化代码", "gofmt -w .", wait=False),
        client.create_shell_task("运行测试", "go test ./...", wait=False),
        client.create_shell_task("构建应用", "go build -v ./...", wait=False),
    ]

    # 等待所有任务完成
    for task in tasks:
        client.wait_for_task(task.id)
        print(f"✓ {task.title} 完成")

    print("\n所有任务完成！")
