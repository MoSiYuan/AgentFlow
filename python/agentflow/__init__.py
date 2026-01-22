"""
AgentFlow - AI Agent Task Collaboration System (Python Version)
A simplified version for local deployment with full feature parity.
"""

__version__ = "1.0.0"
__author__ = "AgentFlow Team"

# Lazy imports to avoid ImportError when dependencies not installed
def Master(*args, **kwargs):
    """Master server factory function"""
    from .master import Master as _Master
    return _Master(*args, **kwargs)

def Worker(*args, **kwargs):
    """Worker factory function"""
    from .worker import Worker as _Worker
    return _Worker(*args, **kwargs)

def Database(*args, **kwargs):
    """Database factory function"""
    from .database import Database as _Database
    return _Database(*args, **kwargs)

# Direct import for database (no external dependencies)
from .database import Database as _DatabaseClass
Database = _DatabaseClass

__all__ = ["Master", "Worker", "Database"]
