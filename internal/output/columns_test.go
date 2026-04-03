package output

import (
	"testing"
)

func TestStripHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<p>Hello</p>", "Hello"},
		{"<p>Hello <b>World</b></p>", "Hello World"},
		{"&amp; &lt; &gt;", "& < >"},
		{"<p>Line1</p>\n<p>Line2</p>", "Line1 Line2"},
		{"No tags here", "No tags here"},
		{"", ""},
		{"<systemEventMessage/>", ""},
		{"<p>Hello &nbsp;World</p>", "Hello World"},
	}

	for _, tt := range tests {
		result := stripHTML(tt.input)
		if result != tt.expected {
			t.Errorf("stripHTML(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"Hello", 10, "Hello"},
		{"Hello World", 8, "Hello..."},
		{"Hi", 2, "Hi"},
		{"Hello", 0, "Hello"},
		{"Hello", 3, "Hel"},
		{"AB", 5, "AB"},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.max)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, expected %q", tt.input, tt.max, result, tt.expected)
		}
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		input  string
		notEmpty bool
	}{
		{"2025-01-15T10:30:00Z", true},
		{"2026-04-03T15:00:00Z", true},
		{"", false},
		{"not-a-date", true}, // returns as-is
	}

	for _, tt := range tests {
		result := formatTime(tt.input)
		if tt.notEmpty && result == "" {
			t.Errorf("formatTime(%q) returned empty string", tt.input)
		}
		if !tt.notEmpty && result != "" {
			t.Errorf("formatTime(%q) = %q, expected empty", tt.input, result)
		}
	}
}

func TestTeamsColumns_Extract(t *testing.T) {
	row := map[string]any{
		"displayName": "My Team",
		"id":          "abc-123",
		"description": "A test team",
	}

	for _, col := range TeamsColumns {
		result := col.Extract(row)
		if result == "" {
			t.Errorf("TeamsColumns[%s] returned empty for valid row", col.Header)
		}
	}
}

func TestChannelsColumns_Extract(t *testing.T) {
	row := map[string]any{
		"displayName":     "General",
		"id":              "19:abc@thread.tacv2",
		"membershipType":  "Standard",
		"createdDateTime": "2025-01-15T10:30:00Z",
	}

	for _, col := range ChannelsColumns {
		result := col.Extract(row)
		if result == "" {
			t.Errorf("ChannelsColumns[%s] returned empty for valid row", col.Header)
		}
	}
}

func TestChatsColumns_Extract_WithTopic(t *testing.T) {
	row := map[string]any{
		"topic":              "Project Chat",
		"id":                 "19:abc@thread.v2",
		"chatType":           "Group",
		"lastUpdatedDateTime": "2025-01-15T10:30:00Z",
	}

	result := ChatsColumns[0].Extract(row) // TOPIC column
	if result != "Project Chat" {
		t.Errorf("expected 'Project Chat', got %q", result)
	}
}

func TestChatsColumns_Extract_NoTopic(t *testing.T) {
	row := map[string]any{
		"topic":    "",
		"id":       "19:abc@thread.v2",
		"chatType": "OneOnOne",
		"members": []any{
			map[string]any{"displayName": "Alice"},
			map[string]any{"displayName": "Bob"},
		},
	}

	result := ChatsColumns[0].Extract(row) // TOPIC column
	if result != "Alice, Bob" {
		t.Errorf("expected 'Alice, Bob', got %q", result)
	}
}

func TestMessagesColumns_Extract(t *testing.T) {
	row := map[string]any{
		"createdDateTime": "2025-01-15T10:30:00Z",
		"from":            map[string]any{"displayName": "Alice"},
		"body":            map[string]any{"content": "<p>Hello <b>World</b></p>", "contentType": "Html"},
		"id":              "12345",
	}

	// DATE
	date := MessagesColumns[0].Extract(row)
	if date == "" {
		t.Error("expected non-empty date")
	}

	// FROM
	from := MessagesColumns[1].Extract(row)
	if from != "Alice" {
		t.Errorf("expected 'Alice', got %q", from)
	}

	// CONTENT (should strip HTML)
	content := MessagesColumns[2].Extract(row)
	if content != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", content)
	}

	// ID
	id := MessagesColumns[3].Extract(row)
	if id != "12345" {
		t.Errorf("expected '12345', got %q", id)
	}
}

