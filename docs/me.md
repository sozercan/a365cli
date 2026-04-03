# Me / Users

Look up user profiles, org charts, and directory information.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `me whoami` | Get your own profile details | _(none)_ |
| `me get` | Get a user's details by UPN or ID | `<user>` (email or GUID) |
| `me search` | Search for multiple users | `<query...>` (names, emails, or IDs) |
| `me manager` | Get a user's manager | `<user-id>` (GUID) |
| `me reports` | Get a user's direct reports | `<user-id>` (GUID) |

## Examples

```bash
# Check your own profile
a365 me whoami

# Look up a colleague by email
a365 me get alice@contoso.com

# Look up a user by ID
a365 me get 00000000-0000-0000-0000-000000000001

# Search for multiple users at once
a365 me search "Alice" "Bob" "carol@contoso.com"

# Find someone's manager
a365 me manager 00000000-0000-0000-0000-000000000001

# List a manager's direct reports
a365 me reports 00000000-0000-0000-0000-000000000001

# Common workflow: find a user, then explore their org chart
a365 me get alice@contoso.com          # get Alice's user ID
a365 me manager ALICE_USER_ID          # who does Alice report to?
a365 me reports ALICE_USER_ID          # who reports to Alice?
```
