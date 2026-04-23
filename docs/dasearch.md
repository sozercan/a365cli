# DASearch

Declarative Agent search — low-level discovery of available M365 Copilot agents.

For the normalized, chat-ready selector view used by `a365 copilot chat --agent`, prefer `a365 copilot agents`. `dasearch agents` remains useful when you want the raw DASearch payload.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `dasearch agents` | List available M365 Copilot agents from the raw DASearch response | _(none)_ |

## Examples

```bash
# List the raw DASearch agent payload
a365 dasearch agents

# Prefer this when you want selectors for copilot chat --agent
a365 copilot agents
```
