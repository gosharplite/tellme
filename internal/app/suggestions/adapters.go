package suggestions

import (
	"context"
	"os"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// TrackerPrompts adapts the shared prompt-log port to the PromptSource seam.
type TrackerPrompts struct {
	Tracker domainhistory.PromptTracker
}

// RecentPrompts returns the newest-first recorded prompts.
func (t TrackerPrompts) RecentPrompts(ctx context.Context, n int) []string {
	if t.Tracker == nil {
		return nil
	}
	entries, err := t.Tracker.Recent(ctx, n)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Prompt)
	}
	return out
}

// RegistryTools adapts a tool registry to the ToolSource seam (registered names,
// the tool-suggestion source — spec.md A4).
type RegistryTools struct {
	Registry domaintools.Registry
}

// ToolNames returns the registered tool names in offer order.
func (r RegistryTools) ToolNames() []string {
	if r.Registry == nil {
		return nil
	}
	ts := r.Registry.Tools()
	names := make([]string, 0, len(ts))
	for _, t := range ts {
		names = append(names, t.Name())
	}
	return names
}

// OSSWorkspace is the real workspace source: it reads a query's directory in
// bounded batches, skipping ignored directories, and stops at the limit
// (round-015 PR #38 review directive ③).
type OSSWorkspace struct{}

// ignoredDir reports whether a directory name is on the noise ignore list.
func ignoredDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", ".idea", ".vscode", "dist", ".cache", "target":
		return true
	}
	return false
}

// Entries returns up to limit entry names under dir matching prefix, reading in
// bounded batches and yielding to ctx (PR #38 review directive ③).
func (OSSWorkspace) Entries(ctx context.Context, dir, prefix string, limit int) []string {
	if dir == "" {
		dir = "."
	}
	f, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()
	var out []string
	for {
		if ctx.Err() != nil {
			return out
		}
		entries, err := f.ReadDir(dirBatch)
		if err != nil {
			break
		}
		for _, e := range entries {
			if e.IsDir() && ignoredDir(e.Name()) {
				continue
			}
			if prefix != "" && !isSubsequence(prefix, e.Name()) {
				continue
			}
			out = append(out, e.Name())
			if len(out) >= limit {
				return out
			}
		}
	}
	return out
}
