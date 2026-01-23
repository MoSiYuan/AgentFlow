# AgentFlow Configuration Directory

This directory contains AgentFlow's core configuration, templates, and examples.

## ⚠️ Port Configuration

Default Master server port: **6767** (changed from 8848 in v2.0.0)

Update your `config.example.json`:
```json
{
  "master": {
    "url": "http://localhost:6767"
  }
}
```

See [Port Migration Guide](../PORT_MIGRATION_GUIDE.md) for details.

## Structure

```
.agentflow/
├── agents/          # Agent templates and configurations
├── skills/          # Skill definitions and examples
├── workflows/       # Workflow templates
├── examples/        # Usage examples
├── rules/           # Agent behavior rules
└── README.md        # This file
```

## Initialization

Run `agentflow init` to initialize this directory with default configurations.

## Usage

- **Agents**: Define specialized agent personas (developer, tester, reviewer, etc.)
- **Skills**: Reusable task patterns and knowledge modules
- **Workflows**: Predefined multi-agent collaboration patterns
- **Examples**: Real-world usage scenarios

## Documentation

See [AI Integration Guide](../docs/AI_INTEGRATION.md) for details.
