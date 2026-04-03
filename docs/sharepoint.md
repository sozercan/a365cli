# SharePoint

Manage SharePoint Online files and folders with Unix-style commands. Also available as `a365 sp`.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `sharepoint find` | Find SharePoint sites | `--query` `--site-url` |
| `sharepoint libs` | List document libraries in a site | `--site-id` `--site-path` `--default` |
| `sharepoint ls` | List folder contents | `<drive-id>` `--folder-path` `--folder-id` |
| `sharepoint get` | Get file/folder metadata | `--drive-id` `--item-path` `--item-id` `--url` |
| `sharepoint cat` | Read a text file | `<drive-id>` `--file-path` `--file-id` `--binary` |
| `sharepoint search` | Search for files or folders | `<query>` `--drive-id` `--site-id` |
| `sharepoint mkdir` | Create a folder | `<drive-id>` `<parent-path>` `<folder-name>` |
| `sharepoint write` | Create a text file | `<drive-id>` `<folder-path>` `<file-name>` `--content` `--content-base64` |
| `sharepoint upload` | Upload a file from URL | `<source-url>` `<drive-id>` `<folder-path>` `<file-name>` |
| `sharepoint rm` | Delete a file or folder | `<drive-id>` `--item-path` `--item-id` |
| `sharepoint mv` | Move a file or folder | `<src-drive>` `<src-path>` `<dst-drive>` `<dst-folder>` |
| `sharepoint cp` | Copy a file or folder | `<src-drive>` `<src-path>` `<dst-drive>` `<dst-folder>` |
| `sharepoint rename` | Rename a file or folder | `<drive-id>` `<item-path>` `<new-name>` |
| `sharepoint share` | Create a sharing link | `<drive-id>` `<item-path>` `<type>` `<scope>` |
| `sharepoint label` | Set sensitivity label on a file | `<drive-id>` `<item-id>` `<label-id>` |
| `sharepoint status` | Check async operation status | `<operation-url>` |

## Examples

```bash
# Find a SharePoint site by name
a365 sp find --query "Engineering"

# Look up a site by its URL
a365 sp find --site-url "https://contoso.sharepoint.com/sites/engineering"

# List document libraries in a site
a365 sp libs --site-path /sites/engineering

# Browse files in a library
a365 sp ls DRIVE_ID --folder-path /Documents/Reports

# Read a text file
a365 sp cat DRIVE_ID --file-path /Documents/README.md

# Search across a site
a365 sp search "quarterly report" --site-id SITE_ID

# Create a folder and write a file into it
a365 sp mkdir DRIVE_ID /Documents "Meeting Notes"
a365 sp write DRIVE_ID /Documents/Meeting\ Notes notes.txt --content "# Notes\n\nAction items..."

# Upload a file from a URL
a365 sp upload "https://example.com/report.pdf" DRIVE_ID /Documents report.pdf

# Move, copy, and rename files
a365 sp mv DRIVE_A /old/path.docx DRIVE_B /new/folder
a365 sp cp DRIVE_ID /src/template.xlsx DRIVE_ID /dst
a365 sp rename DRIVE_ID /Documents/draft.docx "final.docx"

# Share a file (view link, org-wide)
a365 sp share DRIVE_ID /Documents/report.pdf view organization

# Apply a sensitivity label
a365 sp label DRIVE_ID ITEM_ID LABEL_ID

# Delete a file (prompts for confirmation)
a365 sp rm DRIVE_ID --item-path /Documents/old-report.pdf

# Check status of an async copy/move operation
a365 sp status "https://contoso.sharepoint.com/_api/..."

# Get metadata for a file by SharePoint URL
a365 sp get --url "https://contoso.sharepoint.com/sites/eng/Shared Documents/spec.docx"
```
