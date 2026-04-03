package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileConfig_Missing(t *testing.T) {
	// Point HOME at an empty temp dir so no config file exists.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := LoadFileConfig()
	if cfg == nil {
		t.Fatal("LoadFileConfig returned nil")
	}
	if cfg.ClientID != "" || cfg.TenantID != "" || cfg.Output != "" || cfg.Endpoint != "" {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func TestLoadFileConfig_Valid(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, AuthRecordDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(FileConfig{
		ClientID: "test-client",
		TenantID: "test-tenant",
		Output:   "json",
		Endpoint: "https://custom.example.com/",
	})
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := LoadFileConfig()
	if cfg.ClientID != "test-client" {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, "test-client")
	}
	if cfg.TenantID != "test-tenant" {
		t.Errorf("TenantID = %q, want %q", cfg.TenantID, "test-tenant")
	}
	if cfg.Output != "json" {
		t.Errorf("Output = %q, want %q", cfg.Output, "json")
	}
	if cfg.Endpoint != "https://custom.example.com/" {
		t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, "https://custom.example.com/")
	}
}

func TestLoadFileConfig_BadJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, AuthRecordDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{bad json}"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := LoadFileConfig()
	if cfg == nil {
		t.Fatal("LoadFileConfig returned nil on bad JSON")
	}
	// Should return empty config, not crash
	if cfg.ClientID != "" {
		t.Errorf("expected empty ClientID on bad JSON, got %q", cfg.ClientID)
	}
}

func TestSaveFileConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := &FileConfig{
		ClientID: "saved-client",
		Output:   "plain",
	}
	if err := SaveFileConfig(cfg); err != nil {
		t.Fatalf("SaveFileConfig: %v", err)
	}

	// Read it back
	loaded := LoadFileConfig()
	if loaded.ClientID != "saved-client" {
		t.Errorf("ClientID = %q, want %q", loaded.ClientID, "saved-client")
	}
	if loaded.Output != "plain" {
		t.Errorf("Output = %q, want %q", loaded.Output, "plain")
	}
	if loaded.TenantID != "" {
		t.Errorf("TenantID = %q, want empty", loaded.TenantID)
	}
}

func TestSaveFileConfig_CreatesDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Directory should not exist yet
	dir := filepath.Join(tmp, AuthRecordDir)
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("expected directory to not exist")
	}

	cfg := &FileConfig{TenantID: "my-tenant"}
	if err := SaveFileConfig(cfg); err != nil {
		t.Fatalf("SaveFileConfig: %v", err)
	}

	// Directory should exist now
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected a directory")
	}
}

func TestSaveFileConfig_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	original := &FileConfig{
		ClientID: "rt-client",
		TenantID: "rt-tenant",
		Output:   "json",
		Endpoint: "https://rt.example.com/",
	}
	if err := SaveFileConfig(original); err != nil {
		t.Fatalf("SaveFileConfig: %v", err)
	}

	loaded := LoadFileConfig()
	if loaded.ClientID != original.ClientID {
		t.Errorf("ClientID mismatch: %q vs %q", loaded.ClientID, original.ClientID)
	}
	if loaded.TenantID != original.TenantID {
		t.Errorf("TenantID mismatch: %q vs %q", loaded.TenantID, original.TenantID)
	}
	if loaded.Output != original.Output {
		t.Errorf("Output mismatch: %q vs %q", loaded.Output, original.Output)
	}
	if loaded.Endpoint != original.Endpoint {
		t.Errorf("Endpoint mismatch: %q vs %q", loaded.Endpoint, original.Endpoint)
	}
}

func TestConfigPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	p := ConfigPath()
	want := filepath.Join(tmp, AuthRecordDir, "config.json")
	if p != want {
		t.Errorf("ConfigPath() = %q, want %q", p, want)
	}
}
