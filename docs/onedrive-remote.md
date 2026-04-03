# OneDrive Remote

Personal OneDrive file management. Also available as `a365 odr`.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `onedrive-remote info` | Get OneDrive drive info | _(none)_ |
| `onedrive-remote ls` | List files and folders | _(none)_ |
| `onedrive-remote search` | Search for files | `<query>` |
| `onedrive-remote get` | Get file/folder metadata | `--file-or-folder-id` `--url` |
| `onedrive-remote cat` | Read a text file | `<file-id>` |
| `onedrive-remote mkdir` | Create a folder | `<folder-name>` |
| `onedrive-remote write` | Create a text file | `<filename>` `<content-text>` |
| `onedrive-remote rename` | Rename a file or folder | `<file-or-folder-id>` `<new-name>` `<etag>` |
| `onedrive-remote mv` | Move a file | `<file-id>` `<new-parent-folder-id>` `<etag>` |
| `onedrive-remote rm` | Delete a file or folder | `<file-or-folder-id>` `<etag>` |
| `onedrive-remote share` | Share a file or folder | `<file-or-folder-id>` `<emails...>` `--roles` |
| `onedrive-remote label` | Set sensitivity label | `<file-id>` `<sensitivity-label-id>` |

## Examples

```bash
# Get OneDrive info and list root contents
a365 odr info
a365 odr ls

# Search for a file
a365 odr search "quarterly report"

# Get metadata by ID or URL
a365 odr get --file-or-folder-id FILE_ID
a365 odr get --url "https://contoso-my.sharepoint.com/personal/..."

# Read a text file
a365 odr cat FILE_ID

# Create a folder and a text file
a365 odr mkdir "Project Notes"
a365 odr write "notes.txt" "Meeting notes from today"

# Rename a file
a365 odr rename FILE_ID "new-name.txt" ETAG

# Move a file to another folder
a365 odr mv FILE_ID DEST_FOLDER_ID ETAG

# Delete a file (prompts for confirmation)
a365 odr rm FILE_ID ETAG

# Share a file with colleagues
a365 odr share FILE_ID alice@contoso.com bob@contoso.com --roles read

# Set a sensitivity label on a file
a365 odr label FILE_ID LABEL_ID
```