func TestGetString(t *testing.T) {
	row := map[string]any{
		"name":  "Alice",
		"count": 42.0,
		"nil":   nil,
	}

	if getString(row, "name") != "Alice" {
		t.Error("expected 'Alice'")
	}
	if getString(row, "missing") != "" {
		t.Error("expected empty for missing key")
	}
	if getString(row, "nil") != "" {
		t.Error("expected empty for nil value")
	}
	if getString(row, "count") != "42" {
		t.Errorf("expected '42', got %q", getString(row, "count"))
	}
}

func TestGetNestedString(t *testing.T) {
	row := map[string]any{
		"from": map[string]any{"displayName": "Alice"},
	}

	if getNestedString(row, "from", "displayName") != "Alice" {
		t.Error("expected 'Alice'")
	}
	if getNestedString(row, "from", "missing") != "" {
		t.Error("expected empty for missing inner key")
	}
	if getNestedString(row, "missing", "displayName") != "" {
		t.Error("expected empty for missing outer key")
	}
}

func TestMailColumns_Extract(t *testing.T) {
	row := map[string]any{
		"receivedDateTime": "2025-01-15T10:30:00Z",
		"from": map[string]any{
			"emailAddress": map[string]any{
				"name":    "Alice Johnson",
				"address": "alice@example.com",
			},
		},
		"subject": "Weekly Report",
		"isRead":  true,
		"id":      "mail-abc-123",
	}

	// DATE
	date := MailColumns[0].Extract(row)
	if date == "" {
		t.Error("expected non-empty date")
	}

	// FROM — should extract nested from.emailAddress.name
	from := MailColumns[1].Extract(row)
	if from != "Alice Johnson" {
		t.Errorf("expected 'Alice Johnson', got %q", from)
	}

	// SUBJECT
	subject := MailColumns[2].Extract(row)
	if subject != "Weekly Report" {
		t.Errorf("expected 'Weekly Report', got %q", subject)
	}

	// READ
	read := MailColumns[3].Extract(row)
	if read != "yes" {
		t.Errorf("expected 'yes', got %q", read)
	}

	// ID
	id := MailColumns[4].Extract(row)
	if id != "mail-abc-123" {
		t.Errorf("expected 'mail-abc-123', got %q", id)
	}
}

func TestMailColumns_Extract_Unread(t *testing.T) {
	row := map[string]any{
		"isRead": false,
	}
	read := MailColumns[3].Extract(row)
	if read != "no" {
		t.Errorf("expected 'no' for unread mail, got %q", read)
	}
}

func TestMailColumns_Extract_FallbackAddress(t *testing.T) {
	row := map[string]any{
		"from": map[string]any{
			"emailAddress": map[string]any{
				"name":    "",
				"address": "alice@example.com",
			},
		},
	}
	from := MailColumns[1].Extract(row)
	if from != "alice@example.com" {
		t.Errorf("expected email address fallback, got %q", from)
	}
}

func TestMailColumns_Extract_NoFrom(t *testing.T) {
	row := map[string]any{
		"subject": "Test",
	}
	from := MailColumns[1].Extract(row)
	if from != "" {
		t.Errorf("expected empty from, got %q", from)
	}
}

func TestCalendarColumns_Extract(t *testing.T) {
	row := map[string]any{
		"start": map[string]any{
			"dateTime": "2025-01-15T14:00:00Z",
			"timeZone": "UTC",
		},
		"subject": "Team Standup",
		"organizer": map[string]any{
			"emailAddress": map[string]any{
				"name":    "Bob Smith",
				"address": "bob@example.com",
			},
		},
		"id": "event-abc-123",
	}

	// START
	start := CalendarColumns[0].Extract(row)
	if start == "" {
		t.Error("expected non-empty start time")
	}

	// SUBJECT
	subject := CalendarColumns[1].Extract(row)
	if subject != "Team Standup" {
		t.Errorf("expected 'Team Standup', got %q", subject)
	}

	// ORGANIZER — nested organizer.emailAddress.name
	organizer := CalendarColumns[2].Extract(row)
	if organizer != "Bob Smith" {
		t.Errorf("expected 'Bob Smith', got %q", organizer)
	}

	// ID
	id := CalendarColumns[3].Extract(row)
	if id != "event-abc-123" {
		t.Errorf("expected 'event-abc-123', got %q", id)
	}
}

func TestCalendarColumns_Extract_FallbackStartDateTime(t *testing.T) {
	row := map[string]any{
		"startDateTime": "2025-01-15T14:00:00Z",
		"subject":       "Meeting",
	}

	start := CalendarColumns[0].Extract(row)
	if start == "" {
		t.Error("expected non-empty start time from startDateTime fallback")
	}
}

