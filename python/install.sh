#!/bin/bash
# AgentFlow Python - Quick Install Script

set -e

echo "🐍 AgentFlow Python - Installation"
echo ""

# Check Python version
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is not installed"
    echo "   Please install Python 3.8+ first"
    exit 1
fi

PYTHON_VERSION=$(python3 --version | awk '{print $2}')
echo "✓ Python version: $PYTHON_VERSION"

# Create virtual environment (optional but recommended)
if [ "$1" == "--venv" ]; then
    echo ""
    echo "📦 Creating virtual environment..."
    python3 -m venv venv
    source venv/bin/activate
    echo "✓ Virtual environment activated"
fi

# Install dependencies
echo ""
echo "📦 Installing dependencies..."
pip install -r requirements.txt

# Install in editable mode
echo ""
echo "📦 Installing AgentFlow..."
pip install -e .

# Create data directory
echo ""
echo "📁 Creating data directory..."
mkdir -p .claude/cpds-manager

# Verify installation
echo ""
echo "🧪 Verifying installation..."
python3 -c "from agentflow import Master, Worker, Database; print('✓ All modules imported successfully')"

echo ""
echo "✅ Installation complete!"
echo ""
echo "🚀 Quick start:"
echo "   # Terminal 1: Start Master"
echo "   python -m agentflow.cli master --port 8848"
echo ""
echo "   # Terminal 2: Start Worker"
echo "   python -m agentflow.cli worker --auto"
echo ""
echo "   # Terminal 3: Create task"
echo "   curl -X POST http://127.0.0.1:8848/api/tasks/create \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"task_id\": \"T1\", \"title\": \"Test\", \"description\": \"shell:echo Hello\", \"priority\": \"high\"}'"
echo ""
echo "📖 Documentation: python/README.md"
