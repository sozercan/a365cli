# SharePoint Lists

Manage SharePoint list sites, lists, items, and columns.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `sp-lists sites` | Search for SharePoint sites | `<query>` |
| `sp-lists site` | Get a site by hostname and path | `<hostname>` `<server-relative-path>` |
| `sp-lists subsites` | List child sites | `<site-id>` |
| `sp-lists lists` | List SharePoint lists in a site | `<site-id>` |
| `sp-lists items` | List items in a list | `<site-id>` `<list-id>` |
| `sp-lists columns` | List columns in a list | `<site-id>` `<list-id>` |
| `sp-lists create` | Create a new list | `<site-id>` `<display-name>` |
| `sp-lists add-column` | Add a column to a list | `<site-id>` `<list-id>` `<name>` `<type>` |
| `sp-lists add-item` | Add an item to a list | `<site-id>` `<list-id>` `<fields-json>` |
| `sp-lists update-item` | Update a list item | `<site-id>` `<list-id>` `<item-id>` `--fields` |
| `sp-lists edit-column` | Edit a list column | `<site-id>` `<list-id>` `<column-id>` `--name` |
| `sp-lists delete-item` | Delete a list item | `<site-id>` `<list-id>` `<item-id>` |
| `sp-lists delete-column` | Delete a list column | `<site-id>` `<list-id>` `<column-id>` |

## Examples

```bash
# Find SharePoint sites by name
a365 sp-lists sites "Engineering"

# Get a specific site by hostname and path
a365 sp-lists site contoso.sharepoint.com /sites/engineering

# Explore site hierarchy
a365 sp-lists subsites SITE_ID

# List all lists in a site, then inspect columns and items
a365 sp-lists lists SITE_ID
a365 sp-lists columns SITE_ID LIST_ID
a365 sp-lists items SITE_ID LIST_ID

# Create a new list
a365 sp-lists create SITE_ID "Bug Tracker"

# Add columns to define the schema
a365 sp-lists add-column SITE_ID LIST_ID "Priority" choice
a365 sp-lists add-column SITE_ID LIST_ID "DueDate" dateTime
a365 sp-lists add-column SITE_ID LIST_ID "Resolved" boolean

# Add items to the list (fields as JSON)
a365 sp-lists add-item SITE_ID LIST_ID '{"Title":"Login bug","Priority":"High"}'
a365 sp-lists add-item SITE_ID LIST_ID '{"Title":"UI glitch","Priority":"Low"}'

# Update an item's fields
a365 sp-lists update-item SITE_ID LIST_ID ITEM_ID --fields '{"Resolved":true}'

# Rename a column
a365 sp-lists edit-column SITE_ID LIST_ID COL_ID --name "Severity"

# Delete an item (prompts for confirmation)
a365 sp-lists delete-item SITE_ID LIST_ID ITEM_ID

# Delete a column (prompts for confirmation)
a365 sp-lists delete-column SITE_ID LIST_ID COL_ID
```