func TestCalendarColumns_Extract_NoOrganizer(t *testing.T) {
	row := map[string]any{
		"subject": "Event",
	}
	organizer := CalendarColumns[2].Extract(row)
	if organizer != "" {
		t.Errorf("expected empty organizer, got %q", organizer)
	}
}

func TestPlannerPlanColumns_Extract(t *testing.T) {
	row := map[string]any{
		"title":           "Q1 Sprint Plan",
		"id":              "plan-abc-123",
		"createdDateTime": "2025-01-15T10:30:00Z",
	}

	// TITLE
	title := PlannerPlanColumns[0].Extract(row)
	if title != "Q1 Sprint Plan" {
		t.Errorf("expected 'Q1 Sprint Plan', got %q", title)
	}

	// ID
	id := PlannerPlanColumns[1].Extract(row)
	if id != "plan-abc-123" {
		t.Errorf("expected 'plan-abc-123', got %q", id)
	}

	// CREATED
	created := PlannerPlanColumns[2].Extract(row)
	if created == "" {
		t.Error("expected non-empty created time")
	}
}

func TestPlannerTaskColumns_Extract(t *testing.T) {
	tests := []struct {
		name            string
		percentComplete float64
		expectedStatus  string
	}{
		{"completed", 100, "completed"},
		{"in progress", 50, "in progress"},
		{"not started", 0, "not started"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := map[string]any{
				"title":           "Fix bug",
				"id":              "task-abc-123",
				"percentComplete": tt.percentComplete,
				"priority":        5.0,
			}

			// TITLE
			title := PlannerTaskColumns[0].Extract(row)
			if title != "Fix bug" {
				t.Errorf("expected 'Fix bug', got %q", title)
			}

			// ID
			id := PlannerTaskColumns[1].Extract(row)
			if id != "task-abc-123" {
				t.Errorf("expected 'task-abc-123', got %q", id)
			}

			// STATUS — maps percentComplete
			status := PlannerTaskColumns[2].Extract(row)
			if status != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, status)
			}
		})
	}
}

