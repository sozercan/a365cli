package config

import (
	"os"
	"testing"
)

func TestBaseURL_Default(t *testing.T) {
	orig := os.Getenv("A365_ENDPOINT")
	os.Unsetenv("A365_ENDPOINT")
	defer func() {
		if orig != "" {
			os.Setenv("A365_ENDPOINT", orig)
		}
	}()

	got := BaseURL()
	if got != DefaultBaseURL {
		t.Errorf("BaseURL() = %q, want %q", got, DefaultBaseURL)
	}
}

func TestBaseURL_Override(t *testing.T) {
	orig := os.Getenv("A365_ENDPOINT")
	os.Setenv("A365_ENDPOINT", "https://custom.example.com/")
	defer func() {
		if orig == "" {
			os.Unsetenv("A365_ENDPOINT")
		} else {
			os.Setenv("A365_ENDPOINT", orig)
		}
	}()

	got := BaseURL()
	want := "https://custom.example.com/"
	if got != want {
		t.Errorf("BaseURL() = %q, want %q", got, want)
	}
}

func TestEndpoint(t *testing.T) {
	orig := os.Getenv("A365_ENDPOINT")
	os.Unsetenv("A365_ENDPOINT")
	defer func() {
		if orig != "" {
			os.Setenv("A365_ENDPOINT", orig)
		}
	}()

	tests := []struct {
		service string
		want    string
	}{
		{"teams", DefaultBaseURL + "mcp_TeamsServer/"},
		{"mail", DefaultBaseURL + "mcp_MailTools/"},
		{"calendar", DefaultBaseURL + "mcp_CalendarTools/"},
		{"planner", DefaultBaseURL + "mcp_PlannerServer/"},
		{"sharepoint", DefaultBaseURL + "mcp_ODSPRemoteServer/"},
		{"onedrive", DefaultBaseURL + "mcp_OneDriveServer/"},
		{"copilot", DefaultBaseURL + "mcp_M365Copilot/"},
	}
	for _, tt := range tests {
		t.Run(tt.service, func(t *testing.T) {
			got := Endpoint(tt.service)
			if got != tt.want {
				t.Errorf("Endpoint(%q) = %q, want %q", tt.service, got, tt.want)
			}
		})
	}
}

func TestEndpoint_Unknown(t *testing.T) {
	got := Endpoint("nonexistent-service")
	if got != "" {
		t.Errorf("Endpoint('nonexistent-service') = %q, want empty string", got)
	}
}

func TestAuthority_Default(t *testing.T) {
	got := Authority("")
	if got != DefaultAuthority {
		t.Errorf("Authority('') = %q, want %q", got, DefaultAuthority)
	}
}

func TestAuthority_WithTenant(t *testing.T) {
	tenantID := "contoso.onmicrosoft.com"
	got := Authority(tenantID)
	want := "https://login.microsoftonline.com/contoso.onmicrosoft.com/v2.0"
	if got != want {
		t.Errorf("Authority(%q) = %q, want %q", tenantID, got, want)
	}
}

func TestServers_HasExpectedKeys(t *testing.T) {
	expected := []string{
		"teams", "mail", "calendar", "planner", "sharepoint",
		"word", "excel", "powerpoint", "onedrive", "copilot",
		"me", "files", "knowledge", "sp-lists", "dataverse",
		"admin", "nlweb",
	}
	for _, key := range expected {
		if _, ok := Servers[key]; !ok {
			t.Errorf("Servers map missing expected key %q", key)
		}
	}
}
