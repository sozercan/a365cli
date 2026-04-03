# API Explorer

The API explorer is a hidden dev/debug tool for discovering MCP servers, inspecting their tools, and making raw MCP calls. It's not shown in `a365 --help` but is always available.

## Commands

| Command | Description |
|---------|-------------|
| `a365 api servers` | List all configured MCP servers |
| `a365 api servers --probe` | Connect to each server and show tool counts |
| `a365 api discover` | Query the live server catalog from the agent365 gateway |
| `a365 api tools <service>` | List tools with required params and descriptions |
| `a365 api call <service> <tool> '<json>'` | Call any MCP tool with raw JSON arguments |

## Usage

### Discover servers

```bash
# List configured servers (offline, fast)
a365 api servers

# Probe all servers live (connects to each one)
a365 api servers --probe

# Query the gateway for the full catalog (may include servers not yet in config)
a365 api discover
```

### Inspect tools

```bash
# See all tools a service exposes, with required params
a365 api tools teams
a365 api tools mail
a365 api tools calendar

# JSON output for scripting
a365 api tools me -o json
```

### Raw tool calls

```bash
# Call any tool with raw JSON arguments
a365 api call me GetMyDetails '{}'
a365 api call teams ListTeams '{"userId":"alice@contoso.com"}'
a365 api call mail SearchMessagesQueryParameters '{"queryParameters":"?$top=5"}'

# JSON output
a365 api call me GetMyDetails '{}' -o json
```

## When to use it

- **Before implementing a new command** — check the tool name, required params, and response shape
- **Debugging a 403 or error** — call the tool directly with `--verbose` to see the raw request/response
- **Finding new servers** — `api discover` queries the gateway and may return servers not yet wired into the CLI
- **Verifying coverage** — compare `api servers --probe` tool counts against wired commands

## Adding a newly discovered server

If `api discover` shows a server not in the CLI:

1. Add it to `config.go`: `"myservice": "mcp_MyServer"`
2. Probe it: `a365 api tools myservice`
3. If it has tools, create a command package following [CONTRIBUTING.md](../CONTRIBUTING.md)
