# Mail

Read, search, send, and manage Outlook email. Also available as `a365 email`.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `mail search` | Search emails (OData query) | `<query>` `--max` |
| `mail search-nl` | Search emails (natural language) | `<query>` |
| `mail get` | Get an email by ID | `<message-id>` |
| `mail send` | Send an email | `<to...>` `--subject` `--body` `--cc` `--bcc` |
| `mail reply` | Reply to an email | `<message-id>` `<comment>` `--send` |
| `mail reply-all` | Reply all to an email | `<message-id>` `<comment>` `--send` |
| `mail forward` | Forward an email | `<message-id>` `<to...>` `--comment` |
| `mail delete` | Delete an email | `<message-id>` |
| `mail flag` | Flag or unflag an email | `<message-id>` `<status>` (Flagged\|Complete\|NotFlagged) |
| `mail draft` | Create a draft email | `--subject` `--body` `--to` `--cc` |
| `mail send-draft` | Send a saved draft | `<draft-id>` |
| `mail attachments` | List attachments on an email | `<message-id>` |
| `mail update` | Update an email | `<id>` `--subject` `--body` `--importance` `--categories` |
| `mail download` | Download an attachment | `<message-id>` `<attachment-id>` |
| `mail upload-attachment` | Upload an attachment to a message | `<message-id>` `<filename>` `<base64>` `--large` |
| `mail delete-attachment` | Delete an attachment | `<message-id>` `<attachment-id>` |
| `mail update-draft` | Update a draft email | `<message-id>` `--subject` `--body` `--to` `--cc` `--bcc` |
| `mail draft-attachments` | Add attachments to a draft | `<message-id>` `<uris...>` |
| `mail reply-thread` | Reply with full conversation thread | `<message-id>` |
| `mail reply-all-thread` | Reply-all with full conversation thread | `<message-id>` |
| `mail forward-thread` | Forward with full conversation thread | `<message-id>` |

## Examples

```bash
# Search for recent emails about a topic (OData)
a365 mail search '?$search="quarterly review"' --max 5

# Search using natural language
a365 mail search-nl "emails from John about the budget report"

# Read a specific email
a365 mail get AAMkAGI2...

# Send an email with CC
a365 mail send alice@contoso.com bob@contoso.com \
  --subject "Meeting notes" \
  --body "Attached are the notes from today." \
  --cc carol@contoso.com

# Reply to a message
a365 mail reply AAMkAGI2... "Thanks, I'll review this today."

# Forward an email with a note
a365 mail forward AAMkAGI2... dave@contoso.com \
  --comment "FYI - see the original thread below."

# Flag an email for follow-up, then mark it complete
a365 mail flag AAMkAGI2... Flagged
a365 mail flag AAMkAGI2... Complete

# Create a draft, review it, then send
a365 mail draft --subject "Proposal" --body "Draft content" --to alice@contoso.com
a365 mail send-draft AAMkAGI2...

# Check attachments on an email
a365 mail attachments AAMkAGI2...

# Delete an email (prompts for confirmation)
a365 mail delete AAMkAGI2...

# Update an email's subject and importance
a365 mail update AAMkAGI2... --subject "Updated subject" --importance High

# Download an attachment
a365 mail download AAMkAGI2... AAMkAttach...

# Upload an attachment (use --large for files >3MB)
a365 mail upload-attachment AAMkAGI2... report.pdf "$(base64 report.pdf)"
a365 mail upload-attachment AAMkAGI2... bigfile.zip "$(base64 bigfile.zip)" --large

# Delete an attachment (prompts for confirmation)
a365 mail delete-attachment AAMkAGI2... AAMkAttach...

# Update a draft's recipients and body
a365 mail update-draft AAMkAGI2... --to bob@contoso.com --body "Revised content" --cc carol@contoso.com

# Add attachments to a draft by URI
a365 mail draft-attachments AAMkAGI2... https://example.com/file1.pdf https://example.com/file2.docx

# Reply with the full conversation thread included
a365 mail reply-thread AAMkAGI2...

# Reply-all with full thread
a365 mail reply-all-thread AAMkAGI2...

# Forward with full thread
a365 mail forward-thread AAMkAGI2...
```
