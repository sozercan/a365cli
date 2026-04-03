# NLWeb

Search your organization's NLWeb-enabled sites using natural language. Ask questions, find people, and discover available NLWeb sites.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `nlweb ask` | Ask a natural language question | `<query>` |
| `nlweb who` | Find people related to a query | `<query>` |
| `nlweb sites` | List available NLWeb sites | *(none)* |

## Arguments

- **`<query>`** -- Natural language question or search term.

## Examples

```sh
# Ask a question across NLWeb sites
a365 nlweb ask "What is the return policy for enterprise customers?"

# Search for people
a365 nlweb who "engineers working on the authentication service"

# Find people by expertise
a365 nlweb who "who knows about Kubernetes deployments"

# List all available NLWeb sites
a365 nlweb sites

# Output as JSON
a365 nlweb ask "latest product announcements" --output json
```
