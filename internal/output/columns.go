package output

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
)

// Column defines a column for table/TSV output.
type Column struct {
	Header  string                         // Display header, e.g. "DISPLAY NAME"
	Width   int                            // Max chars for table display (0 = unlimited)
	Extract func(row map[string]any) string // Pull value from row
}

// --- Helper extractors ---

func getString(row map[string]any, key string) string {
	v, ok := row[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

func getNestedString(row map[string]any, outer, inner string) string {
	v, ok := row[outer]
	if !ok || v == nil {
		return ""
	}
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	return getString(m, inner)
}

func formatTime(s string) string {
	if s == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try without timezone
		t, err = time.Parse("2006-01-02T15:04:05Z", s)
		if err != nil {
			return s
		}
	}
	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return t.Local().Format("15:04")
	}
	if now.Sub(t) < 7*24*time.Hour {
		return t.Local().Format("Mon 15:04")
	}
	return t.Local().Format("Jan 2")
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// stripHTML removes HTML tags and unescapes HTML entities,
// with special handling for Teams-specific markup.
var (
	htmlTagRe      = regexp.MustCompile(`<[^>]*>`)
	emojiRe        = regexp.MustCompile(`<emoji[^>]*\balt="([^"]*)"[^>]*>.*?</emoji>`)
	attachmentRe   = regexp.MustCompile(`<attachment[^>]*>.*?</attachment>`)
	systemEventRe  = regexp.MustCompile(`<systemEventMessage[^>]*/>`)
	imgAltRe       = regexp.MustCompile(`<img[^>]*\balt="([^"]*)"[^>]*>`)
	codeBlockRe    = regexp.MustCompile(`<codeblock[^>]*><code>|</code></codeblock>`)
	anchorRe       = regexp.MustCompile(`<a[^>]*>(.*?)</a>`)
)

func stripHTML(s string) string {
	// 1. Extract emoji alt attributes (actual emoji characters)
	s = emojiRe.ReplaceAllString(s, "$1")
	// 2. Replace attachment tags with placeholder
	s = attachmentRe.ReplaceAllString(s, "[attachment]")
	// 3. Remove system event messages (join/leave events)
	s = systemEventRe.ReplaceAllString(s, "")
	// 4. Extract img alt text
	s = imgAltRe.ReplaceAllString(s, "[$1]")
	// 5. Preserve anchor text
	s = anchorRe.ReplaceAllString(s, "$1")
	// 6. Strip codeblock/code wrappers (keep code text)
	s = codeBlockRe.ReplaceAllString(s, "")
	// 7. Strip remaining HTML tags (replace with space to avoid word concatenation)
	s = htmlTagRe.ReplaceAllString(s, " ")
	// 8. Unescape HTML entities
	s = html.UnescapeString(s)
	// 9. Replace non-breaking space (U+00A0) with regular space
	s = strings.ReplaceAll(s, "\u00a0", " ")
	// 10. Normalize line endings
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	// 11. Collapse multiple spaces
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}

// --- Column definitions per entity ---

// TeamsColumns defines display columns for teams list.
var TeamsColumns = []Column{
	{Header: "DISPLAY NAME", Width: 40, Extract: func(row map[string]any) string {
		return getString(row, "displayName")
	}},
	{Header: "ID", Width: 36, Extract: func(row map[string]any) string {
		return getString(row, "id")
	}},
	{Header: "DESCRIPTION", Width: 50, Extract: func(row map[string]any) string {
		return truncate(getString(row, "description"), 50)
	}},
}

// ChannelsColumns defines display columns for channels list.
var ChannelsColumns = []Column{
	{Header: "DISPLAY NAME", Width: 35, Extract: func(row map[string]any) string {
		return getString(row, "displayName")
	}},
	{Header: "ID", Width: 0, Extract: func(row map[string]any) string {
		return getString(row, "id")
	}},
	{Header: "TYPE", Width: 10, Extract: func(row map[string]any) string {
		return getString(row, "membershipType")
	}},
	{Header: "CREATED", Width: 10, Extract: func(row map[string]any) string {
		return formatTime(getString(row, "createdDateTime"))
	}},
}

// ChatsColumns defines display columns for chats list.
var ChatsColumns = []Column{
	{Header: "TOPIC", Width: 35, Extract: func(row map[string]any) string {
		topic := getString(row, "topic")
		if topic != "" {
			return topic
		}
		// For 1:1 chats, show member names
		members, ok := row["members"]
		if !ok {
			return "(no topic)"
		}
		arr, ok := members.([]any)
		if !ok || len(arr) == 0 {
			return "(no topic)"
		}
		var names []string
		for _, m := range arr {
			if mm, ok := m.(map[string]any); ok {
				name := getString(mm, "displayName")
				if name != "" {
					names = append(names, name)
				}
			}
		}
		return strings.Join(names, ", ")
	}},
	{Header: "ID", Width: 0, Extract: func(row map[string]any) string {
		return getString(row, "id")
	}},
	{Header: "TYPE", Width: 10, Extract: func(row map[string]any) string {
		return getString(row, "chatType")
	}},
	{Header: "UPDATED", Width: 10, Extract: func(row map[string]any) string {
		return formatTime(getString(row, "lastUpdatedDateTime"))
	}},
}

// MessagesColumns defines display columns for message lists.
var MessagesColumns = []Column{
	{Header: "DATE", Width: 10, Extract: func(row map[string]any) string {
		return formatTime(getString(row, "createdDateTime"))
	}},
	{Header: "FROM", Width: 20, Extract: func(row map[string]any) string {
		return truncate(getNestedString(row, "from", "displayName"), 20)
	}},
	{Header: "CONTENT", Width: 80, Extract: func(row map[string]any) string {
		body, ok := row["body"]
		if !ok || body == nil {
			return ""
		}
		bm, ok := body.(map[string]any)
		if !ok {
			return ""
		}
		content := getString(bm, "content")
		content = stripHTML(content)
		return truncate(content, 80)
	}},
	{Header: "ID", Width: 0, Extract: func(row map[string]any) string {
		return getString(row, "id")
	}},
}

// SearchColumns defines display columns for search results.
var SearchColumns = []Column{
	{Header: "DATE", Width: 10, Extract: func(row map[string]any) string {
		return formatTime(getString(row, "createdDateTime"))
	}},
	{Header: "FROM", Width: 20, Extract: func(row map[string]any) string {
		return truncate(getNestedString(row, "from", "displayName"), 20)
	}},
	{Header: "PREVIEW", Width: 60, Extract: func(row map[string]any) string {
		// Search results may have summary or body
		summary := getString(row, "summary")
		if summary != "" {
			return truncate(stripHTML(summary), 60)
		}
		body, ok := row["body"]
		if !ok || body == nil {
			return ""
		}
		bm, ok := body.(map[string]any)
		if !ok {
			return ""
		}
		return truncate(stripHTML(getString(bm, "content")), 60)
	}},
}

// MembersColumns defines display columns for member lists.
var MembersColumns = []Column{
	{Header: "DISPLAY NAME", Width: 30, Extract: func(row map[string]any) string {
		return getString(row, "displayName")
	}},
	{Header: "EMAIL", Width: 35, Extract: func(row map[string]any) string {
		return getString(row, "email")
	}},
	{Header: "ID", Width: 36, Extract: func(row map[string]any) string {
		return getString(row, "id")
	}},
	{Header: "ROLES", Width: 10, Extract: func(row map[string]any) string {
		roles, ok := row["roles"]
		if !ok || roles == nil {
			return ""
		}
		arr, ok := roles.([]any)
		if !ok {
			return ""
		}
		var rs []string
		for _, r := range arr {
			if s, ok := r.(string); ok {
				rs = append(rs, s)
			}
		}
		return strings.Join(rs, ",")
	}},
}

// --- Mail columns ---

// MailColumns defines display columns for email lists.
var MailColumns = []Column{
	{Header: "DATE", Width: 10, Extract: func(row map[string]any) string {
		return formatTime(getString(row, "receivedDateTime"))
	}},
	{Header: "FROM", Width: 25, Extract: func(row map[string]any) string {
		from, ok := row["from"]
		if !ok || from == nil {
			return ""
		}
		fm, ok := from.(map[string]any)
		if !ok {
			return ""
		}
		ea, ok := fm["emailAddress"]
		if !ok || ea == nil {
			return getString(fm, "name")
		}
		eam, ok := ea.(map[string]any)
		if !ok {
			return ""
		}
		name := getString(eam, "name")
		if name != "" {
			return truncate(name, 25)
		}
		return truncate(getString(eam, "address"), 25)
	}},
	{Header: "SUBJECT", Width: 50, Extract: func(row map[string]any) string {
		return truncate(getString(row, "subject"), 50)
	}},
	{Header: "READ", Width: 4, Extract: func(row map[string]any) string {
		v, ok := row["isRead"]
		if !ok {
			return ""
		}
		if b, ok := v.(bool); ok && b {
			return "yes"
		}
		return "no"
	}},
	{Header: "ID", Width: 0, Extract: func(row map[string]any) string {
		return getString(row, "id")
	}},
}

// --- Calendar columns ---

// CalendarColumns defines display columns for calendar events.
var CalendarColumns = []Column{
	{Header: "START", Width: 16, Extract: func(row map[string]any) string {
		start, ok := row["start"]
		if !ok || start == nil {
			return formatTime(getString(row, "startDateTime"))
		}
		sm, ok := start.(map[string]any)
		if !ok {
			return ""
		}
		return formatTime(getString(sm, "dateTime"))
	}},
	{Header: "SUBJECT", Width: 40, Extract: func(row map[string]any) string {
		return truncate(getString(row, "subject"), 40)
	}},
	{Header: "ORGANIZER", Width: 25, Extract: func(row map[string]any) string {
		org, ok := row["organizer"]
		if !ok || org == nil {
			return ""
		}
		om, ok := org.(map[string]any)
		if !ok {
			return ""
		}
		ea, ok := om["emailAddress"]
		if !ok || ea == nil {
			return ""
		}
		eam, ok := ea.(map[string]any)
		if !ok {
			return ""
		}
		name := getString(eam, "name")
		if name != "" {
			return truncate(name, 25)
		}
		return truncate(getString(eam, "address"), 25)
	}},
	{Header: "ID", Width: 0, Extract: func(row map[string]any) string {
		return getString(row, "id")
	}},
}

// --- Planner columns ---

// PlannerPlanColumns defines display columns for planner plans.
var PlannerPlanColumns = []Column{
	{Header: "TITLE", Width: 40, Extract: func(row map[string]any) string {
		return truncate(getString(row, "title"), 40)
	}},
	{Header: "ID", Width: 36, Extract: func(row map[string]any) string {
		return getString(row, "id")
	}},
	{Header: "CREATED", Width: 10, Extract: func(row map[string]any) string {
		return formatTime(getString(row, "createdDateTime"))
	}},
}

// PlannerTaskColumns defines display columns for planner tasks.
var PlannerTaskColumns = []Column{
	{Header: "TITLE", Width: 40, Extract: func(row map[string]any) string {
		return truncate(getString(row, "title"), 40)
	}},
	{Header: "ID", Width: 36, Extract: func(row map[string]any) string {
		return getString(row, "id")
	}},
	{Header: "STATUS", Width: 12, Extract: func(row map[string]any) string {
		pct := row["percentComplete"]
		if pct == nil {
			return ""
		}
		switch v := pct.(type) {
		case float64:
			if v == 100 {
				return "completed"
			} else if v > 0 {
				return "in progress"
			}
			return "not started"
		default:
			return fmt.Sprintf("%v", v)
		}
	}},
	{Header: "PRIORITY", Width: 8, Extract: func(row map[string]any) string {
		p := row["priority"]
		if p == nil {
			return ""
		}
		switch v := p.(type) {
		case float64:
			switch int(v) {
			case 1:
				return "urgent"
			case 3:
				return "important"
			case 5:
				return "medium"
			case 9:
				return "low"
			default:
				return fmt.Sprintf("%d", int(v))
			}
		default:
			return fmt.Sprintf("%v", v)
		}
	}},
}

// --- User columns ---

// UserColumns defines display columns for user lists.
var UserColumns = []Column{
	{Header: "DISPLAY NAME", Width: 30, Extract: func(row map[string]any) string {
		return getString(row, "displayName")
	}},
	{Header: "UPN", Width: 35, Extract: func(row map[string]any) string {
		upn := getString(row, "userPrincipalName")
		if upn != "" {
			return upn
		}
		return getString(row, "mail")
	}},
	{Header: "JOB TITLE", Width: 25, Extract: func(row map[string]any) string {
		return truncate(getString(row, "jobTitle"), 25)
	}},
	{Header: "ID", Width: 36, Extract: func(row map[string]any) string {
		return getString(row, "id")
	}},
}

// --- API explorer columns ---

// APIServerColumns defines columns for the server list.
var APIServerColumns = []Column{
	{Header: "SERVICE", Width: 15, Extract: func(row map[string]any) string {
		return getString(row, "service")
	}},
	{Header: "SERVER", Width: 0, Extract: func(row map[string]any) string {
		return getString(row, "server")
	}},
}

// APIServerProbeColumns defines columns for the probed server list.
var APIServerProbeColumns = []Column{
	{Header: "SERVICE", Width: 15, Extract: func(row map[string]any) string {
		return getString(row, "service")
	}},
	{Header: "SERVER", Width: 30, Extract: func(row map[string]any) string {
		return getString(row, "server")
	}},
	{Header: "TOOLS", Width: 5, Extract: func(row map[string]any) string {
		return getString(row, "tools")
	}},
	{Header: "STATUS", Width: 0, Extract: func(row map[string]any) string {
		return getString(row, "status")
	}},
}

// APIDiscoverColumns defines columns for discovered servers.
var APIDiscoverColumns = []Column{
	{Header: "NAME", Width: 30, Extract: func(row map[string]any) string {
		return getString(row, "name")
	}},
	{Header: "URL", Width: 60, Extract: func(row map[string]any) string {
		return getString(row, "url")
	}},
	{Header: "SCOPE", Width: 30, Extract: func(row map[string]any) string {
		return getString(row, "scope")
	}},
}

// APIToolColumns defines columns for the tool list.
var APIToolColumns = []Column{
	{Header: "TOOL", Width: 40, Extract: func(row map[string]any) string {
		return getString(row, "name")
	}},
	{Header: "REQUIRED", Width: 30, Extract: func(row map[string]any) string {
		return getString(row, "required")
	}},
	{Header: "DESCRIPTION", Width: 60, Extract: func(row map[string]any) string {
		return truncate(getString(row, "description"), 60)
	}},
}
