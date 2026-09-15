package ui

import (
	"strings"
	"testing"
)

func TestFormatToolUsage(t *testing.T) {
	rows := []ToolUsageRow{
		{Tool: "list_files", OK: 1},
		{Tool: "read_files", OK: 2, Error: 1},
		{Tool: "get_tree"},
		{Tool: "execute_command", Timeout: 1},
	}
	got := FormatToolUsage(rows)
	want := "tool usage (all sessions):\n" +
		"list_files: total=1 ok=1 error=0 timeout=0\n" +
		"read_files: total=3 ok=2 error=1 timeout=0\n" +
		"get_tree: total=0 ok=0 error=0 timeout=0\n" +
		"execute_command: total=1 ok=0 error=0 timeout=1\n"
	if got != want {
		t.Errorf("FormatToolUsage =\n%q\nwant\n%q", got, want)
	}
}

// TestFormatToolUsagePreservesOrder: the report follows the input (registry)
// order exactly — deterministic, no map iteration.
func TestFormatToolUsagePreservesOrder(t *testing.T) {
	rows := []ToolUsageRow{{Tool: "b", OK: 2}, {Tool: "a", OK: 1}}
	lines := strings.Split(strings.TrimSpace(FormatToolUsage(rows)), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines = %v", lines)
	}
	if !strings.HasPrefix(lines[1], "b:") || !strings.HasPrefix(lines[2], "a:") {
		t.Errorf("order not preserved: %v", lines)
	}
	// Two independently-built equivalent row sets render identically
	// (deterministic for the same input).
	a := []ToolUsageRow{{Tool: "b", OK: 2}, {Tool: "a", OK: 1}}
	b := []ToolUsageRow{{Tool: "b", OK: 2}, {Tool: "a", OK: 1}}
	if FormatToolUsage(a) != FormatToolUsage(b) {
		t.Errorf("output must be deterministic for the same input")
	}
}

// TestFormatToolUsageEmpty: every tool shown with zero (the header only when the
// registry is empty).
func TestFormatToolUsageEmpty(t *testing.T) {
	got := FormatToolUsage(nil)
	if got != "tool usage (all sessions):\n" {
		t.Errorf("FormatToolUsage(nil) = %q", got)
	}
}
