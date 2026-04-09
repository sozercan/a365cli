package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

const (
	// DefaultBaseURL is the agent365 MCP gateway base URL.
	DefaultBaseURL = "https://agent365.svc.cloud.microsoft/agents/servers/"

	// DefaultAudience is the Entra ID audience for agent365.
	DefaultAudience = "ea9ffc3e-8a23-4a7d-836d-234d7c7565c1"

	// DefaultScope requests all granted scopes.
	DefaultScope = DefaultAudience + "/.default"

	// DefaultAuthority is the multi-tenant login authority.
	DefaultAuthority = "https://login.microsoftonline.com/organizations/v2.0"

	// AuthRecordDir is the directory for cached auth record.
	AuthRecordDir = ".a365"

	// AuthRecordFile is the filename for cached auth record.
	AuthRecordFile = "auth-record.json"

	// DefaultClientID is the default Entra app client ID (VS Code MCP extension).
	DefaultClientID = "aebc6443-996d-45c2-90f0-388ff96faa56"
)

// Servers maps friendly names to agent365 MCP server names.
var Servers = map[string]string{
	"teams":      "mcp_TeamsServer",
	"mail":       "mcp_MailTools",
	"calendar":   "mcp_CalendarTools",
	"planner":    "mcp_PlannerServer",
	"sharepoint": "mcp_ODSPRemoteServer",
	"word":       "mcp_WordServer",
	"excel":      "mcp_ExcelServer",
	"powerpoint": "mcp_PowerPointServer",
	"onedrive":   "mcp_OneDriveServer",
	"copilot":    "mcp_M365Copilot",
	"me":         "mcp_MeServer",
	"files":      "mcp_FilesServer",
	"knowledge":  "mcp_KnowledgeTools",
	"sp-lists":   "mcp_SharePointListsTools",
	"dataverse":  "mcp_DataverseServer",
	"admin":      "mcp_Admin365_GraphTools",
	"nlweb":      "mcp_NLWeb",
	// Discovered via discoverToolServers
	"websearch":       "mcp_WebSearchTools",
	"w365":            "mcp_W365ComputerUse",
	"dasearch":        "mcp_DASearch",
	"tasks":           "mcp_TaskPersonalizationServer",
	"admin365":        "mcp_AdminTools",
	"onedrive-remote": "mcp_OneDriveRemoteServer",
	"sp-remote":       "mcp_SharePointRemoteServer",
}

// BaseURL returns the agent365 base URL, allowing override via A365_ENDPOINT env var.
// Always ensures the URL ends with a trailing slash.
func BaseURL() string {
	base := DefaultBaseURL
	if v := os.Getenv("A365_ENDPOINT"); v != "" {
		base = v
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	return base
}

// ValidateEndpointURL rejects malformed endpoints and non-loopback plaintext HTTP.
func ValidateEndpointURL(raw string) error {
	if raw == "" {
		return nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("parse endpoint URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("endpoint must be an absolute URL")
	}
	if u.Scheme == "https" {
		return nil
	}
	if u.Scheme != "http" {
		return fmt.Errorf("endpoint must use https or loopback http")
	}
	if !isLoopbackHost(u.Hostname()) {
		return fmt.Errorf("non-loopback endpoints must use https")
	}
	return nil
}

func isLoopbackHost(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// Endpoint returns the full URL for a given service name.
func Endpoint(service string) string {
	server, ok := Servers[service]
	if !ok {
		return ""
	}
	return BaseURL() + server + "/"
}

// Authority returns the Entra ID authority URL. If tenantID is set, uses
// the tenant-specific authority; otherwise uses the "organizations" authority.
func Authority(tenantID string) string {
	if tenantID != "" {
		return "https://login.microsoftonline.com/" + tenantID + "/v2.0"
	}
	return DefaultAuthority
}
