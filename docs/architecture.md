# Architecture

## Overview

a365 is a thin MCP (Model Context Protocol) client that talks JSON-RPC over HTTP+SSE to Microsoft's agent365 gateway. It doesn't use the Graph REST API, Graph SDK, or any MCP SDK — the entire MCP transport layer is hand-written Go (~300 LOC).

```
User
  |
  a365 CLI (kong)
  |
  ├── commands/         Parse args, apply safety guards (dry-run, confirm)
  |
  ├── output/           Extract MCP response → table / JSON / TSV
  |
  ├── mcp/client        JSON-RPC over HTTP+SSE, retry, session cache
  |
  └── auth/             InteractiveBrowserCredential + PKCE + OS keychain
       |
       agent365.svc.cloud.microsoft
       /agents/servers/{mcp_server}/
```

## Request Lifecycle

Every command follows the same path:

1. **Parse** — kong parses CLI args into a typed Go struct
2. **Safety** — `--dry-run` connects to the server, fetches tool schemas via `ListToolsCached()`, validates args against the JSON Schema, and prints the result without executing the tool; destructive ops prompt via `ctx.Confirm()`
3. **Auth** — `EnsureAuth()` loads cached credentials or triggers browser login
4. **Session** — `Initialize()` checks `~/.a365/sessions.json` for a cached session; if valid, skips the MCP handshake
5. **Request** — `CallTool()` sends a JSON-RPC `tools/call` POST with `Authorization: Bearer` and `Mcp-Session-Id` headers
6. **Retry** — on 502/503/429/504, retries up to 2x with exponential backoff (1s, 2s); respects `Retry-After` header
7. **Response** — parses SSE stream (`data:` lines) or plain JSON; extracts the first JSON-RPC message
8. **Extract** — `ExtractContent()` unwraps 3 response patterns (clean JSON, embedded JSON after status text, rawResponse-wrapped)
9. **Render** — `PrintList`/`PrintItem`/`PrintMutation` dispatches to table (tabwriter), JSON, or TSV based on `--output`

## MCP Protocol

### Transport

- **HTTP POST** to `https://agent365.svc.cloud.microsoft/agents/servers/{server}/`
- **Content-Type**: `application/json`
- **Accept**: `application/json, text/event-stream`
- **Response**: either `application/json` (plain) or `text/event-stream` (SSE)

### SSE Format

```
event: message
data: {"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"..."}]}}

```

The parser handles both `data: ` (with space) and `data:` (without space) per the SSE spec.

### Session Management

Each service gets its own MCP session:

1. First request sends `initialize` → server returns `Mcp-Session-Id` header
2. Subsequent requests include the session ID
3. Sessions cached in `~/.a365/sessions.json` with 30-minute TTL (includes tool schemas for `--dry-run` validation)
4. On session errors (401/403, "invalid session"), cache is cleared and session re-established automatically

### Response Patterns

The agent365 servers return data in 3 formats:

| Pattern | Example | How it looks |
|---------|---------|-------------|
| Clean JSON | Teams | `Content[0].Text = {"teams":[...]}` |
| Embedded | Calendar | `Content[0].Text = "Success.\n{\"value\":[...]}"` |
| Wrapped | Mail | `Content[0].Text = {"rawResponse":"{\"value\":[...]}","message":"..."}` |

`ExtractContent()` handles all three transparently.

## Authentication

```
Browser ──PKCE──► Entra ID (login.microsoftonline.com)
                       |
                  Access Token (JWT)
                       |
                  ├── Scope: ea9ffc3e-.../.default (all agent365 scopes)
                  ├── Cached in OS keychain via azidentity/cache
                  └── Auth record in ~/.a365/auth-record.json for silent refresh
```

- **Flow**: Interactive browser + PKCE (passes org Conditional Access on managed devices)
- **Token cache**: OS keychain (macOS Keychain, Windows Credential Manager) via `azidentity/cache`
- **Auth record**: `~/.a365/auth-record.json` enables silent token refresh across CLI invocations
- **Scope**: `ea9ffc3e-8a23-4a7d-836d-234d7c7565c1/.default` — requests all granted agent365 scopes at once

## Output Pipeline

```
MCP JSONRPCResponse
    → ExtractContent()     (extract.go)    → map[string]any
    → ToRows()             (extract.go)    → []map[string]any
    → PrintList()          (formatter.go)  → format dispatch
        ├── FormatHuman → RenderTable()    (render.go)  → text/tabwriter
        ├── FormatJSON  → writeJSON()      (formatter.go) → json.Encoder
        └── FormatPlain → RenderTSV()      (render.go)  → raw tabs
```

Each entity type has column definitions in `columns.go` that extract and format fields:

- **Width** — max chars for table display (0 = unlimited)
- **Extract** — function that pulls a display value from `map[string]any`
- **HTML stripping** — Teams messages have HTML content; `stripHTML()` handles emoji, attachment, codeblock, img tags

## Server Discovery

The server map in `config.go` provides the default mapping of friendly names to MCP server names. The `api discover` command can also query the live catalog:

```
GET https://agent365.svc.cloud.microsoft/agents/discoverToolServers
```

This returns all available servers with their URLs, scopes, and audiences — useful for finding new servers Microsoft has added.

## Key Design Decisions

| Decision | Why |
|----------|-----|
| **Kong over Cobra** | Struct-tag CLI definition, less boilerplate; 173 commands as struct fields |
| **Hand-written MCP client** | ~300 LOC; the protocol is simple enough that an SDK adds complexity without value |
| **`map[string]any` over typed structs** | MCP responses vary across 24 servers; untyped maps are forward-compatible |
| **`text/tabwriter` for tables** | Standard library, zero dependencies |
| **Retry at HTTP layer** | Catches transient 502/503/429 from the gateway without command-level changes |
| **Session cache as JSON file** | Simple, debuggable, no external dependencies |
| **OS keychain for tokens** | Secure, survives reboots, handled by Azure SDK |

## File Layout

| Directory | Purpose |
|-----------|---------|
| `internal/mcp/` | MCP JSON-RPC client, SSE parser, session cache, types, schema validation |
| `internal/auth/` | Entra ID credential, token cache, auth record persistence |
| `internal/config/` | Server endpoint map, constants, `~/.a365/config.json` support |
| `internal/output/` | 3-mode formatter, per-entity columns, table/TSV/JSON renderers, HTML stripping, MCP response extraction |
| `internal/commands/` | Shared context, auth commands, completion |
| `internal/commands/<service>/` | One package per M365 service (18 services + api + config) |
| `internal/version/` | Version/commit vars injected via ldflags at build time |
