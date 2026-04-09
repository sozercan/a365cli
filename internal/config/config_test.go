package config

import "testing"

func TestBaseURL_Default(t *testing.T) {
	t.Setenv("A365_ENDPOINT", "")

	got := BaseURL()
	if got != DefaultBaseURL {
		t.Errorf("BaseURL() = %q, want %q", got, DefaultBaseURL)
	}
}

func TestBaseURL_Override(t *testing.T) {
	t.Setenv("A365_ENDPOINT", "https://custom.example.com/")

	got := BaseURL()
	want := "https://custom.example.com/"
	if got != want {
		t.Errorf("BaseURL() = %q, want %q", got, want)
	}
}

func TestEndpoint(t *testing.T) {
	t.Setenv("A365_ENDPOINT", "")

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

func TestValidateEndpointURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "empty", raw: "", wantErr: false},
		{name: "https", raw: "https://example.com/agents/servers/", wantErr: false},
		{name: "localhost http", raw: "http://127.0.0.1:8080/", wantErr: false},
		{name: "non-loopback http", raw: "http://example.com/agents/servers/", wantErr: true},
		{name: "relative", raw: "/agents/servers/", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEndpointURL(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateEndpointURL(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
		})
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
