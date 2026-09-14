package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// get_tree (round 021 Decision 3): render a read-only connector tree of a
// directory. The reference contract — a `├── `/`└── ` connector tree, `max_depth`
// defaulting to 2, and `.git` listed but never recursed into.

// getTree renders a read-only connector tree of a directory.
type getTree struct{}

// Name is the wire-valid canonical identifier.
func (getTree) Name() string { return "get_tree" }

// Description is the model-facing summary.
func (getTree) Description() string { return "Show a folder tree." }

// Parameters is the JSON-schema for the tool's arguments.
func (getTree) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Directory path to list (default '.')."},"max_depth":{"type":"integer","description":"Depth of the tree (default 2)."},"reason":{"type":"string","description":"Reason for viewing the folder tree."}},"required":["reason"]}`)
}

// Execute renders the folder tree of the given path (default ".") down to
// max_depth (default 2), never recursing into `.git`; the result is bounded by
// the aggregate cap. A directory sitting at depth == max_depth still has its
// child names listed — recursion stops only once depth exceeds max_depth (the
// reference's `depth > maxDepth` cut).
func (getTree) Execute(ctx context.Context, arguments string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var args struct {
		Path     string `json:"path"`
		MaxDepth int    `json:"max_depth"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("get_tree: invalid arguments: %w", err)
	}
	path := args.Path
	if path == "" {
		path = "."
	}
	maxDepth := args.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 2
	}
	var sb strings.Builder
	if err := buildTree(ctx, path, "", 0, maxDepth, &sb); err != nil {
		return "", fmt.Errorf("get_tree: %w", err)
	}
	return truncateToCap(sb.String()), nil
}

// buildTree writes the connector lines for dir at the given depth. Recursion
// stops once depth exceeds maxDepth — so the children of a directory sitting at
// depth == maxDepth are still listed, and only entries more than maxDepth levels
// below the root are omitted (the reference's `depth > maxDepth` cut, round-021
// Decision 3).
func buildTree(ctx context.Context, dir, indent string, depth, maxDepth int, sb *strings.Builder) error {
	if depth > maxDepth {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for i, e := range entries {
		if err := writeTreeEntry(ctx, dir, e, i == len(entries)-1, indent, depth, maxDepth, sb); err != nil {
			return err
		}
	}
	return nil
}

// writeTreeEntry writes one connector line and, for a directory other than
// `.git`, recurses into it.
func writeTreeEntry(ctx context.Context, parent string, e os.DirEntry, isLast bool, indent string, depth, maxDepth int, sb *strings.Builder) error {
	connector, extra := treeConnectors(isLast)
	sb.WriteString(indent + connector + e.Name() + "\n")
	if !e.IsDir() || e.Name() == ".git" {
		return nil
	}
	return buildTree(ctx, filepath.Join(parent, e.Name()), indent+extra, depth+1, maxDepth, sb)
}

// treeConnectors returns the connector glyph and the child-indent unit for an
// entry that is (or is not) the last of its siblings.
func treeConnectors(isLast bool) (connector, extra string) {
	if isLast {
		return "└── ", "    "
	}
	return "├── ", "│   "
}

// Compile-time port conformance.
var _ domaintools.Tool = getTree{}
