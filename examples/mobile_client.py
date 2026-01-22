#!/usr/bin/env python3
"""
AgentFlow 移动端客户端示例

用于从手机或其他移动设备控制 AgentFlow 系统
支持跨平台编译、GUI 操作等任务

用法:
    python mobile_client.py compile --platform windows
    python mobile_client.py compile --all
    python mobile_client.py status --task-id TASK-001
"""

import argparse
import json
import sys
import time
import requests
from typing import Optional, Dict, List


class AgentFlowClient:
    """AgentFlow API 客户端"""

    def __init__(self, master_url: str, api_key: Optional[str] = None, timeout: int = 30):
        """
        初始化客户端

        Args:
            master_url: Master 服务器 URL (如: https://your-server:8848)
            api_key: API 密钥（可选）
            timeout: 请求超时时间（秒）
        """
        self.master_url = master_url.rstrip('/')
        self.api_key = api_key
        self.timeout = timeout

        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json'
        })

        if api_key:
            self.session.headers.update({
                'Authorization': f'Bearer {api_key}'
            })

    def create_task(self, title: str, description: str, group_name: str) -> Dict:
        """
        创建任务

        Args:
            title: 任务标题
            description: 任务描述
            group_name: Worker 组名（windows, macos, linux, cloud）

        Returns:
            任务信息字典
        """
        url = f"{self.master_url}/api/v1/tasks"
        payload = {
            'title': title,
            'description': description,
            'group_name': group_name
        }

        try:
            response = self.session.post(url, json=payload, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            print(f"❌ 创建任务失败: {e}")
            raise

    def get_task(self, task_id: str) -> Dict:
        """
        获取任务详情

        Args:
            task_id: 任务 ID

        Returns:
            任务详情字典
        """
        url = f"{self.master_url}/api/v1/tasks/{task_id}"

        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            print(f"❌ 获取任务失败: {e}")
            raise

    def list_tasks(self, status: Optional[str] = None, group: Optional[str] = None) -> List[Dict]:
        """
        列出任务

        Args:
            status: 任务状态过滤（pending, running, completed, failed）
            group: 组名过滤

        Returns:
            任务列表
        """
        url = f"{self.master_url}/api/v1/tasks"
        params = {}
        if status:
            params['status'] = status
        if group:
            params['group'] = group

        try:
            response = self.session.get(url, params=params, timeout=self.timeout)
            response.raise_for_status()
            data = response.json()
            return data.get('tasks', [])
        except requests.RequestException as e:
            print(f"❌ 列出任务失败: {e}")
            raise

    def list_workers(self, group: Optional[str] = None) -> List[Dict]:
        """
        列出 Workers

        Args:
            group: 组名过滤

        Returns:
            Worker 列表
        """
        url = f"{self.master_url}/api/v1/workers"
        params = {}
        if group:
            params['group'] = group

        try:
            response = self.session.get(url, params=params, timeout=self.timeout)
            response.raise_for_status()
            data = response.json()
            return data.get('workers', [])
        except requests.RequestException as e:
            print(f"❌ 列出 Workers 失败: {e}")
            raise

    def get_stats(self) -> Dict:
        """获取统计信息"""
        url = f"{self.master_url}/api/v1/stats"

        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            print(f"❌ 获取统计信息失败: {e}")
            raise

    def wait_for_task(self, task_id: str, check_interval: int = 5, timeout: int = 600) -> Dict:
        """
        等待任务完成

        Args:
            task_id: 任务 ID
            check_interval: 检查间隔（秒）
            timeout: 超时时间（秒）

        Returns:
            最终任务状态
        """
        start_time = time.time()

        while True:
            if time.time() - start_time > timeout:
                raise TimeoutError(f"等待任务超时: {task_id}")

            task = self.get_task(task_id)
            status = task.get('status')

            if status == 'completed':
                print(f"✅ 任务完成: {task_id}")
                return task
            elif status == 'failed':
                print(f"❌ 任务失败: {task_id}")
                print(f"错误: {task.get('error', 'Unknown error')}")
                return task
            else:
                print(f"⏳ 任务状态: {status}...")
                time.sleep(check_interval)

    def compile_project(self, project_path: str, platform: str, binary_name: str = "app") -> Dict:
        """
        编译项目

        Args:
            project_path: 项目路径
            platform: 目标平台（windows, macos, linux）
            binary_name: 二进制文件名

        Returns:
            创建的任务信息
        """
        platform_map = {
            'windows': ('windows', 'exe'),
            'macos': ('macos', ''),
            'linux': ('linux', ''),
        }

        if platform not in platform_map:
            raise ValueError(f"不支持的平台: {platform}")

        group_name, ext = platform_map[platform]
        binary = f"{binary_name}.{ext}" if ext else binary_name

        description = f"""
编译 {platform.upper()} 版本

项目路径: {project_path}
输出文件: {binary}

步骤:
1. 进入项目目录
2. 设置环境变量
3. 执行编译命令
4. 验证输出文件

编译命令:
cd {project_path} && go build -o bin/{binary}

验证:
ls -lh bin/{binary}
        """.strip()

        return self.create_task(
            title=f"编译 {platform.upper()} 版本: {binary}",
            description=description,
            group_name=group_name
        )

    def cross_compile(self, project_path: str, binary_name: str = "app") -> List[Dict]:
        """
        交叉编译所有平台

        Args:
            project_path: 项目路径
            binary_name: 二进制文件名

        Returns:
            创建的所有任务列表
        """
        platforms = ['windows', 'macos', 'linux']
        tasks = []

        print(f"🚀 开始交叉编译: {len(platforms)} 个平台")

        for platform in platforms:
            try:
                task = self.compile_project(project_path, platform, binary_name)
                tasks.append(task)
                print(f"✅ {platform.upper()} 任务已创建: {task['task_id']}")
            except Exception as e:
                print(f"❌ {platform.upper()} 任务创建失败: {e}")

        return tasks

    def gui_operation(self, operation: str, group: str = 'macos') -> Dict:
        """
        GUI 操作任务

        Args:
            operation: 操作描述
            group: 执行组（macos 或 windows）

        Returns:
            创建的任务信息
        """
        description = f"""
GUI 自动化操作

操作描述: {operation}

使用 VSCode 或其他 GUI 工具执行操作
        """.strip()

        return self.create_task(
            title=f"GUI 操作: {operation[:50]}",
            description=description,
            group_name=group
        )


def main():
    """命令行入口"""
    parser = argparse.ArgumentParser(
        description='AgentFlow 移动端客户端',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  # 编译 Windows 版本
  python mobile_client.py compile --platform windows --path /path/to/project

  # 交叉编译所有平台
  python mobile_client.py compile --all --path /path/to/project

  # 查看任务状态
  python mobile_client.py status --task-id TASK-001

  # 列出所有 Workers
  python mobile_client.py workers

  # 查看统计信息
  python mobile_client.py stats
        """
    )

    parser.add_argument('--master', default='http://localhost:8848',
                       help='Master 服务器 URL (默认: http://localhost:8848)')
    parser.add_argument('--api-key', help='API 密钥（可选）')

    subparsers = parser.add_subparsers(dest='command', help='子命令')

    # compile 命令
    compile_parser = subparsers.add_parser('compile', help='编译项目')
    compile_parser.add_argument('--platform', choices=['windows', 'macos', 'linux'],
                               help='目标平台')
    compile_parser.add_argument('--all', action='store_true', help='交叉编译所有平台')
    compile_parser.add_argument('--path', required=True, help='项目路径')
    compile_parser.add_argument('--output', default='app', help='输出文件名')
    compile_parser.add_argument('--wait', action='store_true', help='等待编译完成')

    # status 命令
    status_parser = subparsers.add_parser('status', help='查看任务状态')
    status_parser.add_argument('--task-id', required=True, help='任务 ID')

    # list 命令
    list_parser = subparsers.add_parser('list', help='列出任务')
    list_parser.add_argument('--status', choices=['pending', 'running', 'completed', 'failed'],
                            help='按状态过滤')
    list_parser.add_argument('--group', help='按组过滤')

    # workers 命令
    workers_parser = subparsers.add_parser('workers', help='列出 Workers')
    workers_parser.add_argument('--group', help='按组过滤')

    # stats 命令
    subparsers.add_parser('stats', help='查看统计信息')

    # gui 命令
    gui_parser = subparsers.add_parser('gui', help='GUI 操作')
    gui_parser.add_argument('--operation', required=True, help='操作描述')
    gui_parser.add_argument('--group', default='macos', help='执行组')

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    # 创建客户端
    client = AgentFlowClient(args.master, args.api_key)

    try:
        if args.command == 'compile':
            if args.all:
                # 交叉编译所有平台
                tasks = client.cross_compile(args.path, args.output)
                print(f"\n✅ 已创建 {len(tasks)} 个编译任务")

                if args.wait:
                    print("\n⏳ 等待所有任务完成...")
                    for task in tasks:
                        client.wait_for_task(task['task_id'])
                    print("\n✅ 所有编译任务完成！")
            else:
                # 单平台编译
                if not args.platform:
                    print("❌ 请指定 --platform 或使用 --all")
                    sys.exit(1)

                task = client.compile_project(args.path, args.platform, args.output)
                task_id = task['task_id']
                print(f"✅ 任务已创建: {task_id}")

                if args.wait:
                    print("\n⏳ 等待编译完成...")
                    result = client.wait_for_task(task_id)
                    print(f"\n📦 编译结果:\n{result.get('result', 'No output')}")

        elif args.command == 'status':
            task = client.get_task(args.task_id)
            print(f"\n任务详情:")
            print(f"  ID: {task.get('id')}")
            print(f"  标题: {task.get('title')}")
            print(f"  状态: {task.get('status')}")
            print(f"  组: {task.get('group_name')}")
            if task.get('result'):
                print(f"  结果: {task['result'][:200]}...")
            if task.get('error'):
                print(f"  错误: {task['error']}")

        elif args.command == 'list':
            tasks = client.list_tasks(args.status, args.group)
            print(f"\n找到 {len(tasks)} 个任务:\n")
            for task in tasks:
                print(f"  {task['id']} | {task['status']} | {task['title']}")

        elif args.command == 'workers':
            workers = client.list_workers(args.group)
            print(f"\n找到 {len(workers)} 个 Workers:\n")
            for worker in workers:
                print(f"  {worker['id']} | {worker['group_name']} | {worker.get('status', 'unknown')}")

        elif args.command == 'stats':
            stats = client.get_stats()
            print(f"\n统计信息:")
            print(f"  总任务数: {stats.get('stats', {}).get('total_tasks', 0)}")
            print(f"  待处理: {stats.get('stats', {}).get('pending_tasks', 0)}")
            print(f"  运行中: {stats.get('stats', {}).get('running_tasks', 0)}")
            print(f"  已完成: {stats.get('stats', {}).get('completed_tasks', 0)}")
            print(f"  失败: {stats.get('stats', {}).get('failed_tasks', 0)}")

        elif args.command == 'gui':
            task = client.gui_operation(args.operation, args.group)
            print(f"✅ GUI 操作任务已创建: {task['task_id']}")

    except Exception as e:
        print(f"\n❌ 错误: {e}")
        sys.exit(1)


if __name__ == '__main__':
    main()
