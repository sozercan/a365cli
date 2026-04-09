# Contributing to a365

## Prerequisites

- Go 1.26+
- An Entra ID app registration with agent365 scopes (or access to one)
- macOS, Linux, or Windows

## Project Structure

See [docs/architecture.md](docs/architecture.md) for the full file layout and component descriptions.

## Building

```bash
make build      # Build
make build-static # Build without CGO / OS-backed token cache
make test       # Run all tests with verbose output
make test-short # Run tests without verbose output
make lint       # Format + vet
make clean      # Remove build artifacts
```

The default native build uses CGO on macOS and Linux so the CLI can persist
refresh tokens in the OS credential store. Use `make build-static` if you need
a pure-Go build; auth still works, but token persistence falls back to the
non-CGO path.

## Adding a New Service

Adding a new M365 service is straightforward. Here's the pattern:

### 1. Add the server mapping to `internal/config/config.go`

```go
var Servers = map[string]string{
    // ...existing...
    "myservice": "mcp_MyServiceServer",
}
```

### 2. Create the command file

```bash
mkdir -p internal/commands/myservice
```

Create `internal/commands/myservice/myservice.go`:

```go
package myservice

import (
    "fmt"
    "github.com/sozercan/a365cli/internal/commands"
    "github.com/sozercan/a365cli/internal/config"
    "github.com/sozercan/a365cli/internal/output"
)

type MyServiceCmd struct {
    List MyServiceListCmd `cmd:"" help:"List items"`
    Get  MyServiceGetCmd  `cmd:"" help:"Get an item"`
}

func endpoint() string {
    return config.Endpoint("myservice")
}

type MyServiceListCmd struct {
    Max int `help:"Maximum results" default:"50"`
}

func (c *MyServiceListCmd) Run(ctx *commands.Context) error {
    client := ctx.NewMCPClient(endpoint())
    if err := client.Initialize(ctx.Ctx); err != nil {
        return fmt.Errorf("initialize: %w", err)
    }

    resp, err := client.CallTool(ctx.Ctx, "ListItems", map[string]any{})
    if err != nil {
        return fmt.Errorf("list items: %w", err)
    }

    data, err := output.ExtractContent(resp)
    if err != nil {
        return err
    }
    rows := output.ToRows(data, "items")
    if rows == nil {
        rows = output.ToRows(data, "value")
    }
    if rows == nil {
        return ctx.Output.PrintItem(data)
    }
    return ctx.Output.PrintList("items", output.MyColumns, rows)
}
```

### 3. Register in `main.go`

```go
import "github.com/sozercan/a365cli/internal/commands/myservice"

type CLI struct {
    // ...existing...
    MyService myservice.MyServiceCmd `cmd:"" help:"My Service"`
}
```

### 4. Add column definitions (optional)

If the service returns list data, add columns to `internal/output/columns.go`:

```go
var MyColumns = []Column{
    {Header: "NAME", Width: 30, Extract: func(row map[string]any) string {
        return getString(row, "name")
    }},
    {Header: "ID", Width: 36, Extract: func(row map[string]any) string {
        return getString(row, "id")
    }},
}
```

### 5. Add documentation

Create `docs/myservice.md` with commands table and usage examples.

That's it — no framework changes, no new dependencies.

## Testing

```bash
make test       # Run all tests
go test ./internal/mcp/... -v    # MCP client tests only
go test ./internal/output/... -v # Output formatting tests only
```

The test suite uses `httptest.NewServer` for MCP client tests and `bytes.Buffer` injection for output formatter tests. Use `testutil.SetupTestServerWithSchemas` for dry-run tests that verify schema validation. No real network calls in tests.

## Discovering MCP Tools

See [docs/api-explorer.md](docs/api-explorer.md) for the full guide. Quick reference:

```bash
a365 api servers --probe    # List servers with tool counts
a365 api tools teams        # List tools + required params
a365 api call me GetMyDetails '{}'  # Raw call
```

## Architecture Decisions

See [docs/architecture.md](docs/architecture.md) for the full architecture guide including request lifecycle, MCP protocol, auth flow, output pipeline, and design rationale.
