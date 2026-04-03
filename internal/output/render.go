package output

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

// RenderTable writes rows as an aligned table with headers using tabwriter.
func RenderTable(w io.Writer, columns []Column, rows []map[string]any) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	defer tw.Flush()

	// Header row
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
	}
	fmt.Fprintln(tw, strings.Join(headers, "\t"))

	// Data rows
	for _, row := range rows {
		vals := make([]string, len(columns))
		for i, col := range columns {
			v := col.Extract(row)
			if col.Width > 0 {
				v = truncate(v, col.Width)
			}
			vals[i] = v
		}
		fmt.Fprintln(tw, strings.Join(vals, "\t"))
	}
}

// RenderTSV writes rows as tab-separated values with a header row. No alignment.
func RenderTSV(w io.Writer, columns []Column, rows []map[string]any) {
	// Header row
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Data rows
	for _, row := range rows {
		vals := make([]string, len(columns))
		for i, col := range columns {
			vals[i] = col.Extract(row)
		}
		fmt.Fprintln(w, strings.Join(vals, "\t"))
	}
}

// RenderKeyValue writes an item as sorted key: value pairs.
func RenderKeyValue(w io.Writer, item map[string]any) {
	// Collect and sort keys
	keys := make([]string, 0, len(item))
	for k := range item {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Find max key length for alignment
	maxLen := 0
	for _, k := range keys {
		if len(k) > maxLen {
			maxLen = len(k)
		}
	}

	for _, k := range keys {
		v := item[k]
		// Format value
		var s string
		switch val := v.(type) {
		case string:
			s = val
		case nil:
			s = ""
		case map[string]any:
			// Flatten nested objects to key=val pairs
			parts := make([]string, 0)
			for nk, nv := range val {
				parts = append(parts, fmt.Sprintf("%s=%v", nk, nv))
			}
			sort.Strings(parts)
			s = strings.Join(parts, ", ")
		case []any:
			parts := make([]string, 0, len(val))
			for _, item := range val {
				parts = append(parts, fmt.Sprintf("%v", item))
			}
			s = strings.Join(parts, ", ")
		default:
			s = fmt.Sprintf("%v", v)
		}

		fmt.Fprintf(w, "%-*s  %s\n", maxLen, k+":", s)
	}
}
