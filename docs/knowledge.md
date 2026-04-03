# Knowledge

Manage federated knowledge sources for Microsoft 365. Query knowledge, list and create configurations, trigger ingestion, and delete configurations. Destructive operations prompt for confirmation before executing.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `knowledge query` | Search federated knowledge | `<consumer-id> <query>` |
| `knowledge list` | List knowledge configurations | `<consumer-id>` |
| `knowledge configure` | Create a knowledge source | `<consumer-id> <source-type> <display-name> <description>` |
| `knowledge ingest` | Trigger ingestion for a source | `<consumer-id> <config-id>` |
| `knowledge delete` | Delete a knowledge configuration | `<consumer-id> <config-id>` |

## Arguments

- **`<consumer-id>`** -- Consumer ID that owns the knowledge configurations.
- **`<query>`** -- Search query to run against federated knowledge.
- **`<source-type>`** -- Type of knowledge source to configure.
- **`<display-name>`** -- Human-readable name for the knowledge source.
- **`<description>`** -- Description of the knowledge source.
- **`<config-id>`** -- Search configuration ID for ingest/delete operations.
- **`--dry-run`** -- Preview write operations without executing them (supported by `configure`, `ingest`, and `delete`).

> **Note:** The `delete` command prompts for confirmation before executing. Use `--dry-run` to preview without being prompted.

## Examples

```sh
# Query federated knowledge
a365 knowledge query my-consumer-id "company vacation policy"

# List all knowledge configurations
a365 knowledge list my-consumer-id

# Configure a new knowledge source
a365 knowledge configure my-consumer-id sharepoint "HR Policies" \
  "Human resources policy documents from SharePoint"

# Preview a configuration without creating it
a365 knowledge configure my-consumer-id sharepoint "IT Runbooks" \
  "Internal IT runbooks" --dry-run

# Trigger ingestion for a knowledge source
a365 knowledge ingest my-consumer-id config-abc-123

# Delete a knowledge configuration (will prompt for confirmation)
a365 knowledge delete my-consumer-id config-abc-123

# Preview deletion without executing
a365 knowledge delete my-consumer-id config-abc-123 --dry-run

# Output as JSON
a365 knowledge list my-consumer-id --output json
```
