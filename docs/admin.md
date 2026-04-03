# Admin

Manage users and licenses in your Microsoft 365 tenant. Search for users, view available licenses, and assign or remove license SKUs.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `admin search-users` | Search for users in the tenant | `<query>` |
| `admin list-licenses` | List available licenses | *(none)* |
| `admin set-license` | Add or remove licenses for a user | `<user-id>`, `--add`, `--remove` |

## Arguments

- **`<query>`** -- Search term to find users (name, email, etc.).
- **`<user-id>`** -- User ID (GUID) of the target user.
- **`--add`** -- One or more license SKU IDs to assign to the user (repeatable).
- **`--remove`** -- One or more license SKU IDs to remove from the user (repeatable).
- **`--dry-run`** -- Preview license changes without executing them (supported by `set-license`).

## Examples

```sh
# Search for a user by name
a365 admin search-users "Jane Doe"

# Search by email
a365 admin search-users "jdoe@contoso.com"

# List all available licenses in the tenant
a365 admin list-licenses

# Assign a license to a user
a365 admin set-license 550e8400-e29b-41d4-a716-446655440000 \
  --add c42b9cae-ea4f-4ab7-9717-81576235ccac

# Remove a license from a user
a365 admin set-license 550e8400-e29b-41d4-a716-446655440000 \
  --remove c42b9cae-ea4f-4ab7-9717-81576235ccac

# Add one license and remove another in a single call
a365 admin set-license 550e8400-e29b-41d4-a716-446655440000 \
  --add 05e9a617-0261-4cee-bb44-138d3ef5d965 \
  --remove c42b9cae-ea4f-4ab7-9717-81576235ccac

# Preview license changes without applying
a365 admin set-license 550e8400-e29b-41d4-a716-446655440000 \
  --add c42b9cae-ea4f-4ab7-9717-81576235ccac --dry-run

# Output as JSON
a365 admin list-licenses --output json
```
