# WebSearch

Search the web from the command line.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `websearch search` | Search the web | `<query>` `--urls` |

## Examples

```bash
# Search the web for a topic
a365 websearch search "Microsoft 365 Copilot release notes"

# Search with specific URLs to focus on
a365 websearch search "deployment guide" --urls https://learn.microsoft.com --urls https://docs.microsoft.com
```
