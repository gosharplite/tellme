// Package tools holds the concrete agent-tool adapters behind
// internal/domain/tools. Round 008 ships two read-only filesystem tools
// (list_files, read_files); there is no path/safety boundary — the tools read
// whatever path the model gives (round-008 research Decision 4; Clarify Q3).
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// readCap bounds a single read_files result so it cannot exhaust the model's
// context window — token-budget pruning is out of scope, so the read tool must
// not inject unbounded input (round-008 research Decision 4 / NFR-006).
const readCap = 1 << 20 // 1 MiB

// listFiles enumerates a directory's entries (read-only).
type listFiles struct{}

// Name is the wire-valid canonical identifier (round-008 BLOCKER-1).
func (listFiles) Name() string { return "list_files" }

// Description is the model-facing summary.
func (listFiles) Description() string { return "List the entries of a directory." }

// Parameters is the JSON-schema for the tool's arguments.
func (listFiles) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Directory path to list."}},"required":["path"]}`)
}

// Execute lists the entries at the given path, directories suffixed with "/".
func (listFiles) Execute(_ context.Context, arguments string) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("list_files: invalid arguments: %w", err)
	}
	if args.Path == "" {
		return "", fmt.Errorf("list_files: missing path")
	}
	entries, err := os.ReadDir(args.Path)
	if err != nil {
		return "", fmt.Errorf("list_files: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, "\n"), nil
}

// readFiles returns a file's contents (read-only), size-bounded.
type readFiles struct{}

// Name is the wire-valid canonical identifier (round-008 BLOCKER-1).
func (readFiles) Name() string { return "read_files" }

// Description is the model-facing summary.
func (readFiles) Description() string { return "Read a file's contents." }

// Parameters is the JSON-schema for the tool's arguments.
func (readFiles) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"File path to read."}},"required":["path"]}`)
}

// Execute reads the file at the given path, bounded by readCap (truncated with
// a marker if exceeded).
func (readFiles) Execute(_ context.Context, arguments string) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("read_files: invalid arguments: %w", err)
	}
	if args.Path == "" {
		return "", fmt.Errorf("read_files: missing path")
	}
	f, err := os.Open(args.Path)
	if err != nil {
		return "", fmt.Errorf("read_files: %w", err)
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, readCap+1))
	if err != nil {
		return "", fmt.Errorf("read_files: %w", err)
	}
	if len(data) > readCap {
		return string(data[:readCap]) + "\n… (truncated at 1 MiB)", nil
	}
	return string(data), nil
}

// NewFilesystemTools returns the two read-only filesystem agent tools in offer
// order (round-008 research Decision 4).
func NewFilesystemTools() []domaintools.Tool {
	return []domaintools.Tool{listFiles{}, readFiles{}}
}
