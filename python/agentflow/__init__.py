"""
AgentFlow - AI Agent Task Collaboration System (Python Version)
A simplified version for local deployment with full feature parity.
"""

__version__ = "1.0.0"
__author__ = "AgentFlow Team"

from .master import Master
from .worker import Worker
from .database import Database

__all__ = ["Master", "Worker", "Database"]
