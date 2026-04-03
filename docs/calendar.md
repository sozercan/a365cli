# Calendar

Manage Outlook calendar events, meeting responses, and room availability. Also available as `a365 cal`.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `calendar list` | List upcoming events | `--max` |
| `calendar view` | List events in a date range | `--max` |
| `calendar create` | Create a calendar event | `<subject>` `--start` `--end` `--attendee` `--body` `--teams` |
| `calendar update` | Update a calendar event | `<event-id>` `--subject` `--start` `--end` `--body` |
| `calendar delete` | Delete a calendar event | `<event-id>` |
| `calendar accept` | Accept a meeting invite | `<event-id>` |
| `calendar tentative` | Tentatively accept a meeting | `<event-id>` |
| `calendar decline` | Decline a meeting invite | `<event-id>` |
| `calendar cancel` | Cancel a meeting you organized | `<event-id>` |
| `calendar forward` | Forward a meeting invite | `<event-id>` `<recipients...>` `--comment` |
| `calendar free-busy` | Find available meeting times | _(none)_ |
| `calendar timezone` | Get user date/time zone settings | _(none)_ |
| `calendar rooms` | List available rooms | _(none)_ |

## Examples

```bash
# See your upcoming events
a365 cal list
a365 cal list --max 5

# View calendar in a date range
a365 cal view --max 30

# Create a 1-hour meeting with a Teams link
a365 cal create "Sprint Planning" \
  --start 2025-07-14T10:00:00 \
  --end 2025-07-14T11:00:00 \
  --attendee alice@contoso.com \
  --attendee bob@contoso.com \
  --teams

# Update the subject and time of an event
a365 cal update EVENT_ID --subject "Sprint Planning (moved)" --start 2025-07-14T14:00:00

# Respond to meeting invites
a365 cal accept EVENT_ID
a365 cal tentative EVENT_ID
a365 cal decline EVENT_ID

# Cancel a meeting you organized (prompts for confirmation)
a365 cal cancel EVENT_ID

# Forward a meeting to additional people
a365 cal forward EVENT_ID dave@contoso.com --comment "Please join us"

# Check your timezone settings
a365 cal timezone

# Find available meeting times and list rooms
a365 cal free-busy
a365 cal rooms

# Delete an event (prompts for confirmation)
a365 cal delete EVENT_ID
```
