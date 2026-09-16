package mcp

import "fmt"

// The single-sourced MCP diagnostic messages (round-032 FR-008/FR-019). They are
// the ONE definition of the operator-facing warn+skip text, referenced by both
// the discovery layer (which emits them on stderr) and the E2E suite (which
// asserts on the exact product text — so the assertion cannot go vacuous if the
// product message changes, mirroring the round-012 cli.MultiLineHint pattern).
//
// They are NOT frozen class phrases: the frozen `tellme: <phrase>` vocabulary and
// the exit-code set are unchanged (FR-015) — a skip is a warning, not a failure.

// UnreachableWarning is the warn+skip message for a server that did not answer
// within the fixed fast-fail bound (or otherwise failed discovery).
func UnreachableWarning(server string) string {
	return fmt.Sprintf("[mcp] the server %q could not be reached; skipping its tools", server)
}

// SchemaSkippedWarning is the warn+skip message for a tool whose advertised
// input schema could not be normalized to a safe object schema (FR-019/B1).
func SchemaSkippedWarning(server, tool string) string {
	return fmt.Sprintf("[mcp] the tool %q from server %q has an unusable schema; skipping it", tool, server)
}

// NameSkippedWarning is the warn+skip message for a tool whose namespaced wire
// name would not satisfy the provider tool-name grammar (round-032 F5).
func NameSkippedWarning(server, tool string) string {
	return fmt.Sprintf("[mcp] the tool %q from server %q has a name that cannot be offered safely; skipping it", tool, server)
}
