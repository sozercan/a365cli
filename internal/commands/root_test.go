package commands

import (
	"context"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/output"
)

func TestConfirm_Force(t *testing.T) {
	ctx := &Context{
		Ctx:   context.Background(),
		Force: true,
	}
	if err := ctx.Confirm("delete something"); err != nil {
		t.Fatalf("Confirm() with Force=true should return nil, got: %v", err)
	}
}

func TestConfirm_NoInput(t *testing.T) {
	ctx := &Context{
		Ctx:     context.Background(),
		NoInput: true,
	}
	err := ctx.Confirm("delete something")
	if err == nil {
		t.Fatal("Confirm() with NoInput=true should return error")
	}
	if got := err.Error(); got != "refusing to delete something without --force (non-interactive)" {
		t.Errorf("unexpected error message: %q", got)
	}
}

func TestNewMCPClient(t *testing.T) {
	tp := func(ctx context.Context) (string, error) {
		return "test-token", nil
	}
	ctx := &Context{
		Ctx:           context.Background(),
		TokenProvider: mcp.TokenProvider(tp),
		Output:        &output.Formatter{Format: output.FormatJSON},
	}
	client := ctx.NewMCPClient("https://example.com/mcp/")
	if client == nil {
		t.Fatal("NewMCPClient() returned nil")
	}
}

func TestNewMCPClient_Verbose(t *testing.T) {
	tp := func(ctx context.Context) (string, error) {
		return "test-token", nil
	}
	ctx := &Context{
		Ctx:           context.Background(),
		TokenProvider: mcp.TokenProvider(tp),
		Output:        &output.Formatter{Format: output.FormatJSON},
		Verbose:       true,
	}
	client := ctx.NewMCPClient("https://example.com/mcp/")
	if client == nil {
		t.Fatal("NewMCPClient() returned nil")
	}
	// The client should have a verbose logger set.
	// We can't directly inspect the private field, but we verify it
	// doesn't panic or error when created with Verbose=true.
}