func TestPlannerTaskColumns_Priority(t *testing.T) {
	tests := []struct {
		name     string
		priority float64
		expected string
	}{
		{"urgent", 1, "urgent"},
		{"important", 3, "important"},
		{"medium", 5, "medium"},
		{"low", 9, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := map[string]any{
				"priority": tt.priority,
			}
			result := PlannerTaskColumns[3].Extract(row)
			if result != tt.expected {
				t.Errorf("expected priority %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPlannerTaskColumns_NilValues(t *testing.T) {
	row := map[string]any{
		"title": "Task",
		"id":    "t1",
	}
	status := PlannerTaskColumns[2].Extract(row)
	if status != "" {
		t.Errorf("expected empty status for nil percentComplete, got %q", status)
	}
	priority := PlannerTaskColumns[3].Extract(row)
	if priority != "" {
		t.Errorf("expected empty priority for nil, got %q", priority)
	}
}

func TestUserColumns_Extract(t *testing.T) {
	row := map[string]any{
		"displayName":       "Jane Doe",
		"userPrincipalName": "jane@example.com",
		"jobTitle":          "Software Engineer",
		"id":                "user-abc-123",
	}

	// DISPLAY NAME
	name := UserColumns[0].Extract(row)
	if name != "Jane Doe" {
		t.Errorf("expected 'Jane Doe', got %q", name)
	}

	// UPN
	upn := UserColumns[1].Extract(row)
	if upn != "jane@example.com" {
		t.Errorf("expected 'jane@example.com', got %q", upn)
	}

	// JOB TITLE
	jobTitle := UserColumns[2].Extract(row)
	if jobTitle != "Software Engineer" {
		t.Errorf("expected 'Software Engineer', got %q", jobTitle)
	}

	// ID
	id := UserColumns[3].Extract(row)
	if id != "user-abc-123" {
		t.Errorf("expected 'user-abc-123', got %q", id)
	}
}

func TestUserColumns_Extract_FallbackMail(t *testing.T) {
	row := map[string]any{
		"displayName":       "Jane Doe",
		"userPrincipalName": "",
		"mail":              "jane@company.com",
		"id":                "user-abc-123",
	}

	upn := UserColumns[1].Extract(row)
	if upn != "jane@company.com" {
		t.Errorf("expected 'jane@company.com' as mail fallback, got %q", upn)
	}
}

func TestMembersColumns_WithRoles(t *testing.T) {
	row := map[string]any{
		"displayName": "Alice",
		"email":       "alice@example.com",
		"id":          "member-123",
		"roles":       []any{"owner", "member"},
	}

	// DISPLAY NAME
	name := MembersColumns[0].Extract(row)
	if name != "Alice" {
		t.Errorf("expected 'Alice', got %q", name)
	}

	// EMAIL
	email := MembersColumns[1].Extract(row)
	if email != "alice@example.com" {
		t.Errorf("expected 'alice@example.com', got %q", email)
	}

	// ID
	id := MembersColumns[2].Extract(row)
	if id != "member-123" {
		t.Errorf("expected 'member-123', got %q", id)
	}

	// ROLES — should join array
	roles := MembersColumns[3].Extract(row)
	if roles != "owner,member" {
		t.Errorf("expected 'owner,member', got %q", roles)
	}
}

func TestMembersColumns_NoRoles(t *testing.T) {
	row := map[string]any{
		"displayName": "Bob",
		"email":       "bob@example.com",
		"id":          "member-456",
	}
	roles := MembersColumns[3].Extract(row)
	if roles != "" {
		t.Errorf("expected empty roles, got %q", roles)
	}
}

func TestChatsColumns_NoMembers(t *testing.T) {
	row := map[string]any{
		"topic":    "",
		"id":       "19:abc@thread.v2",
		"chatType": "OneOnOne",
	}

	result := ChatsColumns[0].Extract(row) // TOPIC column
	if result != "(no topic)" {
		t.Errorf("expected '(no topic)' when no members, got %q", result)
	}
}

func TestChatsColumns_EmptyMembersArray(t *testing.T) {
	row := map[string]any{
		"topic":    "",
		"id":       "19:abc@thread.v2",
		"chatType": "OneOnOne",
		"members":  []any{},
	}

	result := ChatsColumns[0].Extract(row) // TOPIC column
	if result != "(no topic)" {
		t.Errorf("expected '(no topic)' for empty members array, got %q", result)
	}
}

func TestStripHTML_Emoji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"single emoji",
			`<emoji id="smile" alt="😊" title="Smile"></emoji>`,
			"😊",
		},
		{
			"emoji in text",
			`<p>Hello <emoji id="smile" alt="😊" title="Smile"></emoji> World</p>`,
			"Hello 😊 World",
		},
		{
			"multiple emojis",
			`<emoji id="smile" alt="😊" title="Smile"></emoji><emoji id="thumbsup" alt="👍" title="Thumbs Up"></emoji>`,
			"😊👍",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTML(tt.input)
			if result != tt.expected {
				t.Errorf("stripHTML(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripHTML_Attachment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"simple attachment",
			`<attachment id="abc123"></attachment>`,
			"[attachment]",
		},
		{
			"attachment in message",
			`<p>Check this out</p><attachment id="abc123"></attachment>`,
			"Check this out [attachment]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTML(tt.input)
			if result != tt.expected {
				t.Errorf("stripHTML(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripHTML_SystemEvent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"system event self-closing",
			`<systemEventMessage/>`,
			"",
		},
		{
			"system event with attributes",
			`<systemEventMessage type="memberAdded"/>`,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTML(tt.input)
			if result != tt.expected {
				t.Errorf("stripHTML(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripHTML_ImgAlt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"img with alt",
			`<img src="http://example.com/img.png" alt="screenshot">`,
			"[screenshot]",
		},
		{
			"img in message",
			`<p>See this: <img src="http://example.com/img.png" alt="diagram"></p>`,
			"See this: [diagram]",
		},
		{
			"img with empty alt",
			`<img src="http://example.com/img.png" alt="">`,
			"[]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTML(tt.input)
			if result != tt.expected {
				t.Errorf("stripHTML(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripHTML_CodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"codeblock with language",
			`<codeblock class="Bash"><code>echo hello</code></codeblock>`,
			"echo hello",
		},
		{
			"codeblock in message",
			`<p>Run this:</p><codeblock class="Python"><code>print("hi")</code></codeblock>`,
			`Run this: print("hi")`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTML(tt.input)
			if result != tt.expected {
				t.Errorf("stripHTML(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripHTML_Nbsp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"nbsp entity",
			`<p>Hello&nbsp;World</p>`,
			"Hello World",
		},
		{
			"multiple nbsp",
			`<p>A&nbsp;&nbsp;B</p>`,
			"A B",
		},
		{
			"nbsp with other text",
			`Hello&nbsp;World`,
			"Hello World",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTML(tt.input)
			if result != tt.expected {
				t.Errorf("stripHTML(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
