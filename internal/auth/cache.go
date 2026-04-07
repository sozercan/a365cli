package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/sozercan/a365cli/internal/config"
)

// authRecordPath returns the path to the auth record JSON file.
func authRecordPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, config.AuthRecordDir, config.AuthRecordFile), nil
}

// SaveAuthRecord persists the authentication record to disk.
func SaveAuthRecord(record *azidentity.AuthenticationRecord) error {
	path, err := authRecordPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create auth dir: %w", err)
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal auth record: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write auth record: %w", err)
	}

	return nil
}

// LoadAuthRecord reads the cached authentication record from disk.
func LoadAuthRecord() (*azidentity.AuthenticationRecord, error) {
	path, err := authRecordPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var record azidentity.AuthenticationRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("unmarshal auth record: %w", err)
	}

	return &record, nil
}

// RemoveAuthRecord deletes the cached auth record and keychain entries.
func RemoveAuthRecord() error {
	path, err := authRecordPath()
	if err != nil {
		return err
	}

	// Remove auth record file
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove auth record: %w", err)
	}

	return nil
}

// HasCachedAuth checks if there's a cached auth record on disk.
func HasCachedAuth() bool {
	path, err := authRecordPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// GetCachedUsername returns the username from the cached auth record, if any.
func GetCachedUsername() string {
	record, err := LoadAuthRecord()
	if err != nil {
		return ""
	}
	return record.Username
}
