# Word

Create, read, and collaborate on Microsoft Word documents. Supports creating new documents, retrieving content, and adding comments or replies to existing documents.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `word create` | Create a new Word document | `<file-name>` |
| `word get` | Get document content | `<drive-id> <document-id>` |
| `word comment` | Add a comment to a document | `<drive-id> <document-id> <text>` |
| `word reply` | Reply to a document comment | `<comment-id> <drive-id> <document-id> <text>` |

## Arguments

- **`<file-name>`** -- Desired file name for the new document.
- **`<drive-id>`** -- OneDrive or SharePoint drive ID.
- **`<document-id>`** -- Document ID within the drive.
- **`<comment-id>`** -- ID of the comment to reply to.
- **`<text>`** -- Comment or reply text.
- **`--dry-run`** -- Preview write operations without executing them (supported by `create`, `comment`, and `reply`).

## Examples

```sh
# Create a new document
a365 word create "Project Proposal.docx"

# Preview creation without making changes
a365 word create "Draft Notes.docx" --dry-run

# Get document content
a365 word get b!xYzDriveId01 01ABCDEF23456789

# Add a comment to a document
a365 word comment b!xYzDriveId01 01ABCDEF23456789 "Please review section 3"

# Reply to an existing comment
a365 word reply comment-id-456 b!xYzDriveId01 01ABCDEF23456789 "Done, updated."

# Dry-run a reply to see what would happen
a365 word reply comment-id-456 b!xYzDriveId01 01ABCDEF23456789 "Looks good" --dry-run

# Output as JSON
a365 word get b!xYzDriveId01 01ABCDEF23456789 --output json
```
