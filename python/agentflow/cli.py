"""
Command-line interface for AgentFlow (Python version)
"""

import argparse
import sys
from .master import run_master
from .worker import run_worker


def main():
    """Main CLI entry point"""
    parser = argparse.ArgumentParser(
        description="AgentFlow - AI Agent Task Collaboration System (Python)"
    )

    subparsers = parser.add_subparsers(dest="command", help="Command to run")

    # Master command
    master_parser = subparsers.add_parser("master", help="Run master server")
    master_parser.add_argument("--host", default="0.0.0.0", help="Host to bind to")
    master_parser.add_argument("--port", type=int, default=8848, help="Port to listen on")
    master_parser.add_argument("--db", default=".claude/cpds-manager/agentflow-python.db",
                              help="Database path")
    master_parser.add_argument("--auto-shutdown", action="store_true",
                              help="Shutdown when all tasks complete (standalone mode)")
    master_parser.add_argument("--debug", action="store_true", help="Enable debug mode")

    # Worker command
    worker_parser = subparsers.add_parser("worker", help="Run worker")
    worker_parser.add_argument("-m", "--master", default="http://127.0.0.1:8848",
                              help="Master URL")
    worker_parser.add_argument("-n", "--name", help="Worker name")
    worker_parser.add_argument("-a", "--auto", action="store_true", default=True,
                              help="Auto mode (default)")
    worker_parser.add_argument("--oneshot", action="store_true",
                              help="Execute one task and exit")
    worker_parser.add_argument("--manual", action="store_true",
                              help="Manual mode (wait for commands)")

    args = parser.parse_args()

    if args.command == "master":
        run_master(
            host=args.host,
            port=args.port,
            db_path=args.db,
            auto_shutdown=args.auto_shutdown
        )
    elif args.command == "worker":
        auto_mode = args.auto and not args.manual
        run_worker(
            master_url=args.master,
            name=args.name,
            auto=auto_mode,
            oneshot=args.oneshot
        )
    else:
        parser.print_help()
        sys.exit(1)


if __name__ == "__main__":
    main()
