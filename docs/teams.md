# Teams

Manage Microsoft Teams — teams, channels, chats, and message search from the command line.

## Commands

### Teams

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `teams list` | List joined teams | `--user-id` (optional) `--max` |
| `teams get` | Get a team by ID | `<team-id>` |

### Channels

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `teams channels list` | List channels in a team | `<team-id>` `--max` |
| `teams channels get` | Get a channel by ID | `<team-id>` `<channel-id>` |
| `teams channels create` | Create a standard channel | `<team-id>` `<name>` `--description` |
| `teams channels create-private` | Create a private channel | `<team-id>` `<name>` `--description` |
| `teams channels update` | Update channel name/description | `<team-id>` `<channel-id>` `--display-name` `--description` |
| `teams channels messages` | List messages in a channel | `<team-id>` `<channel-id>` `--max` |
| `teams channels post` | Post a message to a channel | `<team-id>` `<channel-id>` `<message>` |
| `teams channels reply` | Reply to a channel message | `<team-id>` `<channel-id>` `<message-id>` `<reply>` |
| `teams channels members` | List members of a channel | `<team-id>` `<channel-id>` `--max` |
| `teams channels add-member` | Add a member to a channel | `<team-id>` `<channel-id>` `<user-id>` |
| `teams channels update-member` | Update a member's role | `<team-id>` `<channel-id>` `<membership-id>` `<role>` |

### Chats

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `teams chats list` | List recent chats | `--topic` `--upn` `--max` |
| `teams chats get` | Get a chat by ID | `<chat-id>` |
| `teams chats create` | Create or get a chat | `<type>` `<members...>` `--topic` |
| `teams chats delete` | Delete a chat | `<chat-id>` |
| `teams chats update` | Update a group chat topic | `<chat-id>` `<topic>` |
| `teams chats messages` | List messages in a chat | `<chat-id>` `--max` |
| `teams chats send` | Send a message to a chat | `<chat-id>` `<message>` |
| `teams chats send-self` | Send a message/note to yourself | `<message>` |
| `teams chats get-message` | Get a specific message | `<chat-id>` `<message-id>` |
| `teams chats update-message` | Update a chat message | `<chat-id>` `<message-id>` `<content>` |
| `teams chats delete-message` | Delete a chat message | `<chat-id>` `<message-id>` |
| `teams chats members` | List members of a chat | `<chat-id>` |
| `teams chats add-member` | Add a member to a chat | `<chat-id>` `<upn>` `--roles` |

### Search

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `teams search` | Search messages (KQL syntax) | `<query>` `--max` |
| `teams search-nl` | Search messages (natural language) | `<query>` `--conversation-id` |

## Examples

```bash
# List all teams you belong to
a365 teams list

# List channels in a team, then read recent messages
a365 teams channels list TEAM_ID
a365 teams channels messages TEAM_ID CHANNEL_ID

# Post a message and reply to it
a365 teams channels post TEAM_ID CHANNEL_ID "Build is green"
a365 teams channels reply TEAM_ID CHANNEL_ID MSG_ID "Nice work!"

# Create a private channel and add a member
a365 teams channels create-private TEAM_ID "Secret Project"
a365 teams channels add-member TEAM_ID CHANNEL_ID USER_GUID

# Start a 1:1 chat and send a message
a365 teams chats create oneOnOne alice@contoso.com
a365 teams chats send CHAT_ID "Hey, got a minute?"

# Create a group chat with a topic
a365 teams chats create group alice@contoso.com bob@contoso.com --topic "Launch planning"

# Send yourself a quick note
a365 teams chats send-self "Remember to update the deck"

# Browse and manage chat messages
a365 teams chats messages CHAT_ID --max 10
a365 teams chats update-message CHAT_ID MSG_ID "Updated text"
a365 teams chats delete-message CHAT_ID MSG_ID

# Search for messages using KQL
a365 teams search "from:alice budget sent>=2025-01-01"

# Search using natural language
a365 teams search-nl "conversations about the Q3 roadmap"
```
