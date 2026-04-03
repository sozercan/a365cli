package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// FileConfig holds optional defaults loaded from ~/.a365/config.json.
type FileConfig struct {
	ClientID string `json:"client_id,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
	Output   string `json:"output,omitempty"`   // "json", "plain", or "" (human default)
	Endpoint string `json:"endpoint,omitempty"` // override base URL
}

// ConfigPath returns the path to the config file (~/.a365/config.json).
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, AuthRecordDir, "config.json")
}

// LoadFileConfig reads ~/.a365/config.json and returns the parsed config.
// Returns an empty config on any error (missing file, bad JSON, etc.).
func LoadFileConfig() *FileConfig {
	cfg := &FileConfig{}
	p := ConfigPath()
	if p == "" {
		return cfg
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, cfg)
	return cfg
}

// SaveFileConfig writes the config to ~/.a365/config.json, creating the
// directory if needed.
func SaveFileConfig(cfg *FileConfig) error {
	p := ConfigPath()
	if p == "" {
		return os.ErrNotExist
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(p, data, 0o600)
}
