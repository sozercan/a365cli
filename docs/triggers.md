# Triggers

Event triggers and automation — create, manage, and evaluate trigger definitions.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `triggers events` | List supported event types | _(none)_ |
| `triggers schema` | Get schema for an event type | `<event-type>` |
| `triggers validate` | Validate a trigger request | `<user-request>` |
| `triggers create` | Create a trigger definition | `<validation-token>` `<name>` `<event-type>` `<logic>` `<conditions>` `<instructions>` |
| `triggers list` | List trigger definitions | _(none)_ |
| `triggers get` | Get a trigger definition | `<id>` |
| `triggers update` | Update a trigger definition | `<validation-token>` `<id>` |
| `triggers delete` | Delete a trigger definition | `<id>` |
| `triggers evaluate` | Evaluate event against triggers | `<event-type>` `<event-data-json>` |

## Examples

```bash
# List all supported event types
a365 triggers events

# Get the schema for a specific event type
a365 triggers schema "mailReceived"

# Validate a trigger before creating it
a365 triggers validate "Notify me when I get an email from my manager"

# Create a trigger definition (use the token from validate)
a365 triggers create TOKEN "Manager emails" "mailReceived" \
  "matchAll" '{"from":"manager@contoso.com"}' "Send a Teams notification"

# List all trigger definitions
a365 triggers list

# Get details of a specific trigger
a365 triggers get TRIGGER_ID

# Update a trigger definition
a365 triggers update TOKEN TRIGGER_ID

# Delete a trigger definition
a365 triggers delete TRIGGER_ID

# Evaluate an event against existing triggers
a365 triggers evaluate "mailReceived" '{"from":"manager@contoso.com","subject":"Q3 Review"}'
```
