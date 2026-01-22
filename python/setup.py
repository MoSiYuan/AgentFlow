"""
Setup script for AgentFlow Python
"""

from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="agentflow",
    version="1.0.0",
    author="AgentFlow Team",
    description="AI Agent Task Collaboration System - Python Edition",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/MoSiYuan/AgentFlow",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "Topic :: Software Development :: Build Tools",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
    ],
    python_requires=">=3.8",
    install_requires=[
        "Flask>=3.0.0",
        "requests>=2.31.0",
    ],
    entry_points={
        "console_scripts": [
            "agentflow=agentflow.cli:main",
            "agentflow-master=agentflow.cli:main",
            "agentflow-worker=agentflow.cli:main",
        ],
    },
)
