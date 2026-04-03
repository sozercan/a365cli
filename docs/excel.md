# Excel

Create, read, and collaborate on Microsoft Excel workbooks. Supports creating new workbooks, retrieving content, and adding comments at specific cell addresses.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `excel create` | Create a new Excel workbook | `<file-name>` |
| `excel get` | Get workbook content | `<drive-id> <document-id>` |
| `excel comment` | Add a comment to a workbook cell | `<drive-id> <document-id> <cell-address> <text>` |
| `excel reply` | Reply to a workbook comment | `<comment-id> <drive-id> <document-id> <text>` |

## Arguments

- **`<file-name>`** -- Desired file name for the new workbook.
- **`<drive-id>`** -- OneDrive or SharePoint drive ID.
- **`<document-id>`** -- Document ID within the drive.
- **`<cell-address>`** -- Cell address for the comment (e.g. `A1`, `B2`, `C10`).
- **`<comment-id>`** -- ID of the comment to reply to.
- **`<text>`** -- Comment or reply text.
- **`--dry-run`** -- Preview write operations without executing them (supported by `create`, `comment`, and `reply`).

## Examples

```sh
# Create a new workbook
a365 excel create "Q3 Budget.xlsx"

# Preview creation without making changes
a365 excel create "Expenses.xlsx" --dry-run

# Get workbook content
a365 excel get b!xYzDriveId01 01ABCDEF23456789

# Add a comment at cell B5
a365 excel comment b!xYzDriveId01 01ABCDEF23456789 B5 "This value looks off"

# Dry-run a comment to preview
a365 excel comment b!xYzDriveId01 01ABCDEF23456789 A1 "Check formula" --dry-run

# Reply to an existing comment
a365 excel reply comment-id-789 b!xYzDriveId01 01ABCDEF23456789 "Fixed the formula"

# Output as JSON
a365 excel get b!xYzDriveId01 01ABCDEF23456789 --output json
```
