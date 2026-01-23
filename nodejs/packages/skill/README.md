# @agentflow/skill

AgentFlow Task Orchestration Skill for Claude CLI.

## Install

```bash
npm install -g @agentflow/skill
# or
npm link @agentflow/skill
```

## Commands

```bash
agentflow create <title>           # Create task
agentflow list [-s <status>]       # List tasks
agentflow status <taskId>          # Task status
agentflow exec <command>           # Execute shell
agentflow health                   # Check Master
```

## Example

```bash
agentflow create "Deploy" -d "Build and deploy"
agentflow list --status pending
```

## API

```typescript
import { AgentFlowSkill } from '@agentflow/skill';

const skill = new AgentFlowSkill();
await skill.createTask({ title: 'Test', description: 'npm test' });
```

## Docs

[Full Documentation](../../docs/SKILL.md)

## License

MIT
