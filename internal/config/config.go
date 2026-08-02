package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
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

	// DefaultMCPResponseHeaderTimeout is the default HTTP response-header timeout for MCP requests.
	DefaultMCPResponseHeaderTimeout = 60 * time.Second

	// DefaultCopilotResponseHeaderTimeout is longer because Copilot requests can take longer to start streaming.
	DefaultCopilotResponseHeaderTimeout = 5 * time.Minute

	// ResolvedTenantEnv is an internal process-local override for the concrete
	// tenant used in first-party MCP endpoint paths. It is distinct from
	// A365_TENANT_ID, which may contain an authority alias such as organizations.
	ResolvedTenantEnv = "A365_RESOLVED_TENANT_ID"
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

// MCPResponseHeaderTimeout returns the HTTP response-header timeout for MCP requests.
//
// A365_MCP_RESPONSE_HEADER_TIMEOUT overrides the default for all services.
// A365_COPILOT_RESPONSE_HEADER_TIMEOUT overrides the Copilot service specifically.
func MCPResponseHeaderTimeout(service string) time.Duration {
	timeout := DefaultMCPResponseHeaderTimeout
	if strings.EqualFold(service, "copilot") {
		timeout = DefaultCopilotResponseHeaderTimeout
	}

	if v := strings.TrimSpace(os.Getenv("A365_MCP_RESPONSE_HEADER_TIMEOUT")); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d >= 0 {
			timeout = d
		}
	}

	if strings.EqualFold(service, "copilot") {
		if v := strings.TrimSpace(os.Getenv("A365_COPILOT_RESPONSE_HEADER_TIMEOUT")); v != "" {
			if d, err := time.ParseDuration(v); err == nil && d >= 0 {
				timeout = d
			}
		}
	}

	return timeout
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

	// Explicit endpoint overrides retain their historical base-URL behavior for
	// local mocks and private gateways.
	if os.Getenv("A365_ENDPOINT") != "" {
		return BaseURL() + server + "/"
	}

	// Agent 365 first-party MCP endpoints are tenant scoped. The tenant is
	// resolved from explicit configuration or the cached authentication record
	// by main before command execution.
	tenantID := strings.TrimSpace(os.Getenv(ResolvedTenantEnv))
	if tenantID == "" {
		tenantID = ResolveEndpointTenantID(os.Getenv("A365_TENANT_ID"), "")
	}
	if tenantID != "" {
		return "https://agent365.svc.cloud.microsoft/agents/tenants/" + url.PathEscape(tenantID) + "/servers/" + server
	}

	// Keep the legacy endpoint as a compatibility fallback for callers that use
	// config.Endpoint outside the authenticated CLI startup path.
	return BaseURL() + server + "/"
}

// ResolveEndpointTenantID chooses a concrete tenant for MCP endpoint routing.
// Authority aliases are valid for authentication but not for tenant-scoped
// Agent 365 endpoint paths, so they fall back to the cached concrete tenant.
func ResolveEndpointTenantID(configured, cached string) string {
	configured = strings.TrimSpace(configured)
	if tenantID, err := uuid.Parse(configured); err == nil {
		return tenantID.String()
	}

	// Tenant-agnostic authority aliases may safely use the concrete directory ID
	// selected during authentication. A configured verified domain must not be
	// paired with a cached ID because the record may belong to another tenant.
	switch strings.ToLower(configured) {
	case "", "common", "organizations", "consumers":
		if tenantID, err := uuid.Parse(strings.TrimSpace(cached)); err == nil {
			return tenantID.String()
		}
	}
	return ""
}

// Authority returns the Entra ID authority URL. If tenantID is set, uses
// the tenant-specific authority; otherwise uses the "organizations" authority.
func Authority(tenantID string) string {
	if tenantID != "" {
		return "https://login.microsoftonline.com/" + tenantID + "/v2.0"
	}
	return DefaultAuthority
}
