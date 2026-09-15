// Package tools holds the concrete agent-tool adapters behind
// internal/domain/tools. Round 021 ships three read-only reader tools —
// list_files, read_files, get_tree — each requiring `reason`; there is no
// path/safety boundary — the tools read whatever path the model gives
// (round-021 Decision 8; round-008 Decision 4 / Clarify Q3).
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// read_files limits (round 021 Decision 5 + D3a): the reference 100000-byte
// per-file cap, the per-call ≤50 cap, and tellme's aggregate 1 MiB result cap
// (restoring the round-008 "cannot exhaust the context window" property, which
// the per-file/≤50 bounds alone would lose).
const (
	readMaxPerFile   = 100000
	readMaxPerCall   = 50
	readAggregateCap = 1 << 20 // 1 MiB
)

// The aggregate truncation markers (round 021 NFR-001): readBudgetMarker
// terminates a read_files result cut by the aggregate cap; capMarker terminates
// a list_files / get_tree result cut by the same cap.
const (
	readBudgetMarker = "\n... (truncated at the read budget)\n"
	capMarker        = "\n... (truncated)\n"
)

// listFiles enumerates a directory's entries (read-only).
type listFiles struct{}

// Name is the wire-valid canonical identifier (round-008 BLOCKER-1).
func (listFiles) Name() string { return "list_files" }

// Description is the model-facing summary.
func (listFiles) Description() string { return "List the entries of a directory." }

// Parameters is the JSON-schema for the tool's arguments.
func (listFiles) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"The directory path to list (defaults to the current directory '.')."},"reason":{"type":"string","description":"Reason for listing files."}},"required":["reason"]}`)
}

// Execute lists the entries at the given path (defaulting to "."), one `[d]` or
// `[f]` line per entry under a `Contents of <path>:` header, in os.ReadDir order.
// The result is bounded by the aggregate cap (round 021 Decision 2 / D3a). It
// honours the per-tool context so a cancelled/expired run aborts rather than
// blocking on I/O (round-008 RF-1).
func (listFiles) Execute(ctx context.Context, arguments string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var args struct {
		Path   string `json:"path"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("list_files: invalid arguments: %w", err)
	}
	path := args.Path
	if path == "" {
		path = "."
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("list_files: failed to list directory: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	lines := make([]string, 0, len(entries)+1)
	lines = append(lines, fmt.Sprintf("Contents of %s:", path))
	for _, e := range entries {
		kind := "f"
		if e.IsDir() {
			kind = "d"
		}
		lines = append(lines, fmt.Sprintf("[%s] %s", kind, e.Name()))
	}
	out := strings.Join(lines, "\n") + "\n"
	return truncateToCap(out), nil
}

// truncateToCap bounds a tool result to the aggregate result cap. The retained
// content is cut at a rune boundary (so a mid-rune cut cannot emit invalid UTF-8
// — get_tree's box-drawing glyphs are 3 bytes each), and the marker is counted
// toward the ceiling, so the whole result — content plus terminator — stays
// within readAggregateCap (round 021 D3a / NFR-001).
func truncateToCap(out string) string {
	if len(out) <= readAggregateCap {
		return out
	}
	out = strings.ToValidUTF8(out[:readAggregateCap-len(capMarker)], "")
	return out + capMarker
}

// readFiles returns a file's contents (read-only), size-bounded.
type readFiles struct{}

// Name is the wire-valid canonical identifier (round-008 BLOCKER-1).
func (readFiles) Name() string { return "read_files" }

// Description is the model-facing summary.
func (readFiles) Description() string { return "Read a file's contents." }

// Parameters is the JSON-schema for the tool's arguments (round 021: the
// multi-file `filepaths` array; `reason` is required by schema but not validated).
func (readFiles) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"filepaths":{"type":"array","items":{"type":"string"},"description":"The list of file paths to read."},"reason":{"type":"string","description":"Reason for reading these files."}},"required":["filepaths","reason"]}`)
}

// Execute reads each requested file in order, framing each block with a header
// line, and bounds the whole result (round 021 Decision 5 / D3a): a 100000-byte
// per-file cap with a truncation marker, inline ERROR/binary/directory handling,
// a ≤50 per-call cap, and a 1 MiB aggregate cap. It honours the per-tool context
// so a cancelled/expired run aborts rather than blocking on I/O (round-008 RF-1).
func (readFiles) Execute(ctx context.Context, arguments string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var args struct {
		FilePaths []string `json:"filepaths"`
		Reason    string   `json:"reason"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("read_files: invalid arguments: %w", err)
	}
	if len(args.FilePaths) == 0 {
		return "", fmt.Errorf("read_files: filepaths argument is required and cannot be empty")
	}
	if len(args.FilePaths) > readMaxPerCall {
		return fmt.Sprintf("Error: requested too many files (%d). Maximum is %d files per call.", len(args.FilePaths), readMaxPerCall), nil
	}
	var sb strings.Builder
	for _, path := range args.FilePaths {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		// Stop before reading (or opening) the next file once the budget is spent,
		// so the block that trips the cap is not wastefully read (round-021 review NIT).
		if sb.Len() >= readAggregateCap-len(readBudgetMarker) {
			sb.WriteString(readBudgetMarker)
			break
		}
		if !appendBounded(&sb, readOneFile(path)) {
			break
		}
	}
	return sb.String(), nil
}

// appendBounded appends block to sb while it fits within the aggregate result
// cap, counting the terminator toward the ceiling (round 021 D3a). It returns
// false — having written readBudgetMarker — when the block would exceed the cap,
// so the caller stops (the omitted file gets no header).
func appendBounded(sb *strings.Builder, block string) bool {
	if sb.Len()+len(block) > readAggregateCap-len(readBudgetMarker) {
		sb.WriteString(readBudgetMarker)
		return false
	}
	sb.WriteString(block)
	return true
}

// readOneFile renders one read_files block: a `--- File: <path> ---` header
// followed by the file's bounded body (`\n\n`-terminated), or an inline ERROR /
// binary / directory message (a recoverable, non-fatal outcome).
func readOneFile(path string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "--- File: %s ---\n", path)
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(&sb, "ERROR: failed to read file: %v\n\n", err)
		return sb.String()
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		fmt.Fprintf(&sb, "ERROR: failed to read file: %v\n\n", err)
		return sb.String()
	}
	if info.IsDir() {
		sb.WriteString("ERROR: path is a directory, use list_files instead\n\n")
		return sb.String()
	}
	data, err := io.ReadAll(io.LimitReader(f, readMaxPerFile+1))
	if err != nil {
		fmt.Fprintf(&sb, "ERROR: failed to read file: %v\n\n", err)
		return sb.String()
	}
	if isBinary(data) {
		sb.WriteString("(Binary file, cannot display as text)\n\n")
		return sb.String()
	}
	if len(data) > readMaxPerFile {
		sb.WriteString(strings.ToValidUTF8(string(data[:readMaxPerFile]), ""))
		sb.WriteString("\n... (truncated)\n\n")
		return sb.String()
	}
	sb.Write(data)
	sb.WriteString("\n\n")
	return sb.String()
}

// NewFilesystemTools returns the read-only filesystem agent tools in offer order
// (round 021: list_files, read_files, get_tree — the reference reader trio).
func NewFilesystemTools() []domaintools.Tool {
	return []domaintools.Tool{listFiles{}, readFiles{}, getTree{}}
}
