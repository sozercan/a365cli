package main

import (
	"context"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/commands"
)

func TestValidateServiceTenant(t *testing.T) {
	tests := []struct {
		name             string
		endpointOverride string
		endpointTenantID string
		wantErr          bool
	}{
		{name: "service requires concrete tenant", wantErr: true},
		{name: "service accepts concrete tenant", endpointTenantID: "00000000-0000-0000-0000-000000000000"},
		{name: "service accepts explicit endpoint", endpointOverride: "http://127.0.0.1:8080/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServiceTenant(tt.endpointOverride, tt.endpointTenantID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateServiceTenant() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrepareAuthenticatedCommandReportsLoginBeforeTenant(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	ctx := &commands.Context{
		Ctx:     context.Background(),
		NoInput: true,
	}
	err := prepareAuthenticatedCommand(ctx, true, "", "")
	if err == nil {
		t.Fatal("expected authentication error")
	}
	if !strings.Contains(err.Error(), "a365 auth login") {
		t.Fatalf("error = %q, want login guidance", err)
	}
	if strings.Contains(err.Error(), "Directory tenant ID") {
		t.Fatalf("authentication error was masked by tenant validation: %q", err)
	}
}

func TestPrepareAuthenticatedCommandSkipsNonServiceCommands(t *testing.T) {
	if err := prepareAuthenticatedCommand(&commands.Context{}, false, "", ""); err != nil {
		t.Fatalf("non-service command error: %v", err)
	}
}
