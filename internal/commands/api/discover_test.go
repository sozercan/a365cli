package api

import "testing"

func TestDiscoverURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
		wantErr bool
	}{
		{
			name:    "servers suffix",
			baseURL: "https://example.com/agents/servers/",
			want:    "https://example.com/agents/discoverToolServers",
		},
		{
			name:    "agents suffix",
			baseURL: "https://example.com/agents/",
			want:    "https://example.com/agents/discoverToolServers",
		},
		{
			name:    "servers no trailing slash",
			baseURL: "https://example.com/agents/servers",
			want:    "https://example.com/agents/discoverToolServers",
		},
		{
			name:    "agents no trailing slash",
			baseURL: "https://example.com/agents",
			want:    "https://example.com/agents/discoverToolServers",
		},
		{
			name:    "invalid server endpoint",
			baseURL: "https://example.com/agents/servers/mcp_TeamsServer/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := discoverURL(tt.baseURL)
			if (err != nil) != tt.wantErr {
				t.Fatalf("discoverURL(%q) error = %v, wantErr %v", tt.baseURL, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("discoverURL(%q) = %q, want %q", tt.baseURL, got, tt.want)
			}
		})
	}
}
