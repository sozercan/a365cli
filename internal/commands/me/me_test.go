package me

import (
	"encoding/json"
	"testing"

	"github.com/sozercan/a365cli/internal/testutil"
)

func TestMeWhoamiCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"GetMyDetails": `{"displayName":"Alice","mail":"alice@contoso.com","jobTitle":"Engineer","id":"00000000-0000-0000-0000-000000000001"}`,
	})

	cmd := &MeWhoamiCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["displayName"] != "Alice" {
		t.Errorf("expected displayName=Alice, got %v", result["displayName"])
	}
	if result["mail"] != "alice@contoso.com" {
		t.Errorf("expected mail=alice@contoso.com, got %v", result["mail"])
	}
	if result["id"] != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("expected id=00000000-0000-0000-0000-000000000001, got %v", result["id"])
	}
}

func TestMeGetCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"GetUserDetails": `{"displayName":"Bob","mail":"bob@contoso.com","jobTitle":"Manager","id":"00000000-0000-0000-0000-000000000002"}`,
	})

	cmd := &MeGetCmd{User: "bob@contoso.com"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["displayName"] != "Bob" {
		t.Errorf("expected displayName=Bob, got %v", result["displayName"])
	}
	if result["mail"] != "bob@contoso.com" {
		t.Errorf("expected mail=bob@contoso.com, got %v", result["mail"])
	}
}

func TestMeSearchCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"GetMultipleUsersDetails": `{"users":[{"displayName":"Alice","mail":"alice@contoso.com","id":"00000000-0000-0000-0000-000000000001"},{"displayName":"Bob","mail":"bob@contoso.com","id":"00000000-0000-0000-0000-000000000002"}]}`,
	})

	cmd := &MeSearchCmd{Query: []string{"Alice", "Bob"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	users, ok := result["users"]
	if !ok {
		t.Fatalf("expected 'users' key in output, got: %s", buf.String())
	}
	arr, ok := users.([]any)
	if !ok {
		t.Fatalf("expected 'users' to be an array, got: %T", users)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 users, got %d", len(arr))
	}
}
