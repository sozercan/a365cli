package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sozercan/a365cli/internal/mcp"
)

// Format represents the output format.
type Format int

const (
	// FormatHuman is the default human-friendly table output.
	FormatHuman Format = iota
	// FormatJSON outputs clean JSON envelopes.
	FormatJSON
	// FormatPlain outputs tab-separated values for scripting.
	FormatPlain
)

// Formatter handles output rendering in one of three modes.
type Formatter struct {
	Format Format
	Writer io.Writer // defaults to os.Stdout
}

// NewFormatter creates a new output formatter from a format string.
// "json" → FormatJSON, "tsv" or "plain" → FormatPlain, anything else → FormatHuman.
func NewFormatter(format string) *Formatter {
	f := FormatHuman
	switch format {
	case "json":
		f = FormatJSON
	case "tsv", "plain":
		f = FormatPlain
	}
	return &Formatter{Format: f, Writer: os.Stdout}
}

// PrintList outputs a list of items with the given entity name and column definitions.
func (f *Formatter) PrintList(entity string, columns []Column, rows []map[string]any) error {
	switch f.Format {
	case FormatJSON:
		return f.writeJSON(map[string]any{entity: rows})
	case FormatPlain:
		RenderTSV(f.Writer, columns, rows)
		return nil
	default:
		RenderTable(f.Writer, columns, rows)
		return nil
	}
}

// PrintItem outputs a single item.
func (f *Formatter) PrintItem(item map[string]any) error {
	switch f.Format {
	case FormatJSON:
		return f.writeJSON(item)
	default:
		RenderKeyValue(f.Writer, item)
		return nil
	}
}

// PrintMutation outputs the result of a write operation.
func (f *Formatter) PrintMutation(msg string, data map[string]any) error {
	switch f.Format {
	case FormatJSON:
		return f.writeJSON(data)
	default:
		fmt.Fprintln(f.Writer, msg)
		return nil
	}
}

// PrintDryRun outputs what WOULD happen without executing.
func (f *Formatter) PrintDryRun(action string, data map[string]any) error {
	switch f.Format {
	case FormatJSON:
		data["dry_run"] = true
		return f.writeJSON(data)
	default:
		fmt.Fprintf(f.Writer, "Dry run: would %s\n", action)
		return nil
	}
}

// PrintRaw outputs a raw MCP response (fallback for unstructured content).
// Used when we can't parse the response into typed data.
func (f *Formatter) PrintRaw(resp *mcp.JSONRPCResponse) error {
	if resp.Error != nil {
		return fmt.Errorf("MCP error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		fmt.Fprintln(f.Writer, "(no result)")
		return nil
	}

	// Try to extract and print clean content
	data, err := ExtractContent(resp)
	if err != nil {
		return err
	}

	return f.PrintItem(data)
}

// writeJSON writes pretty-printed JSON to the writer.
func (f *Formatter) writeJSON(v any) error {
	enc := json.NewEncoder(f.Writer)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// PrintError outputs an error message to stderr.
func PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}
