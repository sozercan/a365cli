package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/output"
)

func TestValidateDryRunUsesExplicitMCPArgsAndSafeDisplay(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "SendMessage",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"chatId":  map[string]any{"type": "string"},
					"content": map[string]any{"type": "string"},
				},
				"required": []any{"chatId", "content"},
			},
		},
	}
	server := setupMockMCPServer(t, schemas)

	var buf bytes.Buffer
	ctx := &Context{
		Ctx:           context.Background(),
		TokenProvider: func(context.Context) (string, error) { return "test-token", nil },
		Output:        &output.Formatter{Format: output.FormatJSON, Writer: &buf},
		DryRun:        true,
	}

	displayData := map[string]any{
		"action":        "chats.send",
		"chatId":        "chat-001",
		"contentLength": len("hello"),
	}
	mcpArgs := map[string]any{
		"chatId":  "chat-001",
		"content": "hello",
	}
	if err := ctx.ValidateDryRun(server.URL+"/", "SendMessage", "send message", displayData, mcpArgs); err != nil {
		t.Fatalf("ValidateDryRun() error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("decode dry-run output: %v", err)
	}
	if got["dry_run"] != true {
		t.Fatalf("dry_run = %v, want true", got["dry_run"])
	}
	if got["action"] != "chats.send" {
		t.Fatalf("action = %v, want chats.send", got["action"])
	}
	if _, ok := got["content"]; ok {
		t.Fatal("raw MCP content leaked into dry-run display")
	}
	validation, ok := got["validation"].(map[string]any)
	if !ok || validation["valid"] != true {
		t.Fatalf("validation = %#v, want valid=true", got["validation"])
	}
}

func TestProductionDryRunsPassExplicitMCPArgs(t *testing.T) {
	fset := token.NewFileSet()
	var missing []string

	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "ValidateDryRun" {
				return true
			}
			if len(call.Args) < 5 {
				pos := fset.Position(call.Pos())
				missing = append(missing, pos.String())
				return true
			}
			if ident, ok := call.Args[4].(*ast.Ident); ok && ident.Name == "nil" {
				pos := fset.Position(call.Pos())
				missing = append(missing, pos.String()+" (nil MCP args)")
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("scan command sources: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("ValidateDryRun calls must pass explicit MCP args:\n%s", strings.Join(missing, "\n"))
	}
}
