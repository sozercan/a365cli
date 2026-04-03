# Admin365

Admin365 agent and Copilot settings — manage tenant-wide agent access, app install policies, and Copilot configuration.

## Commands

### Agent & App Settings (Read)

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `admin365 agent-access` | Get agent access settings | _(none)_ |
| `admin365 agent-sharing` | Get agent sharing settings | _(none)_ |
| `admin365 ms-apps` | Get Microsoft apps install settings | _(none)_ |
| `admin365 third-party` | Get third-party apps settings | _(none)_ |
| `admin365 lob-apps` | Get LOB apps settings | _(none)_ |

### Agent & App Settings (Write)

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `admin365 set-access` | Update agent access settings | `<access-level>` |
| `admin365 set-sharing` | Update agent sharing settings | `<access-level>` |
| `admin365 set-ms-apps` | Update Microsoft apps settings | `<allowed>` (true/false) |
| `admin365 set-third-party` | Update third-party apps settings | `<allowed>` (true/false) |
| `admin365 set-lob-apps` | Update LOB apps settings | `<allowed>` (true/false) |

### Users & Copilot

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `admin365 bulk-add` | Bulk add users to tenant | `<file-content>` (CSV/JSON) |
| `admin365 copilot-readiness` | Check Copilot readiness | _(none)_ |
| `admin365 copilot-status` | Get Copilot admin settings | _(none)_ |
| `admin365 set-copilot` | Enable/disable Copilot for admins | `<is-enabled>` (true/false) |

## Examples

```bash
# Check current agent access and sharing settings
a365 admin365 agent-access
a365 admin365 agent-sharing

# Review app install policies
a365 admin365 ms-apps
a365 admin365 third-party
a365 admin365 lob-apps

# Update agent access level
a365 admin365 set-access "EveryoneInOrg"

# Allow or block third-party app installs
a365 admin365 set-third-party true
a365 admin365 set-lob-apps false

# Bulk add users from a CSV
a365 admin365 bulk-add "$(cat users.csv)"

# Check Copilot readiness and current status
a365 admin365 copilot-readiness
a365 admin365 copilot-status

# Enable Copilot for admins
a365 admin365 set-copilot true
```
